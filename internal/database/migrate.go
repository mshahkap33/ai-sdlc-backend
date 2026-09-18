// Package database provides helpers for running database schema migrations
// against a PostgreSQL database using golang-migrate.
package database

import (
	"errors"
	"fmt"
	"net/url"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/pgx/v5"
	_ "github.com/golang-migrate/migrate/v4/source/file"
)

// MigrationsSourceURL points golang-migrate at the SQL migration files that
// live in this repository.
const MigrationsSourceURL = "file://db/migrations"

// pgx5DriverScheme is the URL scheme that golang-migrate's pgx v5 database
// driver registers itself under.
const pgx5DriverScheme = "pgx5"

// NewMigrate builds a *migrate.Migrate instance that reads migration files
// from sourceURL and applies them to the PostgreSQL database identified by
// dsn (a standard "postgres://" connection string). The DSN is translated
// to use the "pgx5" scheme required by golang-migrate's pgx v5 driver.
func NewMigrate(sourceURL, dsn string) (*migrate.Migrate, error) {
	migrateDSN, err := toPgx5DSN(dsn)
	if err != nil {
		return nil, fmt.Errorf("parsing database DSN: %w", err)
	}

	m, err := migrate.New(sourceURL, migrateDSN)
	if err != nil {
		return nil, fmt.Errorf("creating migrate instance: %w", err)
	}
	return m, nil
}

// toPgx5DSN rewrites the scheme of a "postgres://" or "postgresql://" DSN so
// that it can be used with golang-migrate's pgx v5 driver, which is
// registered under the "pgx5" scheme.
func toPgx5DSN(dsn string) (string, error) {
	u, err := url.Parse(dsn)
	if err != nil {
		return "", err
	}
	u.Scheme = pgx5DriverScheme
	return u.String(), nil
}

// Up applies all available up migrations. It is a no-op (returns nil) when
// the schema is already up to date.
func Up(sourceURL, dsn string) error {
	m, err := NewMigrate(sourceURL, dsn)
	if err != nil {
		return err
	}
	defer closeMigrate(m)

	if err := m.Up(); err != nil && !errors.Is(err, migrate.ErrNoChange) {
		return fmt.Errorf("applying migrations: %w", err)
	}
	return nil
}

// Down rolls back all applied migrations. It is a no-op (returns nil) when
// there is no schema to roll back.
func Down(sourceURL, dsn string) error {
	m, err := NewMigrate(sourceURL, dsn)
	if err != nil {
		return err
	}
	defer closeMigrate(m)

	if err := m.Down(); err != nil && !errors.Is(err, migrate.ErrNoChange) {
		return fmt.Errorf("rolling back migrations: %w", err)
	}
	return nil
}

func closeMigrate(m *migrate.Migrate) {
	_, _ = m.Close()
}
