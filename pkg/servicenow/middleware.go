package servicenow

import (
	"net/http"
)

const (
	// Header names for ServiceNow credentials
	HeaderUsername = "X-ServiceNow-Username"
	HeaderPassword = "X-ServiceNow-Password"
	HeaderAPIKey   = "X-ServiceNow-API-Key"
)

// CredentialsMiddleware extracts ServiceNow credentials from request headers
// and adds them to the request context
type CredentialsMiddleware struct{}

// NewCredentialsMiddleware creates a new credentials middleware
func NewCredentialsMiddleware() *CredentialsMiddleware {
	return &CredentialsMiddleware{}
}

// Wrap wraps an http.Handler with credentials extraction
func (m *CredentialsMiddleware) Wrap(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Extract credentials from headers if present
		username := r.Header.Get(HeaderUsername)
		password := r.Header.Get(HeaderPassword)
		apiKey := r.Header.Get(HeaderAPIKey)

		// Only add to context if at least one credential header is present
		if username != "" || password != "" || apiKey != "" {
			creds := &ContextCredentials{
				Username: username,
				Password: password,
				APIKey:   apiKey,
			}
			r = r.WithContext(ContextWithCredentials(r.Context(), creds))
		}

		next.ServeHTTP(w, r)
	})
}

// WrapFunc wraps an http.HandlerFunc with credentials extraction
func (m *CredentialsMiddleware) WrapFunc(next http.HandlerFunc) http.Handler {
	return m.Wrap(next)
}
