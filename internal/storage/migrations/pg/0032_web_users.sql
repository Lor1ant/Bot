CREATE TABLE IF NOT EXISTS web_users (
    tg_id       BIGINT PRIMARY KEY,
    email       TEXT NOT NULL UNIQUE,
    pass_hash   TEXT NOT NULL,
    created_at  TEXT NOT NULL DEFAULT ''
)
