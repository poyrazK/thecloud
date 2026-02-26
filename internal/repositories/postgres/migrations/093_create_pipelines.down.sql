-- +goose Down

DROP INDEX IF EXISTS idx_build_logs_step_id;
DROP INDEX IF EXISTS idx_build_logs_build_id;
DROP TABLE IF EXISTS build_logs;

DROP INDEX IF EXISTS idx_build_steps_build_id;
DROP TABLE IF EXISTS build_steps;

DROP INDEX IF EXISTS idx_builds_user_id;
DROP INDEX IF EXISTS idx_builds_pipeline_id;
DROP TABLE IF EXISTS builds;

DROP INDEX IF EXISTS idx_pipelines_user_id;
DROP TABLE IF EXISTS pipelines;
