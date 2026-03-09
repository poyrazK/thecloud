-- 098_create_cluster_node_groups.up.sql
CREATE TABLE IF NOT EXISTS cluster_node_groups (
    id UUID PRIMARY KEY,
    cluster_id UUID NOT NULL REFERENCES clusters(id) ON DELETE CASCADE,
    name VARCHAR(255) NOT NULL,
    instance_type VARCHAR(50) NOT NULL,
    min_size INTEGER NOT NULL DEFAULT 1,
    max_size INTEGER NOT NULL DEFAULT 10,
    current_size INTEGER NOT NULL DEFAULT 1,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(cluster_id, name),
    CONSTRAINT chk_cluster_node_groups_min_le_max CHECK (min_size <= max_size),
    CONSTRAINT chk_cluster_node_groups_current_in_range CHECK (current_size BETWEEN min_size AND max_size)
);

-- Index for faster lookups by cluster
CREATE INDEX IF NOT EXISTS idx_cluster_node_groups_cluster_id ON cluster_node_groups(cluster_id);

-- Backfill existing clusters: Create a 'default-pool' for each existing cluster
-- mapping its current worker_count to the node group.
INSERT INTO cluster_node_groups (id, cluster_id, name, instance_type, min_size, max_size, current_size)
SELECT 
    gen_random_uuid(), 
    id, 
    'default-pool', 
    'standard-1', 
    1, 
    CASE WHEN worker_count > 10 THEN worker_count ELSE 10 END, 
    worker_count
FROM clusters
WHERE NOT EXISTS (
    SELECT 1 FROM cluster_node_groups cng 
    WHERE cng.cluster_id = clusters.id AND cng.name = 'default-pool'
);
