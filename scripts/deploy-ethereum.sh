#!/bin/bash
set -e

NAMESPACE="default"
RELEASE_NAME="ethereum"
STORAGE_CLASS="local-path"

BLUE='\033[0;34m'
GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[0;33m'
NC='\033[0m'

echo -e "${BLUE}=== Ethereum Node Deployment ===${NC}"

# Prerequisites check
command -v kubectl >/dev/null || { echo "kubectl missing."; exit 1; }
command -v helm >/dev/null || { echo "helm missing."; exit 1; }

# Get the script directory and repository root
SCRIPT_DIR="$(dirname "$(readlink -f "$0")")"
REPO_ROOT="$(dirname "$SCRIPT_DIR")"

# Set paths using absolute references
CHART_DIR="$REPO_ROOT/helm/ethereum-node"
VALUES_FILE="$REPO_ROOT/helm/ethereum-node/values.yaml"
DASHBOARD_FILE="$REPO_ROOT/helm/ethereum-node/dashboards/ethereum-dashboards.json"

# Add repositories
helm repo add prometheus-community https://prometheus-community.github.io/helm-charts || true
helm repo add grafana https://grafana.github.io/helm-charts || true
helm repo update

# Install kube-prometheus-stack for monitoring capabilities
echo -e "${BLUE}Installing kube-prometheus-stack...${NC}"
kubectl create namespace monitoring --dry-run=client -o yaml | kubectl apply -f -
helm upgrade --install prometheus prometheus-community/kube-prometheus-stack \
  --namespace monitoring \
  --set prometheus.prometheusSpec.serviceMonitorSelectorNilUsesHelmValues=false \
  --wait

# Wait for CRDs to be ready
echo -e "${BLUE}Waiting for ServiceMonitor CRDs to be established...${NC}"
kubectl wait --for=condition=established crd/servicemonitors.monitoring.coreos.com --timeout=60s

## Create the ConfigMap first
kubectl create configmap ethereum-dashboard \
  --from-file=ethereum-dashboards.json="$DASHBOARD_FILE" \
  --namespace "$NAMESPACE" \
  --dry-run=client -o yaml | kubectl apply -f -

# Then label it separately
kubectl label configmap ethereum-dashboard grafana_dashboard=1 -n "$NAMESPACE" --overwrite

# Install local-path provisioner
kubectl apply -f https://raw.githubusercontent.com/rancher/local-path-provisioner/master/deploy/local-path-storage.yaml

# Create local storage directories (only for minikube)
NODE_NAME=$(kubectl get nodes -o jsonpath='{.items[0].metadata.name}')
if [[ "$NODE_NAME" == "minikube" ]]; then
  minikube ssh "sudo mkdir -p /tmp/ethereum/geth /tmp/ethereum/lighthouse && sudo chmod 777 -R /tmp/ethereum"
else
  echo "Ensure directories exist on node: /tmp/ethereum/geth and /tmp/ethereum/lighthouse"
fi

# Apply PVs explicitly from yaml (with environment substitution)
export NODE_NAME=$NODE_NAME
export STORAGE_CLASS=$STORAGE_CLASS
envsubst < "$REPO_ROOT/kubernetes/local-pv.yaml" | kubectl apply -f -

# Build dependencies only if needed
if [ ! -d "$CHART_DIR/charts" ] || [ -z "$(ls -A "$CHART_DIR/charts" 2>/dev/null)" ]; then
  echo -e "${BLUE}Downloading Helm dependencies...${NC}"
  helm dependency build "$CHART_DIR"
else
  echo -e "${GREEN}Helm dependencies already downloaded, skipping...${NC}"
fi

# Deploy the Helm chart
helm upgrade --install "$RELEASE_NAME" "$CHART_DIR" \
  --namespace "$NAMESPACE" \
  --values "$VALUES_FILE" \
  --timeout 600s --wait

# Wait for Pods to be ready
kubectl wait --for=condition=ready pod -l app=geth --timeout=300s
kubectl wait --for=condition=ready pod -l app=lighthouse --timeout=300s

# Check pods
kubectl get pods

echo -e "${GREEN}Deployment successful!${NC}"

# Run health check
echo -e "${BLUE}=== Running Health Check ===${NC}"

# Build Go health check tool
cd "$SCRIPT_DIR"
go build -o eth-health-check eth-health-check.go

# Run health check with port forwarding
kubectl port-forward svc/geth-node 8545:8545 > /dev/null 2>&1 & PID1=$!
kubectl port-forward svc/lighthouse 5052:5052 > /dev/null 2>&1 & PID2=$!
echo "Port forwarding started, waiting 3 seconds..."
sleep 3

./eth-health-check
EXIT_CODE=$?

# Clean up port forwarding
echo "Cleaning up health check port forwarding..."
kill $PID1 $PID2 2>/dev/null || true

echo -e "${GREEN}Deployment and health check complete!${NC}"
exit $EXIT_CODE