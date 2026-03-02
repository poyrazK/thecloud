# ADR 020: RDS Connection Pooling via Sidecar

## Status
Accepted

## Context
Managed Database (RDS) users require efficient connection management, especially for PostgreSQL which has a high per-connection memory overhead. Application-level pooling is often insufficient for distributed serverless or microservices architectures where many small clients connect to a single database.

## Decision
We will implement connection pooling for Managed PostgreSQL instances using a sidecar pattern with **PgBouncer**.

### Architecture
1.  **One Pooler Per Database**: Each database instance with pooling enabled will have its own dedicated PgBouncer container.
2.  **Sidecar Isolation**: The pooler container is isolated to the specific database instance, ensuring no cross-tenant resource contention.
3.  **Transaction Mode**: Default to `transaction` pooling mode to maximize connection reuse.
4.  **Network Placement**: The pooler is attached to the same VPC (Docker network) as the database instance.
5.  **Dynamic Routing**: The `GetConnectionString` API will dynamically return the host port of the PgBouncer sidecar when pooling is enabled, making it transparent to the user.

### Configuration (PgBouncer)
- **Image**: `bitnami/pgbouncer:latest`
- **Port**: Internal `6432` mapped to a dynamic host port.
- **Max Client Connections**: 1000
- **Default Pool Size**: 20
- **Authentication**: Passthrough authentication using the database credentials.

## Consequences
- **Improved Scalability**: Support for hundreds of client connections with minimal backend resource usage.
- **Lifecycle Management**: The pooler sidecar is automatically provisioned, restarted, and deleted alongside the database instance.
- **Port Management**: Requires one additional dynamic host port per database instance.
- **Engine Support**: Initial support is limited to PostgreSQL. MySQL support is deferred.

### Limitations
Connection pooling uses PgBouncer in **transaction mode** to maximize reuse. This introduces several limitations that users must be aware of:
- **Session State**: Parameters set via `SET` or `RESET` are not preserved across transactions.
- **LISTEN/NOTIFY**: Unreliable as notifications are tied to specific backend connections which may change between transactions.
- **Prepared Statements**: Server-side `PREPARE` statements are not supported.
- **Advisory Locks**: Session-level advisory locks may be lost or incorrectly shared.
- **Temporary Tables**: Temporary tables created with `ON COMMIT PRESERVE ROWS` will not persist correctly across transactions.

**Guidance**: If an application relies on persistent session state, explicit advisory locks, or complex temporary table usage, connection pooling should be disabled. Users can enable/disable pooling via the `pooling_enabled` flag at creation or restore time. Note that `GetConnectionString` dynamically routes to the pooler when enabled, making the integration transparent to the application code.
