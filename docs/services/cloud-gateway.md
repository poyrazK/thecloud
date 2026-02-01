CloudGateway provides entry-point routing, rate limiting, and dynamic path matching for your cloud infrastructure.

## Implementation
- **Core Engine**: Built using Go's standard `net/http/httputil.ReverseProxy` with a custom routing engine.
- **Advanced Path Matching**: Supports pattern-based routing with parameter extraction (regex-backed).
- **Dynamic Routing**: Routes are stored in PostgreSQL and cached in-memory. The service pre-compiles route patterns for sub-millisecond matching.
- **Path Stripping**: Optional prefix stripping (e.g., `/gw/v1/users` -> target: `/users`).
- **Rate Limiting**: Per-route rate limiting enforced at the gateway layer.

## Pattern Matching Syntax

CloudGateway supports powerful pattern-based routing:

| Pattern Type | Syntax | Example Path | Extracted Params |
|--------------|--------|--------------|------------------|
| **Wildcard** | `/api/v1/*` | `/api/v1/users/list` | None |
| **Parameter** | `/users/{id}` | `/users/123` | `id=123` |
| **Regex Param**| `/id/{id:[0-9]+}` | `/id/456` | `id=456` |
| **Extension** | `/files/*.{ext}` | `/files/img.png` | `ext=png` |

### Routing Priority
Routes are evaluated based on **Specificity Scoring**:
1. Exact path matches are prioritized.
2. Longest pattern matches win tie-breaks.
3. Explicit `priority` field can be used for manual overrides.

## CLI Usage

```bash
# Basic prefix mapping
cloud gateway create-route identity /auth http://identity-service:8080 --strip

# Parameterized routing
cloud gateway create-route user-api "/users/{id}" http://user-service:8080

# Constrained parameter matching (numbers only)
cloud gateway create-route post-api "/posts/{pid:[0-9]+}" http://posts-service:8080

# List all active patterns
cloud gateway list-routes
```

## Internal Context Injection
When a pattern match occurs, CloudGateway automatically extracts parameters and injects them into the request context. This allows integration with:
- **Downstream Headers**: `X-Path-Param-{Name}` headers can be injected (Future Feature).
- **IAM Policies**: Resource-level permissions using extracted IDs.
