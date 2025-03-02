# Ethereum Node Kubernetes Deployment

This project provides a complete solution for deploying an Ethereum node (execution and consensus clients) on Kubernetes with a focus on simulating a bare metal environment.

## Project Overview

This deployment includes:

1. **Ethereum Clients:**
   - **Execution Client:** Geth (Go Ethereum)
   - **Consensus Client:** Lighthouse

2. **Infrastructure Components:**
   - StatefulSets with persistent storage
   - Local persistent volumes to simulate bare metal storage
   - Services for API access

3. **Monitoring Stack:**
   - Prometheus for metrics collection
   - Grafana with custom dashboards
   - Alert Manager (configurable)

4. **Automation:**
   - Deployment scripts
   - Health check tools
   - Bare metal provisioning playbooks

## Prerequisites

- Kubernetes cluster (local or remote)
- `kubectl` configured to access your cluster
- At least 16GB RAM and 4 CPU cores available
- At least 1.2TB storage space

For local setup with minikube:
```bash
minikube start --driver=virtualbox --cpus=4 --memory=16g --disk-size=1200g
```

## Directory Structure

```
web3-k8s-node/
├── kubernetes/           # Kubernetes manifests
│   ├── namespace.yaml
│   ├── configmaps.yaml
│   ├── secrets.yaml
│   ├── storage.yaml
│   ├── execution-client.yaml
│   ├── consensus-client.yaml
│   ├── services.yaml
│   └── monitoring/
│       ├── prometheus.yaml
│       ├── grafana.yaml
│       └── dashboards/
├── scripts/              # Automation scripts
│   ├── deploy.sh
│   ├── healthcheck.go
│   └── create-jwt-secret.sh
├── ansible/              # Bare metal provisioning
│   └── bare-metal-setup.yaml
├── README.md
└── docs/
    └── bare-metal-optimizations.md
```

## Quick Start

### 1. Prepare Storage Directories

```bash
sudo mkdir -p /mnt/ethereum/geth /mnt/ethereum/lighthouse /mnt/ethereum/prometheus /mnt/ethereum/grafana
sudo chmod -R 777 /mnt/ethereum
```

### 2. Deploy the Ethereum Node

```bash
cd scripts
chmod +x *.sh
./deploy.sh
```

### 3. Check Node Status

```bash
go run healthcheck.go
```

### 4. Access Services

- **Geth RPC:** http://NODE-IP:30545
- **Geth WebSocket:** ws://NODE-IP:30546
- **Prometheus:** http://NODE-IP:30909
- **Grafana:** http://NODE-IP:30300 (default credentials: admin/admin)

## Bare Metal Optimizations

For production deployments on bare metal hardware, several optimizations are recommended. See [Bare Metal Optimizations](docs/bare-metal-optimizations.md) for detailed guidance on:

- **Networking Optimizations:** P2P connection tuning, jumbo frames, TCP optimizations
- **Storage Optimizations:** File system configuration, I/O scheduler tuning
- **System Tuning:** Kernel parameters, file limits, NUMA-aware configurations
- **Security Considerations:** Firewall rules, access control

## Monitoring

The deployment includes a comprehensive monitoring stack with Prometheus and Grafana.

### Key Metrics Monitored:

- **Ethereum Specific:**
  - Sync progress
  - Block height
  - Peer count
  - Transaction pool size

- **System Metrics:**
  - CPU/Memory utilization
  - Disk I/O performance
  - Network traffic

### Dashboards:

Pre-configured Grafana dashboards are included for:
- Ethereum Node Overview
- System Resource Utilization

## Deployment Architecture

The deployment uses StatefulSets to ensure stable network identifiers and persistent storage, which is critical for Ethereum nodes.

### Key Components:

1. **Execution Client (Geth):**
   - Processes transactions and maintains the state of the blockchain
   - Exposes RPC API for interaction

2. **Consensus Client (Lighthouse):**
   - Implements Proof of Stake consensus
   - Communicates with the execution client via Engine API

3. **Persistent Storage:**
   - Local volumes for chain data
   - Designed to simulate bare metal storage performance

4. **Monitoring:**
   - Prometheus for metrics collection
   - Grafana for visualization
   - Optional alerts for node health

## Customization

### Configuration Options:

To modify the default configuration:

1. **Client Configuration:**
   - Edit `configmaps.yaml` to change client parameters

2. **Resource Allocation:**
   - Adjust CPU/Memory requests and limits in `execution-client.yaml` and `consensus-client.yaml`

3. **Storage Size:**
   - Modify storage capacity in `storage.yaml`

### Adding Custom Dashboards:

1. Create your dashboard in Grafana
2. Export as JSON
3. Add to `kubernetes/monitoring/dashboards/`
4. Update the ConfigMap in `grafana.yaml`

## Troubleshooting

### Common Issues:

1. **Persistent Volume Claims Pending:**
   - Ensure storage class exists and volumes are available
   - Check node affinity settings

2. **Clients Not Syncing:**
   - Verify network connectivity (P2P ports open)
   - Check resource allocation (may need more CPU/memory)
   - Inspect logs: `kubectl logs -n ethereum-node <pod-name>`

3. **Poor Performance:**
   - Check disk I/O performance
   - Consider storage optimizations from bare metal guide
   - Adjust resource limits

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

## License

This project is licensed under the MIT License - see the LICENSE file for details.