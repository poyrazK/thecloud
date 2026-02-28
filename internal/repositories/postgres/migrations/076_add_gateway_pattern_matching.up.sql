-- Add pattern matching columns to gateway_routes
ALTER TABLE gateway_routes 
    ADD COLUMN IF NOT EXISTS pattern_type VARCHAR(20) DEFAULT 'prefix' NOT NULL,
    ADD COLUMN IF NOT EXISTS path_pattern TEXT,
    ADD COLUMN IF NOT EXISTS param_names JSONB DEFAULT '[]'::jsonb,
    ADD COLUMN IF NOT EXISTS priority INTEGER DEFAULT 0;

-- Migrate existing routes to use pattern_type='prefix'
UPDATE gateway_routes 
SET path_pattern = path_prefix
WHERE pattern_type = 'prefix';

-- Add index for pattern lookups
CREATE INDEX IF NOT EXISTS idx_gateway_routes_pattern_type ON gateway_routes(pattern_type);

-- Add constraint: pattern_type must be 'prefix' or 'pattern'
DO $$ 
BEGIN 
    IF NOT EXISTS (SELECT 1 FROM information_schema.table_constraints WHERE table_schema = current_schema() AND constraint_name = 'chk_pattern_type') THEN
        ALTER TABLE gateway_routes 
            ADD CONSTRAINT chk_pattern_type 
            CHECK (pattern_type IN ('prefix', 'pattern'));
    END IF;
END $$;
