# ADR 017: Managed Database Configuration (Parameter Groups)

## Status
Accepted

## Context
The Managed Database Service (RDS) previously used default configurations from official Docker images. For production workloads, users need to tune engine-specific settings such as `max_connections`, `shared_buffers`, and `innodb_buffer_pool_size`. 

We needed a way to:
1.  **Inject custom configuration** at runtime.
2.  **Ensure consistency** between primary and replica instances.
3.  **Remain cloud-native** and avoid complex shared filesystem requirements for config files.

## Decision
We implemented dynamic configuration management via "Parameter Groups" (key-value maps) injected through the container entrypoint.

### 1. CLI-based Injection
Rather than generating and mounting configuration files (which requires a shared filesystem or complex sync logic), we leverage the ability of `postgres` and `mysqld` to accept overrides via command-line arguments:
-   **PostgreSQL**: Arguments are passed as `-c key=value`.
-   **MySQL**: Arguments are passed as `--key=value`.

### 2. Implementation in DatabaseService
The `DatabaseService` now generates a custom `Cmd` slice based on the provided `parameters` map. This `Cmd` overrides the default entrypoint command of the container.

### 3. Metadata Persistence
A `parameters` column (JSONB) was added to the `databases` table. This allows the platform to:
-   Track which settings were applied.
-   Ensure replicas inherit exactly the same parameters as their primary counterpart.
-   Enable future "Update Parameter" operations.

### 4. Inheritance
Read replicas automatically inherit the `parameters` map from their primary database during provisioning, ensuring identical performance characteristics and behavior across the cluster.

## Consequences

### Positive
-   **Stateless Configuration**: No host-level files are required, making the system easier to scale across distributed compute nodes.
-   **Atomic Deployment**: Configuration is applied at the same time the container starts.
-   **Cluster Consistency**: Automated inheritance reduces the risk of split-brain or performance degradation due to mismatched settings in replicas.

### Negative
-   **Restart Required**: Changing parameters currently requires a database restart (container recreation), as they are passed at boot time.
-   **Validation Complexity**: Different engines have different parameter naming conventions and valid value ranges; basic string-based passthrough is used for now.
