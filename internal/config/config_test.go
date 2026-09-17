package config

import (
	"os"
	"path/filepath"
	"strings"
	"sync"
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

// clearDBEnv unsets all DB_* variables so LoadDatabaseConfig falls back to
// values sourced from files (or built-in defaults) during the test.
func clearDBEnv(t *testing.T) {
	t.Helper()
	for _, key := range []string{"DB_HOST", "DB_PORT", "DB_USER", "DB_PASSWORD", "DB_NAME", "DB_SSLMODE"} {
		t.Setenv(key, "")
		os.Unsetenv(key)
	}
}

// withWorkingDir switches the process working directory to dir for the
// duration of the test and restores it afterwards.
func withWorkingDir(t *testing.T, dir string) {
	t.Helper()
	original, err := os.Getwd()
	if err != nil {
		t.Fatalf("os.Getwd() error = %v", err)
	}
	if err := os.Chdir(dir); err != nil {
		t.Fatalf("os.Chdir(%q) error = %v", dir, err)
	}
	t.Cleanup(func() {
		_ = os.Chdir(original)
	})
}

// resetLoadEnvFileOnce allows loadEnvFile to run again within the same test
// binary, since it is normally guarded to run only once per process.
func resetLoadEnvFileOnce(t *testing.T) {
	t.Helper()
	original := loadEnvFileOnce
	loadEnvFileOnce = &sync.Once{}
	t.Cleanup(func() {
		loadEnvFileOnce = original
	})
}

func TestLoadDatabaseConfigReadsDotEnvFile(t *testing.T) {
	clearDBEnv(t)
	resetLoadEnvFileOnce(t)
	withWorkingDir(t, t.TempDir())

	writeEnvFile(t, ".env", "DB_NAME=from_dot_env\nDB_HOST=env-file-host\n")

	cfg := LoadDatabaseConfig()

	if cfg.Name != "from_dot_env" {
		t.Errorf("cfg.Name = %q, want %q", cfg.Name, "from_dot_env")
	}
	if cfg.Host != "env-file-host" {
		t.Errorf("cfg.Host = %q, want %q", cfg.Host, "env-file-host")
	}
	// Values not present in the file should still fall back to defaults.
	if cfg.Port != "5432" {
		t.Errorf("cfg.Port = %q, want %q", cfg.Port, "5432")
	}
}

func TestLoadDatabaseConfigDotEnvLocalOverridesDotEnv(t *testing.T) {
	clearDBEnv(t)
	resetLoadEnvFileOnce(t)
	withWorkingDir(t, t.TempDir())

	writeEnvFile(t, ".env", "DB_NAME=from_dot_env\nDB_HOST=env-file-host\n")
	writeEnvFile(t, ".env.local", "DB_NAME=from_dot_env_local\n")

	cfg := LoadDatabaseConfig()

	if cfg.Name != "from_dot_env_local" {
		t.Errorf("cfg.Name = %q, want %q", cfg.Name, "from_dot_env_local")
	}
	// Values only present in .env should be unaffected by .env.local.
	if cfg.Host != "env-file-host" {
		t.Errorf("cfg.Host = %q, want %q", cfg.Host, "env-file-host")
	}
}

func TestLoadDatabaseConfigRealEnvVarOverridesFiles(t *testing.T) {
	resetLoadEnvFileOnce(t)
	withWorkingDir(t, t.TempDir())

	writeEnvFile(t, ".env", "DB_NAME=from_dot_env\n")
	writeEnvFile(t, ".env.local", "DB_NAME=from_dot_env_local\n")
	t.Setenv("DB_NAME", "from_real_env")

	cfg := LoadDatabaseConfig()

	if cfg.Name != "from_real_env" {
		t.Errorf("cfg.Name = %q, want %q", cfg.Name, "from_real_env")
	}
}

func writeEnvFile(t *testing.T, name, contents string) {
	t.Helper()
	if err := os.WriteFile(filepath.Join(".", name), []byte(contents), 0o600); err != nil {
		t.Fatalf("writing %s: %v", name, err)
	}
}
