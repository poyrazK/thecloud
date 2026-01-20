-- +goose Down
DELETE FROM role_permissions WHERE permission IN (
    'cluster:create',
    'cluster:read',
    'cluster:update',
    'cluster:delete',
    'cluster:list',
    'security_group:create',
    'security_group:read',
    'security_group:update',
    'security_group:delete'
) AND role_id = '00000000-0000-0000-0000-000000000002';
