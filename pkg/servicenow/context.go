package servicenow

import (
	"context"
)

// contextKey is a custom type for context keys to avoid collisions
type contextKey string

const (
	// CredentialsContextKey is the context key for ServiceNow credentials
	CredentialsContextKey contextKey = "servicenow_credentials"
)

// ContextCredentials holds ServiceNow credentials from request headers
type ContextCredentials struct {
	Username string
	Password string
	APIKey   string
}

// CredentialsFromContext retrieves ServiceNow credentials from context
func CredentialsFromContext(ctx context.Context) *ContextCredentials {
	if creds, ok := ctx.Value(CredentialsContextKey).(*ContextCredentials); ok {
		return creds
	}
	return nil
}

// ContextWithCredentials adds ServiceNow credentials to context
func ContextWithCredentials(ctx context.Context, creds *ContextCredentials) context.Context {
	return context.WithValue(ctx, CredentialsContextKey, creds)
}
