package mcp

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/elastiflow/go-mcp-servicenow/pkg/auth"
)

// createTestServer creates an MCP server wrapped with HTTP handlers for testing.
// It returns the httptest server and a cleanup function.
func createTestServer(t *testing.T, authorizer auth.Authorizer, enableAuth bool) (*httptest.Server, func()) {
	t.Helper()

	// Store original env value and set test value
	originalToken := os.Getenv("MCP_AUTH_TOKEN")
	if enableAuth {
		os.Setenv("MCP_AUTH_TOKEN", "test-token")
	} else {
		os.Unsetenv("MCP_AUTH_TOKEN")
	}

	// Create MCP server
	mcpServer := NewServer("test-servicenow-mcp", "1.0.0-test")

	// Register a sample tool for testing
	mcpServer.RegisterTool(Tool{
		Name:        "test_tool",
		Description: "A test tool for integration testing",
		InputSchema: JSONSchema{
			Type: "object",
			Properties: map[string]Property{
				"message": {
					Type:        "string",
					Description: "A test message",
				},
			},
		},
	}, func(args map[string]interface{}) (*CallToolResult, error) {
		msg, _ := args["message"].(string)
		if msg == "" {
			msg = "no message provided"
		}
		return &CallToolResult{
			Content: []ContentItem{
				{Type: "text", Text: "Echo: " + msg},
			},
		}, nil
	})

	// Create HTTP mux with the same routes as RunHTTPWithAuthorizer
	mux := http.NewServeMux()

	// Health check endpoint (no auth required)
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"status":  "ok",
			"version": mcpServer.version,
		})
	})

	// MCP endpoint with authentication
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		// Check authentication if enabled
		if auth.IsAuthEnabled() {
			token := r.Header.Get("Authorization")
			if token == "" {
				token = r.Header.Get(auth.AuthHeaderName)
			}

			if token == "" {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusUnauthorized)
				_ = json.NewEncoder(w).Encode(map[string]interface{}{
					"jsonrpc": "2.0",
					"id":      nil,
					"error":   map[string]interface{}{"code": -32001, "message": "Unauthorized: missing Authorization header"},
				})
				return
			}

			var authorized bool
			var authErr error
			if authorizer != nil {
				authorized, authErr = authorizer.Authorize(r.Context(), token)
			} else {
				defaultAuth := auth.NewTokenAuthorizer()
				authorized, authErr = defaultAuth.Authorize(r.Context(), token)
			}

			if authErr != nil || !authorized {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusUnauthorized)
				_ = json.NewEncoder(w).Encode(map[string]interface{}{
					"jsonrpc": "2.0",
					"id":      nil,
					"error":   map[string]interface{}{"code": -32001, "message": "Unauthorized: invalid authentication token"},
				})
				return
			}
		}

		body, err := io.ReadAll(r.Body)
		if err != nil {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusBadRequest)
			_ = json.NewEncoder(w).Encode(map[string]interface{}{
				"jsonrpc": "2.0",
				"id":      nil,
				"error":   map[string]interface{}{"code": -32700, "message": "Parse error"},
			})
			return
		}

		response := mcpServer.handleMessage(body)
		if response != nil {
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(response)
		}
	})

	ts := httptest.NewServer(mux)

	cleanup := func() {
		ts.Close()
		if originalToken != "" {
			os.Setenv("MCP_AUTH_TOKEN", originalToken)
		} else {
			os.Unsetenv("MCP_AUTH_TOKEN")
		}
	}

	return ts, cleanup
}

// TestHTTPHealthEndpoint tests that GET /health returns 200 with status and version
func TestHTTPHealthEndpoint(t *testing.T) {
	ts, cleanup := createTestServer(t, nil, false)
	defer cleanup()

	resp, err := http.Get(ts.URL + "/health")
	if err != nil {
		t.Fatalf("Failed to send GET request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status 200, got %d", resp.StatusCode)
	}

	contentType := resp.Header.Get("Content-Type")
	if contentType != "application/json" {
		t.Errorf("Expected Content-Type 'application/json', got '%s'", contentType)
	}

	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if result["status"] != "ok" {
		t.Errorf("Expected status 'ok', got '%v'", result["status"])
	}

	if result["version"] != "1.0.0-test" {
		t.Errorf("Expected version '1.0.0-test', got '%v'", result["version"])
	}
}

