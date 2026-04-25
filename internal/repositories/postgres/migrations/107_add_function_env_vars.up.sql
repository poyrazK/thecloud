-- +goose Up

-- env_vars stores a JSON object mapping key->value, e.g. {"FOO":"bar"}
ALTER TABLE functions ADD COLUMN env_vars JSONB DEFAULT '{}';

-- +goose Down

ALTER TABLE functions DROP COLUMN IF EXISTS env_vars;
