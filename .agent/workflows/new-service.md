---
description: Create a new service with all required layers (domain, ports, service, repository, handler)
---
# New Service Workflow

This workflow scaffolds a complete new service following the hexagonal architecture pattern.

## Prerequisites
- Know the service name (e.g., "billing", "notification")
- Understand the domain entities needed

## Steps

1. **Create Feature Branch**
```bash
git checkout -b feature/<service-name>
```

2. **Create Domain Entity**
Create `internal/core/domain/<service>.go` with:
- Domain struct(s)
- Status constants (if applicable)
- Value objects

3. **Create Port Interfaces**
Create `internal/core/ports/<service>.go` with:
- `<Service>Repository` interface (CRUD operations)
- `<Service>Service` interface (business operations)

4. **Create Service Implementation**
Create `internal/core/services/<service>.go` with:
- Service struct with dependencies
- `NewServiceParams` struct for DI
- `New<Service>Service()` constructor
- Business logic methods

5. **Create Repository Implementation**
Create `internal/repositories/postgres/<service>.go` with:
- Repository struct with `*pgxpool.Pool`
- SQL queries for all interface methods
- Error wrapping with `internal/errors`

6. **Create Handler**
Create `internal/handlers/<service>_handler.go` with:
- Handler struct with service dependency
- Request/Response DTOs
- Swagger annotations
- Input validation

7. **Wire Up Dependencies**
Update `cmd/api/main.go`:
- Initialize repository
- Initialize service with params
- Initialize handler
- Add routes

8. **Generate Swagger Docs**
// turbo
```bash
make swagger
```

9. **Create Tests**
// turbo
```bash
touch internal/core/services/<service>_test.go
```

10. **Run Tests**
// turbo
```bash
make test
```

11. **Commit Changes**
```bash
git add .
git commit -m "feat: add <service> service with full hexagonal architecture"
```
