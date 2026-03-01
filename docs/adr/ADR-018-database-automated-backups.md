# ADR 018: Managed Database Backups & Snapshots

## Status
Accepted

## Context
The Managed Database Service (RDS) provides persistent storage via block volumes, but lacked a native backup and recovery mechanism. Users needed a way to protect against data loss caused by accidental deletions or volume corruption.

We needed a solution that was:
1.  **Fast and efficient** (incremental or point-in-time).
2.  **Engine-agnostic** (works for both Postgres and MySQL).
3.  **Integrated** with the platform's existing storage management.

## Decision
We implemented automated backups by integrating the `DatabaseService` with the core `SnapshotService`.

### 1. Volume-Level Snapshots
Instead of logical dumps (e.g., `pg_dump`), we utilize block-level snapshots. This ensures crash-consistent backups that are extremely fast to create and restore, regardless of the database size.

### 2. Implementation in DatabaseService
The `DatabaseService` was extended with:
-   `CreateDatabaseSnapshot`: Resolves the database's underlying volume and triggers a snapshot.
-   `ListDatabaseSnapshots`: Retrieves all snapshots associated with a database's volume.
-   `RestoreDatabase`: Provisions a completely new database instance using a volume restored from a snapshot.

### 3. Dynamic Volume Binding
The container launch logic was refactored to support binding existing volumes (restored from snapshots) instead of always creating fresh ones.

### 4. Restore Workflow
Restore is a "provision-from-backup" workflow rather than an "in-place overwrite". This follows immutable infrastructure principles, allowing users to verify the restored data on a new instance before decommissioning the old one.

## Consequences

### Positive
-   **High Performance**: Snapshots are near-instant and handled at the storage layer.
-   **Lower Resource Usage**: No heavy query load on the database engine during backup creation.
-   **Stateless Restore**: Restoring to a new instance simplifies recovery and prevents data corruption on the original instance.

### Negative
-   **Crash Consistency Only**: While storage-level snapshots are highly reliable, they do not guarantee application-level consistency (e.g., in-flight transactions) unless the engine is explicitly flushed/quiesced (future enhancement).
-   **Storage Costs**: Each snapshot consumes additional block storage space.
