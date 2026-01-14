# Observability & Monitoring

The Cloud provides production-grade observability out of the box using Prometheus and Grafana.

## Architecture

1.  **Prometheus**: Scrapes metrics from the API at 30-second intervals.
2.  **Grafana**: Provides visual dashboards for system health and performance.
3.  **Metrics Middleware**: Automatically captures HTTP request rates, status codes, and latency for every endpoint.
4.  **Service Instrumentation**: Core services are instrumented to track compute lifecycle, storage operations, and authentication attempts.

## Metrics Reference

All metrics use the `thecloud_` prefix for consistency.

### API Metrics
- `thecloud_http_requests_total`: Total number of HTTP requests (labels: `method`, `path`, `status`).
- `thecloud_http_request_duration_seconds`: Response latency histogram (labels: `method`, `path`).

### Compute Metrics
- `thecloud_instances_total`: Current instance count (labels: `status`, `backend`).
- `thecloud_instance_operations_total`: Count of launch/stop/terminate actions.

### Storage Metrics
- `thecloud_volumes_total`: Total volumes by status.
- `thecloud_volume_size_bytes`: Total provisioned block storage capacity.
- `thecloud_storage_operations_total`: object and volume operations.

### Managed Services
- `thecloud_rds_instances_total`: Database instances by engine and status.
- `thecloud_cache_instances_total`: Active cache (Redis) instances.
- `thecloud_queue_messages_total`: Queue operations (send/receive/delete).

### Security Metrics
- `thecloud_auth_attempts_total`: Login success/failure tracking.
- `thecloud_api_keys_active`: Current number of active API keys.

## Accessing Dashboards

1.  Start the monitoring stack:
    ```bash
    docker compose -f docker-compose.yml -f docker-compose.prod-full.yml up -d
    ```
2.  Open Grafana: `http://localhost:3000` (Default: `admin`/`admin`)
3.  Navigate to **Dashboards** -> **The Cloud** folder.

### Available Dashboards
- **The Cloud Overview**: High-level health including request rates, p95 latency, and instance counts.
- **The Cloud Compute**: Deep dive into instance lifecycle, scaling group activity, and backend distribution.

## Alerting

Alerts are defined in `prometheus/alerts/api-alerts.yml`. Default alerts include:
- **HighAPIErrorRate**: Triggers if 5xx errors exceed 10% for 2 minutes.
- **HighAPILatency**: Triggers if p95 latency exceeds 2 seconds for 5 minutes.
- **InstanceErrorStatus**: Triggers if any instance enters an `error` state.

## Distributed Tracing

The Cloud supports **Distributed Tracing** via OpenTelemetry and Jaeger to visualize request flows across microservices and databases.

### Enabling Tracing

Tracing is **opt-in** to avoid performance overhead. Enable it via environment variables:

```bash
TRACING_ENABLED=true
JAEGER_ENDPOINT=http://localhost:4318  # Default
```

### Viewing Traces

1. Ensure the Jaeger service is running (included in `docker-compose.yml`).
2. Open the Jaeger UI at [http://localhost:16686](http://localhost:16686).
3. Select the service `thecloud-api` to view traces.

### Features
- **Full Trace Context**: Visualizes the entire lifecycle of a request (API -> Service -> Database/Docker).
- **Database Tracing**: Automatically instrumented PostgreSQL queries with `otelpgx`.
- **Rich Metadata**: Spans include relevant attributes like `instance.id`, `user.id`, and `db.statement`.
