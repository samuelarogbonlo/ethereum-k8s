# Key Decision Making

## Client Selection

### Lighthouse for Consensus
- Lower resource requirements than Prysm (critical for bare metal)
- Better memory safety with Rust implementation
- Contributes to client diversity on the Ethereum network
- More comprehensive Prometheus metrics out-of-the-box

### Geth for Execution
- Industry standard with proven reliability
- Well-documented performance characteristics
- Excellent compatibility with Lighthouse
- Comprehensive metrics exposure

## Infrastructure Choices

### Local Storage Strategy
- Used local-path provisioner for performance-critical blockchain data
- Network storage would introduce unacceptable latency
- Direct alignment with bare metal requirements
- Data locality ensures optimal I/O performance

### StatefulSets with PVCs
- Ensures data persistence across pod restarts
- Provides stable network identity for P2P connections
- Allows for controlled startup/shutdown sequences

## Monitoring Approach

### Prometheus & Grafana
- Native support in both Ethereum clients
- Easy Kubernetes integration via ServiceMonitors
- Customizable dashboards for specific blockchain metrics
- Extensive alerting capabilities

### Service Exposure
- Used ClusterIP with port-forwarding for security during testing
- Simulates how production would be configured with proper load balancers
- Provides controlled access to services

## Automation Tools

### Helm for Deployment
- Powerful templating for environment variations
- Built-in release management and rollbacks
- Manages dependencies between components

### Shell + Go Implementation
- Shell scripts for Kubernetes orchestration
- Go for health checks requiring complex HTTP/JSON handling
- Minimal external dependencies for portability