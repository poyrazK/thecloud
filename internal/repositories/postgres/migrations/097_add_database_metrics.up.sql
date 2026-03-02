-- +goose Up
ALTER TABLE databases ADD COLUMN metrics_enabled BOOLEAN NOT NULL DEFAULT FALSE;
ALTER TABLE databases ADD COLUMN metrics_port INT;
ALTER TABLE databases ADD COLUMN exporter_container_id VARCHAR(255);

-- +goose Down
ALTER TABLE databases DROP COLUMN exporter_container_id;
ALTER TABLE databases DROP COLUMN metrics_port;
ALTER TABLE databases DROP COLUMN metrics_enabled;
