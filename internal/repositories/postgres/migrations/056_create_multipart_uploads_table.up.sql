CREATE TABLE IF NOT EXISTS multipart_uploads (
    id UUID PRIMARY KEY,
    user_id UUID NOT NULL REFERENCES users(id),
    bucket TEXT NOT NULL,
    key TEXT NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS multipart_parts (
    upload_id UUID NOT NULL REFERENCES multipart_uploads(id) ON DELETE CASCADE,
    part_number INTEGER NOT NULL,
    size_bytes BIGINT NOT NULL,
    etag TEXT NOT NULL,
    PRIMARY KEY (upload_id, part_number)
);
