package config

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"os"
	"path/filepath"
	"testing"
)

func generateTestKeyPEM(t *testing.T) string {
	t.Helper()
	key, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatalf("generate key: %v", err)
	}
	der, err := x509.MarshalPKIXPublicKey(&key.PublicKey)
	if err != nil {
		t.Fatalf("marshal public key: %v", err)
	}
	block := &pem.Block{Type: "PUBLIC KEY", Bytes: der}
	return string(pem.EncodeToMemory(block))
}

func TestLoadAuthConfigFromEnvVar(t *testing.T) {
	pemKey := generateTestKeyPEM(t)
	t.Setenv("AUTH_JWT_PUBLIC_KEY", pemKey)
	t.Setenv("AUTH_JWT_PUBLIC_KEY_PATH", "")
	t.Setenv("AUTH_JWT_ISSUER", "issuer")
	t.Setenv("AUTH_JWT_AUDIENCE", "audience")

	cfg, err := LoadAuthConfig()
	if err != nil {
		t.Fatalf("LoadAuthConfig() error = %v", err)
	}

	if cfg.Issuer != "issuer" || cfg.Audience != "audience" {
		t.Fatalf("cfg = %+v, want issuer/audience set", cfg)
	}

	if _, err := cfg.PublicKey(); err != nil {
		t.Fatalf("PublicKey() error = %v", err)
	}
}

func TestLoadAuthConfigFromPublicKeyPath(t *testing.T) {
	pemKey := generateTestKeyPEM(t)
	dir := t.TempDir()
	path := filepath.Join(dir, "key.pem")
	if err := os.WriteFile(path, []byte(pemKey), 0o600); err != nil {
		t.Fatalf("write key file: %v", err)
	}

	t.Setenv("AUTH_JWT_PUBLIC_KEY", "")
	t.Setenv("AUTH_JWT_PUBLIC_KEY_PATH", path)
	t.Setenv("AUTH_JWT_ISSUER", "issuer")
	t.Setenv("AUTH_JWT_AUDIENCE", "audience")

	cfg, err := LoadAuthConfig()
	if err != nil {
		t.Fatalf("LoadAuthConfig() error = %v", err)
	}

	if _, err := cfg.PublicKey(); err != nil {
		t.Fatalf("PublicKey() error = %v", err)
	}
}

func TestAuthConfigPublicKeyInvalidPEM(t *testing.T) {
	cfg := AuthConfig{PublicKeyPEM: "not a pem"}
	if _, err := cfg.PublicKey(); err == nil {
		t.Fatal("expected error for invalid PEM, got nil")
	}
}
