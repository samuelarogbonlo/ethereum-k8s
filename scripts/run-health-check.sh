#!/bin/bash
# Wrapper script for eth-health-check that handles port forwarding

# Start Geth port forwarding
echo "Starting port forwarding for Geth..."
kubectl port-forward svc/geth-node 8545:8545 > /dev/null 2>&1 &
GETH_PID=$!

# Start Lighthouse port forwarding
echo "Starting port forwarding for Lighthouse..."
kubectl port-forward svc/lighthouse 5052:5052 > /dev/null 2>&1 &
LIGHTHOUSE_PID=$!

# Wait for port forwarding to establish
echo "Waiting for port forwarding to establish..."
sleep 3

# Run the health check
echo "Running Ethereum health check..."
go run eth-health-check.go

# Capture the exit code properly
EXIT_CODE=$?

# Clean up port forwarding
echo "Cleaning up port forwarding..."
kill $GETH_PID $LIGHTHOUSE_PID 2>/dev/null || true

# Return the health check status
echo "Health check completed with exit code: $EXIT_CODE"
exit $EXIT_CODE