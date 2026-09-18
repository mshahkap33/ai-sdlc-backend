package auth

import (
	"crypto/rand"
	"crypto/rsa"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

func generateKeyPair(t *testing.T) *rsa.PrivateKey {
	t.Helper()
	key, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatalf("generate key: %v", err)
	}
	return key
}

func signToken(t *testing.T, key *rsa.PrivateKey, c claims) string {
	t.Helper()
	token := jwt.NewWithClaims(jwt.SigningMethodRS256, c)
	signed, err := token.SignedString(key)
	if err != nil {
		t.Fatalf("sign token: %v", err)
	}
	return signed
}

func TestVerifierParseValidToken(t *testing.T) {
	key := generateKeyPair(t)
	v := NewVerifier(&key.PublicKey, "issuer", "audience")

	c := claims{
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   "user-1",
			Issuer:    "issuer",
			Audience:  jwt.ClaimStrings{"audience"},
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour)),
		},
		Roles: []string{"service_staff"},
	}
	token := signToken(t, key, c)

	user, err := v.Parse(token)
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}
	if user.Subject != "user-1" {
		t.Errorf("Subject = %q, want %q", user.Subject, "user-1")
	}
	if !user.HasRole("service_staff") {
		t.Errorf("expected user to have role service_staff, got %v", user.Roles)
	}
}

func TestVerifierParseExpiredToken(t *testing.T) {
	key := generateKeyPair(t)
	v := NewVerifier(&key.PublicKey, "issuer", "audience")

	c := claims{
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   "user-1",
			Issuer:    "issuer",
			Audience:  jwt.ClaimStrings{"audience"},
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(-time.Hour)),
		},
		Roles: []string{"service_staff"},
	}
	token := signToken(t, key, c)

	if _, err := v.Parse(token); err == nil {
		t.Fatal("expected error for expired token, got nil")
	}
}

func TestVerifierParseWrongIssuer(t *testing.T) {
	key := generateKeyPair(t)
	v := NewVerifier(&key.PublicKey, "issuer", "audience")

	c := claims{
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   "user-1",
			Issuer:    "wrong-issuer",
			Audience:  jwt.ClaimStrings{"audience"},
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour)),
		},
	}
	token := signToken(t, key, c)

	if _, err := v.Parse(token); err == nil {
		t.Fatal("expected error for wrong issuer, got nil")
	}
}

func TestAuthenticateMissingHeader(t *testing.T) {
	key := generateKeyPair(t)
	v := NewVerifier(&key.PublicKey, "issuer", "audience")

	handler := v.Authenticate(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Fatal("handler should not be called")
	}))

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusUnauthorized)
	}
}

func TestRequireRoleForbidden(t *testing.T) {
	key := generateKeyPair(t)
	v := NewVerifier(&key.PublicKey, "issuer", "audience")

	c := claims{
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   "user-1",
			Issuer:    "issuer",
			Audience:  jwt.ClaimStrings{"audience"},
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour)),
		},
		Roles: []string{"customer"},
	}
	token := signToken(t, key, c)

	called := false
	handler := v.Authenticate(RequireRole("service_staff", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called = true
	})))

	req := httptest.NewRequest(http.MethodPost, "/", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusForbidden {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusForbidden)
	}
	if called {
		t.Error("expected handler not to be called")
	}
}

func TestRequireAnyRoleAllowsMatchingRole(t *testing.T) {
	key := generateKeyPair(t)
	v := NewVerifier(&key.PublicKey, "issuer", "audience")

	c := claims{
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   "user-1",
			Issuer:    "issuer",
			Audience:  jwt.ClaimStrings{"audience"},
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour)),
		},
		Roles: []string{"operations_manager"},
	}
	token := signToken(t, key, c)

	called := false
	handler := v.Authenticate(RequireAnyRole([]string{"service_staff", "operations_manager"}, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called = true
	})))

	req := httptest.NewRequest(http.MethodPatch, "/", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusOK)
	}
	if !called {
		t.Error("expected handler to be called")
	}
}

func TestRequireAnyRoleForbidden(t *testing.T) {
	key := generateKeyPair(t)
	v := NewVerifier(&key.PublicKey, "issuer", "audience")

	c := claims{
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   "user-1",
			Issuer:    "issuer",
			Audience:  jwt.ClaimStrings{"audience"},
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour)),
		},
		Roles: []string{"customer"},
	}
	token := signToken(t, key, c)

	handler := v.Authenticate(RequireAnyRole([]string{"service_staff", "operations_manager"}, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Fatal("handler should not be called")
	})))

	req := httptest.NewRequest(http.MethodPatch, "/", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusForbidden {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusForbidden)
	}
}
