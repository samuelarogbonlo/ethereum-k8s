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
# setup minikube locally
minikube start --cpus=4 --memory=16g --disk-size=100g --driver=docker

# Deploy everything and run health checks
make

# Other useful commands
make deploy        # Only deploy the Ethereum nodes
make health-check  # Run health checks on existing deployment
make monitoring    # Forward ports for Grafana and Prometheus
make stop-monitoring # Stop port forwarding
```

### Accessing Services

All services are exposed through Kubernetes port forwarding:

| Service | Command | Access |
|---------|---------|--------|
| Ethereum RPC | `kubectl port-forward svc/geth 8545:8545` | `curl -X POST -H "Content-Type: application/json" --data '{"jsonrpc":"2.0","method":"eth_blockNumber","params":[],"id":1}' http://localhost:8545` |
| Lighthouse API | `kubectl port-forward svc/lighthouse 5052:5052` | `curl http://localhost:5052/eth/v1/node/syncing` |
| Grafana | `kubectl port-forward -n monitoring svc/grafana 3000:3000` | http://localhost:3000 (admin/admin123) |
| Prometheus | `kubectl port-forward -n monitoring svc/prometheus 9090:9090` | http://localhost:9090 |

## Additional Documentation

- [Optimization.MD](docs/optimzation.MD) - Detailed performance tuning for bare metal environments
- [Maintenance.md](docs/Maintenance.md) - Troubleshooting and maintenance procedures
- [Prod-Improvements.md](docs/prod-Improvements.md) - Production enhancement suggestions beyond the reference implementation.
- [Decision-process.md](docs/decision-process.md) - Rationale behind key design and technology choices

## Implementation Details

### Architecture

This deployment uses:
- StatefulSets for both Geth and Lighthouse to ensure stable network identity and data persistence
- Local persistent volumes for optimal I/O performance
- Prometheus Operator with ServiceMonitors for automatic metric discovery
- Grafana dashboards customized for Ethereum node monitoring
- Health check script in Go to validate node functionality

### Automation

The included automation:
- `make deploy` - Handles full deployment via Helm
- `make health-check` - Verifies node health including peer count and sync status
- `make monitoring` - Sets up port-forwarding to access Grafana and Prometheus
- Ansible playbook for bare metal server provisioning

## License

This project is licensed under the MIT License.