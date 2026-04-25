-- +goose Down
DELETE FROM role_permissions WHERE permission = 'instance:resize';