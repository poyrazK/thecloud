-- +goose Down
DELETE FROM role_permissions WHERE permission_id = 'instance:resize';
DELETE FROM permissions WHERE id = 'instance:resize';