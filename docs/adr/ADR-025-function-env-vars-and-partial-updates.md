# ADR 025: Function Environment Variables and Partial Updates

## Status
Accepted

## Date
2026-04-25

## Context

Functions (serverless) were created with static configuration — once deployed, users could not change runtime parameters (timeout, memory, handler) without deleting and recreating the function. Additionally, there was no way to inject environment variables into function invocations, making it impossible to configure per-deployment settings (API keys, feature flags, connection strings) without baking them into the code zip artifact.

The previous closed PR #170 (`feat/serverless-function-update-and-env-vars`) attempted to solve this but was not merged. This ADR documents the design decisions made in the subsequent implementation.

## Decision

### 1. Partial Updates via Pointer Nil Pattern

Function configuration updates use a `FunctionUpdate` struct where all fields are pointers:

```go
type FunctionUpdate struct {
    Handler   *string    // nil = do not update
    Timeout   *int       // nil = do not update
    MemoryMB  *int       // nil = do not update
    Status    string     // empty string = do not update
    EnvVars   []*EnvVar  // nil = do not update
}
```

**Why pointer nil over zero-values:**
- Allows distinguishing "not provided" from "set to zero/false/empty"
- SQL UPDATE only includes columns where data was explicitly provided
- `SetColumns()` method returns the list of non-nil/non-empty fields for dynamic SQL generation

**Alternative considered: Separate update methods** (`UpdateTimeout`, `UpdateMemory`, etc.) — rejected because it would require multiple round-trips and complicates atomic updates.

### 2. Environment Variables as JSONB

Function environment variables are stored in a PostgreSQL JSONB column (`env_vars`) on the `functions` table:

```sql
ALTER TABLE functions ADD COLUMN env_vars JSONB DEFAULT '{}';
```

**Why JSONB over a separate table:**
- Simple key-value semantics, no joins needed at read time
- JSONB allows indexing individual keys if needed for debugging
- Schema flexibility — easy to add/remove variables without migrations
- Consistent with how other services (instances, databases) store metadata

**Why map[string]string over []EnvVar in DB:**
- Env vars are inherently a map (key uniqueness)
- JSONB map serialization is simpler than array-of-objects
- At read time, unmarshal to `map[string]string`, convert to `[]*EnvVar` for API response

### 3. Env Var Injection at Runtime

Environment variables are injected into the Docker container at invocation time via the container `Env` field:

```go
env := []string{fmt.Sprintf("PAYLOAD=%s", string(payload))}
for _, e := range f.EnvVars {
    env = append(env, e.Key+"="+e.Value)
}
return ports.RunTaskOptions{Env: env, ...}
```

**Why at runtime over baked-in image:**
- No need to rebuild/redeploy to change env vars
- Supports secrets rotation without code changes
- Same function code can be used with different configurations

**Security note:** Env vars in containers are visible to any user who can inspect the container. Plain-text env vars are suitable for non-sensitive configuration. For secrets, use the SecretService integration described below.

### 4. Validation Rules

| Field | Constraint | Reason |
|-------|-----------|--------|
| `Timeout` | 1–900 seconds | Practical bound; prevents runaway invocations |
| `MemoryMB` | 64–10240 MB | Container resource limits; practical upper bound |

Validation occurs in the service layer before calling the repository.

### 5. Secret Integration

Environment variables can reference secrets stored in the SecretService using `@secretname` syntax. Secrets are resolved dynamically at invocation time (not stored in the function record), enabling rotation without code changes.

**EnvVar structure:**
```go
type EnvVar struct {
    Key       string `json:"key"`
    Value     string `json:"value,omitempty"`
    SecretRef string `json:"secret_ref,omitempty"` // e.g. "@my-api-key"
}
```

**Constraint:** `Value` and `SecretRef` are mutually exclusive — at least one must be set. This is validated in `FunctionUpdate.Validate()`.

**Resolution flow:**
1. At update time, the user sets `--env API_KEY=@my-secret` (CLI) or `{"key":"API_KEY","secret_ref":"@my-secret"}` (API)
2. The `SecretRef` is stored in `env_vars` JSONB — no plaintext secret touches the database
3. At invocation, `buildTaskOptions` calls `SecretService.GetSecretByName(ctx, "my-secret")` to resolve the secret
4. The secret value is injected as a JSON object: `API_KEY={"key":"API_KEY","value":"s3cr3t"}`

**Why dynamic resolution:**
- Secrets can be rotated in SecretService without updating the function
- No plaintext secrets in the database
- If secret resolution fails, the env var is skipped (warning logged) rather than failing the invocation

**Alternative considered: Store encrypted secret at update time**
- Would require storing the encryption key alongside the function
- Rotation would still require function update to re-encrypt
- Rejected in favor of runtime resolution for better secret hygiene

## Consequences

### Positive
- Functions can be reconfigured without recreation
- Environment variables enable configuration-as-code patterns
- Partial updates avoid accidentally overwriting fields
- Dynamic SQL ensures only changed columns are written

### Negative
- JSONB column requires careful handling in scan/update logic
- If a function is updated while simultaneously invoked, the new env vars take effect on the next invocation (by design, but worth noting)

### Neutral
- No changes to invocation performance (env var injection is O(n) where n = env var count, typically small)

## Alternatives Considered

### Alternative 1: Separate `function_env_vars` table
**Why rejected:** Overkill for a simple key-value store; adds join complexity for every read.

### Alternative 2: Env vars baked into code zip as a manifest file
**Why rejected:** Requires user to rebuild/rezip code to change config; no support for secrets rotation.

### Alternative 3: Update methods per field
**Why rejected:** Multiple round-trips for partial updates; harder to make atomic.
