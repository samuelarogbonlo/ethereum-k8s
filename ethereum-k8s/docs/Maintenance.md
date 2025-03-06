## Troubleshooting

### Common Issues

#### Pod Startup Problems

If pods are not starting:

```bash
# Check the pod status
kubectl get pods

# Check events
kubectl get events

# Check PV/PVC binding
kubectl get pv,pvc
```

#### Syncing Issues

If nodes are not syncing:

```bash
# Check Geth sync status
kubectl exec geth-0 -- geth attach --exec eth.syncing

# Check Lighthouse sync status
kubectl exec lighthouse-0 -- lighthouse bn sync_state
```

#### Monitoring Issues

If metrics are not showing in Grafana:

```bash
# Check Prometheus targets
kubectl port-forward -n monitoring svc/prometheus 9090:9090
# Then visit http://localhost:9090/targets

# Check node-exporter
kubectl logs -n monitoring -l app=node-exporter
```

#### Disk I/O Metrics Missing

If disk I/O metrics are not displayed:

```bash
# Make sure node-exporter is running
kubectl get pods -n monitoring -l app=node-exporter

# Check if the disk metrics are being collected
curl http://localhost:9100/metrics | grep node_disk
```

## Maintenance

### Updates

Update client versions:

```bash
# Edit the StatefulSet to update the image tag
kubectl edit statefulset geth
# Change image: ethereum/client-go:latest to the desired version

# Or apply an updated YAML file
kubectl apply -f geth/statefulset.yaml

# Check the rollout status
kubectl rollout status statefulset geth
```

### Backups

Back up blockchain data:

```bash
# Create a snapshot of the data
kubectl exec geth-0 -- geth snapshot

# Back up the PV data from the node (requires exec on the node)
tar -czf geth-data-backup.tar.gz /path/to/pv/data
```