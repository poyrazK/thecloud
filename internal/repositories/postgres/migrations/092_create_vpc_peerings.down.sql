-- +goose Down
DROP INDEX IF EXISTS idx_active_peering_pair;
DROP TABLE IF EXISTS vpc_peerings;
