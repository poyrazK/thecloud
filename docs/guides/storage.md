# Storage Guide

This guide shows you how to use the The Cloud object storage service.

## Prerequisites
- `cloud` CLI installed (`make install`)
- API key configured (`cloud auth create-demo my-user`)
- API server running (`make run`)

## Commands

### Upload a File
```bash
cloud storage upload <bucket> <file>
```
**Example:**
```bash
cloud storage upload photos cat.jpg
```

### List Objects
```bash
cloud storage list <bucket>
```
**Example:**
```bash
cloud storage list photos
```
Output:
```
┌─────────┬──────┬─────────────────────┬──────────────────────────────────────────┐
│   KEY   │ SIZE │     CREATED AT      │                   ARN                    │
├─────────┼──────┼─────────────────────┼──────────────────────────────────────────┤
│ cat.jpg │ 1024 │ 2026-01-01T11:00:00 │ arn:thecloud:storage:local:default:...    │
└─────────┴──────┴─────────────────────┴──────────────────────────────────────────┘
```

### Download a File
```bash
cloud storage download <bucket> <key> <destination>
```
**Example:**
```bash
cloud storage download photos cat.jpg ./local-cat.jpg
```

### Delete a File
```bash
cloud storage delete <bucket> <key>
```
**Example:**
```bash
cloud storage delete photos cat.jpg
```

## How It Works
The Cloud uses a multi-node **Distributed Object Storage (v2)** system:

- **Metadata**: Managed via the API and stored in PostgreSQL (`objects` table).
- **Architecture**: 
  - **Coordinator**: Receives requests, identifies target nodes using a **Consistent Hash Ring**, and handles replication.
  - **Storage Nodes**: Store the actual file bytes and participate in a **Gossip Protocol** for decentralized health tracking.
- **Replication**: Configurable N+M replication with write-quorum for high availability.
- **ARN Format**: `arn:thecloud:storage:distributed:default:object/<bucket>/<key>`
