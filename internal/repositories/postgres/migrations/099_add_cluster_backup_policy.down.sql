-- 099_add_cluster_backup_policy.down.sql
ALTER TABLE clusters DROP COLUMN IF EXISTS backup_schedule;
ALTER TABLE clusters DROP COLUMN IF EXISTS backup_retention_days;
