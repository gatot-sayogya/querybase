-- Chat threads: maps Google Chat threads to approval requests
-- Enables threaded replies and comment syncing
CREATE TABLE IF NOT EXISTS chat_threads (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    approval_id UUID NOT NULL UNIQUE REFERENCES approval_requests(id) ON DELETE CASCADE,
    space_name TEXT NOT NULL,
    thread_name TEXT NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_chat_threads_approval_id ON chat_threads(approval_id);
CREATE INDEX idx_chat_threads_thread_name ON chat_threads(thread_name);
