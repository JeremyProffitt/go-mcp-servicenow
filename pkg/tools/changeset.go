package tools

import (
	"fmt"
	"strings"

	"github.com/elastiflow/go-mcp-servicenow/pkg/mcp"
)

// registerChangesetTools registers all changeset/update set tools
func (r *Registry) registerChangesetTools(server *mcp.Server) int {
	count := 0

	// Helper for limit constraints
	limitMin := float64(1)
	limitMax := float64(1000)

	// List Changesets
	server.RegisterTool(mcp.Tool{
		Name:        "list_changesets",
		Description: "List changesets (update sets) with optional filtering. Update sets are containers for capturing configuration changes.",
		InputSchema: mcp.JSONSchema{
			Type: "object",
			Properties: map[string]mcp.Property{
				"limit": {
					Type:        "number",
					Description: "Maximum number of changesets to return (default: 50)",
					Default:     50,
					Minimum:     &limitMin,
					Maximum:     &limitMax,
				},
				"state": {
					Type:        "string",
					Description: "Filter by state",
					Enum:        []string{"in progress", "complete", "ignore"},
				},
				"created_by": {
					Type:        "string",
					Description: "Filter by creator username",
				},
			},
		},
		Annotations: &mcp.ToolAnnotation{
			Title:        "List Changesets",
			ReadOnlyHint: true,
		},
	}, func(args map[string]interface{}) (*mcp.CallToolResult, error) {
		return r.listChangesets(args)
	})
	count++

	// Get Changeset Details
	server.RegisterTool(mcp.Tool{
		Name:        "get_changeset",
		Description: "Get detailed information about a changeset (update set) including contained changes.",
		InputSchema: mcp.JSONSchema{
			Type: "object",
			Properties: map[string]mcp.Property{
				"changeset_id": {
					Type:        "string",
					Description: "Changeset sys_id (e.g., 'a1b2c3d4e5f6g7h8i9j0k1l2m3n4o5p6') or name. Accepts both formats.",
				},
			},
			Required: []string{"changeset_id"},
		},
		Annotations: &mcp.ToolAnnotation{
			Title:        "Get Changeset",
			ReadOnlyHint: true,
		},
	}, func(args map[string]interface{}) (*mcp.CallToolResult, error) {
		return r.getChangeset(args)
	})
	count++

	// Write operations
	if !r.readOnlyMode {
		// Create Changeset
		server.RegisterTool(mcp.Tool{
			Name:        "create_changeset",
			Description: "Create a new changeset (update set). Use update sets to capture and migrate configuration changes.",
			InputSchema: mcp.JSONSchema{
				Type: "object",
				Properties: map[string]mcp.Property{
					"name": {
						Type:        "string",
						Description: "Changeset name (must be unique)",
					},
					"description": {
						Type:        "string",
						Description: "Changeset description",
					},
					"application": {
						Type:        "string",
						Description: "Application sys_id (e.g., 'a1b2c3d4e5f6g7h8i9j0k1l2m3n4o5p6')",
					},
				},
				Required: []string{"name"},
			},
			Annotations: &mcp.ToolAnnotation{
				Title: "Create Changeset",
			},
		}, func(args map[string]interface{}) (*mcp.CallToolResult, error) {
			return r.createChangeset(args)
		})
		count++

		// Update Changeset
		server.RegisterTool(mcp.Tool{
			Name:        "update_changeset",
			Description: "Update an existing changeset. At least one field besides changeset_id must be provided.",
			InputSchema: mcp.JSONSchema{
				Type: "object",
				Properties: map[string]mcp.Property{
					"changeset_id": {
						Type:        "string",
						Description: "Changeset sys_id (e.g., 'a1b2c3d4e5f6g7h8i9j0k1l2m3n4o5p6')",
					},
					"name": {
						Type:        "string",
						Description: "Changeset name",
					},
					"description": {
						Type:        "string",
						Description: "Changeset description",
					},
				},
				Required: []string{"changeset_id"},
			},
			Annotations: &mcp.ToolAnnotation{
				Title: "Update Changeset",
			},
		}, func(args map[string]interface{}) (*mcp.CallToolResult, error) {
			return r.updateChangeset(args)
		})
		count++

		// Commit Changeset
		server.RegisterTool(mcp.Tool{
			Name:        "commit_changeset",
			Description: "Commit a changeset by marking it as complete. Completed changesets can be exported or deployed to other instances.",
			InputSchema: mcp.JSONSchema{
				Type: "object",
				Properties: map[string]mcp.Property{
					"changeset_id": {
						Type:        "string",
						Description: "Changeset sys_id (e.g., 'a1b2c3d4e5f6g7h8i9j0k1l2m3n4o5p6')",
					},
				},
				Required: []string{"changeset_id"},
			},
			Annotations: &mcp.ToolAnnotation{
				Title: "Commit Changeset",
			},
		}, func(args map[string]interface{}) (*mcp.CallToolResult, error) {
			return r.commitChangeset(args)
		})
		count++
	}

	return count
}

