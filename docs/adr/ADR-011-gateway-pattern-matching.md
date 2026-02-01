# ADR-011: CloudGateway Advanced Path Matching

## Status
Accepted

## Context
CloudGateway initially supported only simple prefix-based routing (e.g., `/api` matches `/api/*`). As the platform grows, we need more sophisticated routing capabilities to support:
- RESTful API patterns (e.g., `/users/{id}`)
- Multi-parameter extraction (e.g., `/orgs/{oid}/projects/{pid}`)
- Regex-constrained parameters (e.g., `/id/{id:[0-9]+}`)
- File extension patterns (e.g., `/assets/*.{ext}`)

The core requirement is to match these patterns dynamically and extract the variable parts to make them available to downstream services or for internal business logic.

## Decison
We will implement an **Advanced Path Matching Engine** using a custom regex-based compiler.

1. **Pattern Syntax**:
   - `{name}`: Matches any non-slash character (`[^/]+`).
   - `{name:regex}`: Matches based on a custom regular expression.
   - `*`: Greedy wildcard matching ( `.*`).
   
2. **Pre-compiled Matchers**:
   - Routes will be compiled into `*regexp.Regexp` instances at the time of route creation or during the periodic `RefreshRoutes` cycle.
   - We will avoid per-request regex compilation to maintain sub-millisecond routing latency.

3. **Routing Priority (Specificity Scoring)**:
   - When multiple routes match a path, we will use a **Specificity Score**.
   - Score = `len(path_pattern)` + `Priority * 1000`.
   - This ensures that more specific patterns (e.g., `/users/me`) win over general ones (e.g., `/users/{id}`), assuming the exact path is longer.

4. **Hexagonal Integration**:
   - **Domain**: `GatewayRoute` will store `PatternType` ("prefix" or "pattern") to remain backward compatible.
   - **Service**: Returns `(proxy, params, found)` to propagate extracted values.
   - **Handler**: Extracted parameters are injected into the Gin context (e.g., `c.Set("path_param_id", value)`), allowing potential header injection in the proxy director.

## Architecture

```
User Request ──▶ Gateway Handler ──▶ Gateway Service (GetProxy)
                                            │
                                            ▼
                                  ┌───────────────────┐
                                  │ Matching Engine   │
                                  ├───────────────────┤
                                  │ 1. Exact Prefix   │
                                  │ 2. Compiled Regex │
                                  │ 3. Score Tie-break│
                                  └─────────┬─────────┘
                                            │
                                            ▼
                                  (ReverseProxy, Params)
```

## Consequences

### Positive
- **Flexibility**: Supports standard RESTful patterns and complex routing requirements.
- **Performance**: Pre-compilation ensures low latency (~0.1ms per match).
- **Backward Compatibility**: Existing prefix routes migrate seamlessly to the new system.
- **Developer Experience**: CLI and SDK now support standard web routing conventions.

### Negative
- **Regex Complexity**: Poorly written custom regexes in patterns could lead to catastrophic backtracking (mitigated by documentation and validation).
- **Database Overhead**: New columns in `gateway_routes` table and JSONB storage for parameter names.
- **Manual Priority**: Users may need to set explicit `priority` for ambiguous overlapping patterns.

## Implementation Notes
- Engine: Custom parser in `internal/routing` using `regexp.QuoteMeta` with surgical unescaping of `{` and `}` tags.
- Scoring: Currently based on pattern length; future iterations might implement a Radix Tree/Trie for O(1) matching if route counts exceed 10k.
- Auth: Path parameters are available *after* the route is identified, allowing for resource-level IAM checks in future phases.
