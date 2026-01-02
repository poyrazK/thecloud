DROP INDEX IF EXISTS idx_scaling_groups_user_id;
DROP INDEX IF EXISTS idx_load_balancers_user_id;
DROP INDEX IF EXISTS idx_objects_user_id;
DROP INDEX IF EXISTS idx_volumes_user_id;
DROP INDEX IF EXISTS idx_vpcs_user_id;
DROP INDEX IF EXISTS idx_instances_user_id;

ALTER TABLE scaling_groups DROP COLUMN IF EXISTS user_id;
ALTER TABLE load_balancers DROP COLUMN IF EXISTS user_id;
ALTER TABLE objects DROP COLUMN IF EXISTS user_id;
ALTER TABLE volumes DROP COLUMN IF EXISTS user_id;
ALTER TABLE vpcs DROP COLUMN IF EXISTS user_id;
ALTER TABLE instances DROP COLUMN IF EXISTS user_id;
