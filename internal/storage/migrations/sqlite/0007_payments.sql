CREATE TABLE IF NOT EXISTS payments (
    id          INTEGER PRIMARY KEY,
    telegram_id INTEGER NOT NULL,
    method      TEXT NOT NULL,
    months      INTEGER NOT NULL,
    amount      TEXT NOT NULL DEFAULT '',
    status      TEXT NOT NULL,
    comment     TEXT NOT NULL DEFAULT '',
    created_at  TEXT NOT NULL DEFAULT ''
)
