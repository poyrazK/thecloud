# Compute Service Roadmap — Updated

> Audited: 2026-04-21 | Status reflects actual codebase inspection

## Status Summary

| PR | Title | Implementation Status |
|----|-------|----------------------|
| PR 1 | Image Management | ✅ COMPLETED |
| PR 2 | Instance Resize/Scale | ✅ COMPLETED |
| PR 3 | Instance Pause/Resume | ❌ NOT IMPLEMENTED |
| PR 4 | Provision Retry & Error Handling | ❌ NOT IMPLEMENTED |
| PR 5 | Instance Tags | ❌ NOT IMPLEMENTED |
| PR 6 | WebSocket Progress Events | ⚠️ PARTIALLY IMPLEMENTED |
| PR 7 | Instance Snapshots | ❌ NOT IMPLEMENTED |
| PR 8 | Instance Live Migration | ❌ NOT IMPLEMENTED |

---

## PR 1: Image Management — Import from URL

**Status: ✅ COMPLETED**

All features implemented:
- `domain.Image` with `SourceURL` field
- `ports.ImageService` interface with all 6 methods including `ImportImage`
- `ImageService` implementation with `RegisterImage`, `UploadImage`, `GetImage`, `ListImages`, `DeleteImage`, `ImportImage`
- `ImageHandler` with all 6 endpoints including `POST /api/v1/images/import`
- `ImageRepository` postgres implementation
- RBAC permissions: `PermissionImageCreate`, `PermissionImageRead`, `PermissionImageDelete`, `PermissionImageReadAll`
- Router wired at `internal/api/setup/router.go:300-304`
- Unit tests passing (`TestImageService_Unit`)
- URL validation (http/https only) to resolve CodeQL go/request-forgery
- Status transitions: PENDING → ACTIVE (or ERROR on failure)
- Format auto-detection from URL extension (.qcow2, .img, .raw, .iso)
- 30-minute HTTP timeout for large downloads
- Returns 202 Accepted on import start

### What exists
- `domain.Image` + `ImageStatus` enum (`internal/core/domain/image.go:10-40`)
- `ports.ImageService` interface with 5 methods (`internal/core/ports/image.go:12-24`)
- `ImageService` implementation (`internal/core/services/image.go:36-192`)
- `ImageHandler` with 5 endpoints (`internal/handlers/image_handler.go:1-162`)
- `ImageRepository` postgres implementation (`internal/repositories/postgres/image_repo.go`)
- RBAC permissions: `PermissionImageCreate`, `PermissionImageRead`, `PermissionImageDelete`, `PermissionImageReadAll` (`internal/core/domain/rbac.go:143-147`)
- Router wired at `internal/api/setup/router.go:300-304`
- Unit tests passing (`TestImageService_Unit`)

### What needs to be built
**`ImportImage`** — Pull an image from an external URL (e.g., cloud-images.ubuntu.com)

#### Implementation plan

**1. `ports.ImageService`** — add method signature
```go
ImportImage(ctx context.Context, name, url, description, os, version string, isPublic bool) (*domain.Image, error)
```

**2. `domain.Image`** — add `SourceURL` field to track import origin
```go
SourceURL string `json:"source_url,omitempty"`
```

**3. `internal/core/services/image.go`** — implement `ImportImage`
- Download from URL using `httputil` or stdlib `http.Get`
- Stream to `FileStore.Write` (same as `UploadImage`)
- Support formats: qcow2, img, raw, iso
- Timeout: 30 minutes for large downloads
- Status transitions: `ImageStatusPending` → `ImageStatusActive` on success; `ImageStatusError` on failure

**4. `internal/handlers/image_handler.go`** — add handler
```go
// POST /api/v1/images/import
ImportImage(c *gin.Context)
```
- Body: `{ "name": "...", "url": "...", "description": "...", "os": "...", "version": "...", "is_public": false }`
- Streams download and upload in one step (no intermediate file storage)
- Returns 202 Accepted (long-running operation)

