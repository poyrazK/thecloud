# Backend Guide

This document explains the Go backend implementation for the The Cloud Compute Service.

## Project Structure (`/internal`)

The codebase follows the layout:

```text
internal/
├── core/
│   ├── domain/       # Pure data structs (Instance, VPC, Event)
│   ├── ports/        # Interface definitions
│   └── services/     # Business logic (Instance, Dashboard, Event)
├── handlers/         # HTTP Handlers (REST)
│   └── ws/           # Real-time WebSocket Hub & Streaming
├── repositories/     # Adapters
│   ├── docker/       # Docker SDK (Compute, Networks, Volumes)
│   └── postgres/     # PostgreSQL (pgx)
└── platform/         # Metrics, Logger, Database Setup
```

## API Design Standards

### JSON Envelopes
Currently API responses are direct JSON objects. Future version 2.0 will introduce envelopes:
```json
{
  "data": { ... },
  "error": null,
  "meta": { "request_id": "..." }
}
```

### Context Usage
Every function in the Service and Repository layers MUST accept `context.Context` as the first argument to support cancellation and timeouts.

## How to Add a New API Feature

1.  **Define Domain**: Add structs to `internal/core/domain/`.
2.  **Define Interface**: Add methods to `internal/core/ports/`.
3.  **Implement Repository**: Add code to `internal/repositories/<adapter>/`.
4.  **Implement Service**: Add business rules to `internal/core/services/`.
5.  **Expose Handler**: Add Gin route in `internal/handlers/`.
6.  **Wire up**: Update `cmd/api/main.go`.

## Testing

The Cloud has comprehensive test coverage (51.3%) across all layers:

### Test Organization

**Unit Tests** (`*_test.go` files):
- **Services**: `internal/core/services/*_test.go` (55.4% coverage)
  - Business logic testing with mocked dependencies
  - Located in `shared_test.go` for reusable mocks
- **Handlers**: `internal/handlers/*_test.go` (52.8% coverage)
  - HTTP endpoint testing using `httptest`
  - Gin test mode for request/response validation

**Integration Tests** (require `//go:build integration` tag):
- **Repositories**: `internal/repositories/postgres/*_test.go` (57.5% coverage)
  - Real PostgreSQL database interactions
  - Uses `setupDB()` helper for test database setup
  - Includes cleanup functions for test isolation

### Running Tests

```bash
# Unit tests only (no database required)
go test ./...

# Integration tests (requires PostgreSQL)
docker compose up -d postgres
go test -tags=integration ./...

# Coverage report
go test -tags=integration -coverprofile=coverage.out ./...
go tool cover -html=coverage.out
```

### Mocking Strategy

We use `testify/mock` for interface mocking:
- Mock definitions in `internal/core/services/shared_test.go`
- Mocks for all repository and service interfaces
- Consistent mock expectations using `On()` and `AssertExpectations()`

### Coverage Goals

| Layer | Current | Target |
|-------|---------|--------|
| Services | 55.4% | 80% |
| Handlers | 52.8% | 80% |
| Repositories | 57.5% | 75% |
| **Overall** | **51.3%** | **80%** |
