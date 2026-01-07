package tools

import (
	"fmt"
	"strings"

	"github.com/elastiflow/go-mcp-servicenow/pkg/mcp"
)

// registerScriptIncludeTools registers all script include tools
func (r *Registry) registerScriptIncludeTools(server *mcp.Server) int {
	count := 0

	// List Script Includes
	server.RegisterTool(mcp.Tool{
		Name:        "list_script_includes",
		Description: "List script includes from ServiceNow",
		InputSchema: mcp.JSONSchema{
			Type: "object",
			Properties: map[string]mcp.Property{
				"limit": {
					Type:        "number",
					Description: "Maximum number of script includes to return (default: 50)",
					Default:     50,
				},
				"active": {
					Type:        "boolean",
					Description: "Filter by active status",
				},
				"query": {
					Type:        "string",
					Description: "Search query for name or API name",
				},
			},
		},
		Annotations: &mcp.ToolAnnotation{
			Title:        "List Script Includes",
			ReadOnlyHint: true,
		},
	}, func(args map[string]interface{}) (*mcp.CallToolResult, error) {
		return r.listScriptIncludes(args)
	})
	count++

	// Get Script Include
	server.RegisterTool(mcp.Tool{
		Name:        "get_script_include",
		Description: "Get a specific script include from ServiceNow",
		InputSchema: mcp.JSONSchema{
			Type: "object",
			Properties: map[string]mcp.Property{
				"script_id": {
					Type:        "string",
					Description: "Script include sys_id or name",
				},
			},
			Required: []string{"script_id"},
		},
		Annotations: &mcp.ToolAnnotation{
			Title:        "Get Script Include",
			ReadOnlyHint: true,
		},
	}, func(args map[string]interface{}) (*mcp.CallToolResult, error) {
		return r.getScriptInclude(args)
	})
	count++

	// Write operations
	if !r.readOnlyMode {
		// Create Script Include
		server.RegisterTool(mcp.Tool{
			Name:        "create_script_include",
			Description: "Create a new script include in ServiceNow",
			InputSchema: mcp.JSONSchema{
				Type: "object",
				Properties: map[string]mcp.Property{
					"name": {
						Type:        "string",
						Description: "Script include name",
					},
					"api_name": {
						Type:        "string",
						Description: "API name (must be unique)",
					},
					"script": {
						Type:        "string",
						Description: "Script content",
					},
					"description": {
						Type:        "string",
						Description: "Script description",
					},
					"client_callable": {
						Type:        "boolean",
						Description: "Whether the script is client callable",
					},
				},
				Required: []string{"name", "api_name", "script"},
			},
		}, func(args map[string]interface{}) (*mcp.CallToolResult, error) {
			return r.createScriptInclude(args)
		})
		count++

		// Update Script Include
		server.RegisterTool(mcp.Tool{
			Name:        "update_script_include",
			Description: "Update an existing script include in ServiceNow",
			InputSchema: mcp.JSONSchema{
				Type: "object",
				Properties: map[string]mcp.Property{
					"script_id": {
						Type:        "string",
						Description: "Script include sys_id",
					},
					"name": {
						Type:        "string",
						Description: "Script include name",
					},
					"script": {
						Type:        "string",
						Description: "Script content",
					},
					"description": {
						Type:        "string",
						Description: "Script description",
					},
					"active": {
						Type:        "boolean",
						Description: "Active status",
					},
				},
				Required: []string{"script_id"},
			},
		}, func(args map[string]interface{}) (*mcp.CallToolResult, error) {
			return r.updateScriptInclude(args)
		})
		count++

		// Delete Script Include
		server.RegisterTool(mcp.Tool{
			Name:        "delete_script_include",
			Description: "Delete a script include from ServiceNow",
			InputSchema: mcp.JSONSchema{
				Type: "object",
				Properties: map[string]mcp.Property{
					"script_id": {
						Type:        "string",
						Description: "Script include sys_id",
					},
				},
				Required: []string{"script_id"},
			},
			Annotations: &mcp.ToolAnnotation{
				Title:           "Delete Script Include",
				DestructiveHint: true,
			},
		}, func(args map[string]interface{}) (*mcp.CallToolResult, error) {
			return r.deleteScriptInclude(args)
		})
		count++
	}

	return count
}

func (r *Registry) listScriptIncludes(args map[string]interface{}) (*mcp.CallToolResult, error) {
	limit := GetIntArg(args, "limit", 50)
	query := GetStringArg(args, "query", "")

	params := map[string]string{
		"sysparm_limit":                  fmt.Sprintf("%d", limit),
		"sysparm_display_value":          "true",
		"sysparm_exclude_reference_link": "true",
	}

	var filters []string
	if active, exists := args["active"]; exists {
		if active.(bool) {
			filters = append(filters, "active=true")
		} else {
			filters = append(filters, "active=false")
		}
	}
	if query != "" {
		filters = append(filters, fmt.Sprintf("nameLIKE%s^ORapi_nameLIKE%s", query, query))
	}

	if len(filters) > 0 {
		params["sysparm_query"] = strings.Join(filters, "^")
	}

	result, err := r.client.Get("/table/sys_script_include", params)
	if err != nil {
		return JSONResult(NewErrorResponse("Failed to list script includes", err)), nil
	}

	scripts := []map[string]interface{}{}
	if resultList, ok := result["result"].([]interface{}); ok {
		for _, item := range resultList {
			if data, ok := item.(map[string]interface{}); ok {
				scripts = append(scripts, map[string]interface{}{
					"sys_id":          data["sys_id"],
					"name":            data["name"],
					"api_name":        data["api_name"],
					"description":     data["description"],
					"active":          data["active"],
					"client_callable": data["client_callable"],
				})
			}
		}
	}

	return JSONResult(map[string]interface{}{
		"success":         true,
		"message":         fmt.Sprintf("Found %d script includes", len(scripts)),
		"script_includes": scripts,
	}), nil
}

