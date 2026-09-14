CREATE TABLE IF NOT EXISTS settings (
    id         INTEGER PRIMARY KEY,
    config     TEXT NOT NULL,
    updated_at TEXT NOT NULL DEFAULT (datetime('now'))
)
