# The Cloud System Implementation Guide

> [!IMPORTANT]
> This document serves as the single source of truth for the technical implementation of the The Cloud Compute Service.

## 1. System Architecture

We follow a strict **Clean Architecture** (Hexagonal) pattern.

```mermaid
graph TD
    User[Clients (CLI/HTTP)] --> Ports[Primary Ports (Handlers)]
    Ports --> Core[Core Domain & Logic]
    Core --> Adapters[Secondary Adapters]
    Adapters --> Infra[Infrastructure]

    subgraph "Internal Core"
        Core
    end

    subgraph "Infrastructure"
        Adapters
        Infra
    end
```

### Layer definitions
- **Primary Ports (`internal/handlers`)**: HTTP/Gin handlers that map JSON -> Domain structs.
- **Core (`internal/core`)**: Contains all business rules. No external dependencies (no docker, no sql imports).
- **Secondary Adapters (`internal/repositories`)**: Implementations of interfaces defined in Core.

---

## 2. Core Domain

### Instance Model
Located in `internal/core/domain/instance.go`.

| Field | Type | Description |
|-------|------|-------------|
| `ID` | `uuid.UUID` | Unique identifier (v4) |
| `Name` | `string` | User-defined name (must be unique) |
| `Image` | `string` | Docker image (e.g., `alpine:latest`) |
| `Status` | `InstanceStatus` | Current lifecycle state |

### Lifecycle States (`InstanceStatus`)
- **STARTING**: Instance created in DB, provisioning in Docker.
- **RUNNING**: Successfully started in Docker.
- **STOPPED**: User requested stop.
- **ERROR**: Provisioning failed.
- **DELETED**: Soft deleted.

---

## 3. Core Ports (Interfaces)

Located in `internal/core/ports/instance.go`.

### Service Interface (Business Logic)
```go
type InstanceService interface {
    LaunchInstance(ctx context.Context, name, image, ports, instanceType string, vpcID, subnetID *uuid.UUID, volumes []domain.VolumeAttachment) (*domain.Instance, error)
    StopInstance(ctx context.Context, idOrName string) error
    GetConsoleURL(ctx context.Context, idOrName string) (string, error)
}
```

### Backend Interfaces
- **ComputeBackend**: Handles instance lifecycle and volume attachment.
- **StorageBackend**: Handles raw volume creation and snapshots.
- **NetworkBackend**: Handles SDN plumbing (OVS).

---

## 4. Adapters Implementation

### Libvirt Adapter
Located in `internal/repositories/libvirt/adapter.go`.
- Manages KVM VMs via dynamic XML generation.
- Supports block device attachment via LVM backend paths.

### LVM Adapter
Located in `internal/repositories/lvm/adapter.go`.
- Uses `os/exec` to call system LVM binaries.
- Provides high-performance block storage for Libvirt instances.

### OVS Adapter
Located in `internal/repositories/ovs/adapter.go`.
- Manages Open vSwitch bridges and flow rules for VPC isolation.

---

## 5. API Reference

### Create Instance
**POST** `/instances`

**Request:**
```json
{
  "name": "web-01",
  "image": "nginx:alpine"
}
```

**Response (201 Created):**
```json
{
  "id": "550e8400-e29b-41d4-a716-446655440000",
  "name": "web-01",
  "status": "STARTING",
  "created_at": "2023-10-27T10:00:00Z"
}
```

### Stop Instance
**POST** `/instances/{id}/stop`

**Response (200 OK):**
```json
{
  "status": "STOPPED"
}
```
