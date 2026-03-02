-- +goose Up
ALTER TABLE databases ADD COLUMN pooling_enabled BOOLEAN NOT NULL DEFAULT FALSE;
ALTER TABLE databases ADD COLUMN pooling_port INT CONSTRAINT pooling_port_check CHECK (pooling_port BETWEEN 1 AND 65535);
ALTER TABLE databases ADD COLUMN pooler_container_id VARCHAR(255);
