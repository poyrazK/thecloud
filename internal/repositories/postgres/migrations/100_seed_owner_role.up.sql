-- +goose Up
-- Seed the owner role which is used for tenant owners
INSERT INTO roles (id, name, description) 
VALUES ('00000000-0000-0000-0000-000000000004', 'owner', 'Tenant owner with full access to tenant resources') 
ON CONFLICT (name) DO UPDATE SET description = EXCLUDED.description;

INSERT INTO role_permissions (role_id, permission) 
VALUES ('00000000-0000-0000-0000-000000000004', '*') 
ON CONFLICT DO NOTHING;

-- +goose Down
DELETE FROM role_permissions WHERE role_id = '00000000-0000-0000-0000-000000000004';
DELETE FROM roles WHERE id = '00000000-0000-0000-0000-000000000004';
