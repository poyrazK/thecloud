# Detailed Implementation Plan: Firecracker MicroVM Backend

## Objective
Implement a new `ComputeBackend` adapter using Amazon Firecracker to support lightweight, high-density MicroVMs.

## Subtasks

### 1. Scout & Setup
- [ ] Analyze `ports.ComputeBackend` interface.
- [ ] Research `firecracker-go-sdk` usage.
- [ ] Add SDK dependency to `go.mod`.

### 2. Core Adapter Implementation (`internal/repositories/firecracker/`)
- [ ] Create `FirecrackerAdapter` struct.
- [ ] Implement `LaunchInstanceWithOptions`:
    - Configure Machine (VCPU, Memory).
    - Configure Kernel (`vmlinux` path).
    - Configure Drives (RootFS).
    - Configure Network (TAP device).
- [ ] Implement Lifecycle Methods:
    - `StartInstance` / `StopInstance` / `DeleteInstance`.
    - `GetInstanceStatus` / `GetInstanceLogs` / `GetInstanceStats`.
- [ ] Implement `Exec` (requires guest agent or serial console interaction).

### 3. Networking Logic
- [ ] Implement TAP device management helper.
- [ ] Integration with host bridge if applicable.

### 4. Integration
- [ ] Modify `internal/api/setup/infrastructure.go` to support `COMPUTE_BACKEND=firecracker`.
- [ ] Add configuration fields for Firecracker binary, kernel path, and rootfs templates.

### 5. Testing & Validation
- [ ] Create `adapter_test.go`.
- [ ] Mock Firecracker binary execution for unit tests.
- [ ] Verify interface compliance.

## Challenges
- **Root Privileges:** Firecracker and network setup (TAP) typically require root.
- **Host Dependencies:** Binary presence of `firecracker` and `jailer`.
- **Kernel/RootFS:** Need compatible artifacts available on the host.

## Success Criteria
- [ ] `internal/repositories/firecracker` implements the `ComputeBackend` interface.
- [ ] System boots with `COMPUTE_BACKEND=firecracker`.
- [ ] Unit tests pass with mocked SDK/Execution.
