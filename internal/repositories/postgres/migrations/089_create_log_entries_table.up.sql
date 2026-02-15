CREATE TABLE IF NOT EXISTS log_entries (
    id UUID PRIMARY KEY,
    tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    resource_id TEXT NOT NULL,
    resource_type TEXT NOT NULL,
    level TEXT NOT NULL,
    message TEXT NOT NULL,
    timestamp TIMESTAMPTZ NOT NULL,
    trace_id TEXT,
    created_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE INDEX idx_logs_tenant_resource ON log_entries(tenant_id, resource_id);
CREATE INDEX idx_logs_timestamp ON log_entries(timestamp);
