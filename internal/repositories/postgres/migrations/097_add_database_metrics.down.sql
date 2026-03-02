-- +goose Down
ALTER TABLE databases DROP COLUMN exporter_container_id;
ALTER TABLE databases DROP COLUMN metrics_port;
ALTER TABLE databases DROP COLUMN metrics_enabled;
