# 🌟 The Cloud - Features & Implementation Deep Dive

This document provides a comprehensive overview of every feature currently implemented in **The Cloud**, including the technologies used and the architectural details of how they are built.

---

## 🏗️ Core Infrastructure

### 1. Compute (Instances)
**What it is**: Launch virtual machines or containers to run applications.

**Tech Stack**: 
- **Docker Engine** (Container Backend - Simulators)
- **Libvirt/KVM** (VM Backend - Production)
- **Go** (Backend with unified `ComputeBackend` interface)

**Implementation**:
- **Multi-Backend Architecture**: The system supports two compute backends via a unified `ComputeBackend` interface:
  
  **Docker Backend** (Simulation/Dev):
  - **Simulation**: Uses Docker Containers to simulate instances with sub-second boot times.
  - **Isolation**: Process-level isolation via namespaces.
  - **Networking**: Instances attach to custom Bridge Networks (simulating VPCs).
  - **Persistence**: Docker Volumes for persistent storage.
  
  **Libvirt Backend** (Production/KVM):
  - **Real VMs**: Uses KVM/QEMU for full hardware virtualization.
  - **Isolation**: Complete VM isolation with dedicated kernels.
  - **Console**: **VNC Console** access via dynamic websockets (`/console` endpoint).
  - **Storage**: Supports both file-based QCOW2 and raw LVM block devices.
  - **Networking**: Integrated with Open vSwitch (OVS) for true SDN.

- **Backend Selection**: Set via `COMPUTE_BACKEND` environment variable (`docker` or `libvirt`).
- **Lifecycle**: The `InstanceService` manages the backend API to Create, Start, Stop, and Remove instances.

### 2. Networking (VPC)
**What it is**: Isolated virtual networks to secure resources.
**Tech Stack**: Docker Networks (Bridge) or Open vSwitch (OVS).
**Implementation**:
- **Docker Mode**: A "VPC" maps directly to a **Docker Bridge Network**.
- **Libvirt Mode**: Uses **Open vSwitch (OVS)** bridges and VXLANs for tenant isolation.
- **Isolation**: strict traffic segregation rules enforced by generic or OVS flow rules.

### 3. Block Storage (Volumes)
**What it is**: Persistent disks that can be attached/detached from instances.
**Tech Stack**: Docker Volumes or Linux LVM.
**Implementation**:
- **Docker Mode**: Maps to `docker volume create`.
- **LVM Mode**:
  - **Creation**: Allocates raw logical volumes (`lvcreate`) from a volume group.
  - **Snapshots**: Instant, copy-on-write snapshots for backups.
  - **Performance**: Near-native block device performance for VMs.
- **Attachment**: Hot-pluggable (in Libvirt mode) or bind-mounted (in Docker mode).
- **Persistence**: Data survives instance termination.

### 4. Object Storage (S3-compatible)
**What it is**: Store and retrieve files (blobs) via API with enterprise-grade features.
**Tech Stack**: Go (IO/FS), Local Filesystem, AES-GCM Encryption.
**Implementation**:

**Bucket Management**:
- **Create/Delete Buckets**: Full lifecycle management with validation.
- **Bucket Versioning**: Enable/disable versioning per bucket to preserve object history.
- **Bucket Listing**: List all buckets for a user/tenant.

**Object Operations**:
- **Upload/Download**: Stream-based PUT/GET with efficient memory usage.
- **List Objects**: Enumerate all objects within a bucket.
- **Delete**: Remove objects with optional version targeting.

**Versioning**:
- **Version History**: List all versions of an object.
- **Version Download**: Retrieve specific historical versions.
- **Version Delete**: Remove specific versions while preserving others.

**Multipart Upload** (for large files):
- **Initiate**: Start a multipart upload session.
- **Upload Parts**: Upload chunks in parallel with part numbers.
- **Complete**: Assemble all parts into final object.
- **Abort**: Cancel and clean up incomplete uploads.

**Security & Access**:
- **Presigned URLs**: Generate temporary signed URLs for time-limited access.
- **Encryption**: Objects encrypted at rest using AES-GCM via EncryptionService.
- **Audit Trail**: All operations logged for compliance.

