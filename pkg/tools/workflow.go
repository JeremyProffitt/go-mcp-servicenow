package tools

import (
	"fmt"
	"strings"

	"github.com/elastiflow/go-mcp-servicenow/pkg/mcp"
)

// registerWorkflowTools registers all workflow management tools
func (r *Registry) registerWorkflowTools(server *mcp.Server) int {
	count := 0

	// Helper for limit constraints
	limitMin := float64(1)
	limitMax := float64(1000)

	// List Workflows
	server.RegisterTool(mcp.Tool{
		Name:        "list_workflows",
		Description: "List workflows with optional filtering by active status or table. Workflows automate business processes.",
		InputSchema: mcp.JSONSchema{
			Type: "object",
			Properties: map[string]mcp.Property{
				"limit": {
					Type:        "number",
					Description: "Maximum number of workflows to return (default: 50)",
					Default:     50,
					Minimum:     &limitMin,
					Maximum:     &limitMax,
				},
				"active": {
					Type:        "boolean",
					Description: "Filter by active status (true = only active workflows, false = only inactive)",
				},
				"table": {
					Type:        "string",
					Description: "Filter by table name (e.g., 'incident', 'change_request', 'sc_req_item')",
				},
			},
		},
		Annotations: &mcp.ToolAnnotation{
			Title:        "List Workflows",
			ReadOnlyHint: true,
		},
	}, func(args map[string]interface{}) (*mcp.CallToolResult, error) {
		return r.listWorkflows(args)
	})
	count++

	// Get Workflow
	server.RegisterTool(mcp.Tool{
		Name:        "get_workflow",
		Description: "Get detailed information about a specific workflow including configuration and activities.",
		InputSchema: mcp.JSONSchema{
			Type: "object",
			Properties: map[string]mcp.Property{
				"workflow_id": {
					Type:        "string",
					Description: "Workflow sys_id (e.g., 'a1b2c3d4e5f6g7h8i9j0k1l2m3n4o5p6') or name. Accepts both formats.",
				},
			},
			Required: []string{"workflow_id"},
		},
		Annotations: &mcp.ToolAnnotation{
			Title:        "Get Workflow",
			ReadOnlyHint: true,
		},
	}, func(args map[string]interface{}) (*mcp.CallToolResult, error) {
		return r.getWorkflow(args)
	})
	count++

	// Write operations
	if !r.readOnlyMode {
		// Create Workflow
		server.RegisterTool(mcp.Tool{
			Name:        "create_workflow",
			Description: "Create a new workflow definition. The workflow is created inactive by default.",
			InputSchema: mcp.JSONSchema{
				Type: "object",
				Properties: map[string]mcp.Property{
					"name": {
						Type:        "string",
						Description: "Workflow name (must be unique)",
					},
					"table": {
						Type:        "string",
						Description: "Table name the workflow applies to (e.g., 'incident', 'change_request', 'sc_req_item')",
					},
					"description": {
						Type:        "string",
						Description: "Workflow description",
					},
				},
				Required: []string{"name", "table"},
			},
			Annotations: &mcp.ToolAnnotation{
				Title: "Create Workflow",
			},
		}, func(args map[string]interface{}) (*mcp.CallToolResult, error) {
			return r.createWorkflow(args)
		})
		count++

		// Update Workflow
		server.RegisterTool(mcp.Tool{
			Name:        "update_workflow",
			Description: "Update an existing workflow. At least one field besides workflow_id must be provided.",
			InputSchema: mcp.JSONSchema{
				Type: "object",
				Properties: map[string]mcp.Property{
					"workflow_id": {
						Type:        "string",
						Description: "Workflow sys_id (e.g., 'a1b2c3d4e5f6g7h8i9j0k1l2m3n4o5p6')",
					},
					"name": {
						Type:        "string",
						Description: "Workflow name",
					},
					"description": {
						Type:        "string",
						Description: "Workflow description",
					},
					"active": {
						Type:        "boolean",
						Description: "Active status (true to activate, false to deactivate)",
					},
				},
				Required: []string{"workflow_id"},
			},
			Annotations: &mcp.ToolAnnotation{
				Title: "Update Workflow",
			},
		}, func(args map[string]interface{}) (*mcp.CallToolResult, error) {
			return r.updateWorkflow(args)
		})
		count++

		// Delete Workflow
		server.RegisterTool(mcp.Tool{
			Name:        "delete_workflow",
			Description: "Permanently delete a workflow. This action cannot be undone.",
			InputSchema: mcp.JSONSchema{
				Type: "object",
				Properties: map[string]mcp.Property{
					"workflow_id": {
						Type:        "string",
						Description: "Workflow sys_id (e.g., 'a1b2c3d4e5f6g7h8i9j0k1l2m3n4o5p6')",
					},
				},
				Required: []string{"workflow_id"},
			},
			Annotations: &mcp.ToolAnnotation{
				Title:           "Delete Workflow",
				DestructiveHint: true,
			},
		}, func(args map[string]interface{}) (*mcp.CallToolResult, error) {
			return r.deleteWorkflow(args)
		})
		count++
	}

	return count
}

