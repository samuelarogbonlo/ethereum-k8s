#!/bin/bash
set -e

# Generate JWT secret for Engine API communication between execution and consensus clients
# This script creates the JWT secret and applies it to the Kubernetes cluster

# Configuration
NAMESPACE="ethereum-node"
SECRET_NAME="jwt-secret"
SECRET_KEY="jwtsecret"

# Create a temporary file for the JWT secret
TMP_FILE=$(mktemp)
trap 'rm -f ${TMP_FILE}' EXIT

# Generate a 32-byte random hex string without newlines
echo -n $(openssl rand -hex 32) > "${TMP_FILE}"

# Base64 encode the JWT secret for Kubernetes
JWT_SECRET=$(cat "${TMP_FILE}" | base64 -w 0)

# Create the Kubernetes secret manifest
cat > "${TMP_FILE}.yaml" << EOF
apiVersion: v1
kind: Secret
metadata:
  name: ${SECRET_NAME}
  namespace: ${NAMESPACE}
type: Opaque
data:
  ${SECRET_KEY}: ${JWT_SECRET}
EOF

# Apply the secret to the cluster
kubectl apply -f "${TMP_FILE}.yaml"

# Clean up
rm -f "${TMP_FILE}.yaml"

echo "JWT secret created and applied to the Kubernetes cluster."
