-- +goose Up
ALTER TABLE gateway_routes ADD COLUMN IF NOT EXISTS methods TEXT[] DEFAULT '{GET,POST,PUT,DELETE,PATCH,OPTIONS}';

DO $$
BEGIN
    IF NOT EXISTS (SELECT 1 FROM information_schema.table_constraints WHERE table_schema = current_schema() AND constraint_name = 'gateway_routes_pattern_methods_key') THEN
        ALTER TABLE gateway_routes ADD CONSTRAINT gateway_routes_pattern_methods_key UNIQUE (path_pattern, methods);
    END IF;
END $$;
