-- +goose Up

-- Add missing cluster permissions for developer role
INSERT INTO role_permissions (role_id, permission)
SELECT '00000000-0000-0000-0000-000000000002', p
FROM (VALUES 
    ('cluster:create'),
    ('cluster:read'),
    ('cluster:update'),
    ('cluster:delete'),
    ('cluster:list'),
    ('security_group:create'),
    ('security_group:read'),
    ('security_group:update'),
    ('security_group:delete')
) AS permissions(p)
ON CONFLICT (role_id, permission) DO NOTHING;
