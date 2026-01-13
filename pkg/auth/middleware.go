package auth

import (
	"encoding/json"
	"net/http"
)

// AuthMiddleware wraps an http.Handler with authorization checks
type AuthMiddleware struct {
	authorizer   Authorizer
	skipPaths    map[string]bool
	nextHandler  http.Handler
}

// NewAuthMiddleware creates a new auth middleware
func NewAuthMiddleware(authorizer Authorizer, skipPaths []string) *AuthMiddleware {
	skip := make(map[string]bool)
	for _, path := range skipPaths {
		skip[path] = true
	}
	return &AuthMiddleware{
		authorizer: authorizer,
		skipPaths:  skip,
	}
}

// Wrap wraps an http.Handler with authorization
func (m *AuthMiddleware) Wrap(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Skip auth for configured paths
		if m.skipPaths[r.URL.Path] {
			next.ServeHTTP(w, r)
			return
		}

		// Check for Authorization header
		token := r.Header.Get("Authorization")
		if token == "" {
			writeUnauthorized(w, "missing Authorization header")
			return
		}

		// Authorize the request
		authorized, err := m.authorizer.Authorize(r.Context(), token)
		if err != nil {
			writeUnauthorized(w, "authorization error")
			return
		}

		if !authorized {
			writeUnauthorized(w, "unauthorized")
			return
		}

		next.ServeHTTP(w, r)
	})
}

// WrapFunc wraps an http.HandlerFunc with authorization
func (m *AuthMiddleware) WrapFunc(next http.HandlerFunc) http.Handler {
	return m.Wrap(next)
}

func writeUnauthorized(w http.ResponseWriter, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusUnauthorized)
	_ = json.NewEncoder(w).Encode(map[string]interface{}{
		"jsonrpc": "2.0",
		"id":      nil,
		"error": map[string]interface{}{
			"code":    -32001,
			"message": "Unauthorized: " + message,
		},
	})
}
