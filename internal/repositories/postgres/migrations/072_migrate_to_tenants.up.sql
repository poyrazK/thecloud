-- 1. Create personal tenants for existing users
-- Use ON CONFLICT if slug is already there (though gen_random_uuid ensures uniqueness of id)
INSERT INTO tenants (id, name, slug, owner_id, plan, status, created_at, updated_at)
SELECT 
    gen_random_uuid(), 
    name || '''s Personal Tenant', 
    -- Normalize slug: lowercase, replace non-alphanumeric with hyphens, trim, append unique ID suffix
    'personal-' || trim(both '-' from regexp_replace(lower(name), '[^a-z0-9]+', '-', 'g')) || '-' || SUBSTR(id::text, 1, 8),
    id, 
    'free', 
    'active', 
    NOW(), 
    NOW()
FROM users
WHERE default_tenant_id IS NULL
ON CONFLICT (slug) DO NOTHING;

-- 2. Add users as members of their new tenants
INSERT INTO tenant_members (tenant_id, user_id, role, joined_at)
SELECT id, owner_id, 'owner', NOW()
FROM tenants
ON CONFLICT (tenant_id, user_id) DO NOTHING;

-- 3. Update users table with default_tenant_id
UPDATE users u
SET default_tenant_id = t.id
FROM tenants t
WHERE u.id = t.owner_id AND u.default_tenant_id IS NULL;

-- 4. Update all resources to match their owner's personal tenant (only if tenant_id is null)
DO $$ 
BEGIN 
    IF EXISTS (SELECT FROM pg_tables WHERE tablename = 'instances') THEN 
        UPDATE instances i SET tenant_id = u.default_tenant_id FROM users u WHERE i.user_id = u.id AND i.tenant_id IS NULL;
    END IF;
    IF EXISTS (SELECT FROM pg_tables WHERE tablename = 'vpcs') THEN 
        UPDATE vpcs v SET tenant_id = u.default_tenant_id FROM users u WHERE v.user_id = u.id AND v.tenant_id IS NULL;
    END IF;
    IF EXISTS (SELECT FROM pg_tables WHERE tablename = 'volumes') THEN 
        UPDATE volumes vol SET tenant_id = u.default_tenant_id FROM users u WHERE vol.user_id = u.id AND vol.tenant_id IS NULL;
    END IF;
    IF EXISTS (SELECT FROM pg_tables WHERE tablename = 'load_balancers') THEN 
        UPDATE load_balancers lb SET tenant_id = u.default_tenant_id FROM users u WHERE lb.user_id = u.id AND lb.tenant_id IS NULL;
    END IF;
    IF EXISTS (SELECT FROM pg_tables WHERE tablename = 'scaling_groups') THEN 
        UPDATE scaling_groups sg SET tenant_id = u.default_tenant_id FROM users u WHERE sg.user_id = u.id AND sg.tenant_id IS NULL;
    END IF;
    IF EXISTS (SELECT FROM pg_tables WHERE tablename = 'buckets') THEN 
        UPDATE buckets b SET tenant_id = u.default_tenant_id FROM users u WHERE b.user_id = u.id AND b.tenant_id IS NULL;
    END IF;
    IF EXISTS (SELECT FROM pg_tables WHERE tablename = 'clusters') THEN 
        UPDATE clusters c SET tenant_id = u.default_tenant_id FROM users u WHERE c.user_id = u.id AND c.tenant_id IS NULL;
    END IF;
    IF EXISTS (SELECT FROM pg_tables WHERE tablename = 'security_groups') THEN 
        UPDATE security_groups s SET tenant_id = u.default_tenant_id FROM users u WHERE s.user_id = u.id AND s.tenant_id IS NULL;
    END IF;
    IF EXISTS (SELECT FROM pg_tables WHERE tablename = 'subnets') THEN 
        UPDATE subnets sub SET tenant_id = u.default_tenant_id FROM users u WHERE sub.user_id = u.id AND sub.tenant_id IS NULL;
    END IF;
    IF EXISTS (SELECT FROM pg_tables WHERE tablename = 'snapshots') THEN 
        UPDATE snapshots snap SET tenant_id = u.default_tenant_id FROM users u WHERE snap.user_id = u.id AND snap.tenant_id IS NULL;
    END IF;
    IF EXISTS (SELECT FROM pg_tables WHERE tablename = 'secrets') THEN 
        UPDATE secrets sec SET tenant_id = u.default_tenant_id FROM users u WHERE sec.user_id = u.id AND sec.tenant_id IS NULL;
    END IF;
    IF EXISTS (SELECT FROM pg_tables WHERE tablename = 'functions') THEN 
        UPDATE functions func SET tenant_id = u.default_tenant_id FROM users u WHERE func.user_id = u.id AND func.tenant_id IS NULL;
    END IF;
    IF EXISTS (SELECT FROM pg_tables WHERE tablename = 'caches') THEN 
        UPDATE caches c SET tenant_id = u.default_tenant_id FROM users u WHERE c.user_id = u.id AND c.tenant_id IS NULL;
    END IF;
    IF EXISTS (SELECT FROM pg_tables WHERE tablename = 'queues') THEN 
        UPDATE queues q SET tenant_id = u.default_tenant_id FROM users u WHERE q.user_id = u.id AND q.tenant_id IS NULL;
    END IF;
    IF EXISTS (SELECT FROM pg_tables WHERE tablename = 'topics') THEN 
        UPDATE topics t SET tenant_id = u.default_tenant_id FROM users u WHERE t.user_id = u.id AND t.tenant_id IS NULL;
    END IF;
    IF EXISTS (SELECT FROM pg_tables WHERE tablename = 'subscriptions') THEN 
        UPDATE subscriptions s SET tenant_id = u.default_tenant_id FROM users u WHERE s.user_id = u.id AND s.tenant_id IS NULL;
    END IF;
    IF EXISTS (SELECT FROM pg_tables WHERE tablename = 'cron_jobs') THEN 
        UPDATE cron_jobs c SET tenant_id = u.default_tenant_id FROM users u WHERE c.user_id = u.id AND c.tenant_id IS NULL;
    END IF;
    IF EXISTS (SELECT FROM pg_tables WHERE tablename = 'gateway_routes') THEN 
        UPDATE gateway_routes g SET tenant_id = u.default_tenant_id FROM users u WHERE g.user_id = u.id AND g.tenant_id IS NULL;
    END IF;
    IF EXISTS (SELECT FROM pg_tables WHERE tablename = 'deployments') THEN 
        UPDATE deployments d SET tenant_id = u.default_tenant_id FROM users u WHERE d.user_id = u.id AND d.tenant_id IS NULL;
    END IF;
    IF EXISTS (SELECT FROM pg_tables WHERE tablename = 'stacks') THEN 
        UPDATE stacks s SET tenant_id = u.default_tenant_id FROM users u WHERE s.user_id = u.id AND s.tenant_id IS NULL;
    END IF;
    IF EXISTS (SELECT FROM pg_tables WHERE tablename = 'images') THEN 
        UPDATE images img SET tenant_id = u.default_tenant_id FROM users u WHERE img.user_id = u.id AND img.tenant_id IS NULL;
    END IF;
END $$;

-- 5. Add default quotas for all tenants
INSERT INTO tenant_quotas (tenant_id, max_instances, max_vpcs, max_storage_gb, max_memory_gb, max_vcpus)
SELECT id, 10, 3, 100, 32, 16 FROM tenants
ON CONFLICT (tenant_id) DO NOTHING;

-- 6. Update API keys with default_tenant_id
UPDATE api_keys ak
SET default_tenant_id = u.default_tenant_id
FROM users u
WHERE ak.user_id = u.id AND ak.default_tenant_id IS NULL;
