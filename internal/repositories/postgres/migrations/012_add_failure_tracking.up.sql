-- +goose Up

-- Migration: 012_add_failure_tracking.up.sql

ALTER TABLE scaling_groups ADD COLUMN failure_count INT DEFAULT 0;
ALTER TABLE scaling_groups ADD COLUMN last_failure_at TIMESTAMPTZ;
