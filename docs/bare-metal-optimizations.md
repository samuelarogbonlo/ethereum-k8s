# Bare Metal Optimizations for Ethereum Nodes

This document outlines specific optimizations for running Ethereum nodes on bare metal hardware. These optimizations go beyond the basic Kubernetes deployment and focus on maximizing performance in a production environment.

## Networking Optimizations

### P2P Connection Optimization

1. **Jumbo Frames**:
   ```bash
   # Increase MTU size to 9000 (jumbo frames)
   ip link set eth0 mtu 9000

   # Make the change permanent
   echo "MTU=9000" >> /etc/sysconfig/network-scripts/ifcfg-eth0  # RHEL/CentOS
   # or add "post-up ip link set eth0 mtu 9000" to /etc/network/interfaces # Debian/Ubuntu
   ```

2. **TCP/IP Stack Tuning**:
   ```bash
   # Add to /etc/sysctl.conf or a file in /etc/sysctl.d/

   # Increase the maximum number of connections
   net.core.somaxconn=65535
   net.core.netdev_max_backlog=50000
   net.ipv4.tcp_max_syn_backlog=30000

   # Improve TCP connection handling
   net.ipv4.tcp_slow_start_after_idle=0
   net.ipv4.tcp_fin_timeout=15

   # Optimize TCP keepalive for P2P networks
   net.ipv4.tcp_keepalive_time=300
   net.ipv4.tcp_keepalive_probes=5
   net.ipv4.tcp_keepalive_intvl=15

   # Apply changes
   sysctl -p
   ```

3. **Port Forwarding and Firewall Rules**:
   - For optimal P2P connectivity, ensure both TCP and UDP ports are open:
     - Execution Client (Geth): Port 30303 (TCP/UDP)
     - Consensus Client (Lighthouse): Port 9000 (TCP/UDP)

   ```bash
   # Using UFW (Ubuntu)
   ufw allow 30303/tcp
   ufw allow 30303/udp
   ufw allow 9000/tcp
   ufw allow 9000/udp

   # Using iptables
   iptables -A INPUT -p tcp --dport 30303 -j ACCEPT
   iptables -A INPUT -p udp --dport 30303 -j ACCEPT
   iptables -A INPUT -p tcp --dport 9000 -j ACCEPT
   iptables -A INPUT -p udp --dport 9000 -j ACCEPT
   ```

4. **Static Peers Configuration**:
   - Configure static peers in your Geth and Lighthouse configurations to improve connectivity
   - Use DNS discovery to find reliable peers

## Data Persistence Optimization

### Storage Configuration

1. **RAID Configuration**:
   - For data redundancy: RAID 10 (mirror + stripe)
   - For pure performance: RAID 0 (stripe) with regular backups

   ```bash
   # Example mdadm RAID 10 setup with 4 drives
   mdadm --create /dev/md0 --level=10 --raid-devices=4 /dev/sda /dev/sdb /dev/sdc /dev/sdd
   ```

2. **Filesystem Optimization**:
   - Use XFS filesystem for better performance with large files
   - Mount options to improve performance:

   ```bash
   # Create XFS filesystem
   mkfs.xfs -f /dev/md0

   # Mount with optimized settings
   mount -o noatime,nodiratime,discard,attr2,nobarrier,logbufs=8 /dev/md0 /mnt/ethereum

   # Add to /etc/fstab for persistence
   echo "/dev/md0 /mnt/ethereum xfs noatime,nodiratime,discard,attr2,nobarrier,logbufs=8 0 0" >> /etc/fstab
   ```

3. **I/O Scheduler Tuning**:
   - Use deadline scheduler for SSDs or none for NVMe devices

   ```bash
   # Set scheduler for SSD
   echo deadline > /sys/block/sda/queue/scheduler

   # Set scheduler for NVMe
   echo none > /sys/block/nvme0n1/queue/scheduler

   # Increase readahead for sequential access patterns
   blockdev --setra 16384 /dev/md0
   ```

4. **Backup Strategy**:
   - Regular snapshots of chain data
   - Incremental backups to minimize downtime
   - Consider using LVM for snapshot capabilities:

   ```bash
   # Create LVM snapshot
   lvcreate -L 50G -s -n ethereum-snap /dev/vg0/ethereum-data

   # Backup snapshot
   rsync -avz /mnt/ethereum-snap/ /backup/ethereum/

   # Remove snapshot
   lvremove /dev/vg0/ethereum-snap
   ```

