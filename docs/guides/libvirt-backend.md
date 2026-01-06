# Libvirt Compute Backend Guide

## Overview

The Cloud supports **Libvirt** as an alternative compute backend to Docker, enabling you to run your cloud infrastructure on KVM/QEMU virtual machines instead of containers. This provides true VM isolation, support for different operating systems, and the ability to run workloads that require kernel-level access.

## Architecture

### Components

- **LibvirtAdapter**: Implements the `ComputeBackend` interface using `go-libvirt`
- **Cloud-Init Integration**: Automatic VM configuration via ISO injection
- **Storage**: QCOW2 volumes in libvirt storage pools
- **Networking**: NAT networks with DHCP and port forwarding via iptables
- **Load Balancing**: Host-level Nginx process for HTTP traffic distribution

### How It Works

1. **Instance Creation**:
   - Creates a QCOW2 root volume in the storage pool
   - Generates a Cloud-Init ISO with environment variables and startup commands
   - Defines a KVM domain with the volume and ISO attached
   - Starts the VM and sets up port forwarding rules

2. **Networking**:
   - VMs connect to libvirt NAT networks
   - DHCP leases are tracked for IP resolution
   - Port forwarding uses iptables DNAT rules on the host

3. **Storage**:
   - Persistent volumes are additional QCOW2 disks attached to VMs
   - Snapshots use `qemu-img convert` to create compressed backups

## Prerequisites

### System Requirements

- **Linux host** with KVM support (check: `egrep -c '(vmx|svm)' /proc/cpuinfo` should return > 0)
- **libvirt daemon** installed and running
- **QEMU/KVM** virtualization packages
- **genisoimage** or **mkisofs** for Cloud-Init ISO generation
- **qemu-img** for volume management

### Installation

#### Ubuntu/Debian
```bash
sudo apt update
sudo apt install -y \
  qemu-kvm \
  libvirt-daemon-system \
  libvirt-clients \
  bridge-utils \
  genisoimage \
  qemu-utils

# Add your user to the libvirt group
sudo usermod -aG libvirt $USER
newgrp libvirt
```

#### Fedora/RHEL/CentOS
```bash
sudo dnf install -y \
  qemu-kvm \
  libvirt \
  libvirt-client \
  bridge-utils \
  genisoimage \
  qemu-img

sudo systemctl enable --now libvirtd
sudo usermod -aG libvirt $USER
newgrp libvirt
```

#### Arch Linux
```bash
sudo pacman -S \
  qemu \
  libvirt \
  bridge-utils \
  cdrtools \
  edk2-ovmf

sudo systemctl enable --now libvirtd
sudo usermod -aG libvirt $USER
newgrp libvirt
```

### Verify Installation

```bash
# Check libvirt is running
systemctl status libvirtd

# Verify KVM module is loaded
lsmod | grep kvm

# Test virsh connection
virsh list --all

# Check default network
virsh net-list --all
```

## Configuration

### 1. Set Up Storage Pool

Create a storage pool for VM volumes:

```bash
# Create directory
sudo mkdir -p /var/lib/libvirt/images

# Define and start the default pool
virsh pool-define-as default dir --target /var/lib/libvirt/images
virsh pool-start default
virsh pool-autostart default

# Verify
virsh pool-list
```

### 2. Set Up Default Network

```bash
# Define default NAT network
cat > /tmp/default-net.xml << 'EOF'
<network>
  <name>default</name>
  <forward mode='nat'/>
  <bridge name='virbr0' stp='on' delay='0'/>
  <ip address='192.168.122.1' netmask='255.255.255.0'>
    <dhcp>
      <range start='192.168.122.2' end='192.168.122.254'/>
    </dhcp>
  </ip>
</network>
EOF

virsh net-define /tmp/default-net.xml
virsh net-start default
virsh net-autostart default

# Verify
virsh net-list
```

### 3. Configure The Cloud

Set the compute backend environment variable:

```bash
export COMPUTE_BACKEND=libvirt
```

Or add it to your `.env` file:

```env
COMPUTE_BACKEND=libvirt
```

### 4. Permissions

Ensure The Cloud process has access to the libvirt socket:

