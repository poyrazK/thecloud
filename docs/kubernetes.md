# Kubernetes as a Service (KaaS)

The Cloud Platform provides a fully managed Kubernetes experience, allowing users to provision and manage highly available Kubernetes clusters on top of our virtualized infrastructure.

## Overview

Our KaaS solution leverages `kubeadm` for standard, compliant cluster bootstrapping, integrating deeply with our Compute, Networking, and Security primitives.

### Key Features
*   **Fully Automated Provisioning**: One-command cluster creation (Control Plane + Workers).
*   **Integrated Networking**: Built-in OVS integration with Calico CNI for Pod networking.
*   **Security by Default**:
    *   Dedicated Security Groups automatically created and attached.
    *   Encrypted Kubeconfig storage.
    *   SSH key management for secure node access.
*   **Standard Compliance**: Uses upstream Kubernetes (v1.29.0 default).
*   **High Availability & Self-Healing**:
    *   **Cluster Reconciliation**: A background `ClusterReconciler` worker periodically audits cluster health.
    *   **Automatic Repair**: Automatically detects and repairs unhealthy clusters (e.g., API server down or nodes not ready) by re-applying configurations and reconciling the desired node count.

## Architecture

A minimal cluster consists of:
1.  **Control Plane Node**: Hosts the API Server, Controller Manager, Scheduler, and Etcd.
2.  **Worker Nodes**: Run your workloads (`pods`).
3.  **Networking**:
    *   **Pod Network**: 192.168.0.0/16 (Calico VXLAN).
    *   **Service Network**: 10.96.0.0/12.
    *   **NodePort Range**: 30000-32767.

### Security Groups

Every cluster gets a dedicated Security Group (e.g., `sg-my-cluster`) with the following rules:

| Protocol | Port / Range | Source      | Description |
| :--- | :--- | :--- | :--- |
| TCP      | 6443         | 0.0.0.0/0   | Kubernetes API Server |
| UDP      | 4789         | 0.0.0.0/0   | VXLAN Overlay (Calico) |
| TCP      | 179          | 0.0.0.0/0   | BGP (Calico) |
| TCP      | 10250        | 0.0.0.0/0   | Kubelet API |
| TCP      | 30000-32767  | 0.0.0.0/0   | NodePort Services |
| UDP      | 30000-32767  | 0.0.0.0/0   | NodePort Services |
| ARP      | N/A          | N/A         | L2 Discovery |

## Getting Started

### Prerequisites
*   The Cloud CLI installed.
*   An active VPC.

### 1. Create a Cluster

```bash
# Create a cluster named "dev-cluster" with 2 worker nodes
cloud k8s create \
  --name dev-cluster \
  --vpc <vpc-id> \
  --workers 2 \
  --version v1.29.0
```

Provisioning typically takes 2-5 minutes as it involves VM booting, package installation, and cluster bootstrapping.

### 2. Monitor Status

```bash
cloud k8s list
```

Wait until the status changes from `PROVISIONING` to `RUNNING`.

### 3. Configure kubectl

Retrieve the `kubeconfig` for your cluster:

```bash
cloud k8s kubeconfig <cluster-id> > kubeconfig.yaml
export KUBECONFIG=./kubeconfig.yaml
```

### 4. Verify Access

```bash
kubectl get nodes
kubectl get pods -A
```

## Advanced Usage

### Persistent Volumes (CSI)
The Cloud supports dynamic volume provisioning via its custom CSI driver.

#### How it works:
1.  **StorageClass**: The `thecloud-block` storage class is pre-configured in every cluster.
2.  **Dynamic Provisioning**: When a user creates a `PersistentVolumeClaim` (PVC), the Cloud's CSI Controller sidecar detects it and calls our platform API to create a block device.
3.  **Attachment**: When a Pod using the PVC is scheduled, the CSI Attacher sidecar calls our API to attach the volume to the corresponding virtual machine.
4.  **Hostname Resolution**: Since Kubernetes identifies nodes by hostname, the CSI driver automatically resolves hostnames to internal platform Instance UUIDs before performing any operations.
5.  **Formatting & Mounting**: The CSI Node service running on the worker node automatically formats the raw device (ext4) and mounts it into the Pod's filesystem.

#### PVC Example:
```yaml
apiVersion: v1
kind: PersistentVolumeClaim
metadata:
  name: mysql-pv-claim
spec:
  storageClassName: thecloud-block
  accessModes:
    - ReadWriteOnce
  resources:
    requests:
      storage: 20Gi
```

### Node Access (SSH)
You can SSH into your cluster nodes using the private key stored securely in The Cloud. Currently, this requires retrieving specific node IP addresses via `cloud instance list`.

### Exposing Services
Use `type: NodePort` to expose services. The Cloud's security groups automatically allow traffic on ports 30000-32767.

```yaml
apiVersion: v1
kind: Service
metadata:
  name: my-service
spec:
  type: NodePort
  selector:
    app: my-app
  ports:
    - port: 80
      targetPort: 80
      nodePort: 30080
```

## Limitations (MVP)
*   **Persistent Volumes**: Dynamic storage provisioning (CSI) is now available in Phase 7.
*   **Multi-Master**: While we have self-healing for single master, multi-control plane HA is a future roadmap item.
