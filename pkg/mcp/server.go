package mcp

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"sync"
	"time"

	"github.com/elastiflow/go-mcp-servicenow/pkg/auth"
	"github.com/elastiflow/go-mcp-servicenow/pkg/servicenow"
)

// ToolHandler is a function that handles a tool call
type ToolHandler func(arguments map[string]interface{}) (*CallToolResult, error)

// ToolHandlerWithContext is a function that handles a tool call with context support
type ToolHandlerWithContext func(ctx context.Context, arguments map[string]interface{}) (*CallToolResult, error)

// ResourceProvider provides resources for the MCP server
type ResourceProvider interface {
	ListResources() []Resource
	ReadResource(uri string) (*ReadResourceResult, error)
}

// PromptProvider provides prompts for the MCP server
type PromptProvider interface {
	ListPrompts() []Prompt
	GetPrompt(name string, arguments map[string]interface{}) (*GetPromptResult, error)
}

// Server represents an MCP server
type Server struct {
	name     string
	version  string
	tools    []Tool
	handlers map[string]ToolHandler
	ctxHandlers map[string]ToolHandlerWithContext
	mu       sync.RWMutex
	stdin    io.Reader
	stdout   io.Writer
	stderr   io.Writer

	// Optional providers
	resourceProvider ResourceProvider
	promptProvider   PromptProvider

	// Rate limiting
	toolCallTimestamps []time.Time
	rateLimitMu        sync.Mutex

	// Callbacks
	onToolCall func(name string, args map[string]interface{}, duration time.Duration, success bool)
	onError    func(err error, context string)
}

// NewServer creates a new MCP server
func NewServer(name, version string) *Server {
	return &Server{
		name:               name,
		version:            version,
		tools:              make([]Tool, 0),
		handlers:           make(map[string]ToolHandler),
		ctxHandlers:        make(map[string]ToolHandlerWithContext),
		stdin:              os.Stdin,
		stdout:             os.Stdout,
		stderr:             os.Stderr,
		toolCallTimestamps: make([]time.Time, 0),
	}
}

// SetToolCallCallback sets a callback for tool calls (for telemetry)
func (s *Server) SetToolCallCallback(cb func(name string, args map[string]interface{}, duration time.Duration, success bool)) {
	s.onToolCall = cb
}

// SetErrorCallback sets a callback for errors
func (s *Server) SetErrorCallback(cb func(err error, context string)) {
	s.onError = cb
}

// RegisterResourceProvider registers a resource provider
func (s *Server) RegisterResourceProvider(provider ResourceProvider) {
	s.resourceProvider = provider
}

// RegisterPromptProvider registers a prompt provider
func (s *Server) RegisterPromptProvider(provider PromptProvider) {
	s.promptProvider = provider
}

// RegisterTool registers a tool with its handler
func (s *Server) RegisterTool(tool Tool, handler ToolHandler) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.tools = append(s.tools, tool)
	s.handlers[tool.Name] = handler
}

// RegisterToolWithContext registers a tool with a context-aware handler
func (s *Server) RegisterToolWithContext(tool Tool, handler ToolHandlerWithContext) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.tools = append(s.tools, tool)
	s.ctxHandlers[tool.Name] = handler
}

// checkRateLimit returns true if the request should be rate limited
func (s *Server) checkRateLimit() bool {
	s.rateLimitMu.Lock()
	defer s.rateLimitMu.Unlock()

	now := time.Now()
	twentySecondsAgo := now.Add(-20 * time.Second)

	// Remove old timestamps
	newTimestamps := make([]time.Time, 0)
	for _, ts := range s.toolCallTimestamps {
		if ts.After(twentySecondsAgo) {
			newTimestamps = append(newTimestamps, ts)
		}
	}
	s.toolCallTimestamps = newTimestamps

	// Check if we have 5 or more calls in the past 20s
	if len(s.toolCallTimestamps) >= 5 {
		return true
	}

	// Record this call
	s.toolCallTimestamps = append(s.toolCallTimestamps, now)
	return false
}

// Run starts the server in stdio mode
func (s *Server) Run() error {
	lines := make(chan string)
	errors := make(chan error)

	go func() {
		reader := bufio.NewReader(s.stdin)
		for {
			line, err := reader.ReadString('\n')
			if err != nil {
				if err == io.EOF {
					if line != "" {
						lines <- line
					}
					errors <- io.EOF
					return
				}
				errors <- err
				return
			}
			lines <- line
		}
	}()

	receivedData := false
	initialTimeout := time.After(30 * time.Second)

	for {
		select {
		case line := <-lines:
			receivedData = true
			line = trimLine(line)
			if line == "" {
				continue
			}

			response := s.handleMessage([]byte(line))
			if response != nil {
				s.sendResponse(response)
			}

		case err := <-errors:
			if err == io.EOF {
				if receivedData {
					return nil
				}
				return fmt.Errorf("stdin closed before receiving any data")
			}
			return fmt.Errorf("read error: %w", err)

		case <-initialTimeout:
			if !receivedData {
				initialTimeout = time.After(24 * time.Hour)
			}
		}
	}
}

