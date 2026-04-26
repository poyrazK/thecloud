-- +goose Up
-- Adds instance:resize permission for developer role
INSERT INTO role_permissions (role_id, permission)
SELECT id, 'instance:resize' FROM roles WHERE name = 'developer'
ON CONFLICT (role_id, permission) DO NOTHING;