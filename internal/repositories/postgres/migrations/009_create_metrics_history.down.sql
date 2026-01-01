-- Migration: 009_create_metrics_history.down.sql
DROP INDEX IF EXISTS idx_metrics_recorded_at;
DROP INDEX IF EXISTS idx_metrics_instance_time;
DROP TABLE IF EXISTS metrics_history;
