-- +goose Up

CREATE TABLE IF NOT EXISTS pipeline_webhook_deliveries (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    pipeline_id UUID NOT NULL REFERENCES pipelines(id) ON DELETE CASCADE,
    provider VARCHAR(32) NOT NULL,
    event VARCHAR(128) NOT NULL,
    delivery_id VARCHAR(255) NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE(provider, delivery_id)
);

CREATE INDEX IF NOT EXISTS idx_pipeline_webhook_deliveries_pipeline_id ON pipeline_webhook_deliveries(pipeline_id);
