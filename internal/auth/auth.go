// Package auth implements bearer-token (JWT) authentication and role-based
// authorization for the Vehicle Onboarding REST API, as required by the
// "Security Requirement" section of the Vehicle Onboarding TRD.
package auth

import (
	"context"
	"crypto/rsa"
	"errors"
	"net/http"
	"strings"

	"github.com/golang-jwt/jwt/v5"
)

// User represents the authenticated caller extracted from a validated JWT.
type User struct {
	// Subject is the staff user identifier (JWT "sub" claim).
	Subject string
	// Roles are the role/permission claims granted to the caller.
	Roles []string
}

// HasRole reports whether the user has been granted the given role.
func (u User) HasRole(role string) bool {
	for _, r := range u.Roles {
		if r == role {
			return true
		}
	}
	return false
}

type contextKey int

const userContextKey contextKey = iota

// UserFromContext returns the authenticated User stored in ctx by
// Middleware, if any.
func UserFromContext(ctx context.Context) (User, bool) {
	u, ok := ctx.Value(userContextKey).(User)
	return u, ok
}

// NewTestContext returns a copy of ctx carrying user, as Middleware would
// after successfully validating a bearer token. It is exported solely to
// let other packages (e.g. vehicle's handler tests) exercise
// authentication-aware code without needing to mint real JWTs.
func NewTestContext(ctx context.Context, user User) context.Context {
	return context.WithValue(ctx, userContextKey, user)
}

// claims is the expected shape of the JWT payload described by the Vehicle
// Onboarding TRD: subject, role/permissions, issuer, audience, issued-at and
// expiration.
type claims struct {
	jwt.RegisteredClaims
	Role  string   `json:"role"`
	Roles []string `json:"roles"`
}

func (c claims) roles() []string {
	if len(c.Roles) > 0 {
		return c.Roles
	}
	if c.Role != "" {
		return []string{c.Role}
	}
	return nil
}

// Verifier validates the signature and standard claims of bearer tokens.
type Verifier struct {
	publicKey *rsa.PublicKey
	issuer    string
	audience  string
}

// NewVerifier builds a Verifier that checks tokens are signed with
// publicKey using an RS256-family algorithm, and, when non-empty, that the
// issuer and audience claims match the given values.
func NewVerifier(publicKey *rsa.PublicKey, issuer, audience string) *Verifier {
	return &Verifier{publicKey: publicKey, issuer: issuer, audience: audience}
}

// ErrInvalidToken is returned when the bearer token is missing, malformed,
// unsigned by the configured key, or missing required claims.
var ErrInvalidToken = errors.New("invalid or missing bearer token")

// Parse validates tokenString and returns the authenticated User it
// describes.
func (v *Verifier) Parse(tokenString string) (User, error) {
	opts := []jwt.ParserOption{jwt.WithValidMethods([]string{"RS256", "RS384", "RS512"})}
	if v.issuer != "" {
		opts = append(opts, jwt.WithIssuer(v.issuer))
	}
	if v.audience != "" {
		opts = append(opts, jwt.WithAudience(v.audience))
	}

	var c claims
	_, err := jwt.ParseWithClaims(tokenString, &c, func(t *jwt.Token) (any, error) {
		return v.publicKey, nil
	}, opts...)
	if err != nil {
		return User{}, ErrInvalidToken
	}

	if c.Subject == "" || len(c.roles()) == 0 {
		return User{}, ErrInvalidToken
	}

	return User{Subject: c.Subject, Roles: c.roles()}, nil
}

// Middleware authenticates each request using its Authorization header
// (a bearer-scheme token, per RFC 6750), rejecting the request with 401
// Unauthorized when the token is missing or invalid. On success, the
// authenticated User is stored in the request context for downstream
// handlers (see UserFromContext).
func Middleware(v *Verifier) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			token, ok := bearerToken(r)
			if !ok {
				unauthorized(w)
				return
			}

			user, err := v.Parse(token)
			if err != nil {
				unauthorized(w)
				return
			}

			ctx := context.WithValue(r.Context(), userContextKey, user)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// authScheme is the RFC 6750 authorization scheme name for bearer tokens.
const authScheme = "Bearer"

func bearerToken(r *http.Request) (string, bool) {
	header := r.Header.Get("Authorization")
	prefix := authScheme + " "
	if !strings.HasPrefix(header, prefix) {
		return "", false
	}
	token := strings.TrimSpace(strings.TrimPrefix(header, prefix))
	if token == "" {
		return "", false
	}
	return token, true
}

func unauthorized(w http.ResponseWriter) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("WWW-Authenticate", authScheme+` realm="ai-sdlc-backend"`)
	w.WriteHeader(http.StatusUnauthorized)
	_, _ = w.Write([]byte(`{"error":"unauthorized: missing or invalid bearer token"}`))
}
