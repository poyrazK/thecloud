-- Add tenant_id column to clusters table for proper tenant isolation
-- Step 1: Add tenant_id as nullable first (skip if already exists)
DO $$
BEGIN
    IF NOT EXISTS (SELECT FROM information_schema.columns WHERE table_name = 'clusters' AND column_name = 'tenant_id') THEN
        ALTER TABLE clusters ADD COLUMN tenant_id UUID;
    END IF;
END $$;

-- Step 2: Backfill tenant_id from users table for existing clusters
UPDATE clusters c
SET tenant_id = u.default_tenant_id
FROM users u
WHERE c.user_id = u.id AND c.tenant_id IS NULL;

-- Step 3: Set default for any remaining (standalone/system clusters without user)
UPDATE clusters SET tenant_id = '00000000-0000-0000-0000-000000000000' WHERE tenant_id IS NULL;

-- Step 4: Add NOT NULL constraint (skip if already NOT NULL)
DO $$
BEGIN
    IF EXISTS (SELECT FROM information_schema.columns WHERE table_name = 'clusters' AND column_name = 'tenant_id' AND is_nullable = 'YES') THEN
        ALTER TABLE clusters ALTER COLUMN tenant_id SET NOT NULL;
    END IF;
END $$;

-- Step 5: Create proper indexes
DROP INDEX IF EXISTS idx_clusters_user_id;
CREATE INDEX idx_clusters_tenant_id ON clusters(tenant_id);
CREATE INDEX idx_clusters_tenant_user ON clusters(tenant_id, user_id);
