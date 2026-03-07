-- 099_add_cluster_backup_policy.up.sql
ALTER TABLE clusters ADD COLUMN backup_schedule VARCHAR(255);
ALTER TABLE clusters ADD COLUMN backup_retention_days INTEGER DEFAULT 7;
