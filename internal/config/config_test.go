package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadExpandsEnvironmentAndAppliesDefaults(t *testing.T) {
	t.Setenv("MYSQL_DSN", "user:pass@tcp(localhost:3306)/stock_lens")
	t.Setenv("ADMIN_TOKEN", "secret-token")

	path := filepath.Join(t.TempDir(), "config.yaml")
	content := []byte(`
database:
  dsn: "${MYSQL_DSN}"
auth:
  admin_token: "${ADMIN_TOKEN}"
`)
	if err := os.WriteFile(path, content, 0o600); err != nil {
		t.Fatalf("write config: %v", err)
	}

	cfg, err := Load(path)
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	if cfg.Server.Addr != "0.0.0.0" {
		t.Fatalf("server addr = %q, want %q", cfg.Server.Addr, "0.0.0.0")
	}
	if cfg.Server.Port != 30078 {
		t.Fatalf("server port = %d, want %d", cfg.Server.Port, 30078)
	}
	if cfg.Server.Timezone != "Asia/Shanghai" {
		t.Fatalf("server timezone = %q, want %q", cfg.Server.Timezone, "Asia/Shanghai")
	}
	if cfg.Database.DSN != "user:pass@tcp(localhost:3306)/stock_lens" {
		t.Fatalf("database dsn = %q", cfg.Database.DSN)
	}
	if cfg.Auth.AdminToken != "secret-token" {
		t.Fatalf("admin token = %q", cfg.Auth.AdminToken)
	}
	if cfg.Market.Provider != "mock" {
		t.Fatalf("market provider = %q, want %q", cfg.Market.Provider, "mock")
	}
	if cfg.Market.BatchSize != 500 {
		t.Fatalf("market batch size = %d, want %d", cfg.Market.BatchSize, 500)
	}
	if cfg.Log.Level != "info" {
		t.Fatalf("log level = %q, want %q", cfg.Log.Level, "info")
	}
}

func TestLoadRejectsNonPositiveServerPort(t *testing.T) {
	t.Setenv("MYSQL_DSN", "user:pass@tcp(localhost:3306)/stock_lens")
	t.Setenv("ADMIN_TOKEN", "secret-token")

	path := filepath.Join(t.TempDir(), "config.yaml")
	content := []byte(`
server:
  port: 0
database:
  dsn: "${MYSQL_DSN}"
auth:
  admin_token: "${ADMIN_TOKEN}"
`)
	if err := os.WriteFile(path, content, 0o600); err != nil {
		t.Fatalf("write config: %v", err)
	}

	if _, err := Load(path); err == nil {
		t.Fatalf("Load() error = nil, want error")
	}
}

func TestLoadRejectsUnsupportedDatabaseDriver(t *testing.T) {
	t.Setenv("MYSQL_DSN", "user:pass@tcp(localhost:3306)/stock_lens")
	t.Setenv("ADMIN_TOKEN", "secret-token")

	path := filepath.Join(t.TempDir(), "config.yaml")
	content := []byte(`
database:
  driver: "postgres"
  dsn: "${MYSQL_DSN}"
auth:
  admin_token: "${ADMIN_TOKEN}"
`)
	if err := os.WriteFile(path, content, 0o600); err != nil {
		t.Fatalf("write config: %v", err)
	}

	if _, err := Load(path); err == nil {
		t.Fatalf("Load() error = nil, want error")
	}
}

func TestLoadProjectTestConfig(t *testing.T) {
	cfg, err := Load("../../configs/config.test.yaml")
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	if cfg.Server.Addr != "127.0.0.1" {
		t.Fatalf("server addr = %q, want %q", cfg.Server.Addr, "127.0.0.1")
	}
	if cfg.Server.Port != 30078 {
		t.Fatalf("server port = %d, want %d", cfg.Server.Port, 30078)
	}
	if cfg.Market.Provider != "mock" {
		t.Fatalf("market provider = %q, want %q", cfg.Market.Provider, "mock")
	}
	if cfg.Auth.AdminToken == "" {
		t.Fatalf("admin token is empty")
	}
}
