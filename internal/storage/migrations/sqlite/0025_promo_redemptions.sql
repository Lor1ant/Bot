CREATE TABLE IF NOT EXISTS promo_redemptions (
    code        TEXT NOT NULL,
    telegram_id INTEGER NOT NULL,
    created_at  TEXT NOT NULL DEFAULT '',
    PRIMARY KEY (code, telegram_id)
)
