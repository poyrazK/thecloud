-- +goose Up
INSERT INTO permissions (id, name, description, created_at)
VALUES ('instance:resize', 'instance:resize', 'Resize an instance', NOW())
ON CONFLICT (id) DO NOTHING;

INSERT INTO role_permissions (role_id, permission_id)
SELECT r.id, p.id FROM roles r, permissions p
WHERE r.name = 'developer' AND p.id = 'instance:resize'
ON CONFLICT (role_id, permission_id) DO NOTHING;