**Distributed Storage**:
- **Cluster Status**: Monitor distributed storage cluster health.
- **Multi-node Support**: Architecture ready for horizontal scaling.

---

## 🛠️ Managed Services

### 5. Managed Databases (RDS)
**What it is**: Provision fully managed PostgreSQL or MySQL databases.
**Tech Stack**: Docker (Official Images), Go.
**Implementation**:
- **Multi-Engine Support**: PostgreSQL and MySQL with configurable versions.
- **Provisioning**: Spawns Docker containers using official images (`postgres:<version>-alpine`, `mysql:<version>`).
- **Credentials**: Auto-generates secure passwords (16-char random) and default usernames.
- **VPC Integration**: Databases can be deployed into specific VPCs for network isolation.
- **Connection Strings**: `GetConnectionString()` API returns ready-to-use connection URLs.
- **Event & Audit**: All operations logged for compliance tracking.
- **Metrics**: Prometheus metrics for RDS instance counts by engine.

### 6. Managed Caches (CloudCache)
**What it is**: Provision fully managed Redis instances.
**Tech Stack**: Redis (Alpine), Go, Docker Exec.
**Implementation**:
- **Provisioning**: Launches `redis-server` with custom configuration (AOF enabled, RDB disabled).
- **Security**: Generates a random password and enables `--requirepass`.
- **VPC Integration**: Caches can be deployed into specific VPCs.
- **Management Operations**:
  - **FlushCache**: Executes `FLUSHALL` via Docker Exec.
  - **GetCacheStats**: Parses `redis-cli INFO` for connected clients, memory usage, keys count, uptime.
- **Connection Strings**: API returns ready-to-use Redis URLs with auth.
- **Lookup**: Get cache by ID or name for flexibility.

### 7. Cloud Functions (Serverless)
**What it is**: Run code snippets without provisioning servers.
**Tech Stack**: Docker (Ephemeral Containers), Go.
**Implementation**:
- **Multi-Runtime Support**:
  - `nodejs20` - Node.js 20 Alpine
  - `python312` - Python 3.12 Alpine
  - `go122` - Go 1.22 Alpine
  - `rust` - Rust 1.75 Alpine
  - `java21` - Eclipse Temurin 21 Alpine
- **Code Deployment**: Supports raw code or ZIP archives with auto-extraction.
- **Execution**: One-shot containers with configurable timeouts.
- **Async Invocation**: Background execution via Go routines.
- **Invocation Logs**: Stored history of function executions with output/errors.
- **Handler Configuration**: Specify entry point file for each function.

### 8. Load Balancer (ELB)
**What it is**: Distribute incoming HTTP traffic across multiple instances.
**Tech Stack**: Go (Reverse Proxy `httputil`), Configurable Algorithms.
**Implementation**:
- **VPC-Aware**: Load balancers are scoped to VPCs for network isolation.
- **Algorithms**: Supports `round-robin` (default) with architecture for additional algorithms.
- **Target Management**:
  - **AddTarget**: Register instances with port and weight.
  - **RemoveTarget**: Deregister instances.
  - **ListTargets**: View all registered targets.
- **Cross-VPC Validation**: Prevents adding instances from different VPCs.
- **Health Tracking**: Target health status tracking (`unknown`, `healthy`, `unhealthy`).
- **Idempotency**: Idempotency keys prevent duplicate LB creation.
- **Versioning**: Optimistic locking via version field for concurrent updates.

### 9. Managed Kubernetes (KaaS) 🆕
**What it is**: Provision and manage production-ready Kubernetes clusters.
**Tech Stack**: Kubeadm, Containerd, LoadBalancer, Go Workers.
**Implementation**:
- **Async Provisioning**: Uses a Redis-backed **Task Queue** to handle cluster creation/deletion asynchronously. 
- **High Availability (HA)**: Supports 3-node HA Control Plane with an automated API Server Load Balancer.
- **Node Management**: Automatically bootstraps nodes with `kubeadm` and required CNI plugins.
- **Isolation**: Each cluster is isolated within its own VPC.

