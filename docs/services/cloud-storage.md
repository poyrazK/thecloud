# Cloud Storage Service

## Overview

Cloud Storage provides an S3-compatible object storage system designed for high availability, durability, and performance. It supports bucket-level isolation, object versioning, and end-to-end streaming encryption.

---

## Features

- **Buckets**: Logical containers for objects with per-user isolation.
- **Objects**: Arbitrary data storage with metadata and custom keys.
- **Versioning**: Preserve, retrieve, and restore every version of every object stored in a bucket.
- **Multipart Uploads**: Efficiently upload large objects in parts.
- **Streaming Encryption**: Authenticated AES-GCM encryption performed on-the-fly without buffering entire files in memory.
- **Distributed Replication**: Automatic replication of data across multiple storage nodes for fault tolerance.
- **Read Repair**: Automatic background synchronization of stale or missing replicas during read operations.

---

## Architecture

Cloud Storage follows a distributed architecture with two primary roles:

### 1. Storage Coordinator
The Coordinator is the entry point for all storage requests. Its responsibilities include:
- **Consistent Hashing**: Determines data placement across the cluster using a hash ring.
- **Replication Management**: Ensures data is written to a quorum of nodes.
- **Streaming Orchestration**: Pipes data from the API layer to multiple storage nodes simultaneously using `io.TeeReader`.
- **Read Repair**: Identifies the latest version of an object and fixes stale replicas.

### 2. Storage Nodes
Storage Nodes are responsible for persistent data storage on local disks. They provide:
- **gRPC Streaming API**: High-performance interface for data transfer.
- **Local Persistence**: Efficiently manages chunks and metadata on disk.
- **Failure Detection**: Participates in a Gossip-based cluster membership protocol to detect and report node health.

---

## Security

### Authenticated Encryption (AES-GCM Chunked)
All data stored in encrypted buckets uses a chunked AEAD scheme:
- Data is split into 64KB chunks.
- Each chunk is sealed with its own nonce and authentication tag.
- Decryption validates the integrity of each chunk before yielding it to the caller.
- This protects against both data theft (confidentiality) and tampering (integrity).

### Isolation
- **Bucket Ownership**: Buckets are owned by specific users and isolated via RBAC policies.
- **Path Sanitization**: Strict path validation prevents directory traversal attacks on storage nodes.

---

## Configuration

The storage service behavior can be adjusted via `platform.Config`:

| Parameter | Default | Description |
|-----------|---------|-------------|
| `ReplicaCount` | `3` | Number of copies per object across the cluster. |
| `WriteQuorum` | `2` | Minimum number of successful writes for a request to succeed. |
| `ChunkSize` | `1MB` | Size of chunks for internal gRPC transfer. |
| `RepairTimeout`| `30s` | Maximum time allowed for background read repairs. |

---

## CLI Usage

```bash
# List buckets
cloud storage ls

# Create a bucket
cloud storage mb my-bucket

# Upload an object
cloud storage put my-bucket/backup.zip ./backup.zip

# Download an object
cloud storage get my-bucket/backup.zip ./restore.zip

# List object versions
cloud storage versions my-bucket/backup.zip
```

---

## Observability

Exposes the following Prometheus metrics:
- `storage_operations_total`: Count of upload, download, and delete operations by bucket and result.
- `storage_bytes_transferred_total`: Total volume of data transferred.
- `storage_cluster_nodes_online`: Number of active nodes in the cluster.
