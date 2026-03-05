# ADR 021: RDS Volume Expansion

## Status
Accepted

## Context
Managed Database (RDS) instances in The Cloud platform utilize persistent block storage via the `VolumeService`. While users can specify the storage size during initial provisioning, application growth often requires increasing the available disk space without manual data migration or significant downtime. The underlying storage backends (LVM) support dynamic resizing, but the integration was missing in the database lifecycle management.

## Decision
We will implement dynamic volume expansion for RDS instances. This allows users to increase the `allocated_storage` of an existing database instance via the `PATCH /databases/:id` endpoint.

### Technical Implementation
1.  **LVM Adapter Enhancement**: The `ResizeVolume` method in the `LVMAdapter` will use the `-r` (or `--resizefs`) flag with the `lvextend` command. This ensures that the underlying filesystem (ext4 or XFS) is automatically grown immediately after the logical volume is extended.
2.  **Database Service Logic**:
    *   The `ModifyDatabase` method will validate that the new requested size is larger than the current allocation (shrinking is prohibited).
    *   It will identify the associated volume and call the `VolumeService.ResizeVolume` method.
    *   Upon successful backend resizing, the database metadata in the repository will be updated to reflect the new `allocated_storage` value.
3.  **Docker Backend**: For the Docker compute backend (primarily used for simulation and development), volume resizing will be implemented as a "no-op" that logs the action. This maintains logical consistency across the service layer while acknowledging the inherent limitations of standard Docker volumes.

## Consequences

### Positive
*   **Operational Scalability**: Users can scale database storage on-demand as their data grows.
*   **Reduced Downtime**: Leveraging LVM's online resizing capabilities allows for disk expansion without requiring database restarts in most common scenarios.
*   **Consistency**: Metadata and physical storage are kept in sync via the atomic `ModifyDatabase` operation.

### Negative
*   **Shrink Limitation**: Disk shrinking is not supported due to the high risk of filesystem corruption and the complexity of moving data blocks.
*   **Backend Support**: Full dynamic resizing is only available on backends that support it (like LVM). Docker simulation remains fixed-size.

### Limitations
*   **Filesystem Ready**: The automatic resize relies on standard Linux utilities (`resize2fs`, `xfs_growfs`) being available on the host.
*   **Quota Management**: Future iterations should ensure that the new volume size is checked against tenant storage quotas before the operation begins.
