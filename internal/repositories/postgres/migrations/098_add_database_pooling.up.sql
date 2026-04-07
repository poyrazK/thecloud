-- +goose Up
ALTER TABLE databases ADD COLUMN IF NOT EXISTS pooling_enabled BOOLEAN NOT NULL DEFAULT FALSE;
ALTER TABLE databases ADD COLUMN IF NOT EXISTS pooling_port INT CONSTRAINT pooling_port_check CHECK (pooling_port IS NULL OR pooling_port = 0 OR pooling_port BETWEEN 1 AND 65535);
ALTER TABLE databases ADD COLUMN IF NOT EXISTS pooler_container_id VARCHAR(255);

-- +goose Down
ALTER TABLE databases DROP COLUMN IF EXISTS pooler_container_id;
ALTER TABLE databases DROP COLUMN IF EXISTS pooling_port;
ALTER TABLE databases DROP COLUMN IF EXISTS pooling_enabled;
