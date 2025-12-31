# Backend Guide

This document explains the Go backend implementation for the Mini AWS Compute Service.

## Project Structure (`/internal`)

The codebase follows the layout:

```text
internal/
├── core/
│   ├── domain/       # Pure data structs (Instance)
│   ├── ports/        # Interface definitions (InstanceRepository, InstanceService)
│   └── services/     # Business logic implementation
├── handlers/         # HTTP Layer (Gin Handlers)
├── repositories/     # Database & External Adapters
│   ├── docker/       # Docker Client implementation
│   └── postgres/     # (Proposed) Postgres implementation
└── platform/         # Cross-cutting concerns (Logger, Config)
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
6.  **Wire up**: Update `cmd/compute-api/main.go`.

## Testing

- **Unit Tests**: Place `_test.go` files next to the code. Mock interfaces using `mockery`.
- **Integration Tests**: (Future) Will use `testcontainers` to spin up real Docker/Postgres instances.
