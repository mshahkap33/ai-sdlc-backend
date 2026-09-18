// Package config provides configuration loading for the ai-sdlc-backend
// service, sourced from environment variables (optionally populated from a
// local .env file).
package config

import (
	"net/url"
	"os"
	"sync"

	"github.com/joho/godotenv"
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

var loadEnvFileOnce = &sync.Once{}

// loadEnvFile loads variables from a .env file (defaults) and an optional,
// git-ignored .env.local file (personal overrides) in the current working
// directory into the process environment. Values already present in the
// real process environment always take precedence over both files, and
// .env.local takes precedence over .env. Missing files are ignored. It only
// attempts to load the files once per process.
func loadEnvFile() {
	loadEnvFileOnce.Do(func() {
		// Record which variables were already set in the real environment
		// before loading any files, so file-based values never clobber them.
		preset := map[string]bool{}
		for _, key := range []string{"DB_HOST", "DB_PORT", "DB_USER", "DB_PASSWORD", "DB_NAME", "DB_SSLMODE"} {
			if _, ok := os.LookupEnv(key); ok {
				preset[key] = true
			}
		}

		// godotenv.Load does not override variables already set in the
		// process environment, so this only fills in unset variables.
		_ = godotenv.Load(".env")

		// .env.local should override values from .env but never values that
		// were already present in the real environment, so apply it manually
		// rather than via godotenv.Overload (which would override everything).
		if local, err := godotenv.Read(".env.local"); err == nil {
			for key, value := range local {
				if !preset[key] {
					_ = os.Setenv(key, value)
				}
			}
		}
	})
}

// LoadDatabaseConfig reads database connection settings from environment
// variables (populated from a .env file when present), applying sensible
// defaults when a variable is not set.
func LoadDatabaseConfig() DatabaseConfig {
	loadEnvFile()

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

// AuthConfig holds the settings required to validate the bearer JWTs
// presented to the REST API, per the Vehicle Onboarding TRD security
// requirement.
type AuthConfig struct {
	// JWTPublicKeyPEM is the PEM-encoded RSA public key (or certificate)
	// used to verify the identity provider's JWT signatures.
	JWTPublicKeyPEM string
	// JWTIssuer, when non-empty, is the required JWT "iss" claim value.
	JWTIssuer string
	// JWTAudience, when non-empty, is the required JWT "aud" claim value.
	JWTAudience string
}

// LoadAuthConfig reads JWT verification settings from environment
// variables: AUTH_JWT_PUBLIC_KEY (or AUTH_JWT_PUBLIC_KEY_PATH to read the
// PEM contents from a file), AUTH_JWT_ISSUER, and AUTH_JWT_AUDIENCE.
func LoadAuthConfig() (AuthConfig, error) {
	loadEnvFile()

	pem := os.Getenv("AUTH_JWT_PUBLIC_KEY")
	if pem == "" {
		if path := os.Getenv("AUTH_JWT_PUBLIC_KEY_PATH"); path != "" {
			data, err := os.ReadFile(path)
			if err != nil {
				return AuthConfig{}, err
			}
			pem = string(data)
		}
	}

	return AuthConfig{
		JWTPublicKeyPEM: pem,
		JWTIssuer:       os.Getenv("AUTH_JWT_ISSUER"),
		JWTAudience:     os.Getenv("AUTH_JWT_AUDIENCE"),
	}, nil
}

// ServerConfig holds settings for the HTTP server entrypoint.
type ServerConfig struct {
	Addr string
}

// LoadServerConfig reads HTTP server settings from environment variables,
// applying sensible defaults when a variable is not set.
func LoadServerConfig() ServerConfig {
	loadEnvFile()

	return ServerConfig{
		Addr: getEnv("SERVER_ADDR", ":8080"),
	}
}
