# API Reference

## Authentication

> **Rate Limiting**: Auth endpoints are rate-limited to 5 requests per minute per IP to prevent brute-force attacks.

### POST /auth/register
Register a new user account.

**Request:**
```json
{
  "email": "user@example.com",
  "password": "password",
  "name": "User Name"
}
```

**Response:**
```json
{
  "message": "user created successfully",
  "user_id": "uuid"
}
```

### POST /auth/login
Login to obtain an API Key.

**Request:**
```json
{
  "email": "user@example.com",
  "password": "password"
}
```

**Response:**
```json
{
  "user": {
    "id": "uuid",
    "email": "user@example.com",
    "name": "User Name"
  },
  "api_key": "thecloud_xxxxx"
}
```

### POST /auth/forgot-password
Request a password reset token (rate limited: 5 requests/minute).

**Request:**
```json
{
  "email": "user@example.com"
}
```

**Response:**
```json
{
  "message": "If the email exists, a reset token has been sent."
}
```

### POST /auth/reset-password
Reset password using a valid token (rate limited: 5 requests/minute).

**Request:**
```json
{
  "token": "reset-token-from-email",
  "new_password": "new-secure-password"
}
```

**Response:**
```json
{
  "message": "password updated successfully"
}
```

---

## API Key Management

**Headers Required:** `X-API-Key: <your-api-key>`

### POST /auth/keys
Create a new API key.

**Request:**
```json
{
  "name": "Production Key"
}
```

**Response:**
```json
{
  "id": "uuid",
  "user_id": "user-uuid",
  "name": "Production Key",
  "key": "thecloud_xxxxx",
  "created_at": "2026-01-05T23:00:00Z",
  "last_used": "2026-01-05T23:00:00Z"
}
```

### GET /auth/keys
List all API keys for the authenticated user.

### DELETE /auth/keys/:id
Revoke an API key.

### POST /auth/keys/:id/rotate
Rotate an API key (creates new key, deletes old one).

### POST /auth/keys/:id/regenerate
Regenerate an API key (alias for rotate).

---

## System Health

### GET /health/live
Liveness probe. Returns 200 OK if the process is running.
**Response:** `{"status": "ok"}`

### GET /health/ready
Readiness probe. Checks connections to Database and Docker daemon.
**Response (200 OK):**
```json
{
  "status": "UP",
  "checks": { "database": "CONNECTED", "docker": "CONNECTED" },
  "time": "..."
}
```
**Response (503 Service Unavailable):**
```json
{
  "status": "DEGRADED",
  "checks": { "database": "DISCONNECTED", ... }
}
```

---

## Compute Instances

**Headers Required:** `X-API-Key: <your-api-key>`

### GET /instances
List all instances owned by the authenticated user.

### POST /instances
Launch a new instance.
```json
{
  "name": "web-01",
  "image": "nginx",
  "instance_type": "basic-2",
  "vpc_id": "vpc-uuid",
  "subnet_id": "subnet-uuid",
  "ports": "80:80",
  "volumes": [
    { "volume_id": "vol-uuid", "mount_path": "/data" }
  ]
}
```

### GET /instance-types
List all available instance types.
**Response:**
```json
[
  {
    "id": "basic-2",
    "name": "Basic 2",
    "vcpus": 1,
    "memory_mb": 1024,
    "disk_gb": 10,
    "network_mbps": 1000,
    "price_per_hour": 0.02,
    "category": "general-purpose"
  }
]
```

### GET /instances/:id
Get details of a specific instance.

### PUT /instances/:id
Update instance (e.g., status).

### DELETE /instances/:id
Terminate an instance.

### GET /instances/:id/console
Get the VNC console URL for the instance.
**Response:**
```json
{
  "console_url": "vnc://127.0.0.1:5901"
}
```

---

## Networks (VPC)

**Headers Required:** `X-API-Key: <your-api-key>`

### GET /vpcs
List all VPCs.

### POST /vpcs
Create a new VPC.
```json
{
  "name": "prod-vpc"
}
```

### DELETE /vpcs/:id
Delete a VPC.

---

## Security Groups

**Headers Required:** `X-API-Key: <your-api-key>`

### GET /security-groups
List all security groups.
Query params: `?vpc_id=<vpc-uuid>` (required)

