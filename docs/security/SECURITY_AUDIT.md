# SQL Injection Security Audit

**Date:** 2025-04
**Updated:** 2026-04-13
**Branch:** fix-sql-injection
**Auditor:** @mamitrkr

## Scope
- `internal/repositories/postgres/` — all repository files
- Target patterns: `fmt.Sprintf` + SQL combination, string concatenation in query construction

## Methodology

1. Automated scan via PowerShell:
```powershell
   Get-ChildItem -Path "internal" -Recurse -Filter "*.go" |
     Select-String -Pattern "fmt\.Sprintf" |
     Where-Object { $_ -match "SELECT|INSERT|UPDATE|DELETE|SCHEMA|search_path" }
```

2. Manual review of dynamic query builders in `log_repo.go` and `cluster_repo.go`

## Findings

### log_repo.go — Dynamic Filter Builder
All dynamic values are parameterized using `$n` placeholders.
String concatenation is used only for static SQL keywords, never for user input.
**Risk: None**

### cluster_repo.go — list() Helper
The `query` parameter is passed as a static SQL literal by all callers.
Arguments are passed separately via `args...`.
**Risk: None**

### Other Repository Files
`accounting`, `audit`, `autoscaling`, `cache`, `dns`, `identity`,
`instance`, `storage`, `vpc` and others — all consistently use
pgx parameterized queries (`$1`, `$2`, ...).
**Risk: None**

## Additional Review (2026-04-13)

### database.go — CLI SQL for credential rotation
`ALTER USER` command strings included user-controlled values.
Escaping was added for Postgres identifiers/literals and MySQL literals.
**Risk: Mitigated**

### test_helpers.go — schema name in test SQL
Schema name was inserted into SQL via `fmt.Sprintf`.
Whitelist validation was added for schema names before use.
**Risk: Mitigated (test-only)**

## Test Notes (2026-04-13)
- `go test ./internal/repositories/postgres` — passed
- `go test ./internal/core/services` — failed on Windows due to Testcontainers (rootless Docker not supported; Docker host detection panics)

## Conclusion
No SQL injection vulnerabilities found in production code.
Parameterized query usage is consistent across the entire repository layer.