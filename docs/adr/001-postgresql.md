# ADR-001: Use PostgreSQL as Primary Datastore

## Status
Accepted

## Context
We need a reliable, persistent datastore for managing cloud resources (instances, users, etc.).

## Decision
We using PostgreSQL (managed via Docker Compose) with the `pgx` driver.

## Consequences
- **Positive**: Industry standard, strong ACID compliance, excellent Go support via pgx.
- **Negative**: Requires Docker to be running locally.
