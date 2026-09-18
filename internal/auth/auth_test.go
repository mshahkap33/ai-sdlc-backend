package auth

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

func generateKeyPair(t *testing.T) (*rsa.PrivateKey, string) {
	t.Helper()
	key, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatalf("generating RSA key: %v", err)
	}

	der, err := x509.MarshalPKIXPublicKey(&key.PublicKey)
	if err != nil {
		t.Fatalf("marshalling public key: %v", err)
	}
	pemBytes := pem.EncodeToMemory(&pem.Block{Type: "PUBLIC KEY", Bytes: der})

	return key, string(pemBytes)
}

func signToken(t *testing.T, key *rsa.PrivateKey, c claims) string {
	t.Helper()
	token := jwt.NewWithClaims(jwt.SigningMethodRS256, c)
	signed, err := token.SignedString(key)
	if err != nil {
		t.Fatalf("signing token: %v", err)
	}
	return signed
}

func TestVerifierParseValidToken(t *testing.T) {
	key, pemBytes := generateKeyPair(t)
	publicKey, err := ParseRSAPublicKeyFromPEM(pemBytes)
	if err != nil {
		t.Fatalf("ParseRSAPublicKeyFromPEM() error = %v", err)
	}
	v := NewVerifier(publicKey, "", "")

	token := signToken(t, key, claims{
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   "staff-1",
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour)),
		},
		Role: "service_staff",
	})

	user, err := v.Parse(token)
	if err != nil {
		t.Fatalf("Parse() error = %v, want nil", err)
	}
	if user.Subject != "staff-1" || !user.HasRole("service_staff") {
		t.Fatalf("Parse() user = %+v, want subject staff-1 with role service_staff", user)
	}
}

func TestVerifierParseExpiredToken(t *testing.T) {
	key, pemBytes := generateKeyPair(t)
	publicKey, _ := ParseRSAPublicKeyFromPEM(pemBytes)
	v := NewVerifier(publicKey, "", "")

	token := signToken(t, key, claims{
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   "staff-1",
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(-time.Hour)),
		},
		Role: "service_staff",
	})

	if _, err := v.Parse(token); err == nil {
		t.Fatalf("Parse() error = nil, want error for expired token")
	}
}

func TestVerifierParseWrongKey(t *testing.T) {
	signKey, _ := generateKeyPair(t)
	_, otherPEM := generateKeyPair(t)
	otherPublicKey, _ := ParseRSAPublicKeyFromPEM(otherPEM)
	v := NewVerifier(otherPublicKey, "", "")

	token := signToken(t, signKey, claims{
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   "staff-1",
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour)),
		},
		Role: "service_staff",
	})

	if _, err := v.Parse(token); err == nil {
		t.Fatalf("Parse() error = nil, want error for token signed by a different key")
	}
}

func TestVerifierParseMissingRole(t *testing.T) {
	key, pemBytes := generateKeyPair(t)
	publicKey, _ := ParseRSAPublicKeyFromPEM(pemBytes)
	v := NewVerifier(publicKey, "", "")

	token := signToken(t, key, claims{
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   "staff-1",
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour)),
		},
	})

	if _, err := v.Parse(token); err == nil {
		t.Fatalf("Parse() error = nil, want error for token missing a role claim")
	}
}

func TestMiddlewareRejectsMissingHeader(t *testing.T) {
	_, pemBytes := generateKeyPair(t)
	publicKey, _ := ParseRSAPublicKeyFromPEM(pemBytes)
	v := NewVerifier(publicKey, "", "")

	called := false
	mw := Middleware(v)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called = true
	}))

	req := httptest.NewRequest(http.MethodPost, "/api/v1/vehicles", nil)
	rr := httptest.NewRecorder()
	mw.ServeHTTP(rr, req)

	if rr.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d, want %d", rr.Code, http.StatusUnauthorized)
	}
	if called {
		t.Fatalf("next handler was called, want it to be skipped")
	}
}

func TestMiddlewareAcceptsValidToken(t *testing.T) {
	key, pemBytes := generateKeyPair(t)
	publicKey, _ := ParseRSAPublicKeyFromPEM(pemBytes)
	v := NewVerifier(publicKey, "", "")

	token := signToken(t, key, claims{
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   "staff-1",
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour)),
		},
		Role: "service_staff",
	})

	var gotUser User
	mw := Middleware(v)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotUser, _ = UserFromContext(r.Context())
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodPost, "/api/v1/vehicles", nil)
	req.Header.Set("Authorization", authScheme+" "+token)
	rr := httptest.NewRecorder()
	mw.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", rr.Code, http.StatusOK)
	}
	if gotUser.Subject != "staff-1" {
		t.Fatalf("gotUser.Subject = %q, want %q", gotUser.Subject, "staff-1")
	}
}
