# Changelog

All notable changes to The Cloud project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Added
- **Integrated Managed Kubernetes (KaaS)**: Launched production-ready Kubernetes cluster management.
- **High-Availability (HA) Control Plane**: Supported 3-node HA control plane with automated API Server Load Balancers.
- **Asynchronous Durable Operations**: Introduced a Redis-backed **Task Queue** and **Cluster Worker** for long-running operations (Provision/Delete/Upgrade).
- **Code Quality Pipeline**: Integrated golangci-lint with comprehensive checks in CI/CD
- **Load Testing Pipeline**: Added k6 performance tests with API server startup
- **SonarQube Analysis**: Static code analysis for quality and security metrics
- **Package Documentation**: Added comprehensive package comments to `setup` package

### Changed
- **Refactored InstanceService**: Introduced `InstanceServiceParams` struct to reduce constructor parameters from 9 to 1
- **Refactored AutoScalingService**: Introduced parameter structs `CreateScalingGroupParams` and `CreateScalingPolicyParams`
- **Reduced Cognitive Complexity**: Extracted helper methods in `InstanceService` and `LibvirtAdapter`
  - `resolveNetworkConfig()` - handles VPC/Subnet resolution and IP allocation
  - `resolveVolumes()` - handles volume retrieval and validation
  - `plumbNetwork()` - handles OVS veth pair creation
  - `formatContainerName()` - centralizes container naming
  - `findAvailableIP()` - reduces complexity in IP allocation
- **Improved Code Quality**: Eliminated duplicate string literals across test files
- **Enhanced Security**: Addressed potential secret exposure in auth tests

### Fixed
- **CI/CD**: Load test workflow now properly starts API server before running k6 tests
- **Linting**: Fixed golangci-lint staticcheck violations (missing package comments)
- **Linting**: Removed unused constants from `cmd/api/main.go`
- **Test Naming**: Standardized test function naming conventions

### Removed
- **sonarqube-mcp-server**: Deleted 297 files of third-party SonarQube MCP server
  - Reduced repository size and complexity
  - SonarQube analysis still functional via GitHub Actions

## Recent PRs and Commits

### Code Quality Improvements (fix/lint-issues branch - MERGED)
- **Commit**: `117637af` - Refactor: reduce NewInstanceService parameter count
- **Commit**: `cf63a6e5` - Refactor: fix duplicated code in autoscaling_handler.go
- **Commit**: `1df815bc` - Refactor: reduce parameter count in AutoScalingService
- **Commit**: `a693cf1f` - Fix: resolve potential secret exposure in auth tests
- **Commit**: `82410e1e` - Refactor: extract IP finding logic in instance service
- **Commit**: `ac33496e` - Refactor: use duplicate string literals in LB handler test
- **Commit**: `c4752939` - Refactor: reduce cognitive complexity of LaunchInstance
- **Commit**: `3126f0a9` - Refactor: reduce complexity in libvirt adapter
- **Commit**: `1a27275c` - Refactor: optimize rbac service and k6 script linting

### CI/CD Improvements
- **Commit**: `ec4b81f5` - Fix: start API server before running load tests
- **Commit**: `bca24cee` - Fix: resolve golangci-lint staticcheck issues
- **Commit**: `05a197d3` - Chore: remove sonarqube-mcp-server directory

## Testing

### Coverage
- **Overall Coverage**: 52.4%
- **Services**: 58.2%
- **Handlers**: 52.8%
- **Repositories**: 57.5%

### Test Types
- ✅ Unit Tests: Core services, handlers, business logic
- ✅ Integration Tests: Database repositories with real PostgreSQL
- ✅ Load Tests: k6 performance validation
- ✅ Static Analysis: SonarQube code quality checks
- ✅ Linting: golangci-lint with 7+ linters

## Infrastructure

### Compute Backends
- **Docker**: Container-based instances with networking and volumes
- **Libvirt/KVM**: Full VM isolation with QCOW2 storage and NAT

### Networking
- **Open vSwitch**: SDN with VXLAN, VPC isolation, and subnet support
- **Network Features**: Port forwarding, DHCP, NAT, OVS integration

### Observability
- **Metrics**: Prometheus metrics collection
- **Monitoring**: Grafana dashboards
- **Real-time**: WebSocket connections for live updates

## Contributors

Special thanks to all contributors who helped improve code quality and CI/CD:
- Code refactoring and complexity reduction
- Test coverage improvements
- Documentation updates
- CI/CD pipeline enhancements

---

For more details, see the [CI/CD Documentation](docs/CI_CD.md).
