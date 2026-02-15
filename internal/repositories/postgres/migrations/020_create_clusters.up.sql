CREATE TABLE IF NOT EXISTS clusters (
    id UUID PRIMARY KEY,
    user_id UUID NOT NULL REFERENCES users(id),
    vpc_id UUID NOT NULL REFERENCES vpcs(id),
    name VARCHAR(255) NOT NULL,
    version VARCHAR(50) NOT NULL DEFAULT 'v1.29.0',
    status VARCHAR(50) NOT NULL DEFAULT 'pending',
    worker_count INT NOT NULL DEFAULT 2,
    ha_enabled BOOLEAN NOT NULL DEFAULT FALSE,
    network_isolation BOOLEAN NOT NULL DEFAULT FALSE,
    -- Networking
    pod_cidr VARCHAR(50) NOT NULL DEFAULT '10.244.0.0/16',
    service_cidr VARCHAR(50) NOT NULL DEFAULT '10.96.0.0/12',
    api_server_lb_address VARCHAR(255),
    -- Secrets (Encrypted at App Layer)
    kubeconfig_encrypted TEXT,
    ssh_private_key_encrypted TEXT,
    join_token TEXT,
    token_expires_at TIMESTAMPTZ,
    ca_cert_hash TEXT,
    -- Metadata
    job_id VARCHAR(255),
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS cluster_nodes (
    id UUID PRIMARY KEY,
    cluster_id UUID NOT NULL REFERENCES clusters(id) ON DELETE CASCADE,
    instance_id UUID NOT NULL REFERENCES instances(id),
    role VARCHAR(50) NOT NULL,
    status VARCHAR(50) NOT NULL DEFAULT 'pending',
    last_heartbeat TIMESTAMPTZ,
    joined_at TIMESTAMPTZ
);

CREATE INDEX IF NOT EXISTS idx_clusters_user_id ON clusters(user_id);
CREATE INDEX IF NOT EXISTS idx_cluster_nodes_cluster_id ON cluster_nodes(cluster_id);
