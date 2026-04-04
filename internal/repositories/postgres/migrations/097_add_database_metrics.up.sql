-- +goose Up
ALTER TABLE databases ADD COLUMN IF NOT EXISTS metrics_enabled BOOLEAN NOT NULL DEFAULT FALSE;
ALTER TABLE databases ADD COLUMN IF NOT EXISTS metrics_port INT;
ALTER TABLE databases ADD COLUMN IF NOT EXISTS exporter_container_id VARCHAR(255);

-- +goose Down
ALTER TABLE databases DROP COLUMN IF EXISTS exporter_container_id;
ALTER TABLE databases DROP COLUMN IF EXISTS metrics_port;
ALTER TABLE databases DROP COLUMN IF EXISTS metrics_enabled;
