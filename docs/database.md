# Database Guide

This document covers the Data Layer of Mini AWS.

## Schema Design

### `instances` Table

Stores the metadata of every compute resource.

```sql
CREATE TABLE instances (
    id UUID PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    image VARCHAR(255) NOT NULL,
    status VARCHAR(50) NOT NULL,
    
    -- Metadata
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    deleted_at TIMESTAMP WITH TIME ZONE, -- Soft Delete support

    CONSTRAINT unique_name UNIQUE (name)
);
```

### Planned Tables

- **`keys`**: Used for SSH key pairs to access instances.
- **`disks`**: Managing persistent volumes (EBS equivalent).
- **`networks`**: Managing VPCs/Subnets.

## Migration Strategy

We use "Up/Down" SQL migration files.
- **Tool**: `golang-migrate` (planned).
- **Location**: `migrations/` directory.

### Rules
1.  **Immutability**: Never change an existing migration file once committed. Create a new one.
2.  **Transactional**: Migrations run inside a transaction.
3.  **Idempotency**: `UP` scripts should be re-runnable or fail safely.

## Repository Pattern

We use the Repository pattern to decouple the database from the service.

```go
// Port Definition
type InstanceRepository interface {
    Create(ctx context.Context, i *domain.Instance) error
    // ...
}
```

This allows us to swap the real Postgres implementation with an In-Memory implementation for unit tests.