func (r *Registry) listChangesets(args map[string]interface{}) (*mcp.CallToolResult, error) {
	limit := GetIntArg(args, "limit", 50)
	state := GetStringArg(args, "state", "")
	createdBy := GetStringArg(args, "created_by", "")

	params := map[string]string{
		"sysparm_limit":                  fmt.Sprintf("%d", limit),
		"sysparm_display_value":          "true",
		"sysparm_exclude_reference_link": "true",
	}

	var filters []string
	if state != "" {
		filters = append(filters, fmt.Sprintf("state=%s", state))
	}
	if createdBy != "" {
		filters = append(filters, fmt.Sprintf("sys_created_by=%s", createdBy))
	}

	if len(filters) > 0 {
		params["sysparm_query"] = strings.Join(filters, "^")
	}

	result, err := r.client.Get("/table/sys_update_set", params)
	if err != nil {
		return JSONResult(NewErrorResponse("Failed to list changesets", err)), nil
	}

	changesets := []map[string]interface{}{}
	if resultList, ok := result["result"].([]interface{}); ok {
		for _, item := range resultList {
			if data, ok := item.(map[string]interface{}); ok {
				changesets = append(changesets, map[string]interface{}{
					"sys_id":         data["sys_id"],
					"name":           data["name"],
					"description":    data["description"],
					"state":          data["state"],
					"application":    data["application"],
					"sys_created_by": data["sys_created_by"],
					"sys_created_on": data["sys_created_on"],
				})
			}
		}
	}

	return JSONResult(map[string]interface{}{
		"success":    true,
		"message":    fmt.Sprintf("Found %d changesets", len(changesets)),
		"changesets": changesets,
	}), nil
}

func (r *Registry) getChangeset(args map[string]interface{}) (*mcp.CallToolResult, error) {
	changesetID := GetStringArg(args, "changeset_id", "")
	if changesetID == "" {
		return JSONResult(NewErrorResponse("changeset_id is required", nil)), nil
	}

	var params map[string]string
	var endpoint string

	if IsSysID(changesetID) {
		endpoint = fmt.Sprintf("/table/sys_update_set/%s", changesetID)
		params = map[string]string{
			"sysparm_display_value":          "true",
			"sysparm_exclude_reference_link": "true",
		}
	} else {
		endpoint = "/table/sys_update_set"
		params = map[string]string{
			"sysparm_query":                  fmt.Sprintf("name=%s", changesetID),
			"sysparm_limit":                  "1",
			"sysparm_display_value":          "true",
			"sysparm_exclude_reference_link": "true",
		}
	}

	result, err := r.client.Get(endpoint, params)
	if err != nil {
		return JSONResult(NewErrorResponse("Failed to get changeset", err)), nil
	}

	var changesetData map[string]interface{}
	if IsSysID(changesetID) {
		changesetData, _ = result["result"].(map[string]interface{})
	} else {
		if resultList, ok := result["result"].([]interface{}); ok && len(resultList) > 0 {
			changesetData, _ = resultList[0].(map[string]interface{})
		}
	}

	if changesetData == nil {
		return JSONResult(map[string]interface{}{
			"success": false,
			"message": fmt.Sprintf("Changeset not found: %s", changesetID),
		}), nil
	}

	return JSONResult(map[string]interface{}{
		"success":   true,
		"message":   "Changeset found",
		"changeset": changesetData,
	}), nil
}

