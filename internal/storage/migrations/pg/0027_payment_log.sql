CREATE TABLE IF NOT EXISTS payment_log (
    id          BIGINT PRIMARY KEY,
    ext_id      TEXT NOT NULL DEFAULT '',
    telegram_id BIGINT NOT NULL DEFAULT 0,
    method      TEXT NOT NULL DEFAULT '',
    stage       TEXT NOT NULL DEFAULT '',
    detail      TEXT NOT NULL DEFAULT '',
    created_at  TEXT NOT NULL DEFAULT ''
)
