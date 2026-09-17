// Package db provides helpers to run database schema migrations for the
// car management service against a PostgreSQL database.
package db

import (
	"database/sql"
	"errors"
	"fmt"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/pgx/v5"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	_ "github.com/jackc/pgx/v5/stdlib"
)

// MigrationsPath is the default filesystem location of the migration files
// relative to the repository root.
const MigrationsPath = "db/migrations"

// Migrate applies all pending "up" migrations found in migrationsPath to the
// database identified by dsn. It returns nil if the schema is already
// up to date.
func Migrate(dsn, migrationsPath string) error {
	m, closeFn, err := newMigrate(dsn, migrationsPath)
	if err != nil {
		return err
	}
	defer closeFn()

	if err := m.Up(); err != nil && !errors.Is(err, migrate.ErrNoChange) {
		return fmt.Errorf("running migrations: %w", err)
	}

	return nil
}

// Rollback reverts the most recently applied "up" migration(s), stepping
// back the given number of migrations found in migrationsPath.
func Rollback(dsn, migrationsPath string, steps int) error {
	m, closeFn, err := newMigrate(dsn, migrationsPath)
	if err != nil {
		return err
	}
	defer closeFn()

	if err := m.Steps(-steps); err != nil && !errors.Is(err, migrate.ErrNoChange) {
		return fmt.Errorf("rolling back migrations: %w", err)
	}

	return nil
}

func newMigrate(dsn, migrationsPath string) (*migrate.Migrate, func(), error) {
	conn, err := sql.Open("pgx", dsn)
	if err != nil {
		return nil, nil, fmt.Errorf("opening database connection: %w", err)
	}

	driver, err := pgx.WithInstance(conn, &pgx.Config{})
	if err != nil {
		conn.Close()
		return nil, nil, fmt.Errorf("creating migration driver: %w", err)
	}

	m, err := migrate.NewWithDatabaseInstance("file://"+migrationsPath, "pgx", driver)
	if err != nil {
		conn.Close()
		return nil, nil, fmt.Errorf("creating migration instance: %w", err)
	}

	closeFn := func() {
		_ = conn.Close()
	}

	return m, closeFn, nil
}
