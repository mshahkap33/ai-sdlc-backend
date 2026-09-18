// Package auth provides JWT bearer-token authentication and role-based
// authorization middleware for the REST API, as required by the security
// section of the vehicle status and availability TRD.
package auth

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/golang-jwt/jwt/v5"
)

// Claims is the expected JWT payload: a subject identifying the calling
// user or system account, and the roles granted to it.
type Claims struct {
	Roles []string `json:"roles"`
	jwt.RegisteredClaims
}

// HasRole reports whether the claims include the given role.
func (c Claims) HasRole(role string) bool {
	for _, r := range c.Roles {
		if r == role {
			return true
		}
	}
	return false
}

// ErrMissingToken is returned when no Authorization bearer token is present
// on the request.
var ErrMissingToken = errors.New("missing bearer token")

type contextKey int

const claimsContextKey contextKey = iota

// ClaimsFromContext returns the authenticated JWT claims previously stored
// on the request context by Middleware, if any.
func ClaimsFromContext(ctx context.Context) (Claims, bool) {
	claims, ok := ctx.Value(claimsContextKey).(Claims)
	return claims, ok
}

// Verifier validates JWT bearer tokens using a shared HMAC secret (HS256 or
// stronger), as required by the platform's authentication standard.
type Verifier struct {
	secret []byte
}

// NewVerifier creates a Verifier for the given HMAC signing secret.
func NewVerifier(secret []byte) *Verifier {
	return &Verifier{secret: secret}
}

// Parse validates the given bearer token string and returns its claims.
func (v *Verifier) Parse(tokenString string) (Claims, error) {
	var claims Claims
	token, err := jwt.ParseWithClaims(tokenString, &claims, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
		}
		return v.secret, nil
	})
	if err != nil {
		return Claims{}, err
	}
	if !token.Valid {
		return Claims{}, errors.New("invalid token")
	}
	return claims, nil
}

// RequireRoles returns middleware that authenticates the request's bearer
// token and rejects it unless the caller has at least one of the allowed
// roles. When allowedRoles is empty, any authenticated caller is permitted.
func (v *Verifier) RequireRoles(allowedRoles ...string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			tokenString, err := bearerToken(r)
			if err != nil {
				http.Error(w, "unauthorized: "+err.Error(), http.StatusUnauthorized)
				return
			}

			claims, err := v.Parse(tokenString)
			if err != nil {
				http.Error(w, "unauthorized: invalid token", http.StatusUnauthorized)
				return
			}

			if len(allowedRoles) > 0 && !hasAnyRole(claims, allowedRoles) {
				http.Error(w, "forbidden: missing required role", http.StatusForbidden)
				return
			}

			ctx := context.WithValue(r.Context(), claimsContextKey, claims)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func hasAnyRole(claims Claims, allowed []string) bool {
	for _, role := range allowed {
		if claims.HasRole(role) {
			return true
		}
	}
	return false
}

func bearerToken(r *http.Request) (string, error) {
	header := r.Header.Get("Authorization")
	if header == "" {
		return "", ErrMissingToken
	}
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
