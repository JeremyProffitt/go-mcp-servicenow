package servicenow

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"

	"github.com/elastiflow/go-mcp-servicenow/pkg/logging"
)

// Client represents a ServiceNow API client
type Client struct {
	config     *Config
	httpClient *http.Client
	logger     *logging.Logger

	// OAuth token caching
	token     string
	tokenType string
	tokenMu   sync.RWMutex
}

// ClientOption is a functional option for the Client
type ClientOption func(*Client)

// WithLogger sets the logger for the client
func WithLogger(logger *logging.Logger) ClientOption {
	return func(c *Client) {
		c.logger = logger
	}
}

// NewClient creates a new ServiceNow API client
func NewClient(config *Config, opts ...ClientOption) (*Client, error) {
	if config == nil {
		return nil, fmt.Errorf("config is required")
	}

	client := &Client{
		config: config,
		httpClient: &http.Client{
			Timeout: time.Duration(config.Timeout) * time.Second,
		},
	}

	for _, opt := range opts {
		opt(client)
	}

	return client, nil
}

// GetHeaders returns the authentication headers for API requests
func (c *Client) GetHeaders() (map[string]string, error) {
	headers := map[string]string{
		"Accept":       "application/json",
		"Content-Type": "application/json",
	}

	switch c.config.Auth.Type {
	case AuthTypeBasic:
		if c.config.Auth.Basic == nil {
			return nil, fmt.Errorf("basic auth configuration is required")
		}
		authStr := fmt.Sprintf("%s:%s", c.config.Auth.Basic.Username, c.config.Auth.Basic.Password)
		encoded := base64.StdEncoding.EncodeToString([]byte(authStr))
		headers["Authorization"] = fmt.Sprintf("Basic %s", encoded)

	case AuthTypeOAuth:
		token, tokenType, err := c.getOAuthToken()
		if err != nil {
			return nil, err
		}
		headers["Authorization"] = fmt.Sprintf("%s %s", tokenType, token)

	case AuthTypeAPIKey:
		if c.config.Auth.APIKey == nil {
			return nil, fmt.Errorf("API key configuration is required")
		}
		headers[c.config.Auth.APIKey.HeaderName] = c.config.Auth.APIKey.APIKey
	}

	return headers, nil
}

