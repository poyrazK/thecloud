-- 099_add_cluster_backup_policy.up.sql
ALTER TABLE clusters ADD COLUMN backup_schedule VARCHAR(255) NOT NULL DEFAULT '';
ALTER TABLE clusters ADD COLUMN backup_retention_days INTEGER NOT NULL DEFAULT 7;
