DO $$
DECLARE
    uid uuid := gen_random_uuid();
    tid uuid := gen_random_uuid();
BEGIN
    INSERT INTO users (id, email, password_hash, name, role, created_at, updated_at)
    VALUES (uid, 'test@example.com', 'hash', 'Test User', 'admin', NOW(), NOW());

    INSERT INTO tenants (id, name, slug, owner_id, plan, status, created_at, updated_at)
    VALUES (tid, 'Test Tenant', 'test-tenant', uid, 'pro', 'active', NOW(), NOW());

    UPDATE users SET default_tenant_id = tid WHERE id = uid;

    INSERT INTO tenant_members (tenant_id, user_id, role, joined_at)
    VALUES (tid, uid, 'owner', NOW());

    INSERT INTO api_keys (id, user_id, key, name, created_at, default_tenant_id)
    VALUES (gen_random_uuid(), uid, 'test-key', 'Test Key', NOW(), tid);
END $$;
