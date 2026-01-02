# API Reference

## Authentication

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
  "token": "api_key...",
  "expires_in": 86400
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
  "vpc_id": "vpc-uuid"
}
```

### GET /instances/:id
Get details of a specific instance.

### PUT /instances/:id
Update instance (e.g., status).

### DELETE /instances/:id
Terminate an instance.

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

## Load Balancers

**Headers Required:** `X-API-Key: <your-api-key>`

### GET /loadbalancers
List load balancers.

### POST /loadbalancers
Create a load balancer.

---

## Auto-Scaling Groups

**Headers Required:** `X-API-Key: <your-api-key>`

### GET /autoscaling/groups
List auto-scaling groups.

### POST /autoscaling/groups
Create an ASG.

---

## Error Codes

| Status Code | Description |
|-------------|-------------|
| 200/201 | Success |
| 400 | Bad Request (Invalid input) |
| 401 | Unauthorized (Missing/Invalid API Key) |
| 403 | Forbidden (Access denied to resource) |
| 404 | Not Found |
| 500 | Internal Server Error |