### POST /security-groups
Create a new security group.
```json
{
  "vpc_id": "vpc-uuid",
  "name": "web-tier",
  "description": "Port 80 and 443 allowed"
}
```

### GET /security-groups/:id
Get details and rules for a security group.

### DELETE /security-groups/:id
Delete a security group.

### POST /security-groups/:id/rules
Add a firewall rule.
```json
{
  "direction": "ingress",
  "protocol": "tcp",
  "port_min": 80,
  "port_max": 80,
  "cidr": "0.0.0.0/0",
  "priority": 100
}
```

### DELETE /security-groups/rules/:rule_id
Remove a firewall rule.

### POST /security-groups/attach
Attach a security group to an instance.
```json
{
  "instance_id": "inst-uuid",
  "group_id": "sg-uuid"
}
```

### POST /security-groups/detach
Detach a security group from an instance.
```json
{
  "instance_id": "inst-uuid",
  "group_id": "sg-uuid"
}
```

---

## Subnets

**Headers Required:** `X-API-Key: <your-api-key>`

### GET /vpcs/:vpc_id/subnets
List all subnets in a VPC.

### POST /vpcs/:vpc_id/subnets
Create a new subnet.
```json
{
  "name": "private-subnet-1",
  "cidr_block": "10.0.1.0/24",
  "availability_zone": "us-east-1a"
}
```

### GET /subnets/:id
Get details of a specific subnet.

### DELETE /subnets/:id
Delete a subnet.

---

## Elastic IPs (Static IPs) 🆕

**Headers Required:** `X-API-Key: <your-api-key>`

### GET /elastic-ips
List all allocated elastic IPs for the tenant.

### POST /elastic-ips
Allocate a new elastic IP. No body required.
**Response:**
```json
{
  "id": "uuid",
  "public_ip": "100.64.x.y",
  "status": "allocated",
  "arn": "arn:thecloud:vpc:local:tenant:eip/uuid"
}
```

### GET /elastic-ips/:id
Get details of a specific elastic IP.

### DELETE /elastic-ips/:id
Release an elastic IP back to the pool. Fails if still associated.

### POST /elastic-ips/:id/associate
Associate an elastic IP with a compute instance.
**Request:**
```json
{
  "instance_id": "inst-uuid"
}
```

### POST /elastic-ips/:id/disassociate
Disassociate an elastic IP from its current instance.

---

## Volumes

**Headers Required:** `X-API-Key: <your-api-key>`

### GET /volumes
List all volumes.

### POST /volumes
Create a new volume.
```json
{
  "name": "data-vol",
  "size_gb": 10
}
```

---

## Global Load Balancers 🆕

**Headers Required:** `X-API-Key: <your-api-key>`

### GET /global-lb
List all global load balancers owned by the authenticated user.

### POST /global-lb
Create a new global load balancer.
```json
{
  "name": "api-global",
  "hostname": "api.myapp.com",
  "policy": "LATENCY",
  "health_check": {
    "protocol": "HTTP",
    "port": 80,
    "path": "/health"
  }
}
```

### GET /global-lb/:id
Get details of a GLB including its endpoints.

### DELETE /global-lb/:id
Delete a global load balancer and its DNS records. Only the owner can delete the resource.

### POST /global-lb/:id/endpoints
Add a regional endpoint to the GLB.
```json
{
  "region": "us-east-1",
  "target_type": "IP",
  "target_ip": "1.2.3.4",
  "weight": 100
}
```

### DELETE /global-lb/:id/endpoints/:epID
Remove an endpoint from the GLB.

---

## Auto-Scaling Groups

**Headers Required:** `X-API-Key: <your-api-key>`

### GET /autoscaling/groups
List auto-scaling groups.

### POST /autoscaling/groups
Create an ASG.

---

## Cloud Gateway

**Headers Required:** `X-API-Key: <your-api-key>`

### GET /routes
List all registered routes.

### POST /routes
Register a new gateway route. Supports advanced pattern matching and HTTP method filtering.

**Request:**
```json
{
  "name": "users-api",
  "path_prefix": "/users/{id}",
  "target_url": "http://user-service:8080",
  "methods": ["GET", "PUT"],
  "strip_prefix": true,
  "priority": 10
}
```

