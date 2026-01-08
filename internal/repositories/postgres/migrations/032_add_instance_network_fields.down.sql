-- +goose Down
DROP INDEX IF EXISTS idx_instances_subnet;
ALTER TABLE instances DROP COLUMN IF EXISTS ovs_port;
ALTER TABLE instances DROP COLUMN IF EXISTS private_ip;
ALTER TABLE instances DROP COLUMN IF EXISTS subnet_id;
