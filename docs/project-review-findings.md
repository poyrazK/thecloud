# Project Review Findings

Date: 2026-04-04

This document records the major findings from a code review of the project.

Scope:
- Backend architecture and runtime wiring
- Authentication and tenancy
- Realtime and websocket behavior
- Async workers and queue reliability
- Migration and test realism
- Frontend integration maturity

Note:
- File and line references reflect the codebase at the time of review.
- This is intentionally blunt and prioritizes risk over politeness.

## Executive Summary

The project is ambitious and already has a wide feature surface. The bigger problem is not missing features. The bigger problem is trustworthiness. Several core paths look feature-complete from the outside but have correctness, security, or reliability gaps underneath.

Current state:
- Strong prototype / strong platform experiment
- Not yet a trustworthy multi-tenant control plane
- Backend is much more mature than the frontend
- Architecture is organized, but some important boundaries are already eroding

## Highest Priority Findings

### 1. Realtime path is broken and unsafe

Severity: Critical

Status: ✅ RESOLVED (PR #113)

What was wrong:
- `internal/api/setup/dependencies.go:214-221` creates a websocket hub for `EventService`.
- `internal/api/setup/router.go:79-80,122-123` creates a different hub for connected websocket clients.
- This means published events and connected clients are likely using different hubs.
- Even if that wiring is fixed, `internal/core/services/event.go:74-79` broadcasts globally.
- `internal/handlers/ws/hub.go:55-65` sends to all clients without tenant or user filtering.
- `internal/handlers/ws/handler.go:16-19` allows every origin.
- `internal/handlers/ws/handler.go:40-60` authenticates via `api_key` query param.

Why it mattered:
- Live events likely do not work correctly.
- If fixed naively, they still become a cross-tenant data leak.
- Query-string credentials are easy to leak through logs, browser history, proxies, and monitoring.

Resolution (PR #113):
- Single hub instance shared via `Services.WsHub`
- `RealtimePublisher` port interface for hexagonal architecture
- `BroadcastEventToTenant` with tenant/user filtering
- Bearer token auth via `Authorization` header
- Origin validation when `AllowedOrigins` is configured

### 2. API keys are handled as plaintext secrets

Severity: High

Status: ✅ RESOLVED (PR #115)

What was wrong:
- API keys are stored in plaintext in Postgres: `internal/repositories/postgres/migrations/002_create_api_keys_table.up.sql:3-12`.
- They are read back directly: `internal/repositories/postgres/identity_repo.go:37-43`.
- The domain model exposes the raw key: `internal/core/domain/identity.go:10-21`.
- Handlers return full keys from create, list, and rotate endpoints: `internal/handlers/identity_handler.go:37-69,104-119`.
- Full key objects are cached in Redis: `internal/core/services/cached_identity.go:40-64`.

Why it mattered:
- A DB leak or Redis leak becomes an immediate credential compromise.
- Listing keys should not return active secrets.

Resolution (PR #115):
- Store SHA-256 hash in `key_hash` column, raw key shown only at create/rotate
- `Key` field uses `omitempty` — empty on list, populated only at create/rotate
- Redis cache keyed by hash (`apikey:hash:{sha256}`) not raw key
- `GetAPIKeyByKey` → `GetAPIKeyByHash` for lookup

### 3. Revoked and rotated API keys can remain valid

Severity: High

What is wrong:
- `internal/core/services/cached_identity.go:40-64` caches successful validations.
- `internal/core/services/cached_identity.go:71-75` does not invalidate the cache on revoke or rotate.

Why it matters:
- A revoked key can continue authenticating until cache TTL expires.
- This weakens the whole credential lifecycle story.

Recommendation:
- Invalidate by key on revoke and rotate.
- If necessary, maintain a reverse mapping from key ID to cached key string.

### 4. API key tenant/default-tenant handling looks incomplete

Severity: High

What is wrong:
- Tenancy was added to `api_keys` in migrations: `internal/repositories/postgres/migrations/070_create_tenants.up.sql:37-41` and `internal/repositories/postgres/migrations/102_add_tenant_id_to_missing_resources.up.sql:12-19`.
- Auth resolves tenant context from the API key: `pkg/httputil/auth.go:35-45,101-125`.
- `internal/core/services/identity.go:71-78` creates keys without setting `TenantID` or `DefaultTenantID`.

Why it matters:
- Default tenant resolution may silently fail for newly created keys.
- Multi-tenant behavior becomes inconsistent across keys.
- This may be hidden by current test setup rather than actually correct.

Recommendation:
- Set key tenant metadata explicitly at creation time.
- Decide whether keys are user-global or tenant-bound, then enforce it consistently.

### 5. Presigned URL flow is inconsistent and likely broken in real use

Severity: High

What is wrong:
- `internal/core/services/storage.go:806-819` signs URLs with `SecretsEncryptionKey`.
- `internal/handlers/storage_handler.go:428-436,469-477` verifies with `StorageSecret`.
- `internal/handlers/storage_handler.go:387-405` parses `expiry_seconds` and ignores it.
- Presigned download/upload call the normal storage service path, which still does RBAC: `internal/core/services/storage.go:84-90,204-209`.
- Presigned `PUT` generation checks read permission only: `internal/core/services/storage.go:790-793`.

Why it matters:
- Signing and verification can disagree.
- Anonymous presigned access is not cleanly modeled.
- Read-only users can mint upload URLs.

Recommendation:
- Use one signing secret and one verification path.
- Make method-specific permission checks.
- Encode the subject/tenant context into the signed token or model presigned operations separately.

### 6. Async task processing is not durable enough for infrastructure control

Severity: High

What is wrong:
- Redis queue is implemented as `LPUSH` + `BRPOP`: `internal/repositories/redis/task_queue.go:22-46`.
- Once a worker pops a message, it is gone before the work is complete.
- `internal/core/services/instance.go:172-179,208-237` reserves quota and creates the instance before enqueue.
- On enqueue failure, quota and DB state are not rolled back.
- `internal/workers/provision_worker.go:63-86` spawns fire-and-forget goroutines from `context.Background()`.

Why it matters:
- Worker crashes lose jobs.
- API failures can still leave leaked state behind.
- Shutdown is not graceful for active jobs.

Recommendation:
- Use ack/retry semantics, visibility timeouts, and dead-letter handling.
- Make instance creation plus enqueue atomic, or add compensating rollback.
- Track active worker goroutines and respect shutdown context.

### 7. Cron jobs can double-execute across workers

Severity: High

What is wrong:
- `internal/repositories/postgres/cron_repo.go:74-80` uses `FOR UPDATE SKIP LOCKED` outside a transaction.
- `internal/core/services/cron_worker.go:52-62,90-115` fetches due jobs, runs them, and updates next execution later.

Why it matters:
- In multi-worker mode, the same due job can be observed and executed more than once.
- Any non-idempotent target becomes dangerous.

Recommendation:
- Claim jobs transactionally.
- Advance state before execution or maintain a durable in-flight claim.

## Important Structural Issues

### 8. Migration strategy is risky

Severity: Medium

What is wrong:
- `internal/repositories/postgres/migrator.go:20-90` re-runs every `.up.sql` on every startup.
- There is no migration history table in the migrator flow itself.

Why it matters:
- This only works if every migration is replay-safe forever.
- The first non-idempotent migration will break deploys.

Recommendation:
- Track applied versions explicitly.
- Use a standard migration table and one-way execution model.

### 9. Tests hide production reality in key areas

Severity: Medium

What is wrong:
- Postgres integration setup disables FK enforcement: `internal/repositories/postgres/test_helpers.go:66-72`.
- Some integration-style tests bypass actual worker execution paths.

Why it matters:
- Real production failures can be invisible in tests.
- Multi-step flows look healthier than they are.

Recommendation:
- Keep relational integrity enabled in realistic integration tests.
- Add end-to-end tests that go through queue and worker paths.

### 10. Multi-tenant model is inconsistent across services

Severity: Medium

What is wrong:
- Pipelines are user-owned rather than tenant-owned: `internal/core/domain/pipeline.go:48-74`, `internal/repositories/postgres/pipeline_repo.go:26-75`.
- Queues are user-owned rather than tenant-owned: `internal/core/domain/queue.go:20-32`, `internal/repositories/postgres/queue_repo.go:27-52`.
- Caches are still keyed and listed by user in the repository: `internal/repositories/postgres/cache_repo.go:24-35,42-73`.

Why it matters:
- Shared-tenant collaboration becomes inconsistent.
- Resource visibility rules differ by module.
- The system story says multi-tenant, but the data model is mixed.

Recommendation:
- Audit every resource type and choose one consistent ownership model.
- Prefer tenant ownership with user attribution where collaboration is intended.

### 11. The architecture is cleaner than average, but not as clean as claimed

Severity: Medium

What is wrong:
- The codebase is organized around domain, ports, services, handlers, and repositories.
- But key core services already depend on transport/platform concerns.
- Example: `internal/core/services/event.go:13-15` imports websocket handler code directly.
- Large files like `internal/core/services/instance.go`, `internal/core/services/storage.go`, `internal/api/setup/router.go`, and `internal/api/setup/dependencies.go` have become concentration points.

Why it matters:
- The "hexagonal" claim is directionally true, but the core is already drifting toward tighter coupling.
- Complex files increase the chance of subtle regressions.

Recommendation:
- Put realtime publishing behind a port.
- Split orchestration-heavy services into smaller units.
- Keep platform and transport concerns out of business services where practical.

## Product Gaps

### 12. The frontend is behind the backend

Severity: Medium

What is missing:
- The backend is the real product today.
- The frontend app contains mostly static or demo-style pages.
- Example: `web/app/(app)/dashboard/page.tsx` is mostly hardcoded UI.
- I found little real API integration in the frontend code.

Why it matters:
- The console does not yet reflect the maturity of the backend.
- The product experience undersells the platform.

Recommendation:
- Decide whether the console is a real operator interface or a marketing/demo shell.
- If real, wire it to live API and tenant-aware data first.

### 13. Documentation drift exists

Severity: Low

Examples:
- `doc.go:54-58` still talks about JWT-based authentication, but the running system is API-key based.
- `README.md:99-108` describes Next.js 14, while `web/package.json:17-19` is Next 16 / React 19.

Why it matters:
- The docs are helpful, but not always authoritative.

Recommendation:
- Treat source as truth and tighten the docs update workflow.

## What Is Missing Most

These are the areas that matter more than adding more features:

1. Trustworthy credential lifecycle
2. Clean tenant isolation across all services
3. Durable async orchestration
4. Safe realtime event delivery
5. Production-grade migration discipline
6. Honest integration testing
7. A frontend that actually exercises the platform

## Suggested Priority Order

### P0

1. Fix websocket hub wiring and tenant/user scoping
2. Hash API keys at rest and stop returning plaintext from list APIs
3. Invalidate key cache on revoke and rotate
4. Repair presigned URL signing and auth model
5. Fix queue durability and instance enqueue rollback behavior

### P1

1. Make cron execution transactional and single-claim
2. Normalize tenant ownership across newer services
3. Replace replay-on-start migration execution with version tracking
4. Remove unrealistic test shortcuts that hide relational and worker failures

### P2

1. Refactor large orchestration services
2. Put realtime publishing behind a proper port
3. Advance the frontend from static UI to real control plane surface
4. Tighten docs so they track source behavior more closely

## Final Assessment

This is a strong and ambitious codebase with a lot of real work in it.

But the gap between "feature breadth" and "operational trust" is still large.

The next stage of improvement should not be adding more cloud services first. It should be hardening the services that already exist so the platform behaves like a real control plane under failure, concurrency, and multi-tenant use.
