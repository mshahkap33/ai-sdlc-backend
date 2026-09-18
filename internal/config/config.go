// Package config provides application configuration loaded from environment
// variables, including the database connection settings used by the
// migration tooling.
package config

import (
	"fmt"
	"net/url"
	"os"
)

// DatabaseConfig holds the settings required to connect to the PostgreSQL
// database.
type DatabaseConfig struct {
	Host     string
	Port     string
	User     string
	Password string
	Name     string
	SSLMode  string
}

// LoadDatabaseConfig builds a DatabaseConfig from environment variables,
// falling back to sensible local-development defaults when a variable is
// not set.
func LoadDatabaseConfig() DatabaseConfig {
	return DatabaseConfig{
		Host:     getEnv("DB_HOST", "localhost"),
		Port:     getEnv("DB_PORT", "5432"),
		User:     getEnv("DB_USER", "postgres"),
		Password: getEnv("DB_PASSWORD", ""),
		Name:     getEnv("DB_NAME", "ai_sdlc"),
		SSLMode:  getEnv("DB_SSLMODE", "disable"),
	}
}

// DSN builds the PostgreSQL connection string (DSN) for this configuration
// using net/url so that special characters in the credentials are properly
// escaped.
func (c DatabaseConfig) DSN() string {
	u := url.URL{
		Scheme: "postgres",
		User:   url.UserPassword(c.User, c.Password),
		Host:   fmt.Sprintf("%s:%s", c.Host, c.Port),
		Path:   "/" + c.Name,
	}

	q := u.Query()
	q.Set("sslmode", c.SSLMode)
	u.RawQuery = q.Encode()

	return u.String()
}

func getEnv(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok && value != "" {
		return value
	}
	return fallback
}
