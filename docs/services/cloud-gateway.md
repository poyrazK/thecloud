# CloudGateway (API Gateway)

CloudGateway provides entry-point routing and rate limiting for your cloud infrastructure.

## Implementation
- **Core Engine**: Built using Go's standard `net/http/httputil.ReverseProxy`.
- **Dynamic Routing**: Routes are stored in PostgreSQL and cached in memory. The service reloads routes periodically or on change.
- **Path Stripping**: Optional prefix stripping (e.g., `/gw/v1/users` -> target: `/users`).

## Design
The gateway is designed to be the "Front Door". It exposes a public `/gw/*proxy` endpoint that matches incoming paths to registered routes.

## CLI Usage
```bash
# Map /auth prefix to internal auth service
cloud gateway create-route auth-proxy /auth http://identity-service:8080 --strip

# List routes
cloud gateway list-routes
```
