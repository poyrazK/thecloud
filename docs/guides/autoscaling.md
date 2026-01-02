# Auto-Scaling Guide

Auto-Scaling in Mini AWS allows you to automatically adjust the number of compute instances in response to changing load. This ensures your applications remain available and performant while optimizing resource usage.

## Concepts

### Scaling Group
A Scaling Group acts as a logical container for a collection of identical instances. It manages the lifecycle of these instances, ensuring the desired number of replicas are running.

- **Min/Max Instances**: The boundaries for the group size. The number of instances will never go below `min` or above `max`.
- **Desired Capacity**: The ideal number of instances the group should maintain.
- **Load Balancer**: Optional integration. New instances are automatically registered as targets with the specified Load Balancer.

### Scaling Policy
A Scaling Policy defines how the group should react to metrics.

- **Target Tracking**: The policy tries to keep a specific metric (e.g., CPU) at a target value (e.g., 50%).
- **Scale Out/In Steps**: How many instances to add or remove when a scaling action is triggered.
- **Cooldown**: A period after a scaling action during which no further actions are taken, preventing oscillation (flapping).

## CLI Commands

### Create a Scaling Group

```bash
cloud autoscaling create \
  --name web-asg \
  --vpc <vpc-id> \
  --image nginx:alpine \
  --ports 80:80 \
  --min 1 \
  --max 5 \
  --desired 2 \
  --lb <lb-id>
```

### List Scaling Groups

```bash
cloud autoscaling list
```

### Create a Scaling Policy

```bash
# Add a CPU policy with target 50%, adding/removing 1 instance at a time
cloud autoscaling add-policy <group-id> \
  --name cpu-policy \
  --metric cpu \
  --target 50 \
  --scale-out 1 \
  --scale-in 1 \
  --cooldown 60
```

### Delete a Scaling Group

```bash
cloud autoscaling rm <group-id>
```
*Note: This will terminate all instances associated with the group.*

## Dynamic Ports

When running locally, binding multiple instances to the same host port (e.g., `80:80`) will cause conflicts. To avoid this, use dynamic host ports by specifying `0` as the host port:

```bash
cloud autoscaling create ... --ports 0:80
```
This will allow Docker to assign a random available port on the host for each instance.

## Metrics
The Auto-Scaling worker runs in the background and evaluates policies every 15 minutes by default (configurable). It queries the metrics history of instances to calculate the average utilization.
