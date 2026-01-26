-- Add tenant_id to core resource tables (with existence checks if possible, but here we just list known tables)
DO $$ 
BEGIN 
    IF EXISTS (SELECT FROM pg_tables WHERE schemaname = 'public' AND tablename = 'instances') THEN 
        ALTER TABLE instances ADD COLUMN IF NOT EXISTS tenant_id UUID REFERENCES tenants(id) ON DELETE RESTRICT;
    END IF;
    IF EXISTS (SELECT FROM pg_tables WHERE schemaname = 'public' AND tablename = 'vpcs') THEN 
        ALTER TABLE vpcs ADD COLUMN IF NOT EXISTS tenant_id UUID REFERENCES tenants(id) ON DELETE RESTRICT;
    END IF;
    IF EXISTS (SELECT FROM pg_tables WHERE schemaname = 'public' AND tablename = 'volumes') THEN 
        ALTER TABLE volumes ADD COLUMN IF NOT EXISTS tenant_id UUID REFERENCES tenants(id) ON DELETE RESTRICT;
    END IF;
    IF EXISTS (SELECT FROM pg_tables WHERE schemaname = 'public' AND tablename = 'load_balancers') THEN 
        ALTER TABLE load_balancers ADD COLUMN IF NOT EXISTS tenant_id UUID REFERENCES tenants(id) ON DELETE RESTRICT;
    END IF;
    IF EXISTS (SELECT FROM pg_tables WHERE schemaname = 'public' AND tablename = 'scaling_groups') THEN 
        ALTER TABLE scaling_groups ADD COLUMN IF NOT EXISTS tenant_id UUID REFERENCES tenants(id) ON DELETE RESTRICT;
    END IF;
    IF EXISTS (SELECT FROM pg_tables WHERE schemaname = 'public' AND tablename = 'buckets') THEN 
        ALTER TABLE buckets ADD COLUMN IF NOT EXISTS tenant_id UUID REFERENCES tenants(id) ON DELETE RESTRICT;
    END IF;
    IF EXISTS (SELECT FROM pg_tables WHERE schemaname = 'public' AND tablename = 'clusters') THEN 
        ALTER TABLE clusters ADD COLUMN IF NOT EXISTS tenant_id UUID REFERENCES tenants(id) ON DELETE RESTRICT;
    END IF;
    IF EXISTS (SELECT FROM pg_tables WHERE schemaname = 'public' AND tablename = 'security_groups') THEN 
        ALTER TABLE security_groups ADD COLUMN IF NOT EXISTS tenant_id UUID REFERENCES tenants(id) ON DELETE RESTRICT;
    END IF;
    IF EXISTS (SELECT FROM pg_tables WHERE schemaname = 'public' AND tablename = 'subnets') THEN 
        ALTER TABLE subnets ADD COLUMN IF NOT EXISTS tenant_id UUID REFERENCES tenants(id) ON DELETE RESTRICT;
    END IF;
    IF EXISTS (SELECT FROM pg_tables WHERE schemaname = 'public' AND tablename = 'snapshots') THEN 
        ALTER TABLE snapshots ADD COLUMN IF NOT EXISTS tenant_id UUID REFERENCES tenants(id) ON DELETE RESTRICT;
    END IF;
    IF EXISTS (SELECT FROM pg_tables WHERE schemaname = 'public' AND tablename = 'secrets') THEN 
        ALTER TABLE secrets ADD COLUMN IF NOT EXISTS tenant_id UUID REFERENCES tenants(id) ON DELETE RESTRICT;
    END IF;
    IF EXISTS (SELECT FROM pg_tables WHERE schemaname = 'public' AND tablename = 'functions') THEN 
        ALTER TABLE functions ADD COLUMN IF NOT EXISTS tenant_id UUID REFERENCES tenants(id) ON DELETE RESTRICT;
    END IF;
    IF EXISTS (SELECT FROM pg_tables WHERE schemaname = 'public' AND tablename = 'caches') THEN 
        ALTER TABLE caches ADD COLUMN IF NOT EXISTS tenant_id UUID REFERENCES tenants(id) ON DELETE RESTRICT;
    END IF;
    IF EXISTS (SELECT FROM pg_tables WHERE schemaname = 'public' AND tablename = 'queues') THEN 
        ALTER TABLE queues ADD COLUMN IF NOT EXISTS tenant_id UUID REFERENCES tenants(id) ON DELETE RESTRICT;
    END IF;
    IF EXISTS (SELECT FROM pg_tables WHERE schemaname = 'public' AND tablename = 'topics') THEN 
        ALTER TABLE topics ADD COLUMN IF NOT EXISTS tenant_id UUID REFERENCES tenants(id) ON DELETE RESTRICT;
    END IF;
    IF EXISTS (SELECT FROM pg_tables WHERE schemaname = 'public' AND tablename = 'subscriptions') THEN 
        ALTER TABLE subscriptions ADD COLUMN IF NOT EXISTS tenant_id UUID REFERENCES tenants(id) ON DELETE RESTRICT;
    END IF;
    IF EXISTS (SELECT FROM pg_tables WHERE schemaname = 'public' AND tablename = 'cron_jobs') THEN 
        ALTER TABLE cron_jobs ADD COLUMN IF NOT EXISTS tenant_id UUID REFERENCES tenants(id) ON DELETE RESTRICT;
    END IF;
    IF EXISTS (SELECT FROM pg_tables WHERE schemaname = 'public' AND tablename = 'gateway_routes') THEN 
        ALTER TABLE gateway_routes ADD COLUMN IF NOT EXISTS tenant_id UUID REFERENCES tenants(id) ON DELETE RESTRICT;
    END IF;
    IF EXISTS (SELECT FROM pg_tables WHERE schemaname = 'public' AND tablename = 'deployments') THEN 
        ALTER TABLE deployments ADD COLUMN IF NOT EXISTS tenant_id UUID REFERENCES tenants(id) ON DELETE RESTRICT;
    END IF;
    IF EXISTS (SELECT FROM pg_tables WHERE schemaname = 'public' AND tablename = 'stacks') THEN 
        ALTER TABLE stacks ADD COLUMN IF NOT EXISTS tenant_id UUID REFERENCES tenants(id) ON DELETE RESTRICT;
    END IF;
    IF EXISTS (SELECT FROM pg_tables WHERE schemaname = 'public' AND tablename = 'images') THEN 
        ALTER TABLE images ADD COLUMN IF NOT EXISTS tenant_id UUID REFERENCES tenants(id) ON DELETE RESTRICT;
    END IF;
