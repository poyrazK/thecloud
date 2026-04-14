# Load Tests

This directory contains k6 load tests for thecloud API.

## Running Tests

### Prerequisites
- k6 installed: `brew install k6` (macOS) or `sudo apt install k6` (Linux)
- API server running on localhost:8080 (or set `BASE_URL` env var)
- PostgreSQL and Redis services (via docker-compose or external)

### Quick Start

```bash
# Run all quick tests
k6 run tests/load/api-smoke.js

# Run with environment variables
BASE_URL=http://localhost:8080 k6 run tests/load/api-smoke.js
```

## Test Categories

### API Quick Tests (~3-4 min)
Fast tests with no infrastructure dependency:
- `api-smoke.js` - Basic API health and auth checks
- `error-paths-unauth.js` - Unauthenticated access rejection
- `error-paths-validation.js` - Input validation errors
- `rate-limit.js` - Rate limiting behavior

### API Storage Tests (~3 min)
Tests for object storage:
- `storage.js` - Bucket and object CRUD operations

### API Slow Infrastructure Tests (~40 min)
Tests requiring Docker compute backend:
- `real-world-lifecycle.js` - VPC → Subnet → Instance lifecycle
- `database-lifecycle.js` - Database provisioning and operations
- `load-balancer.js` - Load balancer with instance targets
- `volumes-lifecycle.js` - Block storage attach/detach
- `security-groups.js` - Security group with rules
- `caches-lifecycle.js` - Redis cache lifecycle
- `secrets-lifecycle.js` - Secrets CRUD
- `clusters-lifecycle.js` - Kubernetes cluster lifecycle

## Infrastructure Requirements

| Test Category | PostgreSQL | Redis | Docker/OVS |
|--------------|------------|-------|------------|
| API Quick | ✓ | ✓ | - |
| API Storage | ✓ | ✓ | - |
| API Slow | ✓ | ✓ | ✓ |

### OVS Requirement

Some tests require OpenVSwitch (OVS) for networking. OVS requires:
- Linux kernel modules (`openvswitch.ko`)
- Running `openvswitch-switch` service

**Without OVS:** Infrastructure tests will fail at VPC creation (OVS bridges are needed for VPC networking).

**Installation (Ubuntu/Debian):**
```bash
sudo apt-get install openvswitch-switch
sudo ovs-vsctl init
sudo service openvswitch-switch start
```

## CI/CD

Tests run automatically on PR via `.github/workflows/load-tests.yml`:

- **API Quick Tests** - Runs on every PR
- **API Storage Tests** - Runs on every PR
- **API Slow Infrastructure Tests** - Runs when `tests/load/**` changes

## Adding New Tests

See existing tests for patterns:
1. Import `BASE_URL`, thresholds from `./common/config.js`
2. Use `getOrCreateApiKey()` for cached authentication
3. Use `check()` for assertions
4. Use `fail('message')` when infrastructure doesn't reach ready state
5. Always clean up resources (VPCs, instances, etc.)

## Troubleshooting

### "Instance never reached running state"
- OVS not installed or kernel module not loaded
- Docker under heavy load - increase timeout in test

### "VPC Creation Failed: failed to create OVS bridge"
- OVS daemon not running
- No kernel module access (e.g., Docker Desktop on macOS)

### Rate limit errors
- `RATE_LIMIT_AUTH` env var set too low
- Run tests with unique API keys per VU

## Test Results Interpretation

k6 thresholds only fail on `http_req_failed` (HTTP errors). Check failures (like "instance not running") don't cause test failure unless `fail()` is called.

This means a test can "pass" even if infrastructure doesn't start - look at check success rates in logs.
