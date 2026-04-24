-- +goose Down
DROP TABLE IF EXISTS route_table_associations;
DROP TABLE IF EXISTS routes;
DROP TABLE IF EXISTS route_tables;