END $$;

-- Create indexes for efficient tenant-scoped queries
CREATE INDEX IF NOT EXISTS idx_instances_tenant ON instances(tenant_id);
CREATE INDEX IF NOT EXISTS idx_vpcs_tenant ON vpcs(tenant_id);
CREATE INDEX IF NOT EXISTS idx_volumes_tenant ON volumes(tenant_id);
CREATE INDEX IF NOT EXISTS idx_load_balancers_tenant ON load_balancers(tenant_id);
CREATE INDEX IF NOT EXISTS idx_scaling_groups_tenant ON scaling_groups(tenant_id);
CREATE INDEX IF NOT EXISTS idx_buckets_tenant ON buckets(tenant_id);
CREATE INDEX IF NOT EXISTS idx_clusters_tenant ON clusters(tenant_id);
CREATE INDEX IF NOT EXISTS idx_security_groups_tenant ON security_groups(tenant_id);
CREATE INDEX IF NOT EXISTS idx_subnets_tenant ON subnets(tenant_id);
CREATE INDEX IF NOT EXISTS idx_snapshots_tenant ON snapshots(tenant_id);
CREATE INDEX IF NOT EXISTS idx_secrets_tenant ON secrets(tenant_id);
CREATE INDEX IF NOT EXISTS idx_functions_tenant ON functions(tenant_id);
CREATE INDEX IF NOT EXISTS idx_caches_tenant ON caches(tenant_id);
CREATE INDEX IF NOT EXISTS idx_queues_tenant ON queues(tenant_id);
CREATE INDEX IF NOT EXISTS idx_topics_tenant ON topics(tenant_id);
CREATE INDEX IF NOT EXISTS idx_subscriptions_tenant ON subscriptions(tenant_id);
CREATE INDEX IF NOT EXISTS idx_cron_jobs_tenant ON cron_jobs(tenant_id);
CREATE INDEX IF NOT EXISTS idx_gateway_routes_tenant ON gateway_routes(tenant_id);
CREATE INDEX IF NOT EXISTS idx_deployments_tenant ON deployments(tenant_id);
CREATE INDEX IF NOT EXISTS idx_stacks_tenant ON stacks(tenant_id);
CREATE INDEX IF NOT EXISTS idx_images_tenant ON images(tenant_id);
