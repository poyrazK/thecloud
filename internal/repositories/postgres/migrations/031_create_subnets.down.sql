-- +goose Down
DROP INDEX IF EXISTS idx_subnets_user;
DROP INDEX IF EXISTS idx_subnets_vpc;
DROP TABLE IF EXISTS subnets;
