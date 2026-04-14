# ADR 022: Multi-Tenancy Isolation

## Status
Accepted

## Context
The platform initially used a flat user-based ownership model (`UserID`). As we scale to support organizations and teams, we need a robust way to group users and isolate resources (Compute, Storage, Networking, Secrets) within organizational boundaries.

## Decision
We decided to implement a first-class Multi-Tenancy model using a `TenantID` (UUID) across all layers of the system.

### 1. Domain Layer
Every resource that belongs to a tenant must include a `TenantID` field.
```go
type Resource struct {
    ID       uuid.UUID
    UserID   uuid.UUID // Individual owner/creator
    TenantID uuid.UUID // Organizational boundary
    // ...
}
```

### 2. Context Propagation
The `appcontext` package was expanded to carry `TenantID` in the `context.Context`, similar to `UserID`. Middleware extracts this from JWT claims or API keys.

### 3. Repository Layer
Repositories MUST filter all read/write operations by `TenantID`.
- `GetByID`, `GetByName`, and `List` queries include `tenant_id = $1` in the `WHERE` clause.
- `Update` and `Delete` operations include `tenant_id = $1` to prevent cross-tenant data corruption or deletion.
- Global resources (e.g., public images) are handled by allowing `tenant_id IS NULL` or using a reserved system tenant ID.

### 4. Service Layer
Services enforce tenant boundaries by ensuring the `TenantID` in the resource matches the `TenantID` in the context before performing sensitive operations.

## Consequences
- **Isolation**: Resources are strictly isolated between tenants at the database level.
- **Security**: Prevents "horizontal" attacks where a user might try to access another tenant's resource by guessing a UUID.
- **Complexity**: Slightly increased boilerplate in repositories and services to handle the extra filtering.
- **Migration**: Existing resources had to be migrated to a default tenant during the refactor.
