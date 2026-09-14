CREATE TABLE IF NOT EXISTS promo_codes (
    code        TEXT PRIMARY KEY,
    kind        TEXT NOT NULL,
    value       INTEGER NOT NULL DEFAULT 0,
    max_uses    INTEGER NOT NULL DEFAULT 0,
    used        INTEGER NOT NULL DEFAULT 0,
    expires_at  TEXT NOT NULL DEFAULT '',
    created_at  TEXT NOT NULL DEFAULT ''
)
