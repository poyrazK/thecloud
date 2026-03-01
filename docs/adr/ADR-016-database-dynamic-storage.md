# ADR 016: Managed Database Dynamic Storage Sizing

## Status
Accepted

## Context
Initially, the Managed Database Service (RDS) hardcoded a 10GB persistent volume for every database instance. While this ensured data persistence, it lacked the flexibility required for production workloads where data size varies significantly. Users needed the ability to specify the amount of storage required for their databases at creation time.

## Decision
We implemented dynamic storage sizing for the Managed Database Service.

### 1. API Changes
The `POST /databases` endpoint was updated to accept an optional `allocated_storage` integer field.
- **Validation**: If the value is omitted or less than 10, it defaults to 10GB.
- **Upper Bound**: Currently, there is no explicit upper bound, but it is limited by the underlying storage backend's capacity.

### 2. Service Integration
The `DatabaseService` now accepts `allocatedStorage` in its `CreateDatabase` method and passes this value to the `VolumeService.CreateVolume` call.

### 3. Replication Inheritance
When creating a read replica via `CreateReplica`, the service automatically inherits the `allocated_storage` size from the primary database. This ensures that replicas always have sufficient capacity to hold the replicated data.

### 4. Persistence
The `allocated_storage` value is stored in the `databases` table in PostgreSQL to maintain an accurate record of the provisioned resources.

## Consequences

### Positive
- **Flexibility**: Users can now provision databases with storage tailored to their specific needs.
- **Consistency**: Replicas are automatically provisioned with matching storage sizes, reducing manual configuration errors.
- **Resource Tracking**: The platform now tracks provisioned storage per database instance, improving resource management.

### Negative
- **Schema Complexity**: Added a new column to the `databases` table.
- **Fixed Size After Creation**: The current implementation only supports sizing at creation time. Scaling an existing volume (Volume Expansion) is out of scope for this decision and will be addressed in a future ADR.
