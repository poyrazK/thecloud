CREATE TABLE IF NOT EXISTS images (
    id UUID PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    description TEXT,
    os VARCHAR(50),
    version VARCHAR(50),
    size_gb INTEGER DEFAULT 0,
    file_path TEXT,
    format VARCHAR(20) DEFAULT 'qcow2',
    is_public BOOLEAN DEFAULT false,
    user_id UUID NOT NULL,
    status VARCHAR(50) NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_images_user_id ON images(user_id);
CREATE INDEX IF NOT EXISTS idx_images_status ON images(status);
