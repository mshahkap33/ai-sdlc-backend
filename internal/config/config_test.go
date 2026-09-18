package config

import (
	"net/url"
	"testing"
)

func TestLoadDatabaseConfigDefaults(t *testing.T) {
	t.Setenv("DB_HOST", "")
	t.Setenv("DB_PORT", "")
	t.Setenv("DB_USER", "")
	t.Setenv("DB_PASSWORD", "")
	t.Setenv("DB_NAME", "")
	t.Setenv("DB_SSLMODE", "")

	cfg := LoadDatabaseConfig()

	if cfg.Host != "localhost" {
		t.Errorf("expected default host 'localhost', got %q", cfg.Host)
	}
	if cfg.Port != "5432" {
		t.Errorf("expected default port '5432', got %q", cfg.Port)
	}
	if cfg.User != "postgres" {
		t.Errorf("expected default user 'postgres', got %q", cfg.User)
	}
	if cfg.Name != "ai_sdlc" {
		t.Errorf("expected default name 'ai_sdlc', got %q", cfg.Name)
	}
	if cfg.SSLMode != "disable" {
		t.Errorf("expected default sslmode 'disable', got %q", cfg.SSLMode)
	}
}

func TestLoadDatabaseConfigFromEnv(t *testing.T) {
	t.Setenv("DB_HOST", "db.example.com")
	t.Setenv("DB_PORT", "6543")
	t.Setenv("DB_USER", "app_user")
	t.Setenv("DB_PASSWORD", "s3cret")
	t.Setenv("DB_NAME", "vehicles")
	t.Setenv("DB_SSLMODE", "require")

	cfg := LoadDatabaseConfig()

	if cfg.Host != "db.example.com" || cfg.Port != "6543" || cfg.User != "app_user" ||
		cfg.Password != "s3cret" || cfg.Name != "vehicles" || cfg.SSLMode != "require" {
		t.Fatalf("unexpected config loaded from environment: %+v", cfg)
	}
}

func TestDatabaseConfigDSN(t *testing.T) {
	cfg := DatabaseConfig{
		Host:     "db.example.com",
		Port:     "5432",
		User:     "app_user",
		Password: "p@ss/word",
		Name:     "vehicles",
		SSLMode:  "require",
	}

	dsn := cfg.DSN()

	parsed, err := url.Parse(dsn)
	if err != nil {
		t.Fatalf("expected DSN to be a valid URL, got error: %v", err)
	}

	if parsed.Scheme != "postgres" {
		t.Errorf("expected scheme 'postgres', got %q", parsed.Scheme)
	}
	if parsed.Host != "db.example.com:5432" {
		t.Errorf("expected host 'db.example.com:5432', got %q", parsed.Host)
	}
	if parsed.Path != "/vehicles" {
		t.Errorf("expected path '/vehicles', got %q", parsed.Path)
	}
	if user := parsed.User.Username(); user != "app_user" {
		t.Errorf("expected user 'app_user', got %q", user)
	}
	if pass, _ := parsed.User.Password(); pass != "p@ss/word" {
		t.Errorf("expected password 'p@ss/word', got %q", pass)
	}
	if parsed.Query().Get("sslmode") != "require" {
		t.Errorf("expected sslmode 'require', got %q", parsed.Query().Get("sslmode"))
	}
}
