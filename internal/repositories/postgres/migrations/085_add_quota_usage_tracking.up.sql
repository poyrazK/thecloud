ALTER TABLE tenant_quotas
    ADD COLUMN used_vcpus INTEGER NOT NULL DEFAULT 0,
    ADD COLUMN used_memory_gb INTEGER NOT NULL DEFAULT 0;

-- Optional: Backfill existing usage if needed, but for now starting at 0 is safe as we will be rigorous going forward
-- or we can do a complex update from instances join instance_types if those tables were fully linked in SQL.
