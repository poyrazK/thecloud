-- +goose Up

ALTER TABLE instances ADD COLUMN container_id TEXT;
