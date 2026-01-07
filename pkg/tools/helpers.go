package tools

import (
	"encoding/json"

	"github.com/elastiflow/go-mcp-servicenow/pkg/mcp"
)

// TextResult creates a successful text result
func TextResult(content string) *mcp.CallToolResult {
	return &mcp.CallToolResult{
		Content: []mcp.ContentItem{{Type: "text", Text: content}},
		IsError: false,
	}
}

// JSONResult creates a successful JSON result
func JSONResult(data interface{}) *mcp.CallToolResult {
	jsonBytes, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return ErrorResult("Failed to serialize result: " + err.Error())
	}
	return &mcp.CallToolResult{
		Content: []mcp.ContentItem{{Type: "text", Text: string(jsonBytes)}},
		IsError: false,
	}
}

// ErrorResult creates an error result
func ErrorResult(message string) *mcp.CallToolResult {
	return &mcp.CallToolResult{
		Content: []mcp.ContentItem{{Type: "text", Text: message}},
		IsError: true,
	}
}

// WriteBlockedResult returns a result for blocked write operations
func WriteBlockedResult() *mcp.CallToolResult {
	return ErrorResult("This operation is blocked in read-only mode. Set READ_ONLY_MODE=false to enable write operations.")
}

// GetStringArg extracts a string argument with default value
func GetStringArg(args map[string]interface{}, key, defaultValue string) string {
	if val, ok := args[key].(string); ok {
		return val
	}
	return defaultValue
}

// GetIntArg extracts an integer argument with default value
func GetIntArg(args map[string]interface{}, key string, defaultValue int) int {
	if val, ok := args[key].(float64); ok {
		return int(val)
	}
	if val, ok := args[key].(int); ok {
		return val
	}
	return defaultValue
}

// GetBoolArg extracts a boolean argument with default value
func GetBoolArg(args map[string]interface{}, key string, defaultValue bool) bool {
	if val, ok := args[key].(bool); ok {
		return val
	}
	return defaultValue
}

// GetStringArrayArg extracts a string array argument
func GetStringArrayArg(args map[string]interface{}, key string) []string {
	if val, ok := args[key].([]interface{}); ok {
		result := make([]string, 0, len(val))
		for _, v := range val {
			if s, ok := v.(string); ok {
				result = append(result, s)
			}
		}
		return result
	}
	return nil
}

// GetMapArg extracts a map argument
func GetMapArg(args map[string]interface{}, key string) map[string]interface{} {
	if val, ok := args[key].(map[string]interface{}); ok {
		return val
	}
	return nil
}

// IsSysID checks if a string looks like a ServiceNow sys_id
func IsSysID(s string) bool {
	if len(s) != 32 {
		return false
	}
	for _, c := range s {
		if !((c >= '0' && c <= '9') || (c >= 'a' && c <= 'f') || (c >= 'A' && c <= 'F')) {
			return false
		}
	}
	return true
}

// SuccessResponse creates a standard success response
type SuccessResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
}

// ErrorResponse creates a standard error response
type ErrorResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
	Error   string `json:"error,omitempty"`
}

// NewSuccessResponse creates a new success response
func NewSuccessResponse(message string) *SuccessResponse {
	return &SuccessResponse{
		Success: true,
		Message: message,
	}
}

// NewErrorResponse creates a new error response
func NewErrorResponse(message string, err error) *ErrorResponse {
	resp := &ErrorResponse{
		Success: false,
		Message: message,
	}
	if err != nil {
		resp.Error = err.Error()
	}
	return resp
}
