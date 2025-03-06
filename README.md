# Ethereum Node on Kubernetes

A production-grade deployment framework for running Ethereum nodes (execution and consensus clients) on Kubernetes, optimized for bare metal environments.

## Overview

This repository contains Kubernetes manifests and supporting scripts for deploying and managing a complete Ethereum node, including both execution layer (Geth) and consensus layer (Lighthouse) clients. The deployment is designed for high performance, reliability, and operational simplicity, with special focus on bare metal environments where performance matters most for Web3 applications.

### Features

- **Complete Ethereum Stack**: Deploys both execution client (Geth) and consensus client (Lighthouse)
- **Persistent Storage**: Configured with local storage volumes for optimal performance
- **Monitoring & Metrics**: Comprehensive Prometheus and Grafana setup with Ethereum-specific dashboards
- **Health Checking**: Automated node health monitoring with alerting capabilities
- **Performance Optimized**: Network, storage, and system configurations tuned for Ethereum workloads
- **Production Ready**: Security configurations, resource management, and high availability options
- **Web3 Support**: Optimized for serving dApps and supporting blockchain applications

## Components

### Execution Client (Geth)

Go Ethereum (Geth) is deployed as a StatefulSet with persistent storage for the blockchain data. Key features:

- Configured for optimal RPC and P2P connectivity
- Persistent volume setup for chain data and state
- Resource requests and limits properly sized for production use
- Service exposures for JSON-RPC and WebSocket endpoints
- JWT authentication for secure communication with the consensus client

### Consensus Client (Lighthouse)

Lighthouse is a high-performance consensus client for Ethereum, deployed with:

- Beacon node configured for the Ethereum Proof of Stake network
- Secure connection to the execution client via JWT
- Optimized peer discovery and sync settings
- Persistent storage for beacon chain data

### Monitoring Stack

The monitoring setup includes:

- **Prometheus**: Collects metrics from Geth, Lighthouse, and node-exporter
- **Grafana**: Pre-configured dashboards for:
  - Blockchain sync status and progress
  - Peer connectivity and network health
  - System resource utilization (CPU, memory, disk I/O)
  - Client-specific metrics for both Geth and Lighthouse
- **Node Exporter**: Collects host-level metrics including disk I/O performance

## Prerequisites

- Kubernetes cluster (v1.23+)
- kubectl configured to interact with your cluster
- Minikube
- For bare metal:
  - Physical servers with sufficient resources (8+ CPU cores, 16GB+ RAM)
  - High-performance storage (NVMe preferred) with 500GB+ capacity
  - Network connectivity with public IP addresses
  - Ansible installed on a control node

## Deployment and Monitoring

### Quick Start

The simplest way to deploy and validate the entire stack:

```bash
# Deploy everything and run health checks
make

# Other useful commands
make deploy        # Only deploy the Ethereum nodes
make health-check  # Run health checks on existing deployment
make help          # Show all available commands
```

### Customization and Manual Deployment

1. **Prepare Environment**: Ensure your Kubernetes enviroemnt is ready with Minikube or preffered local solution
2. **Customize Configuration**: Modify storage, resource allocation, and network settings in the YAML files as you wish
3. **Deploy Manually**: Run `./deploy-ethereum.sh` to deploy the entire stack

### Accessing Services

All services are exposed through Kubernetes port forwarding:

| Service | Command | Access |
|---------|---------|--------|
| Ethereum RPC | `kubectl port-forward svc/geth 8545:8545` | `curl -X POST -H "Content-Type: application/json" --data '{"jsonrpc":"2.0","method":"eth_blockNumber","params":[],"id":1}' http://localhost:8545` |
| Lighthouse API | `kubectl port-forward svc/lighthouse 5052:5052` | `curl http://localhost:5052/eth/v1/node/syncing` |
| Grafana | `kubectl port-forward -n monitoring svc/grafana 3000:3000` | http://localhost:3000 (admin/admin123) |
| Prometheus | `kubectl port-forward -n monitoring svc/prometheus 9090:9090` | http://localhost:9090 |

### Production Deployment

For production environments, replace port forwarding with:

1. **LoadBalancer Services**: Change service type to `LoadBalancer` (cloud environments)
2. **Ingress Controllers**: Configure ingress rules for web access with TLS
3. **NodePort Services**: Expose on fixed ports across all nodes (bare metal)

Important production considerations:
- Implement proper authentication for RPC endpoints
- Enable TLS encryption for all external connections
- Configure network policies and firewall rules
- Set up external monitoring with alerts

## Performance Optimization

See the `optimization.MD` file for detailed information on optimizing your Ethereum node deployment, including:

- Network performance tuning for P2P traffic
- Storage configuration for blockchain data
- System-level optimizations (NUMA, file limits, scheduler settings)
- Client-specific parameter optimization

## Security Considerations

- The JWT secret is used for secure communication between execution and consensus clients
- Proper resource limits prevent DoS attacks
- Network policies can be applied for additional security
- Consider enabling TLS for RPC endpoints in production

## License

This project is licensed under the MIT License.