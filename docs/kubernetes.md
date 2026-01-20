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
*   **Single Master**: Currently supports single control plane node (no HA).
*   **LoadBalancers**: `type: LoadBalancer` is not yet integrated with the platform LB. Use NodePort or Ingress.
*   **Persistent Volumes**: Dynamic storage provisioning (CSI) is planned for Phase 7.
