-- +goose Up

-- Migration: 009_create_metrics_history.up.sql
-- Purpose: Store historical metrics for dashboard charts

CREATE TABLE IF NOT EXISTS metrics_history (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    instance_id UUID NOT NULL REFERENCES instances(id) ON DELETE CASCADE,
    cpu_percent DECIMAL(5,2) NOT NULL,
    memory_bytes BIGINT NOT NULL,
    memory_limit_bytes BIGINT NOT NULL DEFAULT 0,
    network_rx_bytes BIGINT NOT NULL DEFAULT 0,
    network_tx_bytes BIGINT NOT NULL DEFAULT 0,
    recorded_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Index for time-series queries (recent metrics first)
CREATE INDEX idx_metrics_instance_time ON metrics_history(instance_id, recorded_at DESC);

-- Index for cleanup queries (delete old metrics)
CREATE INDEX idx_metrics_recorded_at ON metrics_history(recorded_at);

-- Partitioning hint: In production, consider partitioning by recorded_at
COMMENT ON TABLE metrics_history IS 'Stores time-series metrics for dashboard visualization';
