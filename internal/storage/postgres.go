package storage

import (
	"context"
	"database/sql"
	"fmt"
	"strconv"

	"remnabot/internal/crypto"
	"remnabot/internal/model"
)

type pgStore struct{ base }

func openPostgres(dsn string, crypter *crypto.Crypter) (Storage, error) {
	db, err := sql.Open("pgx", dsn)
	if err != nil {
		return nil, err
	}
	db.SetMaxOpenConns(10)
	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("ping postgres: %w", err)
	}
	return &pgStore{base{
		db:      db,
		kind:    model.DBPostgres,
		ph:      func(n int) string { return "$" + strconv.Itoa(n) },
		crypter: crypter,
	}}, nil
}

func (s *pgStore) Migrate(ctx context.Context) error {
	return runMigrations(ctx, &s.base, "pg")
}

func (s *pgStore) LoadConfig(ctx context.Context) (*model.BotConfig, bool, error) {
	return s.loadConfig(ctx)
}

func (s *pgStore) SaveConfig(ctx context.Context, cfg *model.BotConfig) error {
	const q = `INSERT INTO settings (id, config, updated_at)
	           VALUES (1, $1, now())
	           ON CONFLICT (id) DO UPDATE SET config = EXCLUDED.config, updated_at = now()`
	return s.saveConfig(ctx, cfg, q)
}
