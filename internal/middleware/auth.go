package middleware

import (
	"context"
	"errors"
	"net/http"
	"strings"

	"github.com/RSSembly/rssembly/internal/auth"
)

// ErrUnauthorized is returned when authentication fails.
var ErrUnauthorized = errors.New("unauthorized")

// AuthenticationHandler defines the interface needed by the auth middleware.
type AuthenticationHandler interface {
	VerifyToken(tokenString string) (*auth.AuthenticatedUser, error)
	APIKeyLookup(ctx context.Context, prefix, hash string) (*auth.AuthenticatedUser, error)
}

// Auth enforces JWT or API key authentication on requests.
type Auth struct {
	auth AuthenticationHandler
}

// NewAuth creates an Auth middleware handler.
func NewAuth(auth AuthenticationHandler) *Auth {
	return &Auth{auth: auth}
}

// Middleware returns an HTTP middleware that authenticates requests.
// It checks the Authorization: Bearer <token> header for JWT, then falls back
// to the X-API-Key header for API key authentication.
func (a *Auth) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		user, err := a.authenticate(r)
		if err != nil {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusUnauthorized)
			_, _ = w.Write([]byte(`{"error":{"code":"unauthorized","message":"invalid or missing credentials"}}`))
			return
		}

		ctx := auth.ContextWithUser(r.Context(), user)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// Optional is like Middleware but does not reject unauthenticated requests.
// It sets the authenticated user on the context if a valid credential is present.
func (a *Auth) Optional(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		user, err := a.authenticate(r)
		if err == nil && user != nil {
			ctx := auth.ContextWithUser(r.Context(), user)
			next.ServeHTTP(w, r.WithContext(ctx))
			return
		}
		next.ServeHTTP(w, r)
	})
}

func (a *Auth) authenticate(r *http.Request) (*auth.AuthenticatedUser, error) {
	// Try JWT via Authorization header.
	authHeader := r.Header.Get("Authorization")
	if strings.HasPrefix(authHeader, "Bearer ") {
		token := strings.TrimPrefix(authHeader, "Bearer ")
		return a.auth.VerifyToken(token)
	}

	// Try API key via X-API-Key header.
	apiKey := r.Header.Get("X-API-Key")
	if apiKey != "" {
		return a.auth.APIKeyLookup(r.Context(), apiKey[:8], apiKey)
	}

	return nil, ErrUnauthorized
}