```bash
# Option 1: Run as a user in the libvirt group (recommended)
sudo usermod -aG libvirt $USER

# Option 2: Temporarily allow broader access (testing only)
sudo chmod 666 /var/run/libvirt/libvirt-sock
```

## Usage

### Starting The Cloud with Libvirt

```bash
# Set the backend
export COMPUTE_BACKEND=libvirt

# Run the API server
make run

# Or with docker-compose
docker-compose up -d
```

### Creating Instances

The API remains the same - just set the backend:

```bash
# Create an instance
curl -X POST http://localhost:8080/instances \
  -H "X-API-Key: YOUR_KEY" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "my-vm",
    "image": "ubuntu-22.04",
    "instance_type": "t2.micro"
  }'
```

### Using Cloud-Init

Environment variables and startup commands are automatically injected via Cloud-Init:

```bash
# Create a database with environment configuration
curl -X POST http://localhost:8080/databases \
  -H "X-API-Key: YOUR_KEY" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "mydb",
    "engine": "postgres",
    "version": "14"
  }'
```

The Libvirt adapter will:
1. Generate a Cloud-Init ISO with the database credentials
2. Attach it to the VM as a CD-ROM
3. The VM boots and reads the configuration automatically

## Advanced Features

### Port Forwarding

Port forwarding is handled automatically via iptables:

```bash
# When you create an instance with ports
curl -X POST http://localhost:8080/instances \
  -H "X-API-Key: YOUR_KEY" \
  -d '{
    "name": "web-server",
    "image": "nginx",
    "ports": ["8080:80"]
  }'
```

The adapter will:
1. Wait for the VM to get a DHCP lease
2. Create an iptables DNAT rule: `host:8080 → vm_ip:80`
3. Store the mapping for `GetInstancePort` queries

### Volume Attachment

Persistent volumes are attached as additional virtio disks:

```bash
# Create a volume
curl -X POST http://localhost:8080/volumes \
  -H "X-API-Key: YOUR_KEY" \
  -d '{"name": "data-vol", "size_gb": 100}'

# Attach to instance
curl -X POST http://localhost:8080/instances/my-vm/attach \
  -H "X-API-Key: YOUR_KEY" \
  -d '{"volume_id": "vol-123"}'
```

The volume appears as `/dev/vdb` (or `/dev/vdc`, etc.) inside the VM.

### Snapshots

Volume snapshots use `qemu-img` for efficient QCOW2 compression:

```bash
# Create a snapshot
curl -X POST http://localhost:8080/volumes/vol-123/snapshots \
  -H "X-API-Key: YOUR_KEY" \
  -d '{"name": "backup-2024"}'

# Restore from snapshot
curl -X POST http://localhost:8080/volumes/vol-123/restore \
  -H "X-API-Key: YOUR_KEY" \
  -d '{"snapshot_id": "snap-456"}'
```

### Load Balancing

The Libvirt backend uses a host-level Nginx process:

```bash
# Create a load balancer
curl -X POST http://localhost:8080/loadbalancers \
  -H "X-API-Key: YOUR_KEY" \
  -d '{
    "name": "web-lb",
    "port": 80,
    "algorithm": "round-robin",
    "targets": [
      {"instance_id": "i-001", "port": 8080},
      {"instance_id": "i-002", "port": 8080}
    ]
  }'
```

The adapter will:
1. Resolve instance IDs to VM IPs via DHCP leases
2. Generate an Nginx config with upstream servers
3. Start/reload Nginx on the host

## Troubleshooting

### VMs Not Getting IP Addresses

**Problem**: VMs start but never get DHCP leases.

**Solution**:
```bash
# Check if the default network is active
virsh net-list

# Restart the network
virsh net-destroy default
virsh net-start default

# Check dnsmasq is running
ps aux | grep dnsmasq
```

### Permission Denied Errors

**Problem**: `connect: permission denied` when accessing libvirt socket.

**Solution**:
```bash
# Verify group membership
groups | grep libvirt

# If not in the group, add yourself
sudo usermod -aG libvirt $USER
newgrp libvirt

# Or temporarily fix permissions (testing only)
sudo chmod 666 /var/run/libvirt/libvirt-sock
```

