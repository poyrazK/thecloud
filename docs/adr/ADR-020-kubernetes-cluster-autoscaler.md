# ADR-020: Kubernetes Cluster Autoscaler Strategy

## Status
Accepted

## Context
Kubernetes clusters on "The Cloud" need to dynamically adjust their node count based on pod resource requirements. The industry standard for this is the [Kubernetes Cluster Autoscaler](https://github.com/kubernetes/autoscaler). 

Implementing an "out-of-tree" provider for the Cluster Autoscaler presents two main technical challenges:
1.  **Dependency Conflicts**: Directly importing the Cluster Autoscaler Go library introduces significant version conflicts with Kubernetes internal modules (e.g., `kube-scheduler`, `kubernetes`) that use unversioned internal dependencies.
2.  **Platform Abstraction**: The platform must support multiple types of worker nodes (high-mem, general purpose) which requires a "Node Group" model instead of a single flat worker count.

## Decision
**Implement the Cluster Autoscaler using the gRPC External Provider pattern.**

We will build a standalone gRPC bridge (`autoscaler-server`) that implements the official `external_grpc.proto` specification. The official upstream Cluster Autoscaler image will be deployed as a sidecar to this bridge.

Additionally, we will **refactor the platform domain to support Node Groups (Node Pools)** as the primary unit of scaling.

## Consequences

### Positive
*   **Dependency Isolation**: By using gRPC as the boundary, we avoid importing problematic Kubernetes internal libraries into the main platform codebase.
*   **Standard Compliance**: We can use the official, unmodified Cluster Autoscaler container image, benefiting from its battle-tested scaling algorithms and community fixes.
*   **Flexibility**: The gRPC bridge can be updated independently of the platform API or the Kubernetes cluster version.
*   **Rich Metadata**: Using Node Groups allows users to define heterogeneous clusters with different machine types and scaling boundaries.

### Negative
*   **Operational Overhead**: Requires running an additional gRPC server (`thecloud-bridge`) within the cluster.
*   **Complexity**: Adds a gRPC layer between the Autoscaler and the Cloud API.

### Trade-off Justification
The gRPC External Provider is the recommended way for new cloud providers to integrate with the Cluster Autoscaler without becoming entangled in the Kubernetes repository's complex dependency graph. This ensures long-term maintainability and allows us to focus on the platform-specific scaling logic.
