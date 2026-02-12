# Database Guide

This document covers the data persistence layer of The Cloud platform.

## Overview

**Database**: PostgreSQL 15+  
**Driver**: `pgx/v5` (native Go driver)  
**Connection Pooling**: `pgxpool` with configurable pool size  
**Migrations**: Goose (embedded in binary)  
**Test Coverage**: 57.5% (integration tests)

## Architecture

### Repository Pattern

All database access goes through repository interfaces:

```go
// Port (interface) - defined in internal/core/ports
type InstanceRepository interface {
    Create(ctx context.Context, inst *domain.Instance) error
    GetByID(ctx context.Context, id uuid.UUID) (*domain.Instance, error)
    Update(ctx context.Context, inst *domain.Instance) error
    Delete(ctx context.Context, id uuid.UUID) error
    List(ctx context.Context) ([]*domain.Instance, error)
}

// Adapter (implementation) - in internal/repositories/postgres
type instanceRepository struct {
    db *pgxpool.Pool
}
```

**Benefits**:
- Testable: Easy to mock for unit tests
- Flexible: Can swap PostgreSQL for another database
- Clean: Business logic doesn't know about SQL

### Connection Management

**Connection Pool Configuration**:
```go
config, _ := pgxpool.ParseConfig(databaseURL)
config.MaxConns = 25
config.MinConns = 5
config.MaxConnLifetime = time.Hour
config.MaxConnIdleTime = 30 * time.Minute

pool, _ := pgxpool.NewWithConfig(context.Background(), config)
```

**Connection String**:
```bash
# Development
DATABASE_URL=postgres://cloud:cloud@localhost:5433/thecloud

# Production
DATABASE_URL=postgres://user:pass@db-host:5432/thecloud?sslmode=require
```

**Note**: Port `5433` is used locally to avoid conflicts with system PostgreSQL.

## Schema Design

### Core Tables

#### `users` - User Accounts
```sql
CREATE TABLE users (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    email VARCHAR(255) NOT NULL UNIQUE,
    password_hash VARCHAR(255) NOT NULL,
    name VARCHAR(255) NOT NULL,
    role VARCHAR(50) NOT NULL DEFAULT 'user',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_users_email ON users(email);
```

**Roles**: `admin`, `developer`, `viewer`, `user`

#### `api_keys` - Authentication
```sql
CREATE TABLE api_keys (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    key VARCHAR(64) NOT NULL UNIQUE,
    name VARCHAR(255),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    last_used TIMESTAMPTZ
);

CREATE INDEX idx_api_keys_key ON api_keys(key);
CREATE INDEX idx_api_keys_user_id ON api_keys(user_id);
```

#### `roles` - RBAC Roles
```sql
CREATE TABLE roles (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(100) NOT NULL UNIQUE,
    description TEXT
);

CREATE TABLE role_permissions (
    role_id UUID NOT NULL REFERENCES roles(id) ON DELETE CASCADE,
    permission VARCHAR(100) NOT NULL,
    PRIMARY KEY (role_id, permission)
);
```

**Permissions**: `instance:read`, `instance:launch`, `volume:create`, `full_access`, etc.

### Compute Resources

