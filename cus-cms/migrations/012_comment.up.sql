CREATE TABLE IF NOT EXISTS comments (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    target_type VARCHAR(20) NOT NULL,
    target_id UUID NOT NULL,
    parent_id UUID,
    blogger_id UUID,
    nickname VARCHAR(50) NOT NULL,
    email VARCHAR(100),
    avatar VARCHAR(500),
    content TEXT NOT NULL,
    is_blogger BOOLEAN DEFAULT FALSE,
    status SMALLINT DEFAULT 1,
    ip_address VARCHAR(50),
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW(),
    deleted_at TIMESTAMPTZ
);

CREATE INDEX IF NOT EXISTS idx_comments_target ON comments(target_type, target_id);
CREATE INDEX IF NOT EXISTS idx_comments_parent_id ON comments(parent_id);
CREATE INDEX IF NOT EXISTS idx_comments_status ON comments(status);
CREATE INDEX IF NOT EXISTS idx_comments_deleted_at ON comments(deleted_at);
