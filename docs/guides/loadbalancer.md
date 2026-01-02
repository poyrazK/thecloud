# Load Balancer Guide

The Load Balancer distributes incoming network traffic across a group of backend instances. This increases the availability and fault tolerance of your applications.

## Concepts

### Load Balancer
The entry point for client traffic. It listens on a specific port and routes requests to registered targets.

- **Type**: Layer 7 (HTTP) - currently implemented as a TCP proxy.
- **Port**: The port where the LB listens (e.g., 80 or 8080).
- **Algorithm**: Round Robin (requests are distributed sequentially).

### Target Group
A logical grouping of target instances. Currently simplified: instances are added directly to the LB.

### Targets
The backend instances that process the requests.

## Architecture
The Load Balancer in Mini AWS is implemented using a dedicated proxy container (Nginx based) for each Load Balancer resource. This ensures isolation and mimics real cloud infrastructure behavior.

## CLI Commands

### Create a Load Balancer

```bash
cloud lb create \
  --name web-lb \
  --vpc <vpc-id> \
  --port 8080
```

### List Load Balancers

```bash
cloud lb list
```

### Add Targets

You can manually register instances with a Load Balancer.

```bash
cloud lb add-target   --instance <instance-id>
```

### Remove Targets

```bash
cloud lb remove-target   --instance <instance-id>
```

### Integration with Auto-Scaling
When creating an Auto-Scaling Group, you can specify a Load Balancer ID. The Auto-Scaling Service will automatically register newly launched instances with the LB and deregister terminated ones.

```bash
cloud autoscaling create ... --lb <lb-id>
```
