// Package auth provides JWT bearer token authentication and role-based
// authorization middleware for the car management REST API.
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
	// Roles lists the roles/permissions granted to the user.
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

// UserFromContext returns the authenticated user stored in ctx by the
// Authenticate middleware, if any.
func UserFromContext(ctx context.Context) (User, bool) {
	user, ok := ctx.Value(userContextKey).(User)
	return user, ok
}

// ContextWithUser returns a copy of ctx carrying user, retrievable via
// UserFromContext. It is primarily useful in tests that need to simulate an
// authenticated request without going through the Authenticate middleware.
func ContextWithUser(ctx context.Context, user User) context.Context {
	return context.WithValue(ctx, userContextKey, user)
}

// Verifier validates JWT bearer tokens issued by the platform's identity
// provider.
type Verifier struct {
	publicKey *rsa.PublicKey
	issuer    string
	audience  string
}

// NewVerifier creates a Verifier that validates RS256-signed tokens against
// publicKey, requiring the given issuer and audience claims.
func NewVerifier(publicKey *rsa.PublicKey, issuer, audience string) *Verifier {
	return &Verifier{publicKey: publicKey, issuer: issuer, audience: audience}
}

// claims models the JWT payload fields required by the platform.
type claims struct {
	jwt.RegisteredClaims
	Roles []string `json:"roles"`
}

// ErrMissingToken is returned when the Authorization header is absent or
// malformed.
var ErrMissingToken = errors.New("auth: missing or malformed bearer token")

// Parse validates the given raw JWT and returns the authenticated user.
func (v *Verifier) Parse(rawToken string) (User, error) {
	parsed, err := jwt.ParseWithClaims(rawToken, &claims{}, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodRSA); !ok {
			return nil, jwt.ErrTokenSignatureInvalid
		}
		return v.publicKey, nil
	},
		jwt.WithIssuer(v.issuer),
		jwt.WithAudience(v.audience),
		jwt.WithValidMethods([]string{"RS256"}),
	)
	if err != nil {
		return User{}, err
	}

	c, ok := parsed.Claims.(*claims)
	if !ok || !parsed.Valid {
		return User{}, jwt.ErrTokenInvalidClaims
	}

	return User{Subject: c.Subject, Roles: c.Roles}, nil
}

// Authenticate returns middleware that validates the bearer token on
// incoming requests and stores the resulting User in the request context.
// Requests without a valid token receive a 401 Unauthorized response.
func (v *Verifier) Authenticate(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		token, err := bearerToken(r)
		if err != nil {
			http.Error(w, "missing or malformed Authorization header", http.StatusUnauthorized)
			return
		}

		user, err := v.Parse(token)
		if err != nil {
			http.Error(w, "invalid or expired token", http.StatusUnauthorized)
			return
		}

		ctx := context.WithValue(r.Context(), userContextKey, user)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func bearerToken(r *http.Request) (string, error) {
	header := r.Header.Get("Authorization")
	const prefix = "Bearer "
	if !strings.HasPrefix(header, prefix) {
		return "", ErrMissingToken
	}

	token := strings.TrimSpace(strings.TrimPrefix(header, prefix))
	if token == "" {
		return "", ErrMissingToken
	}

	return token, nil
}

// RequireRole returns middleware that ensures the authenticated user (set
// by Authenticate) has the given role, responding with 403 Forbidden
// otherwise. It must be applied after Authenticate.
func RequireRole(role string, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		user, ok := UserFromContext(r.Context())
		if !ok || !user.HasRole(role) {
			http.Error(w, "forbidden: requires role "+role, http.StatusForbidden)
			return
		}
		next.ServeHTTP(w, r)
	})
}
