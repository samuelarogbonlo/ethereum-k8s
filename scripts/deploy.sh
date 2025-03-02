#!/bin/bash
set -e

# Ethereum Node Kubernetes Deployment Script
# This script deploys a complete Ethereum node (execution + consensus clients) on Kubernetes
# along with monitoring stack

# Set color variables for better output
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
NC='\033[0m' # No Color

# Configuration
NAMESPACE="ethereum-node"
KUBE_DIR="../kubernetes"

# Function to check if a command exists
check_command() {
    if ! command -v $1 &> /dev/null; then
        echo -e "${RED}Error: $1 is required but not installed.${NC}"
        exit 1
    fi
}

# Function to apply Kubernetes manifest and check result
apply_manifest() {
    local file=$1
    echo -e "${YELLOW}Applying $file...${NC}"
    kubectl apply -f "$file"
    if [ $? -ne 0 ]; then
        echo -e "${RED}Failed to apply $file${NC}"
        exit 1
    fi
    echo -e "${GREEN}Successfully applied $file${NC}"
}

# Check for required commands
check_command kubectl
check_command openssl

# Print header
echo -e "${GREEN}=======================================${NC}"
echo -e "${GREEN}  Ethereum Node Kubernetes Deployment  ${NC}"
echo -e "${GREEN}=======================================${NC}"

# Ensure script is run from the scripts directory
if [[ $(basename $(pwd)) != "scripts" ]]; then
    echo -e "${RED}Error: This script must be run from the scripts directory${NC}"
    exit 1
fi

# Create storage directories on host if they don't exist
echo -e "${YELLOW}Creating storage directories on host...${NC}"
mkdir -p /tmp/ethereum/geth /tmp/ethereum/lighthouse /tmp/ethereum/prometheus /tmp/ethereum/grafana
chmod -R 777 /tmp/ethereum
echo -e "${GREEN}Storage directories created${NC}"

# Step 1: Create Namespace
echo -e "${YELLOW}Step 1: Creating namespace...${NC}"
apply_manifest "${KUBE_DIR}/namespace.yaml"

# Step 2: Create Storage Resources
echo -e "${YELLOW}Step 2: Setting up storage...${NC}"
apply_manifest "${KUBE_DIR}/storage.yaml"

# Step 3: Create ConfigMaps
echo -e "${YELLOW}Step 3: Creating ConfigMaps...${NC}"
apply_manifest "${KUBE_DIR}/configmaps.yaml"

# Step 4: Generate and apply JWT secret
echo -e "${YELLOW}Step 4: Creating JWT secret...${NC}"
./create-jwt-secret.sh
echo -e "${GREEN}JWT secret created successfully${NC}"

# Step 5: Deploy Execution Client
echo -e "${YELLOW}Step 5: Deploying Execution Client (Geth)...${NC}"
apply_manifest "${KUBE_DIR}/execution-client.yaml"

# Step 6: Deploy Consensus Client
echo -e "${YELLOW}Step 6: Deploying Consensus Client (Lighthouse)...${NC}"
apply_manifest "${KUBE_DIR}/consensus-client.yaml"

# Step 7: Create Services
echo -e "${YELLOW}Step 7: Creating Services...${NC}"
apply_manifest "${KUBE_DIR}/services.yaml"

# Step 8: Deploy Monitoring Stack
echo -e "${YELLOW}Step 8: Deploying Monitoring Stack...${NC}"
apply_manifest "${KUBE_DIR}/monitoring/prometheus.yaml"
apply_manifest "${KUBE_DIR}/monitoring/grafana.yaml"
apply_manifest "${KUBE_DIR}/monitoring/grafana-deployment.yaml"

# Step 9: Wait for pods to be ready
echo -e "${YELLOW}Step 9: Waiting for pods to become ready...${NC}"
kubectl -n $NAMESPACE wait --for=condition=Ready pod -l app=geth --timeout=300s
kubectl -n $NAMESPACE wait --for=condition=Ready pod -l app=lighthouse --timeout=300s
kubectl -n $NAMESPACE wait --for=condition=Ready pod -l app=prometheus --timeout=300s
kubectl -n $NAMESPACE wait --for=condition=Ready pod -l app=grafana --timeout=300s

# Step 10: Check logs to confirm syncing
echo -e "${YELLOW}Step 10: Checking logs to confirm syncing...${NC}"
GETH_POD=$(kubectl -n $NAMESPACE get pods -l app=geth -o jsonpath="{.items[0].metadata.name}")
LIGHTHOUSE_POD=$(kubectl -n $NAMESPACE get pods -l app=lighthouse -o jsonpath="{.items[0].metadata.name}")

echo -e "${YELLOW}Geth logs:${NC}"
kubectl -n $NAMESPACE logs $GETH_POD --tail=20 | grep -i "sync"

echo -e "${YELLOW}Lighthouse logs:${NC}"
kubectl -n $NAMESPACE logs $LIGHTHOUSE_POD --tail=20 | grep -i "sync"

# Display access information
echo -e "${GREEN}=======================================${NC}"
echo -e "${GREEN}  Deployment Complete!                 ${NC}"
echo -e "${GREEN}=======================================${NC}"
echo -e "${YELLOW}Ethereum Node Services:${NC}"
echo -e "Geth RPC Endpoint: http://$(minikube ip):30545"
echo -e "Geth WebSocket Endpoint: ws://$(minikube ip):30546"
echo -e ""
echo -e "${YELLOW}Monitoring:${NC}"
echo -e "Prometheus UI: http://$(minikube ip):30909"
echo -e "Grafana Dashboard: http://$(minikube ip):30300"
echo -e "  Username: admin"
echo -e "  Password: admin"
echo -e ""
echo -e "${GREEN}Run the health check script to verify the node:${NC}"
echo -e "  go run healthcheck.go"
echo -e "${GREEN}=======================================${NC}"