// RunHTTP starts the server in HTTP mode with optional authentication
func (s *Server) RunHTTP(addr string) error {
	return s.RunHTTPWithAuthorizer(addr, nil)
}

// RunHTTPWithAuthorizer starts the server in HTTP mode with a custom authorizer
func (s *Server) RunHTTPWithAuthorizer(addr string, authorizer auth.Authorizer) error {
	mux := http.NewServeMux()

	// Health check endpoint (no auth required)
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"status":  "ok",
			"version": s.version,
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
			// Check Authorization header first, then fall back to X-MCP-Auth-Token
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

			// Use custom authorizer if provided, otherwise use default token validation
			var authorized bool
			var authErr error
			if authorizer != nil {
				authorized, authErr = authorizer.Authorize(r.Context(), token)
			} else {
				// Default: use TokenAuthorizer for backward compatibility
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

		// Extract ServiceNow credentials from headers and add to context
		ctx := r.Context()
		snUsername := r.Header.Get(servicenow.HeaderUsername)
		snPassword := r.Header.Get(servicenow.HeaderPassword)
		snAPIKey := r.Header.Get(servicenow.HeaderAPIKey)
		if snUsername != "" || snPassword != "" || snAPIKey != "" {
			creds := &servicenow.ContextCredentials{
				Username: snUsername,
				Password: snPassword,
				APIKey:   snAPIKey,
			}
			ctx = servicenow.ContextWithCredentials(ctx, creds)
		}

		response := s.handleMessageWithContext(ctx, body)
		if response != nil {
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(response)
		}
	})

	if auth.IsAuthEnabled() {
		fmt.Fprintf(s.stderr, "MCP Server running on HTTP at %s (authentication enabled)\n", addr)
	} else {
		fmt.Fprintf(s.stderr, "MCP Server running on HTTP at %s (authentication disabled)\n", addr)
	}
	return http.ListenAndServe(addr, mux)
}

func trimLine(s string) string {
	start := 0
	end := len(s)
	for start < end && (s[start] == ' ' || s[start] == '\t' || s[start] == '\n' || s[start] == '\r') {
		start++
	}
	for end > start && (s[end-1] == ' ' || s[end-1] == '\t' || s[end-1] == '\n' || s[end-1] == '\r') {
		end--
	}
	return s[start:end]
}

func (s *Server) handleMessage(data []byte) *JSONRPCResponse {
	return s.handleMessageWithContext(context.Background(), data)
}

func (s *Server) handleMessageWithContext(ctx context.Context, data []byte) *JSONRPCResponse {
	var request JSONRPCRequest
	if err := json.Unmarshal(data, &request); err != nil {
		return &JSONRPCResponse{
			JSONRPC: "2.0",
			Error: &JSONRPCError{
				Code:    ParseError,
				Message: "Parse error",
				Data:    err.Error(),
			},
		}
	}

	// Handle notifications (no ID)
	if request.ID == nil {
		s.handleNotification(&request)
		return nil
	}

	return s.handleRequestWithContext(ctx, &request)
}

func (s *Server) handleNotification(request *JSONRPCRequest) {
	switch request.Method {
	case "notifications/initialized":
		fmt.Fprintln(s.stderr, "Client initialized")
	case "notifications/cancelled":
		// Request cancellation
	}
}

func (s *Server) handleRequest(request *JSONRPCRequest) *JSONRPCResponse {
	return s.handleRequestWithContext(context.Background(), request)
}

func (s *Server) handleRequestWithContext(ctx context.Context, request *JSONRPCRequest) *JSONRPCResponse {
	response := &JSONRPCResponse{
		JSONRPC: "2.0",
		ID:      request.ID,
	}

	switch request.Method {
	case "initialize":
		response.Result = s.handleInitialize(request.Params)
	case "tools/list":
		response.Result = s.handleListTools()
	case "tools/call":
		result, err := s.handleCallToolWithContext(ctx, request.Params)
		if err != nil {
			response.Error = &JSONRPCError{
				Code:    InternalError,
				Message: err.Error(),
			}
		} else {
			response.Result = result
		}
	case "resources/list":
		response.Result = s.handleListResources()
	case "resources/read":
		result, err := s.handleReadResource(request.Params)
		if err != nil {
			response.Error = &JSONRPCError{
				Code:    InternalError,
				Message: err.Error(),
			}
		} else {
			response.Result = result
		}
	case "prompts/list":
		response.Result = s.handleListPrompts()
	case "prompts/get":
		result, err := s.handleGetPrompt(request.Params)
		if err != nil {
			response.Error = &JSONRPCError{
				Code:    InternalError,
				Message: err.Error(),
			}
		} else {
			response.Result = result
		}
	case "ping":
		response.Result = map[string]interface{}{}
	default:
		response.Error = &JSONRPCError{
			Code:    MethodNotFound,
			Message: fmt.Sprintf("Method not found: %s", request.Method),
		}
	}

	return response
}

