# Storage Guide

This guide shows you how to use the Mini AWS object storage service.

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
│ cat.jpg │ 1024 │ 2026-01-01T11:00:00 │ arn:miniaws:storage:local:default:...    │
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
- **Metadata**: Stored in PostgreSQL (`objects` table)
- **File Bytes**: Stored in `./miniaws-data/local/storage/<bucket>/<key>`
- **ARN Format**: `arn:miniaws:storage:local:default:object/<bucket>/<key>`
