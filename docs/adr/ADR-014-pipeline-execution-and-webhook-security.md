# ADR-014: Pipeline Execution and Webhook Security Model

## Status
Accepted

## Context
The platform needed a built-in CI/CD capability so tenants can run build/lint/test workflows directly in The Cloud, instead of relying only on external CI systems.

Key requirements:
1. Asynchronous execution for long-running jobs.
2. User-isolated step execution with observable status and logs.
3. Webhook-triggered automation from GitHub/GitLab.
4. Replay protection for duplicate webhook deliveries.
5. Minimal coupling to specific VCS providers and task content.

## Decision

### 1. Asynchronous queue + worker model
- Pipeline runs are persisted as `builds` and queued in `pipeline_build_queue`.
- `PipelineWorker` consumes queue jobs and orchestrates state transitions:
  - `QUEUED -> RUNNING -> SUCCEEDED|FAILED`
- This decouples API responsiveness from execution duration.

### 2. Step execution in task containers
- Each pipeline step executes in an isolated task container via `ComputeBackend.RunTask`.
- Steps capture:
  - image
  - command list
  - status
  - exit code
  - start/finish timestamps
- Build logs are persisted and queryable per run.

### 3. Public webhook endpoint with provider verification
- Endpoint: `POST /pipelines/:id/webhook/:provider`
- GitHub validation:
  - HMAC SHA-256 signature via `X-Hub-Signature-256`
- GitLab validation:
  - Shared token via `X-Gitlab-Token`
- Only push events are mapped to build triggers.
- Branch matching is enforced against pipeline branch.

### 4. Delivery idempotency
- New persistence table stores webhook deliveries with provider + delivery ID uniqueness.
- Duplicate deliveries are ignored and return accepted/ignored semantics.

### 5. Trigger semantics
- Manual trigger path remains authenticated (`X-API-Key`).
- Webhook trigger path is unauthenticated but cryptographically/secret validated.
- Webhook-triggered runs are stored with trigger type `WEBHOOK`.

## Consequences

### Positive
- Enables tenant self-service CI/CD workflows inside platform boundaries.
- Scales better than synchronous request-bound execution.
- Improves security posture for webhook automation.
- Prevents duplicate run storms from webhook retries.
- Provides auditable, queryable execution history and logs.

### Negative
- Increased operational complexity (worker lifecycle, queue monitoring).
- Requires careful container image selection for toolchain availability (e.g., Go + linter binaries).
- Public webhook endpoint must remain hardened and observable.

## Alternatives Considered
1. Synchronous API execution
   - Rejected due to timeout risk and poor user experience.
2. Polling Git provider APIs instead of webhooks
   - Rejected due to latency, API cost, and complexity.
3. In-memory duplicate suppression
   - Rejected due to non-durable behavior across restarts/replicas.

## Implementation Notes
- Migration numbering introduced:
  - `093_create_pipelines`
  - `094_create_pipeline_webhook_deliveries`
- Core components:
  - pipeline repository/service/handler
  - pipeline worker
  - router and dependency wiring
- Verified with user-style end-to-end API scenarios including webhook replay behavior.