// TestHTTPAuthMiddleware_MissingHeader tests that POST / without Authorization header returns 401
func TestHTTPAuthMiddleware_MissingHeader(t *testing.T) {
	ts, cleanup := createTestServer(t, nil, true) // Enable auth
	defer cleanup()

	// Create a valid JSON-RPC request body
	reqBody := map[string]interface{}{
		"jsonrpc": "2.0",
		"id":      1,
		"method":  "initialize",
		"params": map[string]interface{}{
			"protocolVersion": "2024-11-05",
			"capabilities":    map[string]interface{}{},
			"clientInfo": map[string]interface{}{
				"name":    "test-client",
				"version": "1.0.0",
			},
		},
	}

	body, _ := json.Marshal(reqBody)
	resp, err := http.Post(ts.URL+"/", "application/json", bytes.NewReader(body))
	if err != nil {
		t.Fatalf("Failed to send POST request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusUnauthorized {
		t.Errorf("Expected status 401, got %d", resp.StatusCode)
	}

	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	errObj, ok := result["error"].(map[string]interface{})
	if !ok {
		t.Fatalf("Expected error object in response, got %v", result)
	}

	if errObj["code"].(float64) != -32001 {
		t.Errorf("Expected error code -32001, got %v", errObj["code"])
	}

	errMsg, _ := errObj["message"].(string)
	if errMsg == "" || errMsg != "Unauthorized: missing Authorization header" {
		t.Errorf("Expected 'Unauthorized: missing Authorization header', got '%s'", errMsg)
	}
}

// TestHTTPAuthMiddleware_WithHeader tests that POST / with Authorization header proceeds (using MockAuthorizer)
func TestHTTPAuthMiddleware_WithHeader(t *testing.T) {
	mockAuth := &auth.MockAuthorizer{}
	ts, cleanup := createTestServer(t, mockAuth, true) // Enable auth with MockAuthorizer
	defer cleanup()

	// Create a valid JSON-RPC request body
	reqBody := map[string]interface{}{
		"jsonrpc": "2.0",
		"id":      1,
		"method":  "initialize",
		"params": map[string]interface{}{
			"protocolVersion": "2024-11-05",
			"capabilities":    map[string]interface{}{},
			"clientInfo": map[string]interface{}{
				"name":    "test-client",
				"version": "1.0.0",
			},
		},
	}

	body, _ := json.Marshal(reqBody)
	req, err := http.NewRequest(http.MethodPost, ts.URL+"/", bytes.NewReader(body))
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer any-token-works-with-mock")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		t.Fatalf("Failed to send request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		t.Fatalf("Expected status 200, got %d. Body: %s", resp.StatusCode, string(bodyBytes))
	}

	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	// Should have a result, not an error
	if result["error"] != nil {
		t.Errorf("Expected no error, got %v", result["error"])
	}

	if result["result"] == nil {
		t.Errorf("Expected result in response, got nil")
	}
}