func (r *Registry) getScriptInclude(args map[string]interface{}) (*mcp.CallToolResult, error) {
	scriptID := GetStringArg(args, "script_id", "")
	if scriptID == "" {
		return JSONResult(NewErrorResponse("script_id is required", nil)), nil
	}

	var params map[string]string
	var endpoint string

	if IsSysID(scriptID) {
		endpoint = fmt.Sprintf("/table/sys_script_include/%s", scriptID)
		params = map[string]string{
			"sysparm_display_value":          "true",
			"sysparm_exclude_reference_link": "true",
		}
	} else {
		endpoint = "/table/sys_script_include"
		params = map[string]string{
			"sysparm_query":                  fmt.Sprintf("name=%s^ORapi_name=%s", scriptID, scriptID),
			"sysparm_limit":                  "1",
			"sysparm_display_value":          "true",
			"sysparm_exclude_reference_link": "true",
		}
	}

	result, err := r.client.Get(endpoint, params)
	if err != nil {
		return JSONResult(NewErrorResponse("Failed to get script include", err)), nil
	}

	var scriptData map[string]interface{}
	if IsSysID(scriptID) {
		scriptData, _ = result["result"].(map[string]interface{})
	} else {
		if resultList, ok := result["result"].([]interface{}); ok && len(resultList) > 0 {
			scriptData, _ = resultList[0].(map[string]interface{})
		}
	}

	if scriptData == nil {
		return JSONResult(map[string]interface{}{
			"success": false,
			"message": fmt.Sprintf("Script include not found: %s", scriptID),
		}), nil
	}

	return JSONResult(map[string]interface{}{
		"success":        true,
		"message":        "Script include found",
		"script_include": scriptData,
	}), nil
}

func (r *Registry) createScriptInclude(args map[string]interface{}) (*mcp.CallToolResult, error) {
	if r.readOnlyMode {
		return WriteBlockedResult(), nil
	}

	name := GetStringArg(args, "name", "")
	apiName := GetStringArg(args, "api_name", "")
	script := GetStringArg(args, "script", "")

	if name == "" || apiName == "" || script == "" {
		return JSONResult(NewErrorResponse("name, api_name, and script are required", nil)), nil
	}

	data := map[string]interface{}{
		"name":     name,
		"api_name": apiName,
		"script":   script,
	}

	if v := GetStringArg(args, "description", ""); v != "" {
		data["description"] = v
	}
	if v, exists := args["client_callable"]; exists {
		data["client_callable"] = v
	}

	result, err := r.client.Post("/table/sys_script_include", data)
	if err != nil {
		return JSONResult(NewErrorResponse("Failed to create script include", err)), nil
	}

	if resultData, ok := result["result"].(map[string]interface{}); ok {
		return JSONResult(map[string]interface{}{
			"success":   true,
			"message":   "Script include created successfully",
			"script_id": resultData["sys_id"],
		}), nil
	}

	return JSONResult(NewErrorResponse("Unexpected response from ServiceNow", nil)), nil
}

func (r *Registry) updateScriptInclude(args map[string]interface{}) (*mcp.CallToolResult, error) {
	if r.readOnlyMode {
		return WriteBlockedResult(), nil
	}

	scriptID := GetStringArg(args, "script_id", "")
	if scriptID == "" {
		return JSONResult(NewErrorResponse("script_id is required", nil)), nil
	}

	data := map[string]interface{}{}

	if v := GetStringArg(args, "name", ""); v != "" {
		data["name"] = v
	}
	if v := GetStringArg(args, "script", ""); v != "" {
		data["script"] = v
	}
	if v := GetStringArg(args, "description", ""); v != "" {
		data["description"] = v
	}
	if v, exists := args["active"]; exists {
		data["active"] = v
	}

	result, err := r.client.Put(fmt.Sprintf("/table/sys_script_include/%s", scriptID), data)
	if err != nil {
		return JSONResult(NewErrorResponse("Failed to update script include", err)), nil
	}

	if resultData, ok := result["result"].(map[string]interface{}); ok {
		return JSONResult(map[string]interface{}{
			"success":   true,
			"message":   "Script include updated successfully",
			"script_id": resultData["sys_id"],
		}), nil
	}

	return JSONResult(NewErrorResponse("Unexpected response from ServiceNow", nil)), nil
}

func (r *Registry) deleteScriptInclude(args map[string]interface{}) (*mcp.CallToolResult, error) {
	if r.readOnlyMode {
		return WriteBlockedResult(), nil
	}

	scriptID := GetStringArg(args, "script_id", "")
	if scriptID == "" {
		return JSONResult(NewErrorResponse("script_id is required", nil)), nil
	}

	_, err := r.client.Delete(fmt.Sprintf("/table/sys_script_include/%s", scriptID))
	if err != nil {
		return JSONResult(NewErrorResponse("Failed to delete script include", err)), nil
	}

	return JSONResult(map[string]interface{}{
		"success": true,
		"message": "Script include deleted successfully",
	}), nil
}
