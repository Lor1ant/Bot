package hostctl

import (
	"os"
	"path/filepath"
	"testing"

	"gopkg.in/yaml.v3"
)

func TestEnvDefault(t *testing.T) {
	os.Unsetenv("COMPOSE_PROJECT")
	c := New()
	if c.project != "remnachillbot" {
		t.Fatalf("project по умолчанию = %q", c.project)
	}
}

func TestAddPostgresToCompose(t *testing.T) {
	dir := t.TempDir()
	cf := filepath.Join(dir, "docker-compose.yml")
	initial := "name: remnachillbot\n" +
		"services:\n" +
		"  bot:\n" +
		"    image: x\n" +
		"    environment:\n" +
		"      BOT_TOKEN: \"t\"\n" +
		"volumes:\n" +
		"  bot-data:\n" +
		"networks:\n" +
		"  remnawave-network:\n" +
		"    external: true\n"
	if err := os.WriteFile(cf, []byte(initial), 0o644); err != nil {
		t.Fatal(err)
	}
	c := &Controller{composeFile: cf, project: "remnachillbot", hostDir: "/opt/remnachillbot"}
	if err := c.addPostgresToCompose(); err != nil {
		t.Fatal(err)
	}

	data, _ := os.ReadFile(cf)
	root := map[string]any{}
	if err := yaml.Unmarshal(data, &root); err != nil {
		t.Fatal(err)
	}
	services := root["services"].(map[string]any)
	if _, ok := services["db"]; !ok {
		t.Fatal("сервис db не добавлен")
	}
	bot := services["bot"].(map[string]any)
	env := bot["environment"].(map[string]any)
	if env["DB_KIND"] != "postgres" {
		t.Fatalf("DB_KIND = %v", env["DB_KIND"])
	}
	if env["DATABASE_URL"] != PostgresDSN {
		t.Fatalf("DATABASE_URL = %v", env["DATABASE_URL"])
	}
	vols := root["volumes"].(map[string]any)
	if _, ok := vols["pg-data"]; !ok {
		t.Fatal("том pg-data не добавлен")
	}
}

func TestAvailableNoSock(t *testing.T) {
	c := &Controller{composeFile: "/nonexistent/xyz.yml"}
	if c.Available() {
		t.Fatal("Available должен быть false без compose-файла/сокета")
	}
}
