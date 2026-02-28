# ADR 015: Managed Database Data Persistence

## Status
Accepted

## Context
The Managed Database Service (RDS) initially launched with ephemeral storage. When a database container was restarted or recreated, all data was lost. This made the service unsuitable for production workloads that require high durability and data integrity across container lifecycles.

We needed a solution that:
1.  **Ensures Data Durability**: Data must survive container crashes, restarts, and manual deletions/recreations.
2.  **Integrates with Existing Infrastructure**: Leverage the existing `VolumeService` and block storage backends.
3.  **Supports Replication**: Ensure both primary and replica instances have persistent storage.
4.  **Automates Lifecycle**: Volumes should be provisioned automatically during DB creation and cleaned up during DB deletion to prevent storage leaks.

## Decision
We implemented automated persistent volume integration for the Managed Database Service.

### 1. Automated Provisioning
The `DatabaseService` was updated to interact with the `VolumeService`:
-   **On Create**: Every new database (Primary or Replica) automatically provisions a 10GB persistent block volume.
-   **Naming Convention**: Volumes are named with a predictable prefix `db-vol-<db_id_prefix>` or `db-replica-vol-<db_id_prefix>` for easier identification.

### 2. Volume Binding
The provisioned block volume is mounted into the database container at the engine-specific data directory:
-   **PostgreSQL**: `/var/lib/postgresql/data`
-   **MySQL**: `/var/lib/mysql`

This ensures that the database engine's internal state is stored entirely on the persistent block device.

### 3. Lifecycle Management
-   **Cleanup on Delete**: When a `domain.Database` is deleted via the API, the service now performs an automated lookup for associated volumes and triggers their deletion.
-   **Error Handling**: If volume creation fails, the database provisioning is rolled back. If container launch fails, the provisioned volume is automatically deleted to maintain system hygiene.

### 4. Integration & Testing
-   **Dependency Injection**: `VolumeService` was injected into the `DatabaseService`.
-   **E2E Validation**: A new E2E test suite (`tests/database_persistence_e2e_test.go`) was added to verify the end-to-end lifecycle of persistent databases.

## Consequences

### Positive
-   **Production Readiness**: Managed databases can now safely host critical data as they are no longer ephemeral.
-   **Resource Efficiency**: Automated cleanup ensures that storage resources are released when databases are decommissioned.
-   **Architectural Alignment**: Reuses the core platform's storage abstractions rather than implementing engine-specific hacks.

### Negative
-   **Provisioning Latency**: Database creation now takes slightly longer as it involves a synchronous block volume provisioning step.
-   **Storage Costs**: Every database now consumes at least 10GB of block storage, increasing the infrastructure footprint per instance.
-   **Fixed Sizing**: The initial implementation uses a hardcoded 10GB size; dynamic sizing and volume expansion are deferred to a future version.
