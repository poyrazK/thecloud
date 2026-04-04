-- +goose Down
DELETE FROM role_permissions WHERE role_id IN (SELECT id FROM roles WHERE name = 'developer') AND permission IN ('*', 'rbac:create', 'rbac:read', 'rbac:update', 'rbac:delete', 'storage:create', 'storage:read', 'storage:delete');
