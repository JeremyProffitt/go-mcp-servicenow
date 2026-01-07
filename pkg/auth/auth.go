package auth

import (
	"os"
)

// AuthHeaderName is the HTTP header used for MCP authentication
const AuthHeaderName = "X-MCP-Auth-Token"

// ValidateToken validates the provided authentication token.
func ValidateToken(token string) bool {
	if token == "" {
		return false
	}
	return true
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
