// Package config provides configuration loading for the ai-sdlc-backend
// service, sourced from environment variables (optionally populated from a
// local .env file).
package config

import (
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"errors"
	"fmt"
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

// AuthConfig holds the settings required to verify JWT bearer tokens
// presented to the REST API.
type AuthConfig struct {
	// PublicKeyPEM is the PEM-encoded RSA public key used to verify RS256
	// token signatures.
	PublicKeyPEM string
	Issuer       string
	Audience     string
}

// LoadAuthConfig reads JWT verification settings from environment
// variables (populated from a .env file when present):
//
//   - AUTH_JWT_PUBLIC_KEY: PEM-encoded RSA public key content.
//   - AUTH_JWT_PUBLIC_KEY_PATH: path to a file containing the PEM-encoded
//     RSA public key, used when AUTH_JWT_PUBLIC_KEY is not set.
//   - AUTH_JWT_ISSUER: expected token issuer.
//   - AUTH_JWT_AUDIENCE: expected token audience.
func LoadAuthConfig() (AuthConfig, error) {
	loadEnvFile()

	cfg := AuthConfig{
		PublicKeyPEM: os.Getenv("AUTH_JWT_PUBLIC_KEY"),
		Issuer:       os.Getenv("AUTH_JWT_ISSUER"),
		Audience:     os.Getenv("AUTH_JWT_AUDIENCE"),
	}

	if cfg.PublicKeyPEM == "" {
		if path := os.Getenv("AUTH_JWT_PUBLIC_KEY_PATH"); path != "" {
			data, err := os.ReadFile(path)
			if err != nil {
				return AuthConfig{}, fmt.Errorf("reading AUTH_JWT_PUBLIC_KEY_PATH: %w", err)
			}
			cfg.PublicKeyPEM = string(data)
		}
	}

	return cfg, nil
}

// PublicKey parses the configured PEM-encoded RSA public key.
func (c AuthConfig) PublicKey() (*rsa.PublicKey, error) {
	block, _ := pem.Decode([]byte(c.PublicKeyPEM))
	if block == nil {
		return nil, errors.New("config: no PEM block found in AUTH_JWT_PUBLIC_KEY")
	}

	if pub, err := x509.ParsePKIXPublicKey(block.Bytes); err == nil {
		rsaKey, ok := pub.(*rsa.PublicKey)
		if !ok {
			return nil, errors.New("config: AUTH_JWT_PUBLIC_KEY is not an RSA public key")
		}
		return rsaKey, nil
	}

	cert, err := x509.ParseCertificate(block.Bytes)
	if err == nil {
		rsaKey, ok := cert.PublicKey.(*rsa.PublicKey)
		if !ok {
			return nil, errors.New("config: AUTH_JWT_PUBLIC_KEY certificate does not contain an RSA public key")
		}
		return rsaKey, nil
	}

	return nil, fmt.Errorf("config: failed to parse AUTH_JWT_PUBLIC_KEY: %w", err)
}