func (s *Server) handleInitialize(params interface{}) *InitializeResult {
	caps := ServerCapabilities{
		Tools: &ToolsCapability{ListChanged: false},
	}

	if s.resourceProvider != nil {
		caps.Resources = &ResourcesCapability{Subscribe: false, ListChanged: false}
	}
	if s.promptProvider != nil {
		caps.Prompts = &PromptsCapability{ListChanged: false}
	}

	return &InitializeResult{
		ProtocolVersion: "2024-11-05",
		Capabilities:    caps,
		ServerInfo: ServerInfo{
			Name:    s.name,
			Version: s.version,
		},
	}
}

func (s *Server) handleListTools() *ListToolsResult {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return &ListToolsResult{Tools: s.tools}
}

func (s *Server) handleCallTool(params interface{}) (*CallToolResult, error) {
	return s.handleCallToolWithContext(context.Background(), params)
}

func (s *Server) handleCallToolWithContext(ctx context.Context, params interface{}) (*CallToolResult, error) {
	paramsMap, ok := params.(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("invalid params type")
	}

	name, ok := paramsMap["name"].(string)
	if !ok {
		return nil, fmt.Errorf("missing tool name")
	}

	arguments, _ := paramsMap["arguments"].(map[string]interface{})

	// Check rate limit
	if s.checkRateLimit() {
		return &CallToolResult{
			Content: []ContentItem{{Type: "text", Text: "Rate limit exceeded: Maximum 5 tool calls per 20 seconds. Please try again later."}},
			IsError: true,
		}, nil
	}

	s.mu.RLock()
	handler, handlerExists := s.handlers[name]
	ctxHandler, ctxHandlerExists := s.ctxHandlers[name]
	s.mu.RUnlock()

	if !handlerExists && !ctxHandlerExists {
		return &CallToolResult{
			Content: []ContentItem{{Type: "text", Text: fmt.Sprintf("Unknown tool: %s", name)}},
			IsError: true,
		}, nil
	}

	startTime := time.Now()
	var result *CallToolResult
	var err error

	// Prefer context-aware handler if available
	if ctxHandlerExists {
		result, err = ctxHandler(ctx, arguments)
	} else {
		result, err = handler(arguments)
	}
	duration := time.Since(startTime)

	success := err == nil && (result == nil || !result.IsError)

	// Call telemetry callback
	if s.onToolCall != nil {
		s.onToolCall(name, arguments, duration, success)
	}

	if err != nil {
		if s.onError != nil {
			s.onError(err, fmt.Sprintf("tool_%s", name))
		}
		return &CallToolResult{
			Content: []ContentItem{{Type: "text", Text: fmt.Sprintf("Error: %s", err.Error())}},
			IsError: true,
		}, nil
	}

	return result, nil
}

func (s *Server) handleListResources() *ListResourcesResult {
	if s.resourceProvider == nil {
		return &ListResourcesResult{Resources: []Resource{}}
	}
	return &ListResourcesResult{Resources: s.resourceProvider.ListResources()}
}

func (s *Server) handleReadResource(params interface{}) (*ReadResourceResult, error) {
	if s.resourceProvider == nil {
		return nil, fmt.Errorf("resources not supported")
	}

	paramsMap, ok := params.(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("invalid params type")
	}

	uri, ok := paramsMap["uri"].(string)
	if !ok {
		return nil, fmt.Errorf("missing resource uri")
	}

	return s.resourceProvider.ReadResource(uri)
}

func (s *Server) handleListPrompts() *ListPromptsResult {
	if s.promptProvider == nil {
		return &ListPromptsResult{Prompts: []Prompt{}}
	}
	return &ListPromptsResult{Prompts: s.promptProvider.ListPrompts()}
}

func (s *Server) handleGetPrompt(params interface{}) (*GetPromptResult, error) {
	if s.promptProvider == nil {
		return nil, fmt.Errorf("prompts not supported")
	}

	paramsMap, ok := params.(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("invalid params type")
	}

	name, ok := paramsMap["name"].(string)
	if !ok {
		return nil, fmt.Errorf("missing prompt name")
	}

	arguments, _ := paramsMap["arguments"].(map[string]interface{})
	return s.promptProvider.GetPrompt(name, arguments)
}

func (s *Server) sendResponse(response *JSONRPCResponse) {
	data, err := json.Marshal(response)
	if err != nil {
		fmt.Fprintf(s.stderr, "Error marshaling response: %v\n", err)
		return
	}
	fmt.Fprintln(s.stdout, string(data))
}

// Log writes a message to stderr for debugging
func (s *Server) Log(format string, args ...interface{}) {
	fmt.Fprintf(s.stderr, format+"\n", args...)
}
