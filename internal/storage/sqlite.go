package storage

import (
	"context"
	"database/sql"
	"fmt"

	"remnabot/internal/crypto"
	"remnabot/internal/model"
)

type sqliteStore struct{ base }

func openSQLite(path string, crypter *crypto.Crypter) (Storage, error) {

	dsn := fmt.Sprintf("file:%s?_pragma=foreign_keys(1)&_pragma=journal_mode(WAL)&_pragma=busy_timeout(5000)", path)
	db, err := sql.Open("sqlite", dsn)
	if err != nil {
		return nil, err
	}

	db.SetMaxOpenConns(1)
	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("ping sqlite: %w", err)
	}
	return &sqliteStore{base{
		db:      db,
		kind:    model.DBSQLite,
		ph:      func(int) string { return "?" },
		crypter: crypter,
	}}, nil
}

func (s *sqliteStore) Migrate(ctx context.Context) error {
	return runMigrations(ctx, &s.base, "sqlite")
}

func (s *sqliteStore) LoadConfig(ctx context.Context) (*model.BotConfig, bool, error) {
	return s.loadConfig(ctx)
}

func (s *sqliteStore) SaveConfig(ctx context.Context, cfg *model.BotConfig) error {
	const q = `INSERT INTO settings (id, config, updated_at)
	           VALUES (1, ?, datetime('now'))
	           ON CONFLICT(id) DO UPDATE SET config = excluded.config, updated_at = datetime('now')`
	return s.saveConfig(ctx, cfg, q)
}
