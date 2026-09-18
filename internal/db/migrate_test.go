package db

import (
	"database/sql"
	"os"
	"path/filepath"
	"runtime"
	"testing"

	_ "github.com/jackc/pgx/v5/stdlib"
)

func testDSN(t *testing.T) string {
	t.Helper()

	dsn := os.Getenv("TEST_DATABASE_DSN")
	if dsn == "" {
		t.Skip("TEST_DATABASE_DSN not set; skipping test that requires a live PostgreSQL instance")
	}
	return dsn
}

func repoMigrationsPath(t *testing.T) string {
	t.Helper()

	_, thisFile, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("unable to determine current file path")
	}
	// internal/db/migrate_test.go -> repo root/db/migrations
	return filepath.Join(filepath.Dir(thisFile), "..", "..", "db", "migrations")
}

func expectedTables() []string {
	return []string{
		"vehicle_categories",
		"vehicles",
		"booking_policies",
		"reservations",
		"reservation_vehicle_assignments",
		"vehicle_status_history",
		"vehicle_compliance_documents",
		"vehicle_status_triggers",
		"maintenance_schedules",
		"vehicle_telematics_readings",
		"maintenance_work_orders",
		"maintenance_alerts",
	}
}

func tableExists(t *testing.T, dsn, table string) bool {
	t.Helper()

	conn, err := sql.Open("pgx", dsn)
	if err != nil {
		t.Fatalf("opening connection: %v", err)
	}
	defer conn.Close()

	var exists bool
	err = conn.QueryRow(
		"SELECT EXISTS (SELECT 1 FROM information_schema.tables WHERE table_schema = 'public' AND table_name = $1)",
		table,
	).Scan(&exists)
	if err != nil {
		t.Fatalf("checking table %q existence: %v", table, err)
	}
	return exists
}

func TestMigrateUpCreatesAllTables(t *testing.T) {
	dsn := testDSN(t)
	migrationsPath := repoMigrationsPath(t)

	if err := Migrate(dsn, migrationsPath); err != nil {
		t.Fatalf("Migrate() error = %v", err)
	}
	t.Cleanup(func() {
		if err := Rollback(dsn, migrationsPath, len(expectedTables())+1); err != nil {
			t.Logf("cleanup rollback error: %v", err)
		}
	})

	for _, table := range expectedTables() {
		if !tableExists(t, dsn, table) {
			t.Errorf("expected table %q to exist after migration", table)
		}
	}

	// Running Migrate again should be a no-op and not return an error.
	if err := Migrate(dsn, migrationsPath); err != nil {
		t.Fatalf("Migrate() second call error = %v", err)
	}
}

func TestRollbackRemovesTables(t *testing.T) {
	dsn := testDSN(t)
	migrationsPath := repoMigrationsPath(t)

	if err := Migrate(dsn, migrationsPath); err != nil {
		t.Fatalf("Migrate() error = %v", err)
	}

	if err := Rollback(dsn, migrationsPath, len(expectedTables())+1); err != nil {
		t.Fatalf("Rollback() error = %v", err)
	}

	for _, table := range expectedTables() {
		if tableExists(t, dsn, table) {
			t.Errorf("expected table %q to not exist after rollback", table)
		}
	}
}
