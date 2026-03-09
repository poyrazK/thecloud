-- +goose Up
ALTER TABLE databases ADD COLUMN parameters JSONB DEFAULT '{}'::jsonb;
