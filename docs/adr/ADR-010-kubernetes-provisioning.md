# ADR-010: Kubernetes Provisioning Strategy

## Status
Accepted

## Context
The Cloud Platform requires a capability to provision managed Kubernetes clusters ("KaaS") for users. This feature must allow users to spin up a control plane and multiple worker nodes, networked together securely.

We evaluated several approaches for bootstrapping Kubernetes on our virtualized infrastructure:

1.  **K3s**: A lightweight Kubernetes distribution.
    *   *Pros*: Fast, single,binary, low resource footprint.
    *   *Cons*: Deviates slightly from upstream (e.g., uses SQLite by default, different flag arguments), which might confuse users looking for a "standard" K8s experience.
2.  **Kubeadm**: The official tool for bootstrapping Kubernetes clusters.
    *   *Pros*: Industry standard, widely documented, creates a conformant cluster, flexible configuration.
    *   *Cons*: Slightly higher complexity (requires pre-installing container runtimes, CNI configuration).
3.  **Kind / Minikube**: Tools for running K8s inside Docker.
    *   *Pros*: Trivial to start.
    *   *Cons*: Not designed for multi-node VM-based orchestration or production workloads; creates a "cluster in a container" rather than a cluster of infrastructure.

## Decision
We decided to use **kubeadm** to provision Kubernetes clusters, augmented by a **Redis-backed Task Queue** for asynchronous durability.

The provisioning workflow is implemented as follows:
1.  **Job Enqueue**: The API enqueues a `ClusterJob` and returns a `202 Accepted` status.
2.  **Infrastructure Provisioning**: The `ClusterWorker` handles the job:
    *   For HA clusters: Creates an **API Server Load Balancer**.
    *   Creates control plane and worker instances.
3.  **Bootstrapping**: A `KubeadmProvisioner` installs prerequisites and kernel modules.
4.  **Clustering**:
    *   Standard: `kubeadm init` on 1 master.
    *   HA: `kubeadm init` on the first master with `--upload-certs`, and `kubeadm join --control-plane` on subsequent masters.
5.  **Networking & Storage**: Install Calico CNI and provision local storage drivers.
6.  **Security**: Admin `kubeconfig` is encrypted and stored using `SecretService`.

## Consequences

### Positive
*   **Standardization**: Users get a standard Kubernetes environment compatible with all standard tools.
*   **High Availability**: Production-ready 3-node HA control plane support.
*   **Durable Operations**: Operations can survive API restarts thanks to the Task Queue.

### Negative
*   **Provisioning Time**: Bootstrapping a full K8s cluster takes minutes (VM boot + package install) compared to seconds for K3s.
*   **Complexity**: The `KubeadmProvisioner` must handle SSH connectivity, retries, and potential script failures on remote hosts.

## Compliance
This approach fully isolates clusters using our `VPC` and `SecurityGroup` primitives, ensuring multi-tenancy security.
