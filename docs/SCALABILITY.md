# System Scalability Architecture

This document describes the high-scalability architecture implemented to support 1000+ concurrent users with low latency.

## Architecture Overview

The system employs a multi-layered scaling strategy covering caching, asynchronous processing, horizontal scaling, and database optimizations.

### 1. API Key Caching (Redis)
- **Problem:** Database CPU saturation due to validiting API keys on every request.
- **Solution:** `IdentityService` is decorated with a Redis-backed cache layer.
- **Mechanism:**
  - Valid API keys are cached in Redis with a TTL (e.g., 5 minutes).
  - Subsequent requests hit Redis (sub-millisecond latency) instead of PostgreSQL.
  - Cache invalidation occurs on key rotation or revocation.

### 2. Asynchronous Instance Provisioning
- **Problem:** `POST /instances` requests blocked for 500ms-2s while waiting for Docker/OVS operations, leading to worker pool exhaustion under load.
- **Solution:** Decoupled request acceptance from processing logic using a Redis Task Queue.
- **Mechanism:**
  - **API Layer:** `POST /instances` enqueues a `ProvisionJob` to Redis and returns `202 Accepted` immediately with a `job_id`.
  - **Worker Layer:** Detailed `ProvisionWorker` consumes jobs and executes the heavy lifting (container creation, network plumbing).
  - **Status Polling:** Clients poll `/instances/{id}` to check for `status: "running"`.

### 3. Horizontal API Scaling
- **Problem:** Single API container becomes a CPU bottleneck for request parsing/routing.
- **Solution:** Run multiple stateless API replicas behind a load balancer.
- **Mechanism:**
  - **Nginx Load Balancer:** Distributes incoming traffic across API replicas using round-robin.
  - **Replica Set:** Configurable via `docker-compose.scale.yml`. Default is 3 replicas.
  - **Statelessness:** No session state stored in API memory; all state is in Redis or Postgres.

### 4. Database Optimization
- **Problem:** Connection limits and read contention on the primary database node.
- **Solution:** Connection pooling and Read/Write splitting.
- **Mechanism:**
  - **Pooling:** `pgxpool` manages efficient connection reuse with configurable `DB_MAX_CONNS`.
  - **DualDB:** Custom DB implementation routes `Exec` (writes) to Primary and `Query` (reads) to a Read Replica if `DATABASE_READ_URL` is configured.
  - **Resilience:** Automatic fallback to Primary for reads if Replica is unavailable.

## Configuration

### Environment Variables
| Variable | Description | Default |
|----------|-------------|---------|
| `REDIS_URL` | Redis connection string | `redis:6379` |
| `DB_MAX_CONNS` | Max allowed DB connections per replica | `20` |
| `DATABASE_READ_URL` | Optional Read Replica URL | (empty) |
| `APP_ENV` | Environment mode | `development` |

### Running in Scalable Mode

To run with full scalability stack:

```bash
docker-compose -f docker-compose.yml -f docker-compose.scale.yml up -d --scale api=3
```
