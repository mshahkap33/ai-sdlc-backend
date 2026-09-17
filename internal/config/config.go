// Package config provides configuration loading for the ai-sdlc-backend
// service, sourced from environment variables.
package config

import (
	"net/url"
	"os"
)

// DatabaseConfig holds the settings required to connect to the PostgreSQL
// database used by the car management service.
type DatabaseConfig struct {
	Host     string
	Port     string
	User     string
	Password string
	Name     string
	SSLMode  string
}

// DSN builds a PostgreSQL connection string (DSN) from the configuration.
func (c DatabaseConfig) DSN() string {
	dsn := url.URL{
		Scheme: "postgres",
		User:   url.UserPassword(c.User, c.Password),
		Host:   hostPort(c.Host, c.Port),
		Path:   "/" + c.Name,
	}

	query := url.Values{}
	query.Set("sslmode", c.SSLMode)
	dsn.RawQuery = query.Encode()

	return dsn.String()
}

func hostPort(host, port string) string {
	return host + ":" + port
}

// LoadDatabaseConfig reads database connection settings from environment
// variables, applying sensible defaults when a variable is not set.
func LoadDatabaseConfig() DatabaseConfig {
	return DatabaseConfig{
		Host:     getEnv("DB_HOST", "localhost"),
		Port:     getEnv("DB_PORT", "5432"),
		User:     getEnv("DB_USER", "postgres"),
		Password: getEnv("DB_PASSWORD", "postgres"),
		Name:     getEnv("DB_NAME", "ai_sdlc"),
		SSLMode:  getEnv("DB_SSLMODE", "disable"),
	}
}

func getEnv(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok && value != "" {
		return value
	}
	return fallback
}
