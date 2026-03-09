-- +goose Down

DROP INDEX IF EXISTS idx_user_policies_user_id;
DROP TABLE IF EXISTS user_policies;
DROP TABLE IF EXISTS policies;
