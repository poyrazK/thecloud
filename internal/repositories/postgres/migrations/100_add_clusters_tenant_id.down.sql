-- Remove tenant_id column from clusters table
ALTER TABLE clusters DROP COLUMN tenant_id;

-- Restore original index
DROP INDEX IF EXISTS idx_clusters_tenant_id;
DROP INDEX IF EXISTS idx_clusters_tenant_user;
CREATE INDEX idx_clusters_user_id ON clusters(user_id);