func (r *Registry) createChangeset(args map[string]interface{}) (*mcp.CallToolResult, error) {
	if r.readOnlyMode {
		return WriteBlockedResult(), nil
	}

	name := GetStringArg(args, "name", "")
	if name == "" {
		return JSONResult(NewErrorResponse("name is required", nil)), nil
	}

	data := map[string]interface{}{
		"name": name,
	}

	if v := GetStringArg(args, "description", ""); v != "" {
		data["description"] = v
	}
	if v := GetStringArg(args, "application", ""); v != "" {
		data["application"] = v
	}

	result, err := r.client.Post("/table/sys_update_set", data)
	if err != nil {
		return JSONResult(NewErrorResponse("Failed to create changeset", err)), nil
	}

	if resultData, ok := result["result"].(map[string]interface{}); ok {
		return JSONResult(map[string]interface{}{
			"success":      true,
			"message":      "Changeset created successfully",
			"changeset_id": resultData["sys_id"],
		}), nil
	}

	return JSONResult(NewErrorResponse("Unexpected response from ServiceNow", nil)), nil
}

func (r *Registry) updateChangeset(args map[string]interface{}) (*mcp.CallToolResult, error) {
	if r.readOnlyMode {
		return WriteBlockedResult(), nil
	}

	changesetID := GetStringArg(args, "changeset_id", "")
	if changesetID == "" {
		return JSONResult(NewErrorResponse("changeset_id is required", nil)), nil
	}

	data := map[string]interface{}{}

	if v := GetStringArg(args, "name", ""); v != "" {
		data["name"] = v
	}
	if v := GetStringArg(args, "description", ""); v != "" {
		data["description"] = v
	}

	result, err := r.client.Put(fmt.Sprintf("/table/sys_update_set/%s", changesetID), data)
	if err != nil {
		return JSONResult(NewErrorResponse("Failed to update changeset", err)), nil
	}

	if resultData, ok := result["result"].(map[string]interface{}); ok {
		return JSONResult(map[string]interface{}{
			"success":      true,
			"message":      "Changeset updated successfully",
			"changeset_id": resultData["sys_id"],
		}), nil
	}

	return JSONResult(NewErrorResponse("Unexpected response from ServiceNow", nil)), nil
}

func (r *Registry) commitChangeset(args map[string]interface{}) (*mcp.CallToolResult, error) {
	if r.readOnlyMode {
		return WriteBlockedResult(), nil
	}

	changesetID := GetStringArg(args, "changeset_id", "")
	if changesetID == "" {
		return JSONResult(NewErrorResponse("changeset_id is required", nil)), nil
	}

	data := map[string]interface{}{
		"state": "complete",
	}

	result, err := r.client.Put(fmt.Sprintf("/table/sys_update_set/%s", changesetID), data)
	if err != nil {
		return JSONResult(NewErrorResponse("Failed to commit changeset", err)), nil
	}

	if resultData, ok := result["result"].(map[string]interface{}); ok {
		return JSONResult(map[string]interface{}{
			"success":      true,
			"message":      "Changeset committed successfully",
			"changeset_id": resultData["sys_id"],
		}), nil
	}

	return JSONResult(NewErrorResponse("Unexpected response from ServiceNow", nil)), nil
}
