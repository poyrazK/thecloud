# ADR-015: Kubernetes CSI Driver Implementation

## Status
Accepted

## Context
As the platform evolves towards supporting stateful workloads in Kubernetes (KaaS), there is a requirement for dynamic block storage provisioning. Standard Kubernetes clusters expect a Container Storage Interface (CSI) driver to bridge the gap between PersistentVolumeClaims (PVCs) and the underlying cloud infrastructure.

Without a CSI driver, users must manually manage volume attachments and mounts, which is error-prone and non-scalable.

## Decision
We will implement a custom CSI Driver for The Cloud platform.

### Architecture
The driver follows the standard CSI 1.5+ specification and consists of:
1.  **Controller Service**: Responsible for creating, deleting, and attaching/detaching volumes via the Cloud REST API.
2.  **Node Service**: Runs on every worker node. Responsible for formatting devices and mounting them into pod directories.
3.  **Identity Service**: Provides metadata about the driver.

### Key Implementation Details
- **Sidecars**: We use standard Kubernetes CSI sidecars (`csi-provisioner`, `csi-attacher`, `node-driver-registrar`) to handle the heavy lifting of watching Kubernetes objects and calling our gRPC driver.
- **NodeID Resolution**: Kubernetes uses hostnames as `NodeId`. Our driver implements a resolution logic that calls `GET /instances/:hostname` to find the corresponding Cloud Instance UUID before performing attachments.
- **Mounter Abstraction**: To ensure 100% testability without root access, we introduced a `Mounter` interface. This allows mocking OS-level operations like `mkfs`, `mount`, and `blkid` in unit tests.
- **Protocol**: Communication between sidecars and the driver happens over a Unix Domain Socket (UDS).

### Security
- The driver requires a `CLOUD_API_KEY` to authenticate with the platform API.
- Node components run as `privileged: true` to allow mounting devices from the host namespace.

## Consequences
- **Pros**: Enables standard stateful applications (MySQL, Postgres, etc.) to run on KaaS. Provides seamless, automatic volume lifecycle management.
- **Cons**: Increases the maintenance surface area. Requires privileged pods on worker nodes.
- **Performance**: Volume attachment involves an API call and a Libvirt domain update, typically taking 2-5 seconds.
