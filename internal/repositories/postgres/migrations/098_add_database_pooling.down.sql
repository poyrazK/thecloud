-- +goose Down
ALTER TABLE databases DROP COLUMN pooler_container_id;
ALTER TABLE databases DROP COLUMN pooling_port;
ALTER TABLE databases DROP COLUMN pooling_enabled;
