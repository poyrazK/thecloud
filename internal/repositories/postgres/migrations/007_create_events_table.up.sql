-- +goose Up

CREATE TABLE IF NOT EXISTS events (
    id UUID PRIMARY KEY,
    action VARCHAR(50) NOT NULL,
    resource_id VARCHAR(50) NOT NULL,
    resource_type VARCHAR(50) NOT NULL,
    metadata JSONB NOT NULL DEFAULT '{}'::jsonb,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_events_created_at ON events(created_at DESC);
