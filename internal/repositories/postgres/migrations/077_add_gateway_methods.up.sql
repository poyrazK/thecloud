-- Add methods column to gateway_routes
ALTER TABLE gateway_routes ADD COLUMN methods TEXT[] DEFAULT '{}';

-- Remove the unique constraint on path_prefix to allow multiple methods on the same path
ALTER TABLE gateway_routes DROP CONSTRAINT IF EXISTS gateway_routes_path_prefix_key;

-- Create a composite unique constraint (path_pattern, methods). 
-- Note: This is an approximation since methods is an array.
ALTER TABLE gateway_routes ADD CONSTRAINT gateway_routes_pattern_methods_key UNIQUE (path_pattern, methods);

-- Create an index to speed up route lookups by method if needed
CREATE INDEX idx_gateway_routes_methods ON gateway_routes USING GIN (methods);