// getOAuthToken gets or refreshes the OAuth token
func (c *Client) getOAuthToken() (string, string, error) {
	c.tokenMu.RLock()
	if c.token != "" {
		token, tokenType := c.token, c.tokenType
		c.tokenMu.RUnlock()
		return token, tokenType, nil
	}
	c.tokenMu.RUnlock()

	c.tokenMu.Lock()
	defer c.tokenMu.Unlock()

	// Double-check after acquiring write lock
	if c.token != "" {
		return c.token, c.tokenType, nil
	}

	if c.config.Auth.OAuth == nil {
		return "", "", fmt.Errorf("OAuth configuration is required")
	}

	oauthConfig := c.config.Auth.OAuth

	// Determine token URL
	tokenURL := oauthConfig.TokenURL
	if tokenURL == "" {
		// Extract instance name from URL
		instanceURL := c.config.InstanceURL
		parts := strings.Split(instanceURL, ".")
		if len(parts) < 2 {
			return "", "", fmt.Errorf("invalid instance URL: %s", instanceURL)
		}
		instanceName := strings.TrimPrefix(parts[0], "https://")
		instanceName = strings.TrimPrefix(instanceName, "http://")
		tokenURL = fmt.Sprintf("https://%s.service-now.com/oauth_token.do", instanceName)
	}

	// Prepare Authorization header
	authStr := fmt.Sprintf("%s:%s", oauthConfig.ClientID, oauthConfig.ClientSecret)
	authHeader := base64.StdEncoding.EncodeToString([]byte(authStr))

	// Try client_credentials grant first
	data := url.Values{}
	data.Set("grant_type", "client_credentials")

	req, err := http.NewRequest("POST", tokenURL, strings.NewReader(data.Encode()))
	if err != nil {
		return "", "", fmt.Errorf("failed to create token request: %w", err)
	}
	req.Header.Set("Authorization", fmt.Sprintf("Basic %s", authHeader))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return "", "", fmt.Errorf("failed to get OAuth token: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusOK {
		var tokenResp struct {
			AccessToken string `json:"access_token"`
			TokenType   string `json:"token_type"`
		}
		if err := json.NewDecoder(resp.Body).Decode(&tokenResp); err != nil {
			return "", "", fmt.Errorf("failed to decode token response: %w", err)
		}
		c.token = tokenResp.AccessToken
		c.tokenType = tokenResp.TokenType
		if c.tokenType == "" {
			c.tokenType = "Bearer"
		}
		return c.token, c.tokenType, nil
	}

	// Try password grant if client_credentials failed
	if oauthConfig.Username != "" && oauthConfig.Password != "" {
		data = url.Values{}
		data.Set("grant_type", "password")
		data.Set("username", oauthConfig.Username)
		data.Set("password", oauthConfig.Password)

		req, err = http.NewRequest("POST", tokenURL, strings.NewReader(data.Encode()))
		if err != nil {
			return "", "", fmt.Errorf("failed to create token request: %w", err)
		}
		req.Header.Set("Authorization", fmt.Sprintf("Basic %s", authHeader))
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

		resp, err = c.httpClient.Do(req)
		if err != nil {
			return "", "", fmt.Errorf("failed to get OAuth token: %w", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode == http.StatusOK {
			var tokenResp struct {
				AccessToken string `json:"access_token"`
				TokenType   string `json:"token_type"`
			}
			if err := json.NewDecoder(resp.Body).Decode(&tokenResp); err != nil {
				return "", "", fmt.Errorf("failed to decode token response: %w", err)
			}
			c.token = tokenResp.AccessToken
			c.tokenType = tokenResp.TokenType
			if c.tokenType == "" {
				c.tokenType = "Bearer"
			}
			return c.token, c.tokenType, nil
		}
	}

	return "", "", fmt.Errorf("failed to get OAuth token using both client_credentials and password grants")
}

// RefreshToken refreshes the OAuth token
func (c *Client) RefreshToken() error {
	if c.config.Auth.Type != AuthTypeOAuth {
		return nil
	}

	c.tokenMu.Lock()
	c.token = ""
	c.tokenType = ""
	c.tokenMu.Unlock()

	_, _, err := c.getOAuthToken()
	return err
}

// Request makes an HTTP request to the ServiceNow API
func (c *Client) Request(method, endpoint string, body interface{}) (map[string]interface{}, error) {
	apiURL := fmt.Sprintf("%s%s", c.config.APIURL(), endpoint)

	var bodyReader io.Reader
	if body != nil {
		bodyBytes, err := json.Marshal(body)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal request body: %w", err)
		}
		bodyReader = bytes.NewReader(bodyBytes)
	}

	req, err := http.NewRequest(method, apiURL, bodyReader)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	headers, err := c.GetHeaders()
	if err != nil {
		return nil, fmt.Errorf("failed to get headers: %w", err)
	}

	for k, v := range headers {
		req.Header.Set(k, v)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	if resp.StatusCode >= 400 {
		return nil, fmt.Errorf("API error (status %d): %s", resp.StatusCode, string(respBody))
	}

	var result map[string]interface{}
	if len(respBody) > 0 {
		if err := json.Unmarshal(respBody, &result); err != nil {
			return nil, fmt.Errorf("failed to parse response: %w", err)
		}
	}

	return result, nil
}

// Get makes a GET request to the ServiceNow API
func (c *Client) Get(endpoint string, params map[string]string) (map[string]interface{}, error) {
	apiURL := fmt.Sprintf("%s%s", c.config.APIURL(), endpoint)

	if len(params) > 0 {
		values := url.Values{}
		for k, v := range params {
			values.Set(k, v)
		}
		apiURL = fmt.Sprintf("%s?%s", apiURL, values.Encode())
	}

	req, err := http.NewRequest("GET", apiURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	headers, err := c.GetHeaders()
	if err != nil {
		return nil, fmt.Errorf("failed to get headers: %w", err)
	}

	for k, v := range headers {
		req.Header.Set(k, v)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	if resp.StatusCode >= 400 {
		return nil, fmt.Errorf("API error (status %d): %s", resp.StatusCode, string(respBody))
	}

	var result map[string]interface{}
	if len(respBody) > 0 {
		if err := json.Unmarshal(respBody, &result); err != nil {
			return nil, fmt.Errorf("failed to parse response: %w", err)
		}
	}

	return result, nil
}

// Post makes a POST request to the ServiceNow API
func (c *Client) Post(endpoint string, body interface{}) (map[string]interface{}, error) {
	return c.Request("POST", endpoint, body)
}

// Put makes a PUT request to the ServiceNow API
func (c *Client) Put(endpoint string, body interface{}) (map[string]interface{}, error) {
	return c.Request("PUT", endpoint, body)
}

// Patch makes a PATCH request to the ServiceNow API
func (c *Client) Patch(endpoint string, body interface{}) (map[string]interface{}, error) {
	return c.Request("PATCH", endpoint, body)
}

// Delete makes a DELETE request to the ServiceNow API
func (c *Client) Delete(endpoint string) (map[string]interface{}, error) {
	return c.Request("DELETE", endpoint, nil)
}

// Config returns the client configuration
func (c *Client) Config() *Config {
	return c.config
}
