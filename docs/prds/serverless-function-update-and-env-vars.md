# PRD: CloudFunctions — Function Update & Environment Variables

## Problem Statement

CloudFunctions currently only supports two operations after creation: invoke and delete. Users cannot:

- Adjust timeout or memory without deleting and recreating the function (which changes the function ID and breaks any references)
- Set environment variables, meaning functions cannot connect to databases, access secrets, or receive runtime configuration without hardcoding values

This blocks production use cases where runtime parameters change across environments (dev/staging/prod).

## User Stories

- **As a Developer**, I want to update a function's timeout and memory without deleting it, so that I can tune performance without breaking existing integrations.
- **As a Developer**, I want to set environment variables for my function, so that it can connect to my database using `DATABASE_URL` without hardcoding credentials.
- **As an Operator**, I want to enable/disable a function by toggling its status, so that I can temporarily halt execution without deleting the function.

## Acceptance Criteria

- [x] `PATCH /functions/:id` endpoint accepts partial updates (timeout, memory, handler, status, env_vars)
- [x] New `env_vars` JSONB column stored in the `functions` table
- [x] Environment variables are injected into the Docker container at invocation time as `KEY=value`
- [x] `cloud fn update` CLI command supports `--timeout`, `--memory`, `--handler`, `--status`, `--env KEY=VALUE`
- [x] Go SDK exposes `UpdateFunction` method
- [x] Timeout validated 1–900 seconds; memory validated 64–10240 MB
- [x] Audit log entry written on function update
- [x] Unit and handler tests added

## UX / CLI Design

```bash
# Update timeout and memory
cloud fn update my-func --timeout 120 --memory 512

# Set environment variables
cloud fn update my-func --env DATABASE_URL=postgres://prod/db --env DEBUG=false

# Disable a function
cloud fn update my-func --status INACTIVE

# View current configuration
cloud fn list
```

**API:**

```http
PATCH /functions/:id
Content-Type: application/json

{
  "timeout": 120,
  "memory_mb": 512,
  "env_vars": [
    { "key": "DATABASE_URL", "value": "postgres://prod/db" }
  ]
}
```

## Out of Scope

- Secret references in env vars (e.g. `$SECRET:my-api-key`) — tracked in a separate PR
- Per-invocation timeout override — tracked in the concurrency limits PR
- Function code re-deploy without ID change — tracked in a separate PR
