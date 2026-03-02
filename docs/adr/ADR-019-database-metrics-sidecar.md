# ADR 019: Managed Database Observability (Sidecar Exporters)

## Status
Accepted

## Context
The Managed Database Service (RDS) provided limited visibility into internal engine performance. Users needed access to native database metrics (e.g., connection counts, buffer cache hit ratios, transaction rates) to effectively monitor and scale their workloads.

We needed a solution that was:
1.  **Low Impact**: Minimal interference with the database engine itself.
2.  **Scalable**: Automatically provisioned with every database instance.
3.  **Standardized**: Exporting metrics in a format compatible with the platform's Prometheus infrastructure.

## Decision
We implemented a "Sidecar" pattern for database observability.

### 1. Automated Sidecar Injection
When a user enables metrics for a database (via the `metrics_enabled` flag), the `DatabaseService` automatically provisions a secondary "exporter" container in the same network space as the database instance.

### 2. Standard Exporters
We utilize industry-standard Prometheus exporters:
-   **PostgreSQL**: `prometheuscommunity/postgres-exporter`
-   **MySQL**: `prom/mysqld-exporter`

### 3. Secure Internal Communication
The sidecar connects to the database using its internal VPC/network IP. Authentication is handled using the same administrative credentials generated for the database instance.

### 4. Lifecycle Management
The sidecar container is linked to the database metadata. When a database is deleted, the platform ensures the associated exporter container is also terminated and cleaned up.

## Consequences

### Positive
-   **Deep Visibility**: Users gain access to hundreds of engine-specific metrics without manual configuration.
-   **Isolation**: Metric collection runs in a separate process/container, preventing exporter issues from affecting database availability.
-   **Seamless Integration**: Metrics are immediately available for scraping by the platform's central Prometheus instance.

### Negative
-   **Resource Overhead**: Each database instance now potentially uses two containers, increasing the overall memory and CPU footprint slightly.
-   **Port Consumption**: Each metrics sidecar requires an additional host-mapped port for scraping.
