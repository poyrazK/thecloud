-- +goose Down
-- Reverse tenant migration (best effort)
-- Note: We generally don't want to delete user data, but for rollback testing we need to clear side effects
DELETE FROM tenant_quotas;
DELETE FROM tenant_members;
UPDATE users SET default_tenant_id = NULL;
DELETE FROM tenants WHERE slug LIKE 'personal-%';
