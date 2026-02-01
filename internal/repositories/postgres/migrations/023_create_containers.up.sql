-- +goose Up

-- CloudContainers: Container Orchestration Service
CREATE TABLE IF NOT EXISTS deployments (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    name VARCHAR(255) NOT NULL,
    image TEXT NOT NULL,
    replicas INT NOT NULL DEFAULT 1,
    current_count INT NOT NULL DEFAULT 0,
    ports TEXT,
    status VARCHAR(50) NOT NULL DEFAULT 'SCALING',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE(user_id, name)
);

CREATE INDEX IF NOT EXISTS idx_deployments_user_id ON deployments(user_id);

CREATE TABLE IF NOT EXISTS deployment_containers (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    deployment_id UUID NOT NULL REFERENCES deployments(id) ON DELETE CASCADE,
    instance_id UUID NOT NULL REFERENCES instances(id) ON DELETE CASCADE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE(instance_id) -- An instance can only belong to one deployment
);

CREATE INDEX IF NOT EXISTS idx_deployment_containers_deployment_id ON deployment_containers(deployment_id);
