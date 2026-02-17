CREATE TABLE IF NOT EXISTS videos (
    id UUID PRIMARY KEY,
    title VARCHAR(255) NOT NULL DEFAULT '',
    description TEXT NOT NULL DEFAULT '',
    file_path VARCHAR(1024) NOT NULL DEFAULT '',
    hls_path VARCHAR(1024) NOT NULL DEFAULT '',
    manifest_path VARCHAR(1024) NOT NULL DEFAULT '',
    se_manifest_url VARCHAR(1024) NOT NULL DEFAULT '',
    s3_url VARCHAR(1024) NOT NULL DEFAULT '',
    status VARCHAR(50) NOT NULL DEFAULT 'pending',
    upload_status VARCHAR(50) NOT NULL DEFAULT 'pending_s3',
    error_message TEXT NOT NULL DEFAULT '',
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMP WITH TIME ZONE
);

CREATE INDEX IF NOT EXISTS idx_videos_created_at ON videos(created_at);
CREATE INDEX IF NOT EXISTS idx_videos_status ON videos(status);
CREATE INDEX IF NOT EXISTS idx_videos_deleted_at ON videos(deleted_at) WHERE deleted_at IS NULL;