### 10. Auto-Scaling
**What it is**: Automatically add/remove instances based on CPU load.
**Tech Stack**: Go Background Workers.
**Implementation**:
- **Metrics Loop**: A background worker polls Docker Stats for every instance in a scaling group.
- **Decision Engine**: Checks if average metric > target (e.g., CPU > 50%).
- **Scale Out**: Calls `InstanceService` to clone the template instance.
- **Scale In**: Terminates the oldest instance in the group.

### 11. Secrets Manager
**What it is**: Store sensitive data (API keys, passwords) securely.
**Tech Stack**: AES-GCM Encryption, Go `crypto/aes`, HKDF Key Derivation.
**Implementation**:
- **Encryption**: Secrets encrypted at rest using AES-256-GCM.
- **Per-User Key Derivation**: Master key + user ID via HKDF for user-isolated encryption.
- **Access Tracking**: `LastAccessedAt` updated on every secret read.
- **Lookup**: Get secret by ID or by name.
- **Redaction**: List operations return `[REDACTED]` values for security.
- **Event & Audit Logging**: CREATE, ACCESS, DELETE events tracked.
- **Production Safety**: Enforces encryption key requirement in production mode.

---

## 🧩 Platform Services

### 12. Identity & Auth (IAM)
**What it is**: Secure access to the platform.
**Tech Stack**: JWT (JSON Web Tokens), BCrypt, RBAC.
**Implementation**:
- **Passwords**: Hashed using `bcrypt` cost 12.
- **Tokens**: Stateless JWTs signed with HMAC-SHA256.
- **API Keys**: Alternative authentication method with tenant context.
- **Middleware**: Go middleware validates API Key or JWT on every authenticated route.

**Role-Based Access Control (RBAC)**:
- **Built-in Roles**: `admin`, `developer`, `viewer` with default permissions.
- **Custom Roles**: Create, update, delete custom roles via API.
- **Permission System**: Fine-grained permissions (e.g., `instance:read`, `volume:write`).
- **Role Binding**: Assign roles to users by ID or email.
- **Authorization**: `Authorize()` checks user permissions before operations.
- **Fallback Logic**: Default permissions apply if role not in DB.

### 13. Observability
**What it is**: Monitor system health and logs.
**Tech Stack**: Docker API, WebSockets, Prometheus, Grafana.
**Implementation**:
- **Real-time Stats**: Stream Docker container stats (CPU/Mem/Net) via WebSocket.
- **Logs**: Attach to container `stdout/stderr` streams.
- **Dashboard Service**: Aggregated system metrics and health status.
- **Prometheus Metrics**: Custom metrics for instances, databases, caches, functions.
- **Grafana Dashboards**: Pre-configured dashboards for visualization.
- **Event System**: Event recording for all resource state changes.
- **Audit Logs**: Comprehensive audit trail for compliance.

### 13. CLI (Command Line Interface)
**What it is**: Terminal tool to manage "The Cloud".
**Tech Stack**: Cobra (CLI framework), Viper (Config).
**Implementation**:
- **Structure**: Command-based (`cloud <resource> <action>`).
- **State**: Stateless client; talks to the Backend API via HTTP.

### 14. Console (Frontend)
**What it is**: Visual dashboard.
**Tech Stack**: Next.js 14, Tailwind CSS, TypeScript.
**Implementation**:
- **SSR**: Server-Side Rendering for main dashboard views.
- **Architecture**: Component-based modern React.

---

## 💾 Storage & Database
**State Management**:
- **Primary DB**: **PostgreSQL 16** holds all metadata (users, instance configs, VPCs, permissions).
- **Object Storage**: Local file system (simulating S3).

---

## 🚀 How It's Built
We use a **Clean / Hexagonal Architecture**:
1.  **Adapters (Handlers/Repositories)**: External facing code (HTTP, SQL, Docker).
2.  **Core (Services/Ports)**: Pure business logic (e.g., "Create Instance", "Hash Password").
3.  **Domain**: Shared Go structs and entities.

This ensures that we can swap out Docker for Kubernetes, or Postgres for SQLite, without changing the core business logic.
