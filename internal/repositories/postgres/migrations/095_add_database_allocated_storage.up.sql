-- +goose Up
ALTER TABLE databases ADD COLUMN allocated_storage INT NOT NULL DEFAULT 10;
