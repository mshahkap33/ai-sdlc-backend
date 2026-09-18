package auth

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

func signToken(t *testing.T, secret []byte, subject string, roles []string, expiresIn time.Duration) string {
	t.Helper()
	claims := Claims{
		Roles: roles,
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   subject,
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(expiresIn)),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signed, err := token.SignedString(secret)
	if err != nil {
		t.Fatalf("signing token: %v", err)
	}
	return signed
}

func TestRequireRolesAllowsAuthorizedRole(t *testing.T) {
	secret := []byte("test-secret")
	verifier := NewVerifier(secret)
	token := signToken(t, secret, "system-1", []string{"system_service"}, time.Hour)

	var gotSubject string
	handler := verifier.RequireRoles("system_service")(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		claims, _ := ClaimsFromContext(r.Context())
		gotSubject = claims.Subject
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodPost, "/", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusOK)
	}
	if gotSubject != "system-1" {
		t.Errorf("subject = %q, want %q", gotSubject, "system-1")
	}
}

func TestRequireRolesRejectsMissingToken(t *testing.T) {
	verifier := NewVerifier([]byte("test-secret"))
	handler := verifier.RequireRoles("system_service")(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodPost, "/", nil)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusUnauthorized)
	}
}

func TestRequireRolesRejectsWrongRole(t *testing.T) {
	secret := []byte("test-secret")
	verifier := NewVerifier(secret)
	token := signToken(t, secret, "user-1", []string{"service_staff"}, time.Hour)

	handler := verifier.RequireRoles("system_service")(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodPost, "/", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusForbidden {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusForbidden)
	}
}

func TestRequireRolesRejectsExpiredToken(t *testing.T) {
	secret := []byte("test-secret")
	verifier := NewVerifier(secret)
	token := signToken(t, secret, "user-1", []string{"system_service"}, -time.Hour)

	handler := verifier.RequireRoles("system_service")(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodPost, "/", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusUnauthorized)
	}
}

func TestRequireRolesRejectsWrongSecret(t *testing.T) {
	token := signToken(t, []byte("other-secret"), "user-1", []string{"system_service"}, time.Hour)
	verifier := NewVerifier([]byte("test-secret"))

	handler := verifier.RequireRoles("system_service")(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodPost, "/", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusUnauthorized)
	}
}
