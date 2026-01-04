# CloudContainers (Container Service)

CloudContainers is a long-lived container orchestration service, similar to ECS or K8s Deployments.

## How it Works
- **Deployment**: A high-level resource defining `image`, `replicas`, and `ports`.
- **Reconciliation Worker**: A background loop (`ContainerWorker`) ensures the current count of running instances matches the desired `replicas`.
- **Instance Management**: Uses the core `InstanceService` to launch/terminate the actual Docker containers.

## Features
- **Auto-healing**: If an instance is lost, the worker detects the mismatch and launches a new one.
- **Scaling**: Simply update the replica count, and the worker will scale up or out on the next tick (15s).

## CLI Usage
```bash
# Deploy 3 replicas of nginx
cloud container deploy my-web nginx:latest --replicas 3 --ports 80:80

# Scale up
cloud container scale <deployment-id> 5

# List status
cloud container list
```
