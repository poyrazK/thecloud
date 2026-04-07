-- +goose Up
-- Seed anonymous user for tests and unauthenticated actions
-- This prevents foreign key violations in audit_logs when userID is uuid.Nil
INSERT INTO users (id, email, password_hash, name, role)
VALUES ('00000000-0000-0000-0000-000000000000', 'anonymous@thecloud.local', 'NO_LOGIN', 'Anonymous User', 'user')
ON CONFLICT DO NOTHING;

-- Seed a default tenant for the anonymous user if needed
INSERT INTO tenants (id, name, slug, owner_id, plan, status, created_at, updated_at)
VALUES ('00000000-0000-0000-0000-000000000000', 'Anonymous Tenant', 'anonymous', '00000000-0000-0000-0000-000000000000', 'free', 'active', NOW(), NOW())
ON CONFLICT DO NOTHING;

-- +goose Down
DELETE FROM users WHERE id = '00000000-0000-0000-0000-000000000000';
DELETE FROM tenants WHERE id = '00000000-0000-0000-0000-000000000000';
