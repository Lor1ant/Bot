CREATE TABLE IF NOT EXISTS whitelist (
    telegram_id BIGINT PRIMARY KEY,
    created_at  TEXT NOT NULL DEFAULT ''
)
