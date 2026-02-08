-- +goose Up
CREATE TABLE IF NOT EXISTS messages (
    id UUID PRIMARY KEY,
    text TEXT NOT NULL,
    status VARCHAR(50) NOT NULL,
    scheduled_at TIMESTAMPTZ NOT NULL,
    user_id BIGINT NOT NULL,
    telegram_chat_id BIGINT NOT NULL,
    created_at TIMESTAMPTZ DEFAULT NOW() NOT NULL,
    updated_at TIMESTAMPTZ DEFAULT NOW() NOT NULL
);

-- +goose Down
DROP TABLE IF EXISTS messages;
