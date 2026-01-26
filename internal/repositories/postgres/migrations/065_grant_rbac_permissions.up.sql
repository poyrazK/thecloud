-- +goose Up
-- Grant RBAC permissions to the developer role via role_permissions table
INSERT INTO role_permissions (role_id, permission)
SELECT id, p FROM roles CROSS JOIN (VALUES ('*'), ('rbac:create'), ('rbac:read'), ('rbac:update'), ('rbac:delete'), ('storage:create'), ('storage:read'), ('storage:delete')) AS perms(p)
WHERE name = 'developer'
ON CONFLICT DO NOTHING;