// TestHTTPMCPInitialize tests POST / with valid JSON-RPC initialize request returns valid response
func TestHTTPMCPInitialize(t *testing.T) {
	ts, cleanup := createTestServer(t, nil, false) // No auth
	defer cleanup()

	reqBody := map[string]interface{}{
		"jsonrpc": "2.0",
		"id":      1,
		"method":  "initialize",
		"params": map[string]interface{}{
			"protocolVersion": "2024-11-05",
			"capabilities":    map[string]interface{}{},
			"clientInfo": map[string]interface{}{
				"name":    "test-client",
				"version": "1.0.0",
			},
		},
	}

	body, _ := json.Marshal(reqBody)
	resp, err := http.Post(ts.URL+"/", "application/json", bytes.NewReader(body))
	if err != nil {
		t.Fatalf("Failed to send POST request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status 200, got %d", resp.StatusCode)
	}

	var result JSONRPCResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if result.JSONRPC != "2.0" {
		t.Errorf("Expected jsonrpc '2.0', got '%s'", result.JSONRPC)
	}

	if result.ID == nil || result.ID.(float64) != 1 {
		t.Errorf("Expected id 1, got %v", result.ID)
	}

	if result.Error != nil {
		t.Errorf("Expected no error, got %v", result.Error)
	}

	// Check the result structure
	resultMap, ok := result.Result.(map[string]interface{})
	if !ok {
		t.Fatalf("Expected result to be a map, got %T", result.Result)
	}

	if resultMap["protocolVersion"] != "2024-11-05" {
		t.Errorf("Expected protocolVersion '2024-11-05', got '%v'", resultMap["protocolVersion"])
	}

	serverInfo, ok := resultMap["serverInfo"].(map[string]interface{})
	if !ok {
		t.Fatalf("Expected serverInfo to be a map, got %T", resultMap["serverInfo"])
	}

	if serverInfo["name"] != "test-servicenow-mcp" {
		t.Errorf("Expected server name 'test-servicenow-mcp', got '%v'", serverInfo["name"])
	}

	if serverInfo["version"] != "1.0.0-test" {
		t.Errorf("Expected server version '1.0.0-test', got '%v'", serverInfo["version"])
	}

	// Check capabilities
	capabilities, ok := resultMap["capabilities"].(map[string]interface{})
	if !ok {
		t.Fatalf("Expected capabilities to be a map, got %T", resultMap["capabilities"])
	}

	if capabilities["tools"] == nil {
		t.Errorf("Expected tools capability to be present")
	}
}

// TestHTTPMCPToolsList tests POST / with tools/list request returns list of tools
func TestHTTPMCPToolsList(t *testing.T) {
	ts, cleanup := createTestServer(t, nil, false) // No auth
	defer cleanup()

	reqBody := map[string]interface{}{
		"jsonrpc": "2.0",
		"id":      2,
		"method":  "tools/list",
		"params":  map[string]interface{}{},
	}

	body, _ := json.Marshal(reqBody)
	resp, err := http.Post(ts.URL+"/", "application/json", bytes.NewReader(body))
	if err != nil {
		t.Fatalf("Failed to send POST request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status 200, got %d", resp.StatusCode)
	}

	var result JSONRPCResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if result.JSONRPC != "2.0" {
		t.Errorf("Expected jsonrpc '2.0', got '%s'", result.JSONRPC)
	}

	if result.ID == nil || result.ID.(float64) != 2 {
		t.Errorf("Expected id 2, got %v", result.ID)
	}

	if result.Error != nil {
		t.Errorf("Expected no error, got %v", result.Error)
	}

	// Check the result structure
	resultMap, ok := result.Result.(map[string]interface{})
	if !ok {
		t.Fatalf("Expected result to be a map, got %T", result.Result)
	}

	tools, ok := resultMap["tools"].([]interface{})
	if !ok {
		t.Fatalf("Expected tools to be an array, got %T", resultMap["tools"])
	}

	// We registered one test tool
	if len(tools) != 1 {
		t.Errorf("Expected 1 tool, got %d", len(tools))
	}

	// Verify the test tool structure
	if len(tools) > 0 {
		tool, ok := tools[0].(map[string]interface{})
		if !ok {
			t.Fatalf("Expected tool to be a map, got %T", tools[0])
		}

		if tool["name"] != "test_tool" {
			t.Errorf("Expected tool name 'test_tool', got '%v'", tool["name"])
		}

		if tool["description"] != "A test tool for integration testing" {
			t.Errorf("Expected tool description 'A test tool for integration testing', got '%v'", tool["description"])
		}

		inputSchema, ok := tool["inputSchema"].(map[string]interface{})
		if !ok {
			t.Fatalf("Expected inputSchema to be a map, got %T", tool["inputSchema"])
		}

		if inputSchema["type"] != "object" {
			t.Errorf("Expected inputSchema type 'object', got '%v'", inputSchema["type"])
		}
	}
}

