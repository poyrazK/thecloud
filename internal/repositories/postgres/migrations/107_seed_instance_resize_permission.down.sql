-- +goose Down
DELETE FROM role_permissions WHERE permission = 'instance:resize' AND role_id = (SELECT id FROM roles WHERE name = 'developer');