**5. `internal/api/setup/router.go`** — wire endpoint
```go
imageGroup.POST("/import", httputil.Permission(svcs.RBAC, domain.PermissionImageCreate), handlers.Image.ImportImage)
```

**6. Tests**
- Unit: `image_unit_test.go` — add `ImportImage` subtests
- E2E: `tests/compute_e2e_test.go` if applicable

### Files
| Action | File |
|--------|------|
| Modify | `internal/core/domain/image.go` (add SourceURL field) |
| Modify | `internal/core/ports/image.go` (add ImportImage method) |
| Modify | `internal/core/services/image.go` (implement ImportImage) |
| Modify | `internal/handlers/image_handler.go` (add ImportImage handler) |
| Modify | `internal/api/setup/router.go` (wire /import endpoint) |
| Modify | `internal/core/services/image_unit_test.go` (add test cases) |

---

## PR 2: Instance Resize/Scale

**Status: ✅ COMPLETED** (PR #175 merged 2026-04-26)

### What was built

- `POST /api/v1/instances/:id/resize` — changes instance type (CPU/memory)
- `PermissionInstanceResize` RBAC permission (`internal/core/domain/rbac.go`)
- `ResizeInstance` on `ComputeBackend` and `InstanceService` interfaces
- **Docker backend**: live resize via `docker ContainerUpdate` (NanoCPUs + memory bytes)
- **Libvirt backend**: cold resize (stop → update XML → start) with full rollback on failure
- Quota enforcement: `prepareResize()` checks/increments quota *before* backend resize (fail-fast for upsize); `completeResize()` releases quota on downsize
- Memory unit: quota APIs use GB internally (`deltaMemMB/1024`), not MB
- DomainCreate rollback: if `DomainCreate(newDom)` fails, undefines new DOM, redefines original XML, restarts original domain
- ADR-025: `docs/adr/ADR-025-instance-resize.md`

### Files
| File | Change |
|------|--------|
| `internal/core/ports/instance.go` | `ResizeInstance` method added |
| `internal/core/ports/compute.go` | `ResizeInstance` added to `ComputeBackend` |
| `internal/core/domain/rbac.go` | `PermissionInstanceResize` added |
| `internal/core/services/instance.go` | Full implementation with quota enforcement |
| `internal/handlers/instance_handler.go` | `ResizeInstance` HTTP handler + 429 swagger |
| `internal/repositories/docker/adapter.go` | `ResizeInstance` via `ContainerUpdate` |
| `internal/repositories/libvirt/adapter.go` | Cold resize with XML regex patching + rollback |
| `internal/platform/resilient_compute.go` | `ResizeInstance` passthrough with circuit breaker |
| `docs/adr/ADR-025-instance-resize.md` | Architecture decision record |

### Tests
- Unit (`instance_unit_test.go`): success, downsize, quota exceeded, invalid type, DB failure, rollback
- Unit (`libvirt/adapter_test.go`): cold resize path, DomainCreate rollback
- Unit (`resilient_compute_test.go`): `TestResilientComputeResizeInstance`
- Handler (`instance_handler_test.go`): validation, error mapping
- E2E (`compute_e2e_test.go`): upsize, downsize, invalid type

---

## PR 3: Instance Pause/Resume

**Status: ❌ NOT IMPLEMENTED**

### What exists
- `PermissionInstanceUpdate` in RBAC (`internal/core/domain/rbac.go:18`)

### What needs to be built

**Goal**: Pause an instance (freeze CPU, retain memory and network) without full shutdown.

#### 1. `internal/core/domain/instance.go` — add `StatusPaused`
```go
StatusPaused InstanceStatus = "PAUSED"  // Add to existing enum
```

#### 2. `internal/core/ports/instance.go` — add methods
```go
PauseInstance(ctx context.Context, idOrName string) error
ResumeInstance(ctx context.Context, idOrName string) error
```

#### 3. `internal/core/services/instance.go` — implement both methods
**`PauseInstance`**:
- RBAC: `PermissionInstanceUpdate`
- Get instance by idOrName
- Validate status == `StatusRunning`
- Call `compute.PauseInstance(ctx, containerID)` or `compute.SuspendInstance(ctx, containerID)`
- Update status to `StatusPaused`
- Audit log

**`ResumeInstance`**:
- RBAC: `PermissionInstanceUpdate`
- Get instance by idOrName
- Validate status == `StatusPaused`
- Call `compute.ResumeInstance(ctx, containerID)` or `compute.ResumeInstance(ctx, containerID)`
- Update status to `StatusRunning`
- Audit log

#### 4. `internal/core/ports/compute.go` — add to `ComputeBackend` interface
```go
PauseInstance(ctx context.Context, id string) error
ResumeInstance(ctx context.Context, id string) error
```

#### 5. `internal/repositories/docker/adapter.go` — implement
```go
func (a *Adapter) PauseInstance(ctx context.Context, id string) error {
    return a.client.ContainerPause(ctx, id)
}
func (a *Adapter) ResumeInstance(ctx context.Context, id string) error {
    return a.client.ContainerUnpause(ctx, id)
}
```

#### 6. `internal/repositories/libvirt/adapter.go` — implement
```go
// Uses virDomainSuspend (pause) and virDomainResume
```

#### 7. `internal/handlers/instance_handler.go` — add endpoints
```go
// POST /api/v1/instances/:id/pause
PauseInstance(c *gin.Context)
// POST /api/v1/instances/:id/resume
ResumeInstance(c *gin.Context)
```

#### 8. Tests
- Unit tests for pause/resume state transitions
- Error: instance not running → cannot pause
- Error: instance not paused → cannot resume

### Files
| Action | File |
|--------|------|
| Modify | `internal/core/domain/instance.go` — add StatusPaused |
| Modify | `internal/core/ports/compute.go` — add PauseInstance/ResumeInstance |
| Modify | `internal/core/ports/instance.go` — add interface methods |
| Modify | `internal/core/services/instance.go` — implement methods |
| Modify | `internal/handlers/instance_handler.go` — add pause/resume handlers |
| Modify | `internal/repositories/docker/adapter.go` — implement docker pause/unpause |
| Modify | `internal/repositories/libvirt/adapter.go` — implement libvirt suspend/resume |
| Modify | `internal/api/setup/router.go` — wire endpoints |

---

## PR 4: Instance Provision Retry & Improved Error Handling

**Status: ❌ NOT IMPLEMENTED**

### What exists
- `Provision()` method in `InstanceService` (`internal/core/services/instance.go:310`)
- `ProvisionJob` domain type (`internal/core/domain/instance.go`)

### What needs to be built

**Goal**: Make provisioning more reliable with structured retries and better error reporting.

#### 1. `internal/core/domain/instance.go` — add fields
```go
ProvisionAttempts  int               `json:"provision_attempts"`
StatusMessage      string            `json:"status_message,omitempty"`  // Human-readable error/info
```

#### 2. `internal/core/services/instance.go` — modify `Provision()`
- Track attempt count with exponential backoff
- Retry on transient errors: Docker API timeout, network errors, storage I/O errors
- Max 3 attempts
- Sleep between retries: 2s, 4s, 8s (exponential backoff)
- On final failure: set `StatusError` with `StatusMessage` describing failure
- Log each attempt: `s.logger.Info("provision attempt N", "instance_id", ...)`
- Increment metric: `instance_provision_retries_total`

#### 3. `internal/platform/metrics.go` — add metrics
```go
InstanceProvisionRetriesTotal prometheus.Counter
```

#### 4. `internal/workers/pipeline_worker.go` — ensure retry loop
Check if provision job handling already retries or just calls `Provision()` once. May need to modify the worker loop to respect attempt count.

#### 5. Tests
- Unit: verify retry on transient error, no retry on fatal error
- Verify attempt count increments

### Files
| Action | File |
|--------|------|
| Modify | `internal/core/domain/instance.go` — add ProvisionAttempts, StatusMessage |
| Modify | `internal/core/services/instance.go` — add retry logic to Provision() |
| Modify | `internal/platform/metrics.go` — add provision retry counter |
| Modify | `internal/workers/pipeline_worker.go` — ensure retry behavior |

---

## PR 5: Instance Tags & Resource Grouping

**Status: ❌ NOT IMPLEMENTED**

Note: `Instance.Labels` (map[string]string) already exists at `domain/instance.go:81` but provides key-value labels, not a simple tag list.

### What needs to be built

**Goal**: Add simple `Tags []string` for organizing/filtering instances.

#### 1. `internal/core/domain/instance.go` — add `Tags` field
```go
Tags []string `json:"tags,omitempty"`
```

#### 2. `internal/repositories/postgres/instance_repo.go` — update `List`
Add optional tag-based filtering to the List query. Implementation options:
- Store tags as JSON array in a `tags` column (PostgreSQL JSONB)
- Or store in separate `instance_tags` join table

#### 3. `internal/core/ports/instance.go` — update interface
```go
ListInstances(ctx context.Context, tags []string) ([]*domain.Instance, error)
// Or add filtering to existing method via query params
```

#### 4. `internal/core/services/instance.go` — update `ListInstances`
Pass tags through to repo. Update `LaunchInstance` to accept tags in `LaunchParams`.

#### 5. `internal/handlers/instance_handler.go` — update `List`
```go
// GET /api/v1/instances?tag=team:backend&tag=env:prod
func (h *InstanceHandler) List(c *gin.Context) {
    tags := c.QueryArray("tag")
    // pass to service
}
```

#### 6. CLI support (`cmd/cloud/main.go`)
Add `--tag` flag to `instance list` command.

#### 7. Tests
- Tag filtering in List
- Tags on Launch

### Files
| Action | File |
|--------|------|
| Modify | `internal/core/domain/instance.go` — add Tags field |
| Modify | `internal/repositories/postgres/instance_repo.go` — add tag filtering |
| Modify | `internal/core/ports/instance.go` — update List interface |
| Modify | `internal/core/services/instance.go` — implement tag list in LaunchParams + ListInstances |
| Modify | `internal/handlers/instance_handler.go` — parse ?tag= query params |
| Modify | `cmd/cloud/main.go` — add --tag flag to instance list/create |

---

## PR 6: WebSocket Progress Events

**Status: ⚠️ PARTIALLY IMPLEMENTED**

Infrastructure partially exists: `ws_event.go` has event types, `ws/` has Hub+broadcast.

### What exists
- `WSEventInstanceCreated/Started/Stopped/Terminated` event types (`internal/core/domain/ws_event.go:15-22`)
- `Hub` struct for broadcasting (`internal/handlers/ws/hub.go`)
- `INSTANCE_LAUNCH` event recorded in `finalizeProvision` (`instance.go:432`)

### What needs to be built

**Goal**: Real-time progress (0-100%) during instance provisioning lifecycle.

#### 1. `internal/core/domain/ws_event.go` — add new event types
```go
WSEventInstanceProvisioning = "instance.provisioning"
WSEventInstanceError        = "instance.error"
// Add Progress field to WSEvent struct (0-100)
```

#### 2. `internal/core/services/instance.go` — emit progress events

In `Provision()` method (called by worker):
```go
// At start of provision
s.eventSvc.RecordEvent(ctx, "INSTANCE_PROVISIONING", inst.ID.String(), "INSTANCE",
    map[string]interface{}{"progress": 0, "message": "Starting provision..."})

// After network setup (25%)
// After volume resolution (50%)
// After container launch (75%)
// After finalize (90%)
// After finalize complete (100%)
```

Or emit via WebSocket hub directly (check if hub is accessible from service).

#### 3. `internal/handlers/ws/handler.go` — add instance-specific subscription
```go
// GET /ws/instances/:id/events
// Subscribes to events for a specific instance
```

#### 4. WebSocket event flow
```
LaunchInstance → enqueue job → return 202
Worker picks up job → calls Provision()
  → emits progress events (0%, 25%, 50%, 75%, 100%)
  → on error: WSEventInstanceError
Client connects to /ws/instances/{id}/events and receives real-time updates
```

### Files
| Action | File |
|--------|------|
| Modify | `internal/core/domain/ws_event.go` — add provisioning event type + progress field |
| Modify | `internal/core/services/instance.go` — emit events in Provision() and finalizeProvision() |
| Modify | `internal/handlers/ws/handler.go` — add instance-filtered subscription |
| Modify | `internal/api/setup/router.go` — wire /ws/instances/:id/events |

---

## PR 7: Instance Snapshots

**Status: ❌ NOT IMPLEMENTED**

Important: **Volume snapshots are fully implemented** (`SnapshotService`, `SnapshotRepository`). This PR is for **instance snapshots** (full VM/container state including memory).

### What needs to be built

**Goal**: Create, list, delete, and restore full instance snapshots (memory + disk).

#### 1. `internal/core/domain/instance_snapshot.go` — new domain model
```go
type InstanceSnapshot struct {
    ID          uuid.UUID
    Name        string
    InstanceID  uuid.UUID
    Status      SnapshotStatus  // PENDING, ACTIVE, ERROR, DELETING
    Description string
    SizeBytes   int64
    Format      string  // "docker-image", "qcow2", "raw"
    FilePath    string  // Path to stored snapshot
    CreatedAt   time.Time
}
```

#### 2. `internal/core/ports/instance_snapshot.go` — new port interface
```go
type InstanceSnapshotService interface {
    CreateSnapshot(ctx context.Context, instanceID uuid.UUID, name, description string) (*InstanceSnapshot, error)
    ListSnapshots(ctx context.Context, instanceID uuid.UUID) ([]*InstanceSnapshot, error)
    DeleteSnapshot(ctx context.Context, snapshotID uuid.UUID) error
    RestoreSnapshot(ctx context.Context, snapshotID uuid.UUID) error  // stop → restore → start
}
```

#### 3. `internal/core/services/instance_snapshot.go` — new service
**`CreateSnapshot`**:
- For **Docker**: `docker commit <container> <image>`, then `docker save` to tar/qcow2
- For **Libvirt**: `virDomainSnapshotCreateXML` with flag `VIR_DOMAIN_SNAPSHOT_CREATE_ALL`
- Store snapshot file in `thecloud-data/snapshots/`
- Create `InstanceSnapshot` record in new repository

**`RestoreSnapshot`**:
- Get snapshot and validate instance is stopped
- Stop instance if running
- For Docker: `docker load` then recreate container
- For Libvirt: revert to snapshot via `virDomainRevertToSnapshot`
- Start instance
- Emit events: `instance.restored`

#### 4. `internal/repositories/postgres/instance_snapshot_repo.go` — new repository

#### 5. `internal/handlers/instance_snapshot_handler.go` — new handler
```go
// POST /api/v1/instances/:id/snapshots
CreateSnapshot(c *gin.Context)
// GET /api/v1/instances/:id/snapshots
ListSnapshots(c *gin.Context)
// DELETE /api/v1/snapshots/:id
DeleteSnapshot(c *gin.Context)
// POST /api/v1/snapshots/:id/restore
RestoreSnapshot(c *gin.Context)
```

#### 6. RBAC permissions
Add `PermissionInstanceSnapshotCreate`, `PermissionInstanceSnapshotDelete`, `PermissionInstanceSnapshotRestore` in `domain/rbac.go`.

### Files
| Action | File |
|--------|------|
| New | `internal/core/domain/instance_snapshot.go` |
| New | `internal/core/ports/instance_snapshot.go` |
| New | `internal/core/services/instance_snapshot.go` |
| New | `internal/handlers/instance_snapshot_handler.go` |
| New | `internal/repositories/postgres/instance_snapshot_repo.go` |
| Modify | `internal/core/domain/rbac.go` — add snapshot permissions |
| Modify | `internal/api/setup/router.go` — wire snapshot endpoints |
| Modify | `internal/repositories/docker/adapter.go` — implement docker commit/save |
| Modify | `internal/repositories/libvirt/adapter.go` — implement snapshot APIs |

---

## PR 8: Instance Live Migration (Libvirt)

**Status: ❌ NOT IMPLEMENTED**

### What needs to be built

**Goal**: Live migrate a Libvirt VM from one host to another.

#### 1. `internal/core/domain/rbac.go` — add permission
```go
PermissionInstanceMigrate Permission = "instance:migrate"
```

#### 2. `internal/core/ports/instance.go` — add method
```go
MigrateInstance(ctx context.Context, idOrName, targetHost string) error
```

#### 3. `internal/core/services/instance.go` — implement `MigrateInstance`
- RBAC check: `PermissionInstanceMigrate`
- Validate instance is `StatusRunning`
- Pre-migration validation:
  - Check target host is reachable (ping or libvirt connection test)
  - Check target has sufficient vCPU/memory capacity
  - Check target host is compatible (same storage pool)
- Call `compute.MigrateInstance` on Libvirt backend

#### 4. `internal/core/ports/compute.go` — add to interface
```go
MigrateInstance(ctx context.Context, id, targetHost string) error
```

#### 5. `internal/repositories/libvirt/adapter.go` — implement
```go
func (a *Adapter) MigrateInstance(ctx context.Context, id, targetHost string) error {
    // virDomainMigrate3 with VIR_MIGRATE_LIVE flag
    // Handle VIR_MIGRATE_ABORT_ON_ERROR for reliability
    // Fallback to cold migration if live fails
}
```

#### 6. `internal/handlers/instance_handler.go` — add endpoint
```go
// POST /api/v1/instances/:id/migrate
// Body: { "target_host": "libvirt-host-2" }
MigrateInstance(c *gin.Context)
```

### Files
| Action | File |
|--------|------|
| Modify | `internal/core/domain/rbac.go` — add PermissionInstanceMigrate |
| Modify | `internal/core/ports/instance.go` — add MigrateInstance method |
| Modify | `internal/core/ports/compute.go` — add MigrateInstance to interface |
| Modify | `internal/core/services/instance.go` — implement MigrateInstance |
| Modify | `internal/handlers/instance_handler.go` — add migrate handler |
| Modify | `internal/repositories/libvirt/adapter.go` — implement virDomainMigrate3 |
| Modify | `internal/api/setup/router.go` — wire /migrate endpoint |

---

## Completed Features Not In Original Roadmap

These exist in the codebase but weren't in the original roadmap:

| Feature | File | Notes |
|---------|------|-------|
| `Instance.Labels` (map) | `internal/core/domain/instance.go:81` | Key-value labels, not simple tags |
| `Instance.Metadata` (map) | `internal/core/domain/instance.go:80` | Arbitrary key-value metadata |
| `UpdateInstanceMetadata` | `services/instance.go:1182` | Patch metadata/labels on instance |
| `Exec` | `services/instance.go:1148` | Run command inside instance |
| Volume Snapshots | `services/snapshot.go`, `repos/postgres/snapshot_repo.go` | Fully implemented volume snapshots |
| Image RBAC perms | `domain/rbac.go:143` | All 4 image permissions exist |
| WebSocket infrastructure | `handlers/ws/` | Hub, broadcast, event types exist |

---

## Implementation Order Recommendation

| Order | PR | Reason |
|-------|----|--------|
| 1 | PR 1 (Import) | Small addition, completes existing work |
| 2 | PR 5 (Tags) | Small, low risk — adds filtering |
| 3 | PR 3 (Pause/Resume) | Low effort, adds useful state |
| 4 | PR 4 (Retry) | Low effort, improves reliability |
| 5 | PR 2 (Resize) | Medium effort, high utility |
| 6 | PR 6 (WebSocket) | Medium, depends on PR 4 |
| 7 | PR 7 (Snapshots) | High effort, complex |
| 8 | PR 8 (Migration) | Very high effort, Libvirt-only |
