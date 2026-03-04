# Kubernetes Cluster Autoscaler

The Cloud Platform provides native support for Kubernetes Cluster Autoscaling, enabling your clusters to automatically adjust the number of worker nodes based on the resource requirements of your workloads.

## Overview

The Cluster Autoscaler (CA) automatically increases the size of a Node Group when pods fail to schedule due to insufficient resources and decreases it when nodes are underutilized for an extended period.

Our implementation uses the **gRPC External Provider** model, running a sidecar bridge that communicates between the upstream Autoscaler and The Cloud's scaling APIs.

## Key Concepts

### Node Groups (Node Pools)
A Node Group is a logical grouping of identical worker nodes within a cluster. Scaling operations are performed at the Node Group level.
- **Min Size**: The minimum number of nodes the group must maintain.
- **Max Size**: The upper limit for scaling.
- **Desired Size**: The current target set by the platform or CA.

## Configuration

When a cluster is created, a `default-pool` is automatically provisioned. You can manage additional node pools via the Cloud CLI or API.

### Automatic Metadata
Every worker node launched by the platform is injected with the following metadata used by the autoscaler:
- `thecloud.io/cluster-id`: The unique ID of the cluster.
- `thecloud.io/node-group`: The name of the node pool the instance belongs to.

## Architecture

The Autoscaler is deployed as a Deployment in the `kube-system` namespace consisting of two containers:
1.  **cluster-autoscaler**: The official upstream image configured with the `externalgrpc` provider.
2.  **thecloud-bridge**: Our gRPC server that translates scaling commands into The Cloud API calls.

### Scaling Up Flow
1.  A pod is in `Pending` status due to CPU/Memory constraints.
2.  Upstream CA identifies that a Node Group can satisfy the requirement.
3.  Upstream CA calls `NodeGroupIncreaseSize` via gRPC.
4.  `thecloud-bridge` calls the Node Group update endpoint (`PUT /clusters/{id}/nodegroups/{name}`) on The Cloud API.
5.  Cloud API launches a new VM; the VM joins the cluster via cloud-init.

### Scaling Down Flow
1.  Upstream CA identifies a node has been unneeded for >5 minutes.
2.  Upstream CA drains the node (evicts pods).
3.  Upstream CA calls `NodeGroupDeleteNodes` via gRPC.
4.  `thecloud-bridge` calls `DELETE /instances/{id}` on The Cloud API.

## Usage via CLI

### List Node Groups
```bash
cloud k8s show <cluster-id> --json | jq '.node_groups'
```

### Add a Node Group
```bash
# Currently supported via API, CLI command coming soon
curl -X POST $CLOUD_API_URL/clusters/$CLUSTER_ID/nodegroups \
  -H "X-API-Key: $CLOUD_API_KEY" \
  -d '{
    "name": "high-mem-pool",
    "instance_type": "high-mem-1",
    "min_size": 1,
    "max_size": 5,
    "desired_size": 1
  }'
```
