package servicenow

import (
	"fmt"
	"os"
	"strconv"
	"strings"
)

// AuthType represents the authentication type for ServiceNow
type AuthType string

const (
	AuthTypeBasic  AuthType = "basic"
	AuthTypeOAuth  AuthType = "oauth"
	AuthTypeAPIKey AuthType = "api_key"
)

// BasicAuthConfig holds basic authentication credentials
type BasicAuthConfig struct {
	Username string
	Password string
}

// OAuthConfig holds OAuth authentication configuration
type OAuthConfig struct {
	ClientID     string
	ClientSecret string
	Username     string
	Password     string
	TokenURL     string
}

// APIKeyConfig holds API key authentication configuration
type APIKeyConfig struct {
	APIKey     string
	HeaderName string
}

// AuthConfig holds the authentication configuration
type AuthConfig struct {
	Type   AuthType
	Basic  *BasicAuthConfig
	OAuth  *OAuthConfig
	APIKey *APIKeyConfig
}

// Config holds the ServiceNow server configuration
type Config struct {
	InstanceURL string
	Auth        AuthConfig
	Debug       bool
	Timeout     int
}

// APIURL returns the base API URL for ServiceNow
func (c *Config) APIURL() string {
	return fmt.Sprintf("%s/api/now", strings.TrimSuffix(c.InstanceURL, "/"))
}

// LoadConfigFromEnv loads configuration from environment variables
func LoadConfigFromEnv() (*Config, error) {
	instanceURL := os.Getenv("SERVICENOW_INSTANCE_URL")
	if instanceURL == "" {
		return nil, fmt.Errorf("SERVICENOW_INSTANCE_URL is required")
	}

	authType := AuthType(strings.ToLower(os.Getenv("SERVICENOW_AUTH_TYPE")))
	if authType == "" {
		authType = AuthTypeBasic
	}

	timeout := 30
	if t := os.Getenv("SERVICENOW_TIMEOUT"); t != "" {
		if parsed, err := strconv.Atoi(t); err == nil {
			timeout = parsed
		}
	}

	debug := strings.ToLower(os.Getenv("SERVICENOW_DEBUG")) == "true"

	config := &Config{
		InstanceURL: instanceURL,
		Debug:       debug,
		Timeout:     timeout,
		Auth: AuthConfig{
			Type: authType,
		},
	}

	switch authType {
	case AuthTypeBasic:
		username := os.Getenv("SERVICENOW_USERNAME")
		password := os.Getenv("SERVICENOW_PASSWORD")
		if username == "" || password == "" {
			return nil, fmt.Errorf("SERVICENOW_USERNAME and SERVICENOW_PASSWORD are required for basic auth")
		}
		config.Auth.Basic = &BasicAuthConfig{
			Username: username,
			Password: password,
		}

	case AuthTypeOAuth:
		clientID := os.Getenv("SERVICENOW_CLIENT_ID")
		clientSecret := os.Getenv("SERVICENOW_CLIENT_SECRET")
		if clientID == "" || clientSecret == "" {
			return nil, fmt.Errorf("SERVICENOW_CLIENT_ID and SERVICENOW_CLIENT_SECRET are required for OAuth")
		}
		config.Auth.OAuth = &OAuthConfig{
			ClientID:     clientID,
			ClientSecret: clientSecret,
			Username:     os.Getenv("SERVICENOW_USERNAME"),
			Password:     os.Getenv("SERVICENOW_PASSWORD"),
			TokenURL:     os.Getenv("SERVICENOW_TOKEN_URL"),
		}

	case AuthTypeAPIKey:
		apiKey := os.Getenv("SERVICENOW_API_KEY")
		if apiKey == "" {
			return nil, fmt.Errorf("SERVICENOW_API_KEY is required for API key auth")
		}
		headerName := os.Getenv("SERVICENOW_API_KEY_HEADER")
		if headerName == "" {
			headerName = "X-ServiceNow-API-Key"
		}
		config.Auth.APIKey = &APIKeyConfig{
			APIKey:     apiKey,
			HeaderName: headerName,
		}

	default:
		return nil, fmt.Errorf("unsupported auth type: %s", authType)
	}

	return config, nil
}
