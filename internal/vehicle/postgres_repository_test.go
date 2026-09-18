package vehicle

import (
	"context"
	"os"
	"path/filepath"
	"runtime"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/mshahkap33/ai-sdlc-backend/internal/db"
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
	// internal/vehicle/postgres_repository_test.go -> repo root/db/migrations
	return filepath.Join(filepath.Dir(thisFile), "..", "..", "db", "migrations")
}

func newTestPool(t *testing.T) *pgxpool.Pool {
	t.Helper()
	dsn := testDSN(t)

	if err := db.Migrate(dsn, repoMigrationsPath(t)); err != nil {
		t.Fatalf("applying migrations: %v", err)
	}
	t.Cleanup(func() {
		_ = db.Rollback(dsn, repoMigrationsPath(t), 13)
	})

	pool, err := pgxpool.New(context.Background(), dsn)
	if err != nil {
		t.Fatalf("connecting to database: %v", err)
	}
	t.Cleanup(pool.Close)

	return pool
}

func insertTestCategory(t *testing.T, pool *pgxpool.Pool, name string, active bool) string {
	t.Helper()
	var id string
	err := pool.QueryRow(context.Background(),
		`INSERT INTO vehicle_categories (name, active, created_by, updated_by)
		 VALUES ($1, $2, 'test', 'test') RETURNING id`,
		name, active,
	).Scan(&id)
	if err != nil {
		t.Fatalf("inserting test category: %v", err)
	}
	return id
}

func TestPostgresRepositoryCreateVehicleAndCategoryStatus(t *testing.T) {
	pool := newTestPool(t)
	repo := NewPostgresRepository(pool)
	ctx := context.Background()

	activeCategory := insertTestCategory(t, pool, "Economy-"+time.Now().Format("150405.000000000"), true)
	inactiveCategory := insertTestCategory(t, pool, "Retired-"+time.Now().Format("150405.000000000"), false)

	found, active, err := repo.CategoryStatus(ctx, activeCategory)
	if err != nil {
		t.Fatalf("CategoryStatus() error = %v", err)
	}
	if !found || !active {
		t.Fatalf("CategoryStatus() = (%v, %v), want (true, true)", found, active)
	}

	found, active, err = repo.CategoryStatus(ctx, inactiveCategory)
	if err != nil {
		t.Fatalf("CategoryStatus() error = %v", err)
	}
	if !found || active {
		t.Fatalf("CategoryStatus() = (%v, %v), want (true, false)", found, active)
	}

	found, _, err = repo.CategoryStatus(ctx, "00000000-0000-0000-0000-000000000000")
	if err != nil {
		t.Fatalf("CategoryStatus() error = %v", err)
	}
	if found {
		t.Fatalf("CategoryStatus() found = true, want false for unknown category")
	}

	v := &Vehicle{
		VIN:         "1HGCM82633A004352",
		PlateNumber: "TEST-0001",
		Make:        "Honda",
		Model:       "Accord",
		Year:        2023,
		CategoryID:  activeCategory,
		Mileage:     100,
		FuelLevel:   90,
		Location:    "Depot A",
		CreatedBy:   "staff-1",
		UpdatedBy:   "staff-1",
	}
	if err := repo.CreateVehicle(ctx, v); err != nil {
		t.Fatalf("CreateVehicle() error = %v", err)
	}
	if v.ID == "" {
		t.Fatalf("CreateVehicle() left ID empty")
	}
	if v.CreatedAt.IsZero() {
		t.Fatalf("CreateVehicle() left CreatedAt empty")
	}

	dup := *v
	dup.PlateNumber = "TEST-0002"
	if err := repo.CreateVehicle(ctx, &dup); err != ErrDuplicateVIN {
		t.Fatalf("CreateVehicle() duplicate VIN error = %v, want ErrDuplicateVIN", err)
	}

	dup2 := *v
	dup2.VIN = "2HGCM82633A004352"
	if err := repo.CreateVehicle(ctx, &dup2); err != ErrDuplicatePlateNumber {
		t.Fatalf("CreateVehicle() duplicate plate error = %v, want ErrDuplicatePlateNumber", err)
	}
}
