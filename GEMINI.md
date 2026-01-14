# The Cloud - AI Assistant Guidelines

## Project Overview
Go cloud platform with hexagonal architecture. Frontend: Next.js 14 + TailwindCSS.

## Git Workflow

> **CRITICAL**: When developing a new feature, always create a new branch and commit there often.

```bash
git checkout -b feature/<feature-name>
git commit -m "feat: description" # Commit frequently!
```

## Architecture Rules

### Layer Dependencies (DO NOT VIOLATE)
```
Handlers → Services → Ports ← Repositories
              ↓
           Domain
```

- **Handlers** (`internal/handlers/`) import **Services only**
- **Services** (`internal/core/services/`) import **Ports and Domain only**
- **Repositories** (`internal/repositories/`) implement **Ports**

## DO NOT ❌

### Code
- ❌ Put business logic in handlers
- ❌ Import repositories directly in handlers
- ❌ Use global variables (`var DB *sql.DB`)
- ❌ Panic in production code - return errors
- ❌ Silent failures (`_ = someFunc()`)
- ❌ Magic numbers - use constants
- ❌ Skip `context.Context` as first parameter
- ❌ Create circular dependencies between packages

### Testing
- ❌ Skip tests for new services
- ❌ Test with real external dependencies in unit tests

### Git
- ❌ Commit directly to `main` branch
- ❌ Large, infrequent commits
- ❌ Commit binaries or coverage files

## DO ✅

### Code
- ✅ Use constructor injection for dependencies
- ✅ Propagate `context.Context` to all blocking calls
- ✅ Use `internal/errors` package for domain errors
- ✅ Follow existing patterns in the codebase

### New Service Pattern
1. Domain: `internal/core/domain/<name>.go`
2. Ports: `internal/core/ports/<name>.go`
3. Service: `internal/core/services/<name>.go`
4. Repository: `internal/repositories/postgres/<name>.go`
5. Handler: `internal/handlers/<name>_handler.go`
6. Wire up in `cmd/api/main.go`

### Testing
- ✅ Table-driven tests
- ✅ Mock repositories in service tests
- ✅ Use `testify/mock` for mocks

## Key Commands
```bash
make run          # Start services
make test         # Run all tests
make build        # Build binaries
make swagger      # Generate API docs
```

## File Naming
- Handlers: `*_handler.go`
- Tests: `*_test.go`
- Domain: singular (`instance.go`, `user.go`)

## Naming Conventions

### Go Code
| Element | Style | Example |
|---------|-------|---------|
| Types/Structs | PascalCase | `InstanceService`, `LaunchRequest` |
| Interfaces | PascalCase | `InstanceRepository`, `ComputeBackend` |
| Functions | PascalCase (exported), camelCase (private) | `LaunchInstance`, `parsePort` |
| Constants | SCREAMING_SNAKE (status), PascalCase (limits) | `StatusRunning`, `MaxPortsPerInstance` |
| Variables | camelCase | `instanceRepo`, `vpcID` |

### Struct Tags
```go
// Always include json tags with omitempty for optional fields
type Instance struct {
    ID        uuid.UUID  `json:"id"`
    VpcID     *uuid.UUID `json:"vpc_id,omitempty"`
    CreatedAt time.Time  `json:"created_at"`
}
```

## Error Handling Pattern
```go
// Service layer - use internal/errors
if user == nil {
    return nil, errors.New(errors.NotFound, "user not found")
}
return nil, errors.Wrap(errors.Internal, "failed to create", err)

// Handler layer - use httputil
httputil.Error(c, err)  // Maps to HTTP status automatically
httputil.Success(c, http.StatusOK, data)
```

## Handler Pattern
```go
// @Summary Short description
// @Tags resourceName
// @Security APIKeyAuth
// @Router /path [method]
func (h *Handler) Method(c *gin.Context) {
    var req Request
    if err := c.ShouldBindJSON(&req); err != nil {
        httputil.Error(c, errors.New(errors.InvalidInput, "invalid request"))
        return
    }
    result, err := h.svc.Method(c.Request.Context(), req)
    if err != nil {
        httputil.Error(c, err)
        return
    }
    httputil.Success(c, http.StatusOK, result)
}
```

## Service Constructor Pattern
```go
// Use params struct for 3+ dependencies
type ServiceParams struct {
    Repo      ports.Repository
    EventSvc  ports.EventService
    Logger    *slog.Logger
}

func NewService(params ServiceParams) *Service {
    return &Service{
        repo:     params.Repo,
        eventSvc: params.EventSvc,
        logger:   params.Logger,
    }
}
```

## Port Interface Pattern
```go
// Repository interfaces - CRUD operations with context
type Repository interface {
    Create(ctx context.Context, entity *domain.Entity) error
    GetByID(ctx context.Context, id uuid.UUID) (*domain.Entity, error)
    Update(ctx context.Context, entity *domain.Entity) error
    Delete(ctx context.Context, id uuid.UUID) error
}

// Service interfaces - business operations
type Service interface {
    CreateEntity(ctx context.Context, name string) (*domain.Entity, error)
    ListEntities(ctx context.Context) ([]*domain.Entity, error)
}
```
