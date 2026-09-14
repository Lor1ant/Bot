package app

import (
	"testing"

	"remnabot/internal/config"
	"remnabot/internal/model"
)

func TestDsnForEnv(t *testing.T) {
	a := &App{cfg: &config.Config{DataDir: "/data", DatabaseURL: "postgres://x"}}
	if got := a.dsnForEnv(model.DBSQLite); got != "/data/bot.db" {
		t.Fatalf("sqlite dsn = %q", got)
	}
	if got := a.dsnForEnv(model.DBPostgres); got != "postgres://x" {
		t.Fatalf("pg dsn = %q", got)
	}
}

func TestInstalled(t *testing.T) {
	a := &App{}
	if a.installed() {
		t.Fatal("при nil botCfg должно быть «не установлен»")
	}
	a.botCfg = &model.BotConfig{Installed: true}
	if !a.installed() {
		t.Fatal("должно быть «установлен»")
	}
}

func TestLangSelection(t *testing.T) {
	a := &App{wiz: map[int64]*wizard{}}
	if a.lang(1) != "ru" {
		t.Fatalf("fallback lang = %q", a.lang(1))
	}
	a.botCfg = &model.BotConfig{Language: "en"}
	if a.lang(1) != "en" {
		t.Fatalf("config lang = %q", a.lang(1))
	}
	a.wiz[1] = &wizard{cfg: model.BotConfig{Language: "ru"}}
	if a.lang(1) != "ru" {
		t.Fatalf("wizard lang = %q", a.lang(1))
	}
}
