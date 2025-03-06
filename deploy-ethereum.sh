#!/bin/bash
# Ethereum Node Deployment Script for current directory structure
# This script deploys components in the correct order and checks status

set -e

# Configuration
NAMESPACE="default"
WAIT_TIMEOUT=300  # 5 minutes timeout for pod readiness
LOG_WAIT_TIME=30  # How long to wait for sync logs after pods are ready

echo "=== Ethereum Node Deployment Script ==="
echo "Starting deployment process..."

# Function to check if a command exists
command_exists() {
  command -v "$1" >/dev/null 2>&1
}

# Check for required tools
if ! command_exists kubectl; then
  echo "Error: kubectl is required but not installed."
  exit 1
fi

# Deploy in the correct order using existing files
echo "Deploying Storage Class..."
kubectl apply -f storage-class.yaml

echo "Creating JWT Secret ConfigMap..."
kubectl apply -f geth/jwt-secret-configmap.yaml  # Updated path to match your directory

# Deploy Geth (Execution Client)
echo "Deploying Geth (Execution Client)..."
kubectl apply -f geth/pv.yaml
kubectl apply -f geth/pvc.yaml
kubectl apply -f geth/statefulset.yaml
kubectl apply -f geth/service.yaml

# Deploy Lighthouse (Consensus Client)
echo "Deploying Lighthouse (Consensus Client)..."
kubectl apply -f lighthouse/pv.yaml
kubectl apply -f lighthouse/pvc.yaml
kubectl apply -f lighthouse/statefulset.yaml
kubectl apply -f lighthouse/service.yaml

# Deploy monitoring infrastructure
echo "Deploying monitoring infrastructure..."
kubectl apply -f monitoring/namespace.yaml # Using the dedicated namespace file

# Deploy Prometheus
kubectl apply -f monitoring/prometheus/prometheus-rbac.yaml
kubectl apply -f monitoring/prometheus/prometheus-configmap.yaml
kubectl apply -f monitoring/prometheus/prometheus-deployment.yaml
kubectl apply -f monitoring/prometheus/prometheus-service.yaml

# Deploy Grafana
kubectl apply -f monitoring/grafana/grafana-configmap.yaml
kubectl apply -f monitoring/grafana/grafana-dashboard-config.yaml # File name adjustment
kubectl apply -f monitoring/grafana/grafana-dashboard-configmap.yaml
kubectl apply -f monitoring/grafana/grafana-deployment.yaml
kubectl apply -f monitoring/grafana/grafana-service.yaml

# Deploy node-exporter for disk I/O metrics
echo "Deploying Node Exporter for disk metrics..."
kubectl apply -f monitoring/node-exporter.yaml

# Wait for pods to be ready
echo "Waiting for Geth pod to be ready..."
kubectl wait --for=condition=ready pod/geth-0 --timeout=${WAIT_TIMEOUT}s

echo "Waiting for Lighthouse pod to be ready..."
kubectl wait --for=condition=ready pod/lighthouse-0 --timeout=${WAIT_TIMEOUT}s

echo "Waiting for Prometheus pod to be ready..."
kubectl wait --for=condition=ready pod -l app=prometheus -n monitoring --timeout=${WAIT_TIMEOUT}s

echo "Waiting for Grafana pod to be ready..."
kubectl wait --for=condition=ready pod -l app=grafana -n monitoring --timeout=${WAIT_TIMEOUT}s

# Check logs to confirm syncing
echo "Checking Geth logs for sync status..."
echo "Waiting ${LOG_WAIT_TIME} seconds for logs to appear..."
sleep ${LOG_WAIT_TIME}

GETH_SYNC_LOG=$(kubectl logs geth-0 | grep -i "sync" | tail -5)
if [ -z "$GETH_SYNC_LOG" ]; then
  echo "Warning: No sync logs found in Geth. This could be normal if sync hasn't started yet."
else
  echo "Geth sync logs:"
  echo "$GETH_SYNC_LOG"
fi

echo "Checking Lighthouse logs for sync status..."
LIGHTHOUSE_SYNC_LOG=$(kubectl logs lighthouse-0 | grep -i "sync\|slot" | tail -5)
if [ -z "$LIGHTHOUSE_SYNC_LOG" ]; then
  echo "Warning: No sync logs found in Lighthouse. This could be normal if sync hasn't started yet."
else
  echo "Lighthouse sync logs:"
  echo "$LIGHTHOUSE_SYNC_LOG"
fi

# Set up port forwarding for Grafana
echo "Setting up port forwarding for Grafana on port 3000..."
kubectl port-forward -n monitoring svc/grafana 3000:3000 > /dev/null 2>&1 &
GRAFANA_PID=$!
echo "Grafana port forwarding PID: $GRAFANA_PID"
echo "Access Grafana at: http://localhost:3000"
echo "Default credentials: admin/admin123"

echo "=== Ethereum Node Deployment Complete ==="
echo "To stop Grafana port forwarding: kill $GRAFANA_PID"