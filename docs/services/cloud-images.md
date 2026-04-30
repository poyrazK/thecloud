# Cloud Images

Cloud Images provides a managed image registry for bootable OS templates (qcow2, img, raw, iso). Images can be registered and uploaded manually, or imported directly from remote URLs.

---

## Overview

Cloud Images is a service for managing custom OS images that can be used to launch compute instances. Images go through a lifecycle: `PENDING` → `ACTIVE` (or `ERROR` on failure).

---

## Features

- **Register + Upload**: Create image metadata first, then stream the binary file via multipart upload.
- **Import from URL**: Pull an OS image directly from a remote URL (e.g., cloud-images.ubuntu.com). The image is downloaded and stored without intermediate buffering.
- **Public Images**: Mark images as public to share them across all users in a tenant.
- **Format Detection**: Format is auto-detected from URL extension (`.qcow2`, `.img`, `.raw`, `.iso`).
- **Status Tracking**: Images transition through `PENDING` → `ACTIVE` or `ERROR` states.

---

## Internal Workings

### Register + Upload Flow

```
1. POST /images          → Create image record (status=PENDING)
2. POST /images/:id/upload → Stream file to FileStore (status=ACTIVE on success)
```

The two-step flow allows large files to be uploaded without blocking the API server, since the binary transfer happens in a separate request.

### Import from URL Flow

```
POST /images/import
  → Create image record (status=PENDING)
  → HTTP GET remote URL (30-minute timeout)
  → Validate Content-Length (max 10 GB)
  → Validate Content-Type header against allowlist
  → Read first 512 bytes for magic byte validation (qcow2, iso formats)
  → Stream response body to FileStore with size limit
  → On success: status=ACTIVE
  → On failure: status=ERROR
```

Import is synchronous and returns `202 Accepted` immediately after the image metadata is created. The download and storage happens in the foreground; for very large images (>1GB) this may take several minutes.

### Import Security Validations

URL imports are protected against SSRF and content-spoofing attacks:

| Validation | Detail |
|------------|--------|
| **Scheme restriction** | Only `http://` and `https://` allowed — no `file://`, `gopher://`, etc. |
| **Size limit** | `Content-Length` header must be ≤ 10 GB. Streaming is capped at the same limit. |
| **Content-Type allowlist** | Must be one of: `image/jpeg`, `image/png`, `image/gif`, `application/x-iso9660-image`, `application/octet-stream` |
| **Magic byte validation** | First 512 bytes are verified against expected format signature (QCOW2: `QFD¿`, ISO: `CD001`) |
| **Format-to-content consistency** | The declared URL extension (`.qcow2`, `.iso`, etc.) must match magic bytes — a `.qcow2` URL returning JPEG data is rejected |

---

## Image Model

| Field | Type | Description |
|-------|------|-------------|
| `id` | UUID | Unique identifier |
| `name` | string | Human-readable name |
| `description` | string | Optional description |
| `os` | string | OS family (e.g., `linux`, `windows`) |
| `version` | string | OS version (e.g., `22.04`) |
| `format` | string | Disk format: `qcow2`, `img`, `raw`, `iso` |
| `size_gb` | int | Image size in GB (populated after upload/import) |
| `file_path` | string | Path in object storage |
| `source_url` | string | URL the image was imported from (if applicable) |
| `is_public` | bool | Whether the image is visible to all users |
| `status` | enum | `PENDING`, `ACTIVE`, `ERROR`, `DELETING` |
| `user_id` | UUID | Owner |
| `tenant_id` | UUID | Tenant scope |

---

## CLI Usage

```bash
# Register an image (metadata only)
cloud image register my-ubuntu "Ubuntu 22.04 LTS" --os linux --version 22.04

# Upload the qcow2 file
cloud image upload <image-id> --file ./ubuntu-22.04.qcow2

# Import directly from URL (e.g., Ubuntu cloud images)
cloud image import my-ubuntu \
  --url "https://cloud-images.ubuntu.com/releases/22.04/release/ubuntu-22.04-server-cloudimg-amd64.img" \
  --os linux --version 22.04

# List images
cloud image list

# Get image details
cloud image get <image-id>

# Delete image
cloud image delete <image-id>
```

---

## Common Import Sources

| Source | URL Pattern |
|--------|-------------|
| Ubuntu Cloud Images | `https://cloud-images.ubuntu.com/releases/<version>/release/ubuntu-<version>-server-cloudimg-amd64.img` |
| CentOS Cloud | `https://cloud.centos.org/centos/<version>/images/CentOS-<version>-GenericCloud.qcow2` |
| Debian Cloud | `https://cloud.debian.org/images/cloud/<version>/latest/debian-<version>-nocloud-amd64.qcow2` |
| Fedora Cloud | `https://download.fedoraproject.org/pub/fedora/linux/releases/<version>/Cloud/x86_64/images/Fedora-Cloud-Base-<version>.qcow2` |

---

## Error Handling

| Scenario | Result |
|----------|--------|
| Remote URL returns non-200 | `ERROR` status; error message includes HTTP status code |
| Remote URL unreachable / timeout | `ERROR` status; 30-minute timeout per request |
| Content-Length exceeds 10 GB | `ERROR` status; `"image exceeds max size"` |
| Content-Type not in allowlist | `ERROR` status; `"invalid content-type"` |
| Magic bytes don't match declared format | `ERROR` status; `"invalid magic bytes for format"` |
| Invalid URL format | Returns `400 Bad Request` before image creation |
| File store write fails | `ERROR` status; partial file may remain |

---

## Relationship to Compute

Images are referenced by `LaunchInstance` via the `image` field. After an image is imported and active, it can be used as the `image` value when launching instances.

---

## Storage

Image binaries are stored in the `images` bucket in the underlying `FileStore` (local filesystem by default). The storage path is `<image-id>.<format>` (e.g., `abc123.qcow2`).
