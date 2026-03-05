-- +goose Up
ALTER TABLE objects ADD COLUMN IF NOT EXISTS checksum VARCHAR(64);