#### `instances` - Compute Instances
```sql
CREATE TABLE instances (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    name VARCHAR(255) NOT NULL,
    image VARCHAR(255) NOT NULL,
    status VARCHAR(50) NOT NULL,
    container_id VARCHAR(255),
    ports TEXT,
    vpc_id UUID REFERENCES vpcs(id) ON DELETE SET NULL,
    instance_type TEXT NOT NULL DEFAULT 'basic-2',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    version INT NOT NULL DEFAULT 1
);

CREATE INDEX idx_instances_user_id ON instances(user_id);
CREATE INDEX idx_instances_status ON instances(status);

#### `instance_types` - Predefined Instance Specs
```sql
CREATE TABLE instance_types (
    id TEXT PRIMARY KEY,
    name TEXT NOT NULL,
    vcpus INT NOT NULL,
    memory_mb INT NOT NULL,
    disk_gb INT NOT NULL,
    network_mbps INT NOT NULL DEFAULT 1000,
    price_per_hour NUMERIC(10,4) NOT NULL DEFAULT 0.01,
    category TEXT NOT NULL DEFAULT 'general',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
```

**Seeded Types**: `basic-1`, `basic-2`, `standard-1`, `performance-2`, etc.
```

**Optimistic Locking**: `version` column prevents concurrent update conflicts

#### `volumes` - Block Storage
```sql
CREATE TABLE volumes (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    name VARCHAR(255) NOT NULL,
    size_gb INT NOT NULL,
    status VARCHAR(50) NOT NULL,
    attached_to UUID REFERENCES instances(id) ON DELETE SET NULL,
    mount_path VARCHAR(512),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_volumes_user_id ON volumes(user_id);
CREATE INDEX idx_volumes_attached_to ON volumes(attached_to);
```

#### `snapshots` - Volume Snapshots
```sql
CREATE TABLE snapshots (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    volume_id UUID NOT NULL REFERENCES volumes(id) ON DELETE CASCADE,
    name VARCHAR(255) NOT NULL,
    size_gb INT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
```

### Networking

#### `vpcs` - Virtual Private Clouds
```sql
CREATE TABLE vpcs (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    name VARCHAR(255) NOT NULL,
    cidr_block VARCHAR(18) NOT NULL,
    network_id VARCHAR(255) NOT NULL,
    gateway_id VARCHAR(255),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_vpcs_user_id ON vpcs(user_id);
```

#### `elastic_ips` - Static/Elastic IPs
```sql
CREATE TABLE elastic_ips (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    instance_id UUID REFERENCES instances(id) ON DELETE SET NULL,
    vpc_id UUID REFERENCES vpcs(id) ON DELETE SET NULL,
    public_ip VARCHAR(45) NOT NULL UNIQUE,
    status VARCHAR(50) NOT NULL DEFAULT 'allocated',
    arn VARCHAR(255) NOT NULL UNIQUE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE UNIQUE INDEX idx_elastic_ips_instance_id_unique 
ON elastic_ips(instance_id) 
WHERE instance_id IS NOT NULL;
CREATE INDEX idx_elastic_ips_tenant_id ON elastic_ips(tenant_id);
```
#### `load_balancers` - Load Balancers
```sql
CREATE TABLE load_balancers (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    vpc_id UUID REFERENCES vpcs(id) ON DELETE SET NULL,
    name VARCHAR(255) NOT NULL,
    type VARCHAR(50) DEFAULT 'HTTP',
    status VARCHAR(50) NOT NULL,
    listener_port INT NOT NULL,
    container_id VARCHAR(255),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE lb_targets (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    lb_id UUID NOT NULL REFERENCES load_balancers(id) ON DELETE CASCADE,
    instance_id UUID NOT NULL REFERENCES instances(id) ON DELETE CASCADE,
    port INT NOT NULL,
    health_status VARCHAR(50) DEFAULT 'healthy',
    last_health_check TIMESTAMPTZ
);
```

### Auto-Scaling

#### `scaling_groups` - Auto-Scaling Groups
```sql
CREATE TABLE scaling_groups (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    name VARCHAR(255) NOT NULL,
    vpc_id UUID REFERENCES vpcs(id) ON DELETE SET NULL,
    min_instances INT NOT NULL,
    max_instances INT NOT NULL,
    desired_count INT NOT NULL,
    current_count INT NOT NULL DEFAULT 0,
    status VARCHAR(50) DEFAULT 'ACTIVE',
    image VARCHAR(255) NOT NULL,
    failure_count INT DEFAULT 0,
    last_failure_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE scaling_policies (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    group_id UUID NOT NULL REFERENCES scaling_groups(id) ON DELETE CASCADE,
    name VARCHAR(255) NOT NULL,
    metric_name VARCHAR(100) NOT NULL,
    threshold DOUBLE PRECISION NOT NULL,
    comparison_operator VARCHAR(20) NOT NULL,
    adjustment_type VARCHAR(50) NOT NULL,
    adjustment_value INT NOT NULL,
    cooldown_seconds INT DEFAULT 300
);

CREATE TABLE scaling_group_instances (
    group_id UUID NOT NULL REFERENCES scaling_groups(id) ON DELETE CASCADE,
    instance_id UUID NOT NULL REFERENCES instances(id) ON DELETE CASCADE,
    PRIMARY KEY (group_id, instance_id)
);
```

### Managed Services

#### `databases` - RDS (Managed Databases)
```sql
CREATE TABLE databases (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    vpc_id UUID REFERENCES vpcs(id) ON DELETE SET NULL,
    name VARCHAR(255) NOT NULL,
    engine VARCHAR(50) NOT NULL,
    version VARCHAR(50) NOT NULL,
    status VARCHAR(50) NOT NULL,
    host VARCHAR(255),
    port INT,
    username VARCHAR(255),
    password TEXT,
    container_id VARCHAR(255),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
```

**Engines**: `postgres`, `mysql`

#### `caches` - Redis Instances
```sql
CREATE TABLE caches (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    vpc_id UUID REFERENCES vpcs(id) ON DELETE SET NULL,
    name VARCHAR(255) NOT NULL,
    status VARCHAR(50) NOT NULL,
    host VARCHAR(255),
    port INT,
    password TEXT,
    container_id VARCHAR(255),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
```

#### `functions` - Serverless Functions
```sql
CREATE TABLE functions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    name VARCHAR(255) NOT NULL,
    runtime VARCHAR(50) NOT NULL,
    handler VARCHAR(255) NOT NULL,
    code_path VARCHAR(512) NOT NULL,
    timeout_seconds INT DEFAULT 30,
    memory_mb INT DEFAULT 256,
    status VARCHAR(50) NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE invocations (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    function_id UUID NOT NULL REFERENCES functions(id) ON DELETE CASCADE,
    status VARCHAR(50) NOT NULL,
    started_at TIMESTAMPTZ NOT NULL,
    ended_at TIMESTAMPTZ,
    duration_ms BIGINT,
    status_code INT,
    logs TEXT
);

CREATE INDEX idx_invocations_function_id ON invocations(function_id);
```

### Messaging & Events

#### `queues` - Message Queues
```sql
CREATE TABLE queues (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    name VARCHAR(255) NOT NULL,
    arn VARCHAR(512) NOT NULL,
    visibility_timeout INT DEFAULT 30,
    retention_days INT DEFAULT 4,
    max_message_size INT DEFAULT 262144,
    status VARCHAR(50) DEFAULT 'ACTIVE',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE queue_messages (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    queue_id UUID NOT NULL REFERENCES queues(id) ON DELETE CASCADE,
    body TEXT NOT NULL,
    receipt_handle VARCHAR(255) UNIQUE,
    visible_after TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_queue_messages_queue_id ON queue_messages(queue_id);
CREATE INDEX idx_queue_messages_visible_after ON queue_messages(visible_after);
```

#### `topics` - Pub/Sub Topics
```sql
CREATE TABLE topics (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    name VARCHAR(255) NOT NULL,
    arn VARCHAR(512) NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE subscriptions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    topic_id UUID NOT NULL REFERENCES topics(id) ON DELETE CASCADE,
    protocol VARCHAR(50) NOT NULL,
    endpoint TEXT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
```

**Protocols**: `queue`, `webhook`

#### `cron_jobs` - Scheduled Tasks
```sql
CREATE TABLE cron_jobs (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    name VARCHAR(255) NOT NULL,
    schedule VARCHAR(100) NOT NULL,
    function_id UUID REFERENCES functions(id) ON DELETE SET NULL,
    http_endpoint TEXT,
    status VARCHAR(50) DEFAULT 'ACTIVE',
    last_run_at TIMESTAMPTZ,
    next_run_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE cron_job_runs (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    job_id UUID NOT NULL REFERENCES cron_jobs(id) ON DELETE CASCADE,
    status VARCHAR(50) NOT NULL,
    started_at TIMESTAMPTZ NOT NULL,
    ended_at TIMESTAMPTZ,
    output TEXT,
    error TEXT
);
```

### Security & Secrets

#### `secrets` - Encrypted Secrets
```sql
CREATE TABLE secrets (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    name VARCHAR(255) NOT NULL,
    encrypted_value TEXT NOT NULL,
    description TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    last_accessed_at TIMESTAMPTZ,
    CONSTRAINT secrets_name_user_key UNIQUE (name, user_id)
);
```

### Storage

#### `objects` - Object Storage Metadata
```sql
CREATE TABLE objects (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    arn VARCHAR(512) NOT NULL UNIQUE,
    bucket VARCHAR(255) NOT NULL,
    key VARCHAR(512) NOT NULL,
    size_bytes BIGINT NOT NULL,
    content_type VARCHAR(255),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ,
    UNIQUE (bucket, key)
);

CREATE INDEX idx_objects_bucket ON objects(bucket);
CREATE INDEX idx_objects_user_id ON objects(user_id);
```

**Note**: Actual file bytes stored on filesystem, not in database.

### Observability

#### `audit_logs` - Audit Trail
```sql
CREATE TABLE audit_logs (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    action VARCHAR(100) NOT NULL,
    resource_type VARCHAR(100) NOT NULL,
    resource_id VARCHAR(255),
    metadata JSONB,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_audit_logs_user_id ON audit_logs(user_id);
CREATE INDEX idx_audit_logs_created_at ON audit_logs(created_at);
```

#### `events` - System Events
```sql
CREATE TABLE events (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    event_type VARCHAR(100) NOT NULL,
    resource_id VARCHAR(255) NOT NULL,
    resource_type VARCHAR(100) NOT NULL,
    metadata JSONB,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_events_resource_id ON events(resource_id);
CREATE INDEX idx_events_created_at ON events(created_at);
```

## Migrations

### Migration System

**Tool**: Goose  
**Location**: `internal/repositories/postgres/migrations/`  
**Format**: SQL files with up/down migrations

**Migration File Naming**:
```
001_create_users.sql
002_create_instances.sql
003_add_rbac_tables.sql
```

**Migration Structure**:
```sql
-- +goose Up
CREATE TABLE my_table (
    id UUID PRIMARY KEY,
    name VARCHAR(255) NOT NULL
);

-- +goose Down
DROP TABLE my_table;
```

### Running Migrations

**Automatic** (on API startup):
```bash
go run cmd/api/main.go
```

**Manual** (migrate only, then exit):
```bash
go run cmd/api/main.go -migrate-only
```

**Rollback** (down migration):
```bash
goose -dir internal/repositories/postgres/migrations postgres "connection-string" down
```

### Creating New Migrations

1. Create new file: `XXX_description.sql`
2. Add `-- +goose Up` section
3. Add `-- +goose Down` section
4. Test both up and down migrations

**Example**:
```sql
-- +goose Up
CREATE TABLE new_feature (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id),
    data TEXT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_new_feature_user_id ON new_feature(user_id);

-- +goose Down
DROP TABLE new_feature;
```

## Testing

### Integration Tests (57.5% Coverage)

All repository tests use real PostgreSQL:

```go
//go:build integration

func TestInstanceRepository_Create(t *testing.T) {
    db := setupDB(t)
    defer db.Close()
    repo := NewInstanceRepository(db)
    ctx := setupTestUser(t, db)
    
    // Cleanup
    _, _ = db.Exec(ctx, "DELETE FROM instances")
    
    inst := &domain.Instance{
        ID:     uuid.New(),
        UserID: userID,
        Name:   "test-instance",
    }
    
    err := repo.Create(ctx, inst)
    require.NoError(t, err)
}
```

**Test Database Setup**:
```go
func setupDB(t *testing.T) *pgxpool.Pool {
    dbURL := os.Getenv("DATABASE_URL")
    if dbURL == "" {
        dbURL = "postgres://cloud:cloud@localhost:5433/thecloud"
    }
    
    db, err := pgxpool.New(context.Background(), dbURL)
    require.NoError(t, err)
    
    err = db.Ping(context.Background())
    if err != nil {
        t.Skip("Skipping integration test: database not available")
    }
    
    return db
}
```

### Running Integration Tests

```bash
# Start PostgreSQL
docker compose up -d postgres

# Run integration tests
go test -tags=integration ./internal/repositories/postgres/...

# With coverage
go test -tags=integration -coverprofile=coverage.out ./internal/repositories/postgres/...
```

## Performance Optimization

### Indexes

All foreign keys have indexes:
```sql
CREATE INDEX idx_instances_user_id ON instances(user_id);
CREATE INDEX idx_volumes_attached_to ON volumes(attached_to);
CREATE INDEX idx_audit_logs_created_at ON audit_logs(created_at);
```

### Connection Pooling

```go
config.MaxConns = 25        // Maximum connections
config.MinConns = 5         // Minimum idle connections
config.MaxConnLifetime = 1h // Recycle connections
config.MaxConnIdleTime = 30m
```

### Query Optimization

- Use `EXPLAIN ANALYZE` for slow queries
- Add indexes on frequently queried columns
- Use `LIMIT` for pagination
- Avoid `SELECT *`, specify columns

### Prepared Statements

pgx automatically uses prepared statements for repeated queries.

## Security

### SQL Injection Prevention

**Always use parameterized queries**:
```go
// ✅ Good
query := "SELECT * FROM instances WHERE id = $1"
row := db.QueryRow(ctx, query, instanceID)

// ❌ Bad
query := fmt.Sprintf("SELECT * FROM instances WHERE id = '%s'", instanceID)
```

### Password Storage

- User passwords: bcrypt hashed
- Database passwords: Encrypted (future: use Vault)
- Secrets: AES-256 encrypted

### Row-Level Security

User ID filtering in all queries:
```go
func (r *instanceRepository) List(ctx context.Context) ([]*domain.Instance, error) {
    userID := appcontext.UserIDFromContext(ctx)
    query := `SELECT * FROM instances WHERE user_id = $1`
    rows, err := r.db.Query(ctx, query, userID)
    // ...
}
```

## Backup & Recovery

### Backup Strategy

**Development**:
```bash
pg_dump -h localhost -p 5433 -U cloud thecloud > backup.sql
```

**Production**:
- Automated daily backups
- Point-in-time recovery (PITR)
- Backup retention: 30 days
- Cross-region replication

### Restore

```bash
psql -h localhost -p 5433 -U cloud thecloud < backup.sql
```

## Monitoring

### Key Metrics

- Connection pool utilization
- Query execution time
- Slow query log
- Table sizes
- Index usage

### Prometheus Metrics

```go
platform.DBConnectionsTotal.Set(float64(pool.Stat().TotalConns()))
platform.DBQueryDuration.Observe(duration.Seconds())
```

## Best Practices

1. **Always use context**: Pass `context.Context` to all queries
2. **Use transactions**: For multi-step operations
3. **Handle errors**: Check and wrap all database errors
4. **Close resources**: Defer `rows.Close()` and `tx.Rollback()`
5. **Use connection pooling**: Don't create new connections per request
6. **Add indexes**: On foreign keys and frequently queried columns
7. **Test with real DB**: Use integration tests
8. **Version your schema**: Use migrations, never manual changes

## Further Reading

- [Backend Guide](backend.md) - Repository pattern implementation
- [Architecture Guide](architecture.md) - Hexagonal architecture
- [Testing Guide](backend.md#testing) - Integration test patterns
