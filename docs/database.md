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

#### `tenants` - Organizations/Teams
```sql
CREATE TABLE tenants (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(255) NOT NULL,
    slug VARCHAR(255) NOT NULL UNIQUE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE tenant_members (
    tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    role VARCHAR(50) NOT NULL DEFAULT 'member',
    PRIMARY KEY (tenant_id, user_id)
);

CREATE TABLE tenant_quotas (
    tenant_id UUID PRIMARY KEY REFERENCES tenants(id) ON DELETE CASCADE,
    max_instances INT NOT NULL DEFAULT 10,
    max_vcpus INT NOT NULL DEFAULT 20,
    max_memory_mb INT NOT NULL DEFAULT 40960,
    max_storage_gb INT NOT NULL DEFAULT 500,
    used_instances INT NOT NULL DEFAULT 0,
    -- ...
);
```

#### `users` - User Accounts
```sql
CREATE TABLE users (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    email VARCHAR(255) NOT NULL UNIQUE,
    password_hash VARCHAR(255) NOT NULL,
    name VARCHAR(255) NOT NULL,
    role VARCHAR(50) NOT NULL DEFAULT 'user',
    default_tenant_id UUID REFERENCES tenants(id) ON DELETE SET NULL,
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
    tenant_id UUID REFERENCES tenants(id) ON DELETE SET NULL,
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
    tenant_id UUID REFERENCES tenants(id) ON DELETE SET NULL,
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
    tenant_id UUID REFERENCES tenants(id) ON DELETE SET NULL,
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

#### `clusters` - Kubernetes Clusters
```sql
CREATE TABLE clusters (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    vpc_id UUID REFERENCES vpcs(id) ON DELETE SET NULL,
    name VARCHAR(255) NOT NULL,
    version VARCHAR(50) NOT NULL,
    status VARCHAR(50) NOT NULL,
    worker_count INT NOT NULL DEFAULT 2,
    ha_enabled BOOLEAN DEFAULT FALSE,
    network_isolation BOOLEAN DEFAULT FALSE,
    pod_cidr VARCHAR(18),
    service_cidr VARCHAR(18),
    api_server_lb_address TEXT,
    kubeconfig_encrypted TEXT,
    ssh_private_key_encrypted TEXT,
    join_token TEXT,
    token_expires_at TIMESTAMPTZ,
    ca_cert_hash TEXT,
    job_id TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE cluster_node_groups (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    cluster_id UUID NOT NULL REFERENCES clusters(id) ON DELETE CASCADE,
    name VARCHAR(255) NOT NULL,
    instance_type VARCHAR(50) NOT NULL,
    min_size INTEGER NOT NULL DEFAULT 1,
    max_size INTEGER NOT NULL DEFAULT 10,
    current_size INTEGER NOT NULL DEFAULT 1,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE(cluster_id, name)
);

CREATE INDEX idx_cluster_node_groups_cluster_id ON cluster_node_groups(cluster_id);
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
    role VARCHAR(20) NOT NULL DEFAULT 'PRIMARY',
    primary_id UUID REFERENCES databases(id) ON DELETE CASCADE,
    host VARCHAR(255),
    port INT,
    username VARCHAR(255),
    password TEXT,
    credential_path TEXT,
    container_id VARCHAR(255),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    allocated_storage INT NOT NULL DEFAULT 10,
    parameters JSONB DEFAULT '{}'::jsonb,
    metrics_enabled BOOLEAN NOT NULL DEFAULT FALSE,
    metrics_port INT,
    exporter_container_id VARCHAR(255),
    pooling_enabled BOOLEAN NOT NULL DEFAULT FALSE,
    pooling_port INT,
    pooler_container_id VARCHAR(255)
);
```

**Engines**: `postgres`, `mysql`  
**Roles**: `PRIMARY`, `REPLICA`

### Managed Database Security & Credential Rotation

The platform integrates with **HashiCorp Vault** to securely manage database credentials.

#### Vault Integration
The platform is Vault-backed with database fallback. Credentials are primarily stored in Vault at `secret/data/thecloud/rds/:db_id/credentials`, while the `password` field in the `databases` table remains persisted for legacy support and fallback during Vault unavailability.
- **Metadata**: The `credential_path` field in the `databases` table stores the reference to the Vault secret.

#### Credential Rotation
Users can trigger automated password rotation for their database instances.
- **Endpoint**: `POST /databases/:id/rotate-credentials`
- **Workflow**:
    1.  **Generate Password**: A new 16-character secure password is generated.
    2.  **Engine Update**: The `ALTER USER` command is executed inside the database container first to apply the new password.
    3.  **Update Vault**: The new secret is written to Vault. If Vault store fails after the DB password has been changed, the system automatically rolls back the database password to the original value to maintain consistency.
    4.  **Sidecar Update**: Sidecars such as PgBouncer/pooler are automatically recreated to apply the new credentials only when they are present.
    5.  **Audit**: The rotation event is recorded in the system events and audit logs.

This mechanism ensures that database access remains secure and meets compliance requirements for periodic credential updates.

### Managed Database Encryption at Rest

The platform supports **encryption at rest** for managed database volumes using HashiCorp Vault Transit Secrets Engine for key management.

#### Overview
When a `KmsKeyID` is provided during database creation, the platform:
1. Generates a unique 256-bit DEK (Data Encryption Key) for the volume
2. Encrypts the DEK with Vault Transit using the specified key
3. Stores the encrypted DEK in the `volume_encryption_keys` table
4. Uses the DEK for transparent volume encryption/decryption

#### Schema
```sql
CREATE TABLE volume_encryption_keys (
    volume_id      UUID PRIMARY KEY REFERENCES volumes(id) ON DELETE CASCADE,
    encrypted_dek  BYTEA NOT NULL,
    kms_key_id     VARCHAR(500) NOT NULL,
    algorithm      VARCHAR(50) NOT NULL DEFAULT 'AES-256-GCM',
    created_at     TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
```

#### API Usage
```bash
# Create encrypted database
POST /databases
{
  "name": "secure-db",
  "engine": "postgres",
  "version": "15",
  "kms_key_id": "vault:transit/my-master-key"
}

# Response includes encryption status
{
  "id": "...",
  "name": "secure-db",
  "encrypted_volume": true,
  "kms_key_id": "vault:transit/my-master-key"
}
```

#### Architecture
- **DEK Pattern**: Each volume has a unique DEK encrypted by Vault Transit (master key)
- **Application-Level**: Encryption is transparent to the database engine
- **Vault Transit**: Handles all cryptographic operations; DEKs are never persisted unencrypted

#### Implementation
- **Service**: `VolumeEncryptionService` in `internal/core/services/volume_encryption.go`
- **Repository**: `VolumeEncryptionRepository` in `internal/repositories/postgres/volume_encryption_repo.go`
- **Adapter**: `TransitKMSAdapter` in `internal/adapters/vault/transit_kms_adapter.go`
- **Interface**: `KMSClient` in `internal/core/ports/kms_client.go`

See [ADR-024](./adr/ADR-024-database-encryption-at-rest.md) for full architecture details.

### Managed Database Persistence

Managed databases in The Cloud platform utilize persistent block storage to ensure data durability across container lifecycles.

#### Storage Architecture
When a managed database is provisioned, the service automatically:
1.  **Creates a Block Volume**: A persistent volume is provisioned via the `VolumeService`. The size is determined by the `allocated_storage` parameter (default 10GB).
2.  **Mounts the Volume**: The volume is attached to the compute instance and mounted to the appropriate data directory:
    -   **PostgreSQL**: `/var/lib/postgresql/data`
    -   **MySQL**: `/var/lib/mysql`

This integration ensures that all database state (tables, indexes, logs) is stored on durable block storage rather than the container's ephemeral layer.

#### Volume Lifecycle
-   **Automated Provisioning**: Volumes are created synchronously during the `CreateDatabase` and `CreateReplica` calls. Replicas inherit the `allocated_storage` size of their primary.
-   **Automated Cleanup**: When a database is deleted via the API, the service identifies and deletes the associated block volumes to prevent storage leaks.

### Managed Database Backups & Snapshots

The platform provides a native backup and recovery mechanism leveraging volume snapshots.

#### Backup Creation (Snapshots)
Users can create manual point-in-time backups of their databases. The system takes a crash-consistent snapshot of the underlying block volume, ensuring all persisted data is captured.
- **Endpoint**: `POST /databases/:id/snapshots`
- **Mechanism**: Integrated with core `SnapshotService`.

#### Data Recovery (Restore)
Databases can be restored from any valid snapshot. The restore process provisions a **completely new database instance** using a volume initialized from the snapshot data. This allows for safe verification of restored data without affecting the source database.
- **Endpoint**: `POST /databases/restore`
- **Flexibility**: Users can specify new names, VPCs, and configurations during the restore process.

### Managed Database Configuration (Parameter Groups)

The platform supports dynamic engine configuration via a `parameters` map provided at creation time.

#### Configuration Mechanism
Parameters are injected directly into the database engine entrypoint via CLI arguments:
-   **PostgreSQL**: Passed as `-c key=value`.
-   **MySQL**: Passed as `--key=value`.

#### Replication Consistency
Read replicas automatically inherit the exact same parameter set as their primary instance, ensuring consistent behavior and performance across the database cluster.

### Managed Database Observability

The Managed Database Service includes built-in observability via sidecar exporters.

#### Metrics Sidecars
Users can enable native engine metrics by setting the `metrics_enabled` flag to `true` during provisioning. The platform will automatically launch a Prometheus-compatible exporter sidecar:
-   **PostgreSQL**: Uses `postgres-exporter` (port 9187).
-   **MySQL**: Uses `mysqld-exporter` (port 9104).

#### Scraping & Monitoring
Once enabled, the exporter's port is mapped to a host port (available in the `metrics_port` field of the database object). These endpoints are automatically registered with the platform's central Prometheus instance for dashboarding and alerting.

### Managed Database Connection Pooling

The platform supports high-performance connection pooling via sidecar containers for PostgreSQL.

#### Pooling Architecture
When `pooling_enabled` is set to `true`, the service provisions a **PgBouncer** sidecar:
1.  **Dedicated Instance**: Each database gets a private pooler instance.
2.  **Transaction Mode**: Optimized for high-throughput, short-lived connections common in web applications.
3.  **Automatic Routing**: The `GetConnectionString` API automatically returns the pooler's endpoint instead of the direct database port.

#### Lifecycle & Configuration
- **Support**: Currently exclusive to **PostgreSQL**.
- **Dynamic Toggling**: Pooling can be enabled or disabled on an existing database via the `PATCH /databases/:id` endpoint.
- **Sidecar Management**: When disabled, the pooler container is automatically terminated and cleaned up.
- **Port Mapping**: The pooler's host port is stored in the `pooling_port` field.
- **Defaults**:
    - **Max Client Connections**: 1000
    - **Default Pool Size**: 20
    - **Pool Mode**: `transaction`
    - **Image**: `edoburu/pgbouncer:latest`

This ensures that applications can scale to hundreds of concurrent clients without exhausting the database engine's backend connection limit.

### Managed Database Volume Expansion

The platform supports dynamic storage scaling for managed database instances to accommodate data growth.

#### Resizing Mechanism
Users can increase the `allocated_storage` of an existing database via the `PATCH /databases/:id` endpoint.
- **Support**: Available for both PostgreSQL and MySQL.
- **Constraints**: Storage can only be increased; shrinking volumes is prohibited to prevent data loss.

#### Implementation Details
1.  **Storage Layer**: For LVM-backed instances, the system extends the logical volume and automatically grows the underlying filesystem (ext4 or XFS) using the `lvextend -r` command.
2.  **Simulation Layer**: In Docker mode, resizing is simulated by updating the metadata and logging the action, as standard Docker volumes do not support native online resizing.
3.  **Metadata Sync**: The database record is updated atomically upon successful storage expansion to ensure consistent reporting in the API and CLI.

### Database Replication

The Cloud platform supports asynchronous replication for managed databases to provide high availability and read scaling.

#### Replication Architecture
- **Primary**: The main read-write instance. All databases start as Primary by default.
- **Replica**: Read-only instances that follow a specific Primary. Replicas are provisioned with engine-specific configurations (e.g., `PRIMARY_HOST` for PostgreSQL) to establish the replication stream.

#### Automated Failover
The `DatabaseFailoverWorker` provides automated recovery for failed primary instances:
1. **Health Monitoring**: Performs periodic TCP health checks on all instances with the `PRIMARY` role.
2. **Failure Detection**: If a Primary is unreachable, it is marked as failed.
3. **Replica Selection**: The worker identifies all healthy replicas linked to the failed Primary.
4. **Promotion**: The first available healthy replica is automatically promoted to the `PRIMARY` role using the `PromoteToPrimary` logic, which reconfigures the underlying engine and updates the metadata.

#### Manual Promotion
Replicas can be promoted manually via the API. Promoting a replica removes its link to the previous primary and converts it into a standalone Primary instance.

### Managed Database Stop/Start Lifecycle

The platform supports stopping and starting managed database instances to pause usage and save costs (similar to AWS RDS Stop/Start).

#### Stop Operation
The `POST /databases/:id/stop` endpoint stops a running database:
1. **Validation**: Database must be RUNNING and not a REPLICA (replicas must be promoted first).
2. **Compute Stop**: The database container and all sidecars (PgBouncer, exporter) are stopped via the configured compute backend. If the main container fails to stop, the operation returns an error without updating status.
3. **Status Update**: Database status transitions to `STOPPED`.
4. **Data Persistence**: The underlying volume is retained — no data is lost.

#### Start Operation
The `POST /databases/:id/start` endpoint restarts a stopped database:
1. **Validation**: Database must be in `STOPPED` state. The container ID must be present in the database record.
2. **Compute Start**: The database container is started with the existing volume.
3. **Readiness Wait**: The service polls until the instance has a non-empty IP address assigned. If readiness fails, the operation returns an error without updating status.
4. **Sidecar Start**: PgBouncer and/or metrics exporter are started if enabled.
5. **Status Update**: Database status transitions to `RUNNING`.

#### Use Cases
- **Cost Savings**: Stop dev/test databases overnight or when not in use.
- **Pause Workflows**: Halt batch processing without deleting the database.
- **Graceful Maintenance**: Stop before performing maintenance on the underlying host.

#### Constraints
- Cannot stop a REPLICA (must promote to primary first).
- Cannot stop databases in CREATING, DELETING, or STOPPED state.
- Cannot start databases that are not STOPPED.
- Volumes persist indefinitely; manually delete the database to clean up storage.

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
    tenant_id UUID NOT NULL DEFAULT auth.jwt_token().tenant_id,
    name VARCHAR(255) NOT NULL,
    schedule VARCHAR(100) NOT NULL,
    target_url TEXT NOT NULL,
    target_method VARCHAR(10) NOT NULL DEFAULT 'POST',
    target_payload TEXT,
    status VARCHAR(50) NOT NULL DEFAULT 'ACTIVE',
    last_run_at TIMESTAMPTZ,
    next_run_at TIMESTAMPTZ,
    claimed_until TIMESTAMPTZ,  -- visibility timeout for distributed claiming
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE(user_id, name)
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
    tenant_id UUID REFERENCES tenants(id) ON DELETE SET NULL,
    name VARCHAR(255) NOT NULL,
    encrypted_value TEXT NOT NULL,
    description TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    last_accessed_at TIMESTAMPTZ,
    CONSTRAINT secrets_name_tenant_key UNIQUE (name, tenant_id)
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
    version_id VARCHAR(64),
    is_latest BOOLEAN DEFAULT TRUE,
    size_bytes BIGINT NOT NULL,
    content_type VARCHAR(255),
    checksum VARCHAR(64),
    upload_status VARCHAR(20) NOT NULL DEFAULT 'AVAILABLE',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ,
    UNIQUE (bucket, key, version_id)
);

CREATE INDEX idx_objects_bucket ON objects(bucket);
CREATE INDEX idx_objects_user_id ON objects(user_id);
CREATE INDEX idx_objects_pending ON objects(created_at) WHERE upload_status = 'PENDING';
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

**Tool**: Custom embed-based migrator (`internal/repositories/postgres/migrator.go`)  
**Location**: `internal/repositories/postgres/migrations/`  
**Format**: `.up.sql` files only (one-way, forward-only)

**Migration File Naming**:
```
001_create_users.up.sql
002_create_instances.up.sql
003_add_rbac_tables.up.sql
```

**Migration Structure**:
```sql
-- +goose Up
CREATE TABLE my_table (
    id UUID PRIMARY KEY,
    name VARCHAR(255) NOT NULL
);
```

**Version Tracking**: Applied versions are tracked in `schema_migrations` table (version, dirty, created_at). Each migration runs exactly once — subsequent startups skip already-applied migrations.

### Running Migrations

**Automatic** (on API startup):
```bash
go run cmd/api/main.go
```

**Manual** (migrate only, then exit):
```bash
go run cmd/api/main.go -migrate-only
```

**Rollback**: Not supported. Migrations are forward-only and version-tracked.

### Creating New Migrations

1. Create new file: `XXX_description.up.sql` (use `.up.sql` suffix)
2. Add SQL statements (no `-- +goose Up` marker required, but harmless as SQL comment)
3. Version is extracted from numeric prefix (e.g., `072_` → version 72)
4. Test the migration on a fresh database

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

### Row-Level Security (Multi-Tenancy)

Tenant and User ID filtering in all queries:
```go
func (r *instanceRepository) List(ctx context.Context) ([]*domain.Instance, error) {
    tenantID := appcontext.TenantIDFromContext(ctx)
    query := `SELECT * FROM instances WHERE tenant_id = $1`
    rows, err := r.db.Query(ctx, query, tenantID)
    // ...
}
```

This ensures that resources are strictly isolated within their organizational boundary (Tenant), preventing cross-tenant access even if resource UUIDs are known.

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
