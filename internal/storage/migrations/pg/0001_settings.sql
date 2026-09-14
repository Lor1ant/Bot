CREATE TABLE IF NOT EXISTS settings (
    id         INTEGER PRIMARY KEY,
    config     TEXT NOT NULL,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
)
