-- Remove tenant_id column from clusters table (skip if already removed)
-- Acquire advisory lock to prevent concurrent migration runs
PERFORM pg_advisory_lock(hashtext('100_add_clusters_tenant_id')::bigint);

DO $$
BEGIN
    IF EXISTS (SELECT FROM information_schema.columns WHERE table_name = 'clusters' AND column_name = 'tenant_id') THEN
        ALTER TABLE clusters DROP COLUMN tenant_id;
    END IF;
END $$;

-- Restore original index (only if not exists)
DROP INDEX IF EXISTS idx_clusters_tenant_id;
DROP INDEX IF EXISTS idx_clusters_tenant_user;
CREATE INDEX IF NOT EXISTS idx_clusters_user_id ON clusters(user_id);

-- Release advisory lock
PERFORM pg_advisory_unlock(hashtext('100_add_clusters_tenant_id')::bigint);