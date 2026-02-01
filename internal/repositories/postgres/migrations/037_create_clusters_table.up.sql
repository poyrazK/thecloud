CREATE TABLE IF NOT EXISTS clusters (
    id UUID PRIMARY KEY,
    user_id UUID NOT NULL,
    vpc_id UUID NOT NULL REFERENCES vpcs(id),
    name TEXT NOT NULL,
    version TEXT NOT NULL,
    control_plane_ips TEXT[],
    worker_count INTEGER NOT NULL DEFAULT 0,
    status TEXT NOT NULL,
    ssh_key TEXT,
    kubeconfig TEXT, -- Encrypted
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS cluster_nodes (
    id UUID PRIMARY KEY,
    cluster_id UUID NOT NULL REFERENCES clusters(id) ON DELETE CASCADE,
    instance_id UUID NOT NULL REFERENCES instances(id),
    role TEXT NOT NULL,
    status TEXT NOT NULL,
    joined_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_clusters_user_id ON clusters(user_id);
CREATE INDEX IF NOT EXISTS idx_cluster_nodes_cluster_id ON cluster_nodes(cluster_id);