// TestHTTPMCPToolsCall tests POST / with tools/call request executes the tool
func TestHTTPMCPToolsCall(t *testing.T) {
	ts, cleanup := createTestServer(t, nil, false) // No auth
	defer cleanup()

	reqBody := map[string]interface{}{
		"jsonrpc": "2.0",
		"id":      3,
		"method":  "tools/call",
		"params": map[string]interface{}{
			"name": "test_tool",
			"arguments": map[string]interface{}{
				"message": "Hello, World!",
			},
		},
	}

	body, _ := json.Marshal(reqBody)
	resp, err := http.Post(ts.URL+"/", "application/json", bytes.NewReader(body))
	if err != nil {
		t.Fatalf("Failed to send POST request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status 200, got %d", resp.StatusCode)
	}

	var result JSONRPCResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if result.Error != nil {
		t.Errorf("Expected no error, got %v", result.Error)
	}

	// Check the result structure
	resultMap, ok := result.Result.(map[string]interface{})
	if !ok {
		t.Fatalf("Expected result to be a map, got %T", result.Result)
	}

	content, ok := resultMap["content"].([]interface{})
	if !ok {
		t.Fatalf("Expected content to be an array, got %T", resultMap["content"])
	}

	if len(content) != 1 {
		t.Errorf("Expected 1 content item, got %d", len(content))
	}

	if len(content) > 0 {
		item, ok := content[0].(map[string]interface{})
		if !ok {
			t.Fatalf("Expected content item to be a map, got %T", content[0])
		}

		if item["type"] != "text" {
			t.Errorf("Expected content type 'text', got '%v'", item["type"])
		}

		if item["text"] != "Echo: Hello, World!" {
			t.Errorf("Expected content text 'Echo: Hello, World!', got '%v'", item["text"])
		}
	}
}

// TestHTTPMethodNotAllowed tests that non-POST methods to / return 405
func TestHTTPMethodNotAllowed(t *testing.T) {
	ts, cleanup := createTestServer(t, nil, false)
	defer cleanup()

	resp, err := http.Get(ts.URL + "/")
	if err != nil {
		t.Fatalf("Failed to send GET request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusMethodNotAllowed {
		t.Errorf("Expected status 405, got %d", resp.StatusCode)
	}
}

// TestHTTPInvalidJSON tests that invalid JSON returns parse error
func TestHTTPInvalidJSON(t *testing.T) {
	ts, cleanup := createTestServer(t, nil, false)
	defer cleanup()

	resp, err := http.Post(ts.URL+"/", "application/json", bytes.NewReader([]byte("not valid json")))
	if err != nil {
		t.Fatalf("Failed to send POST request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status 200 (JSON-RPC error in body), got %d", resp.StatusCode)
	}

	var result JSONRPCResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if result.Error == nil {
		t.Error("Expected error in response")
	}

	if result.Error != nil && result.Error.Code != ParseError {
		t.Errorf("Expected error code %d (ParseError), got %d", ParseError, result.Error.Code)
	}
}

// TestHTTPUnknownMethod tests that unknown JSON-RPC methods return method not found error
func TestHTTPUnknownMethod(t *testing.T) {
	ts, cleanup := createTestServer(t, nil, false)
	defer cleanup()

	reqBody := map[string]interface{}{
		"jsonrpc": "2.0",
		"id":      99,
		"method":  "unknown/method",
		"params":  map[string]interface{}{},
	}

	body, _ := json.Marshal(reqBody)
	resp, err := http.Post(ts.URL+"/", "application/json", bytes.NewReader(body))
	if err != nil {
		t.Fatalf("Failed to send POST request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status 200, got %d", resp.StatusCode)
	}

	var result JSONRPCResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if result.Error == nil {
		t.Error("Expected error in response")
	}

	if result.Error != nil && result.Error.Code != MethodNotFound {
		t.Errorf("Expected error code %d (MethodNotFound), got %d", MethodNotFound, result.Error.Code)
	}
}
