-- Create media table for polymorphic file attachments
CREATE TABLE IF NOT EXISTS media (
    id SERIAL PRIMARY KEY,

    -- Polymorphic relationship
    attachable_type VARCHAR(100) NOT NULL,
    attachable_id INTEGER NOT NULL,
    collection VARCHAR(50) NOT NULL DEFAULT 'default',

    -- File metadata
    filename VARCHAR(255) NOT NULL,
    original_name VARCHAR(255) NOT NULL,
    mime_type VARCHAR(100) NOT NULL,
    size BIGINT NOT NULL,

    -- Storage info
    disk VARCHAR(50) NOT NULL DEFAULT 's3',
    path VARCHAR(500) NOT NULL,

    -- Media specific
    type VARCHAR(20) NOT NULL,
    hash VARCHAR(64),

    -- Image specific (nullable)
    width INTEGER,
    height INTEGER,

    -- JSON fields
    variants JSONB,
    metadata JSONB,

    -- Timestamps
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP WITH TIME ZONE
);

-- Create indexes for common queries
CREATE INDEX idx_media_attachable ON media(attachable_type, attachable_id);
CREATE INDEX idx_media_collection ON media(attachable_type, attachable_id, collection);
CREATE INDEX idx_media_hash ON media(hash);
CREATE INDEX idx_media_deleted_at ON media(deleted_at);
CREATE INDEX idx_media_type ON media(type);

-- Add comment to table
COMMENT ON TABLE media IS 'Polymorphic file attachments for any entity';
