-- Create approval_comments table
CREATE TABLE IF NOT EXISTS approval_comments (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    approval_request_id UUID NOT NULL REFERENCES approval_requests(id) ON DELETE CASCADE,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    comment TEXT NOT NULL,
    created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP
);

-- Create index for faster lookups
CREATE INDEX IF NOT EXISTS idx_approval_comments_approval_request_id ON approval_comments(approval_request_id);
CREATE INDEX IF NOT EXISTS idx_approval_comments_user_id ON approval_comments(user_id);
CREATE INDEX IF NOT EXISTS idx_approval_comments_created_at ON approval_comments(created_at DESC);
