package config

import (
	"strings"
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

	want := DatabaseConfig{
		Host:     "localhost",
		Port:     "5432",
		User:     "postgres",
		Password: "postgres",
		Name:     "ai_sdlc",
		SSLMode:  "disable",
	}

	if cfg != want {
		t.Fatalf("LoadDatabaseConfig() = %+v, want %+v", cfg, want)
	}
}

func TestLoadDatabaseConfigOverrides(t *testing.T) {
	t.Setenv("DB_HOST", "db.example.com")
	t.Setenv("DB_PORT", "6543")
	t.Setenv("DB_USER", "car_admin")
	t.Setenv("DB_PASSWORD", "test_password")
	t.Setenv("DB_NAME", "car_management")
	t.Setenv("DB_SSLMODE", "require")

	cfg := LoadDatabaseConfig()

	want := DatabaseConfig{
		Host:     "db.example.com",
		Port:     "6543",
		User:     "car_admin",
		Password: "test_password",
		Name:     "car_management",
		SSLMode:  "require",
	}

	if cfg != want {
		t.Fatalf("LoadDatabaseConfig() = %+v, want %+v", cfg, want)
	}
}

func TestDatabaseConfigDSN(t *testing.T) {
	cfg := DatabaseConfig{
		Host:     "db.example.com",
		Port:     "5432",
		User:     "car_admin",
		Password: "test_password",
		Name:     "car_management",
		SSLMode:  "disable",
	}

	got := cfg.DSN()

	wantScheme := "postgres"
	if scheme := got[:len(wantScheme)]; scheme != wantScheme {
		t.Fatalf("DSN() scheme = %q, want %q", scheme, wantScheme)
	}

	for _, part := range []string{cfg.User, cfg.Password, cfg.Host, cfg.Port, cfg.Name, "sslmode=" + cfg.SSLMode} {
		if !strings.Contains(got, part) {
			t.Fatalf("DSN() = %q, expected it to contain %q", got, part)
		}
	}
}
