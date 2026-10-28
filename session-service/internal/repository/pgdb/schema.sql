CREATE TABLE IF NOT EXISTS agent_chats (
    chat_id TEXT PRIMARY KEY,
    user_id TEXT NOT NULL,
    summary TEXT NOT NULL DEFAULT '',
    created_at TIMESTAMPTZ NOT NULL,
    updated_at TIMESTAMPTZ NOT NULL
);

CREATE INDEX IF NOT EXISTS agent_chats_user_updated_idx
    ON agent_chats (user_id, updated_at DESC);

CREATE TABLE IF NOT EXISTS agent_messages (
    id BIGSERIAL PRIMARY KEY,
    chat_id TEXT NOT NULL REFERENCES agent_chats (chat_id),
    role TEXT NOT NULL,
    body TEXT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL
);

CREATE INDEX IF NOT EXISTS agent_messages_chat_idx
    ON agent_messages (chat_id, id);

CREATE TABLE IF NOT EXISTS trip_configs (
    chat_id TEXT PRIMARY KEY,
    user_id TEXT NOT NULL,
    config JSONB NOT NULL,
    updated_at TIMESTAMPTZ NOT NULL
);