func (r *Registry) listWorkflows(args map[string]interface{}) (*mcp.CallToolResult, error) {
	limit := GetIntArg(args, "limit", 50)
	table := GetStringArg(args, "table", "")

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
	if table != "" {
		filters = append(filters, fmt.Sprintf("table=%s", table))
	}

	if len(filters) > 0 {
		params["sysparm_query"] = strings.Join(filters, "^")
	}

	result, err := r.client.Get("/table/wf_workflow", params)
	if err != nil {
		return JSONResult(NewErrorResponse("Failed to list workflows", err)), nil
	}

	workflows := []map[string]interface{}{}
	if resultList, ok := result["result"].([]interface{}); ok {
		for _, item := range resultList {
			if data, ok := item.(map[string]interface{}); ok {
				workflows = append(workflows, map[string]interface{}{
					"sys_id":      data["sys_id"],
					"name":        data["name"],
					"table":       data["table"],
					"description": data["description"],
					"active":      data["active"],
				})
			}
		}
	}

	return JSONResult(map[string]interface{}{
		"success":   true,
		"message":   fmt.Sprintf("Found %d workflows", len(workflows)),
		"workflows": workflows,
	}), nil
}

func (r *Registry) getWorkflow(args map[string]interface{}) (*mcp.CallToolResult, error) {
	workflowID := GetStringArg(args, "workflow_id", "")
	if workflowID == "" {
		return JSONResult(NewErrorResponse("workflow_id is required", nil)), nil
	}

	var params map[string]string
	var endpoint string

	if IsSysID(workflowID) {
		endpoint = fmt.Sprintf("/table/wf_workflow/%s", workflowID)
		params = map[string]string{
			"sysparm_display_value":          "true",
			"sysparm_exclude_reference_link": "true",
		}
	} else {
		endpoint = "/table/wf_workflow"
		params = map[string]string{
			"sysparm_query":                  fmt.Sprintf("name=%s", workflowID),
			"sysparm_limit":                  "1",
			"sysparm_display_value":          "true",
			"sysparm_exclude_reference_link": "true",
		}
	}

	result, err := r.client.Get(endpoint, params)
	if err != nil {
		return JSONResult(NewErrorResponse("Failed to get workflow", err)), nil
	}

	var workflowData map[string]interface{}
	if IsSysID(workflowID) {
		workflowData, _ = result["result"].(map[string]interface{})
	} else {
		if resultList, ok := result["result"].([]interface{}); ok && len(resultList) > 0 {
			workflowData, _ = resultList[0].(map[string]interface{})
		}
	}

	if workflowData == nil {
		return JSONResult(map[string]interface{}{
			"success": false,
			"message": fmt.Sprintf("Workflow not found: %s", workflowID),
		}), nil
	}

	return JSONResult(map[string]interface{}{
		"success":  true,
		"message":  "Workflow found",
		"workflow": workflowData,
	}), nil
}

func (r *Registry) createWorkflow(args map[string]interface{}) (*mcp.CallToolResult, error) {
	if r.readOnlyMode {
		return WriteBlockedResult(), nil
	}

	name := GetStringArg(args, "name", "")
	table := GetStringArg(args, "table", "")

	if name == "" || table == "" {
		return JSONResult(NewErrorResponse("name and table are required", nil)), nil
	}

	data := map[string]interface{}{
		"name":  name,
		"table": table,
	}

	if v := GetStringArg(args, "description", ""); v != "" {
		data["description"] = v
	}

	result, err := r.client.Post("/table/wf_workflow", data)
	if err != nil {
		return JSONResult(NewErrorResponse("Failed to create workflow", err)), nil
	}

	if resultData, ok := result["result"].(map[string]interface{}); ok {
		return JSONResult(map[string]interface{}{
			"success":     true,
			"message":     "Workflow created successfully",
			"workflow_id": resultData["sys_id"],
		}), nil
	}

	return JSONResult(NewErrorResponse("Unexpected response from ServiceNow", nil)), nil
}

func (r *Registry) updateWorkflow(args map[string]interface{}) (*mcp.CallToolResult, error) {
	if r.readOnlyMode {
		return WriteBlockedResult(), nil
	}

	workflowID := GetStringArg(args, "workflow_id", "")
	if workflowID == "" {
		return JSONResult(NewErrorResponse("workflow_id is required", nil)), nil
	}

	data := map[string]interface{}{}

	if v := GetStringArg(args, "name", ""); v != "" {
		data["name"] = v
	}
	if v := GetStringArg(args, "description", ""); v != "" {
		data["description"] = v
	}
	if v, exists := args["active"]; exists {
		data["active"] = v
	}

	result, err := r.client.Put(fmt.Sprintf("/table/wf_workflow/%s", workflowID), data)
	if err != nil {
		return JSONResult(NewErrorResponse("Failed to update workflow", err)), nil
	}

	if resultData, ok := result["result"].(map[string]interface{}); ok {
		return JSONResult(map[string]interface{}{
			"success":     true,
			"message":     "Workflow updated successfully",
			"workflow_id": resultData["sys_id"],
		}), nil
	}

	return JSONResult(NewErrorResponse("Unexpected response from ServiceNow", nil)), nil
}

func (r *Registry) deleteWorkflow(args map[string]interface{}) (*mcp.CallToolResult, error) {
	if r.readOnlyMode {
		return WriteBlockedResult(), nil
	}

	workflowID := GetStringArg(args, "workflow_id", "")
	if workflowID == "" {
		return JSONResult(NewErrorResponse("workflow_id is required", nil)), nil
	}

	_, err := r.client.Delete(fmt.Sprintf("/table/wf_workflow/%s", workflowID))
	if err != nil {
		return JSONResult(NewErrorResponse("Failed to delete workflow", err)), nil
	}

	return JSONResult(map[string]interface{}{
		"success": true,
		"message": "Workflow deleted successfully",
	}), nil
}