### Cloud-Init Not Working

**Problem**: Environment variables or commands not executing in VMs.

**Solution**:
```bash
# Check if genisoimage is installed
which genisoimage || which mkisofs

# Verify the ISO was created
ls -lh /tmp/cloud-init-*.iso

# Check ISO contents
isoinfo -l -i /tmp/cloud-init-*.iso

# Ensure your VM image supports Cloud-Init
# (Ubuntu Cloud Images, Fedora Cloud, etc.)
```

### Port Forwarding Not Working

**Problem**: Cannot access VM services via host ports.

**Solution**:
```bash
# Check iptables rules
sudo iptables -t nat -L PREROUTING -n -v

# Verify the VM has an IP
virsh net-dhcp-leases default

# Test connectivity from host
ping <vm-ip>
curl http://<vm-ip>:<port>
```

### Storage Pool Issues

**Problem**: `Storage pool not found: no storage pool with matching name 'default'`

**Solution**:
```bash
# List all pools
virsh pool-list --all

# If default doesn't exist, create it
virsh pool-define-as default dir --target /var/lib/libvirt/images
virsh pool-start default
virsh pool-autostart default

# Refresh the pool
virsh pool-refresh default
```

## Performance Tuning

### CPU Pinning

For better performance, pin VMs to specific CPU cores:

```xml
<!-- In domain XML -->
<vcpu placement='static' cpuset='0-3'>4</vcpu>
```

### Huge Pages

Enable huge pages for large memory VMs:

```bash
# Reserve huge pages
echo 1024 > /proc/sys/vm/nr_hugepages

# In domain XML
<memoryBacking>
  <hugepages/>
</memoryBacking>
```

### Virtio Drivers

Always use virtio for best I/O performance:
- Network: `<model type='virtio'/>`
- Disk: `<target dev='vda' bus='virtio'/>`

## Comparison: Docker vs Libvirt

| Feature | Docker | Libvirt |
|---------|--------|---------|
| **Isolation** | Process-level | Full VM isolation |
| **Boot Time** | Instant | 10-30 seconds |
| **Resource Overhead** | Minimal | Moderate (hypervisor) |
| **OS Support** | Linux containers only | Any OS (Windows, BSD, etc.) |
| **Kernel Access** | Shared kernel | Dedicated kernel |
| **Networking** | Bridge/overlay | NAT/bridge/macvtap |
| **Storage** | Overlay2/volumes | QCOW2/raw images |
| **Use Cases** | Microservices, CI/CD | Legacy apps, multi-OS, security |

## Best Practices

1. **Use Cloud Images**: Always use official cloud images (Ubuntu Cloud, Fedora Cloud) that support Cloud-Init
2. **Storage Pools**: Keep pools on fast storage (SSD/NVMe) for better VM performance
3. **Network Isolation**: Create separate networks for different VPCs
4. **Resource Limits**: Set appropriate CPU and memory limits in domain XML
5. **Monitoring**: Use `virsh domstats` for VM metrics
6. **Backups**: Regularly snapshot volumes before major changes
7. **Security**: Use AppArmor/SELinux profiles for VM isolation

## Migration from Docker

To migrate existing workloads from Docker to Libvirt:

1. **Export Docker volumes** to tarballs
2. **Create QCOW2 volumes** in libvirt
3. **Restore data** using the snapshot restore mechanism
4. **Update environment** variable: `COMPUTE_BACKEND=libvirt`
5. **Recreate instances** using the same API calls

## Further Reading

- [Libvirt Documentation](https://libvirt.org/docs.html)
- [Cloud-Init Documentation](https://cloudinit.readthedocs.io/)
- [KVM Performance Tuning](https://www.linux-kvm.org/page/Tuning_KVM)
- [QEMU Documentation](https://www.qemu.org/documentation/)

## Support

For issues specific to the Libvirt backend:
1. Check the troubleshooting section above
2. Review logs: `journalctl -u libvirtd -f`
3. Test with the manual test script: `go run scripts/test_libvirt.go`
4. Open an issue on GitHub with libvirt version and error logs