**Fields:**
- `path_prefix`: The pattern to match (e.g., `/api/*`, `/users/{id}`, `/id/{id:[0-9]+}`).
- `methods`: Array of allowed HTTP methods (empty/null = all).
- `priority`: Higher values take precedence when multiple patterns match.
- `strip_prefix`: If true, the matched part of the path is removed before forwarding.

**Extracted Parameters:**
Matched parameters like `{id}` are made available to downstream services as headers (in the future) and are currently injected into the gateway context.

### DELETE /routes/:id
Remove a route.

---

## Cloud Functions

**Headers Required:** `X-API-Key: <your-api-key>`

### GET /functions
List all deployed functions.

### POST /functions
Create a new function.
```json
{
  "name": "hello-world",
  "runtime": "nodejs20",
  "code_zip": "<base64>"
}
```

### POST /functions/:id/invoke
Invoke a function.
```json
{
  "payload": { "foo": "bar" },
  "async": false
}
```

### GET /functions/:id/logs
Get execution logs.

---

## Cloud Cache (Redis)

**Headers Required:** `X-API-Key: <your-api-key>`

### GET /caches
List all cache instances.

### POST /caches
Provision a new Redis cache.
```json
{
  "name": "my-cache",
  "memory_mb": 256
}
```

### DELETE /caches/:id
Terminate a cache instance.

---

## Cloud Queue

**Headers Required:** `X-API-Key: <your-api-key>`

### GET /queues
List all queues.

### POST /queues
Create a new message queue.
```json
{
  "name": "task-queue",
  "visibility_timeout": 30
}
```

### POST /queues/:id/messages
Send a message.
```json
{
  "body": "payload-data"
}
```

### GET /queues/:id/messages
Receive messages.
Query params: `?count=1`

---

## Cloud Notify (Pub/Sub)

**Headers Required:** `X-API-Key: <your-api-key>`

### GET /topics
List topics.

### POST /topics
Create a new topic.

### POST /subscriptions
Subscribe to a topic.
```json
{
  "topic_id": "uuid",
  "protocol": "webhook",
  "endpoint": "http://my-api/hook"
}
```

---

## Cloud Cron

**Headers Required:** `X-API-Key: <your-api-key>`

### GET /cron
List scheduled jobs.

### POST /cron
Create a scheduled job.
```json
{
  "name": "daily-cleanup",
  "schedule": "0 0 * * *",
  "target_url": "http://my-service/cleanup"
}
```

### POST /cron/:id/pause
Pause a job.

### POST /cron/:id/resume
Resume a job.

---

## Managed Kubernetes (KaaS)

**Headers Required:** `X-API-Key: <your-api-key>`

### GET /clusters
List all Kubernetes clusters.

### POST /clusters
Create a new Kubernetes cluster (Asynchronous).
```json
{
  "name": "my-cluster",
  "vpc_id": "vpc-uuid",
  "version": "v1.29.0",
  "workers": 3,
  "ha": true
}
```

### GET /clusters/:id
Get detailed information and status.

### DELETE /clusters/:id
Delete a cluster (Asynchronous).

### GET /clusters/:id/kubeconfig
Download the admin kubeconfig.

### GET /clusters/:id/health
Get operational health (nodes ready, api server reachability).

### POST /clusters/:id/upgrade
Upgrade cluster version (Asynchronous).
```json
{
  "version": "v1.30.0"
}
```

### POST /clusters/:id/scale
Scale worker nodes.
```json
{
  "workers": 5
}
```

### POST /clusters/:id/repair
Re-run bootstrap scripts on nodes.

### POST /clusters/:id/backups
Trigger an etcd backup.

### POST /clusters/:id/restore
Restore etcd state from path.
```json
{
  "backup_path": "/backups/etcd-snapshot.db"
}
```

## Error Codes

| Status Code | Description |
|-------------|-------------|
| 200/201 | Success |
| 400 | Bad Request (Invalid input) |
| 401 | Unauthorized (Missing/Invalid API Key) |
| 403 | Forbidden (Access denied to resource) |
| 404 | Not Found |
| 429 | Too Many Requests (Rate limit exceeded) |
| 500 | Internal Server Error |
