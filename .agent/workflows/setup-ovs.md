---
description: Setup and verify Open vSwitch networking environment
---
# Setup OVS Workflow

Setup Open vSwitch for VPC/SDN networking.

## Prerequisites
- Linux host with OVS installed
- Root/sudo access

## Steps

1. **Check OVS Installation**
// turbo
```bash
which ovs-vsctl || echo "OVS not installed"
```

2. **Start OVS Service**
```bash
sudo systemctl start openvswitch-switch
```

3. **Verify OVS Running**
// turbo
```bash
sudo ovs-vsctl show
```

4. **Create Test Bridge**
```bash
sudo ovs-vsctl add-br test-br0
```

5. **List Bridges**
// turbo
```bash
sudo ovs-vsctl list-br
```

6. **Cleanup Test Bridge**
```bash
sudo ovs-vsctl del-br test-br0
```

## Troubleshooting
- If OVS not found: `apt install openvswitch-switch` (Ubuntu) or `brew install openvswitch` (Mac)
- If permission denied: ensure running with sudo