## System Tuning

### Kernel Parameters

1. **File Descriptors and Limits**:
   ```bash
   # Add to /etc/security/limits.conf
   * soft nofile 1048576
   * hard nofile 1048576
   * soft nproc 1048576
   * hard nproc 1048576

   # Add to /etc/sysctl.conf
   fs.file-max=1000000
   fs.nr_open=1000000
   ```

2. **Virtual Memory Optimization**:
   ```bash
   # Add to /etc/sysctl.conf

   # Reduce swappiness (minimize swap usage)
   vm.swappiness=1

   # Optimize dirty page handling
   vm.dirty_ratio=80
   vm.dirty_background_ratio=5
   vm.dirty_expire_centisecs=12000
   ```

3. **Enable Transparent Huge Pages**:
   ```bash
   # For database-like workloads (chain data)
   echo always > /sys/kernel/mm/transparent_hugepage/enabled
   ```

### NUMA-Aware Configuration

1. **CPU Pinning**:
   - Pin Ethereum processes to specific CPU cores to avoid context switching

   ```bash
   # Get NUMA topology
   numactl --hardware

   # Run Geth pinned to NUMA node 0
   numactl --cpunodebind=0 --membind=0 geth --config /etc/geth/config.toml

   # For Kubernetes, use CPU manager policy and topology manager
   # Edit kubelet config at /var/lib/kubelet/config.yaml
   ```

2. **Memory Allocation**:
   - Ensure memory is allocated from the same NUMA node as the CPUs:

   ```bash
   # Check current policy
   cat /proc/sys/kernel/numa_balancing

   # Disable automatic NUMA balancing
   echo 0 > /proc/sys/kernel/numa_balancing
   ```

3. **Network Interface Alignment**:
   - Align network processing with CPU cores handling Ethereum clients

   ```bash
   # Check NIC interrupt affinity
   cat /proc/interrupts | grep eth0

   # Set IRQ affinity for network interface
   # Example: assign IRQs to cores 0-3
   for i in $(grep eth0 /proc/interrupts | awk '{print $1}' | sed 's/://'); do
     echo "f" > /proc/irq/$i/smp_affinity
   done
   ```

## Time Synchronization

Accurate timekeeping is crucial for Ethereum nodes, especially for consensus clients:

```bash
# Install and configure chrony
apt install -y chrony

# Configure NTP servers
cat > /etc/chrony/chrony.conf << EOF
server 0.pool.ntp.org iburst
server 1.pool.ntp.org iburst
server 2.pool.ntp.org iburst
server 3.pool.ntp.org iburst
driftfile /var/lib/chrony/drift
makestep 1.0 3
rtcsync
EOF

# Restart chrony
systemctl restart chronyd
systemctl enable chronyd
```

## Kubernetes-Specific Optimizations

1. **Pod Quality of Service**:
   - Set appropriate resource requests and limits
   - Use Guaranteed QoS class for Ethereum clients

2. **Node Affinity and Anti-Affinity Rules**:
   - Use node affinity to place pods on specific hardware
   - Use pod anti-affinity to distribute clients properly

3. **Topology Manager Configuration**:
   - Enable Topology Manager in kubelet config
   - Set policy to "best-effort" or "single-numa-node"

4. **Priority Classes**:
   - Create custom PriorityClass for Ethereum nodes
   - Ensure clients are not evicted during resource pressure

## Monitoring and Alerting

1. **Prometheus Node Exporter**:
   - Install node_exporter with textfile collector
   - Add custom metrics for Ethereum-specific monitoring

2. **Custom Alert Rules**:
   - Set up alerts for chain sync issues
   - Monitor disk space and I/O latency
   - Alert on peer count drops

3. **Log Management**:
   - Centralize logs with Loki or similar
   - Create dashboards for quick issue identification

## Security Hardening

1. **Restrict RPC Access**:
   - Only expose RPC interfaces on private networks
   - Use TLS for all RPC connections
   - Implement API key authentication

2. **Regular Updates**:
   - Automated security patches for OS
   - Scheduled updates for Ethereum clients

3. **Network Segmentation**:
   - Separate management, monitoring, and P2P traffic

## Additional Considerations

1. **Client Diversity**:
   - Consider running multiple client implementations for redundancy

2. **Fallback Nodes**:
   - Configure standby nodes for critical applications
   - Set up automatic failover mechanisms

3. **Load Testing**:
   - Periodically test performance under high load conditions
   - Simulate network congestion and recovery