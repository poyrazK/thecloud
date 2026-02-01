-- Remove methods column and its index
ALTER TABLE gateway_routes DROP CONSTRAINT IF EXISTS gateway_routes_pattern_methods_key;
DROP INDEX IF EXISTS idx_gateway_routes_methods;
ALTER TABLE gateway_routes DROP COLUMN IF EXISTS methods;

-- Restore the unique constraint on path_prefix
ALTER TABLE gateway_routes ADD CONSTRAINT gateway_routes_path_prefix_key UNIQUE (path_prefix);
