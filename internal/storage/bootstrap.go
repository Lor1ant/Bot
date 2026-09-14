package storage

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
)

type Bootstrap struct {
	DBKind string `json:"db_kind"`
	DSN    string `json:"dsn"`
}

func bootstrapPath(dataDir string) string {
	return filepath.Join(dataDir, "bootstrap.json")
}

func LoadBootstrap(dataDir string) (*Bootstrap, error) {
	data, err := os.ReadFile(bootstrapPath(dataDir))
	if errors.Is(err, os.ErrNotExist) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	var b Bootstrap
	if err := json.Unmarshal(data, &b); err != nil {
		return nil, err
	}
	return &b, nil
}

func SaveBootstrap(dataDir string, b *Bootstrap) error {
	if err := os.MkdirAll(dataDir, 0o700); err != nil {
		return err
	}
	data, err := json.MarshalIndent(b, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(bootstrapPath(dataDir), data, 0o600)
}
