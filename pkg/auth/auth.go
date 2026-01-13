package auth

import (
	"context"
	"os"
	"strings"
)

// AuthHeaderName is the HTTP header used for MCP authentication
const AuthHeaderName = "X-MCP-Auth-Token"

// ValidateToken validates the provided authentication token.
func ValidateToken(token string) bool {
	return token != ""
}

// GetExpectedToken returns the expected token from environment variable.
func GetExpectedToken() string {
	return os.Getenv("MCP_AUTH_TOKEN")
}

// IsAuthEnabled returns true if MCP authentication is enabled (token is configured)
func IsAuthEnabled() bool {
	return GetExpectedToken() != ""
}

// ValidateAgainstExpected validates the provided token against the expected token.
func ValidateAgainstExpected(providedToken string) bool {
	expectedToken := GetExpectedToken()
	if expectedToken == "" {
		return true
	}
	return providedToken == expectedToken
}

// TokenAuthorizer implements Authorizer using the MCP_AUTH_TOKEN environment variable
type TokenAuthorizer struct{}

// Authorize validates the token against the expected MCP_AUTH_TOKEN
func (t *TokenAuthorizer) Authorize(ctx context.Context, token string) (bool, error) {
	// Extract token from "Bearer <token>" format if present
	if strings.HasPrefix(token, "Bearer ") {
		token = strings.TrimPrefix(token, "Bearer ")
	}
	return ValidateAgainstExpected(token), nil
}

// NewTokenAuthorizer creates a new TokenAuthorizer
func NewTokenAuthorizer() *TokenAuthorizer {
	return &TokenAuthorizer{}
}
