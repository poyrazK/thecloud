# ADR-009: Distributed Tracing with OpenTelemetry and Jaeger

## Status
**Proposed**

## Context
As The Cloud grows with 30+ services, debugging performance issues and understanding request flows becomes increasingly difficult. When a user reports "instance launch is slow," we currently have no visibility into which component (VPC lookup, Docker pull, database insert) is the bottleneck.

**Current State:**
- Structured logging exists but doesn't correlate across service boundaries
- Prometheus metrics show aggregates but not per-request details
- No way to trace a single request through the entire system

## Decision
We will implement **distributed tracing** using:
- **OpenTelemetry SDK** - Industry-standard, CNCF-graduated instrumentation library
- **Jaeger** - Open-source tracing backend with excellent UI
- **OTLP Protocol** - Modern, vendor-neutral trace export format

### Why OpenTelemetry?
| Option | Pros | Cons |
|--------|------|------|
| OpenTelemetry | Standard, multi-vendor, active community | Slightly more setup |
| Jaeger SDK (legacy) | Simple | Deprecated, vendor lock-in |
| Datadog/NewRelic | Easy | Proprietary, costly |

### Why Jaeger?
- Open-source, self-hosted (fits our philosophy)
- Beautiful UI for trace visualization
- Used by Uber, Netflix, and major companies
- Easy Docker deployment

## Implementation Approach
1. Create `pkg/tracing` package for initialization
2. Add Gin middleware for automatic HTTP span creation
3. Manually instrument key service methods with spans
4. Add Jaeger to docker-compose for local dev
5. Make tracing opt-in via `TRACING_ENABLED` env var

## Consequences

### Positive
- **Visibility**: See exactly where time is spent in each request
- **Debugging**: Quickly identify failing components
- **Learning**: Deep understanding of OpenTelemetry (valuable skill)
- **Portfolio**: Demonstrates production-grade observability

### Negative
- **Overhead**: Small performance cost (~1-2% latency)
- **Complexity**: More code to maintain
- **Storage**: Jaeger needs disk space for traces

### Risks
- **Cardinality explosion** if we add too many span attributes
- **Security**: Traces may contain sensitive data (mitigate with sampling)

## References
- [OpenTelemetry Go SDK](https://opentelemetry.io/docs/instrumentation/go/)
- [Jaeger Documentation](https://www.jaegertracing.io/docs/)
- [Google Dapper Paper](https://research.google/pubs/pub36356/) (original tracing paper)
