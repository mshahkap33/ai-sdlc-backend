package database

import (
	"database/sql"
	"fmt"
	"net/url"
	"testing"

	"github.com/mshahkap33/ai-sdlc-backend/internal/config"

	_ "github.com/jackc/pgx/v5/stdlib"
)

func testDSN(t *testing.T) string {
	t.Helper()

	cfg := config.LoadDatabaseConfig()
	dsn := cfg.DSN()

	db, err := sql.Open("pgx", dsn)
	if err != nil {
		t.Skipf("skipping: unable to open database connection: %v", err)
	}
	defer db.Close()

	if err := db.Ping(); err != nil {
		t.Skipf("skipping: database is not reachable: %v", err)
	}

	return dsn
}

func TestUpAndDownMigrations(t *testing.T) {
	dsn := testDSN(t)

	db, err := sql.Open("pgx", dsn)
	if err != nil {
		t.Fatalf("opening database connection: %v", err)
	}
	defer db.Close()

	t.Cleanup(func() {
		if err := Down("file://../../db/migrations", dsn); err != nil {
			t.Errorf("cleanup: rolling back migrations: %v", err)
		}
	})

	if err := Up("file://../../db/migrations", dsn); err != nil {
		t.Fatalf("applying migrations: %v", err)
	}

	// Applying again should be a no-op and must not return an error.
	if err := Up("file://../../db/migrations", dsn); err != nil {
		t.Fatalf("re-applying migrations should be a no-op: %v", err)
	}

	for _, table := range []string{"vehicle_categories", "vehicles"} {
		var exists bool
		query := fmt.Sprintf("SELECT EXISTS (SELECT 1 FROM information_schema.tables WHERE table_name = '%s')", table)
		if err := db.QueryRow(query).Scan(&exists); err != nil {
			t.Fatalf("checking table %q exists: %v", table, err)
		}
		if !exists {
			t.Errorf("expected table %q to exist after migrations", table)
		}
	}

	if err := Down("file://../../db/migrations", dsn); err != nil {
		t.Fatalf("rolling back migrations: %v", err)
	}

	for _, table := range []string{"vehicle_categories", "vehicles"} {
		var exists bool
		query := fmt.Sprintf("SELECT EXISTS (SELECT 1 FROM information_schema.tables WHERE table_name = '%s')", table)
		if err := db.QueryRow(query).Scan(&exists); err != nil {
			t.Fatalf("checking table %q exists: %v", table, err)
		}
		if exists {
			t.Errorf("expected table %q to be dropped after rolling back migrations", table)
		}
	}
}

func TestToPgx5DSN(t *testing.T) {
	input := url.URL{
		Scheme:   "postgres",
		User:     url.UserPassword("app_user", "app_password"),
		Host:     "localhost:5432",
		Path:     "/mydb",
		RawQuery: "sslmode=disable",
	}

	want := input
	want.Scheme = pgx5DriverScheme

	dsn, err := toPgx5DSN(input.String())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if dsn != want.String() {
		t.Errorf("expected %q, got %q", want.String(), dsn)
	}
}

func TestToPgx5DSNInvalid(t *testing.T) {
	if _, err := toPgx5DSN("://not-a-valid-url"); err == nil {
		t.Fatal("expected an error for an invalid DSN")
	}
}
