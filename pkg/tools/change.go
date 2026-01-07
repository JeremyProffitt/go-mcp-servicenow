package tools

import (
	"fmt"
	"strings"

	"github.com/elastiflow/go-mcp-servicenow/pkg/mcp"
)

// registerChangeTools registers all change management tools
func (r *Registry) registerChangeTools(server *mcp.Server) int {
	count := 0

	// List Change Requests
	server.RegisterTool(mcp.Tool{
		Name:        "list_change_requests",
		Description: "List change requests from ServiceNow with filtering options",
		InputSchema: mcp.JSONSchema{
			Type: "object",
			Properties: map[string]mcp.Property{
				"limit": {
					Type:        "number",
					Description: "Maximum number of change requests to return (default: 10)",
					Default:     10,
				},
				"offset": {
					Type:        "number",
					Description: "Offset for pagination",
					Default:     0,
				},
				"state": {
					Type:        "string",
					Description: "Filter by state",
				},
				"type": {
					Type:        "string",
					Description: "Filter by change type (normal, standard, emergency)",
					Enum:        []string{"normal", "standard", "emergency"},
				},
				"assigned_to": {
					Type:        "string",
					Description: "Filter by assigned user",
				},
			},
		},
		Annotations: &mcp.ToolAnnotation{
			Title:        "List Change Requests",
			ReadOnlyHint: true,
		},
	}, func(args map[string]interface{}) (*mcp.CallToolResult, error) {
		return r.listChangeRequests(args)
	})
	count++

	// Get Change Request Details
	server.RegisterTool(mcp.Tool{
		Name:        "get_change_request",
		Description: "Get detailed information about a specific change request",
		InputSchema: mcp.JSONSchema{
			Type: "object",
			Properties: map[string]mcp.Property{
				"change_id": {
					Type:        "string",
					Description: "Change request number (e.g., CHG0010001) or sys_id",
				},
			},
			Required: []string{"change_id"},
		},
		Annotations: &mcp.ToolAnnotation{
			Title:        "Get Change Request",
			ReadOnlyHint: true,
		},
	}, func(args map[string]interface{}) (*mcp.CallToolResult, error) {
		return r.getChangeRequest(args)
	})
	count++

	// Write operations
	if !r.readOnlyMode {
		// Create Change Request
		server.RegisterTool(mcp.Tool{
			Name:        "create_change_request",
			Description: "Create a new change request in ServiceNow",
			InputSchema: mcp.JSONSchema{
				Type: "object",
				Properties: map[string]mcp.Property{
					"short_description": {
						Type:        "string",
						Description: "Short description of the change",
					},
					"type": {
						Type:        "string",
						Description: "Type of change (normal, standard, emergency)",
						Enum:        []string{"normal", "standard", "emergency"},
					},
					"description": {
						Type:        "string",
						Description: "Detailed description of the change",
					},
					"category": {
						Type:        "string",
						Description: "Category of the change",
					},
					"priority": {
						Type:        "string",
						Description: "Priority (1-5)",
					},
					"risk": {
						Type:        "string",
						Description: "Risk level",
					},
					"impact": {
						Type:        "string",
						Description: "Impact level",
					},
					"assigned_to": {
						Type:        "string",
						Description: "Assigned user",
					},
					"assignment_group": {
						Type:        "string",
						Description: "Assignment group",
					},
					"start_date": {
						Type:        "string",
						Description: "Planned start date",
					},
					"end_date": {
						Type:        "string",
						Description: "Planned end date",
					},
				},
				Required: []string{"short_description", "type"},
			},
		}, func(args map[string]interface{}) (*mcp.CallToolResult, error) {
			return r.createChangeRequest(args)
		})
		count++

		// Update Change Request
		server.RegisterTool(mcp.Tool{
			Name:        "update_change_request",
			Description: "Update an existing change request",
			InputSchema: mcp.JSONSchema{
				Type: "object",
				Properties: map[string]mcp.Property{
					"change_id": {
						Type:        "string",
						Description: "Change request number or sys_id",
					},
					"short_description": {
						Type:        "string",
						Description: "Short description",
					},
					"description": {
						Type:        "string",
						Description: "Detailed description",
					},
					"state": {
						Type:        "string",
						Description: "State of the change",
					},
					"priority": {
						Type:        "string",
						Description: "Priority",
					},
					"assigned_to": {
						Type:        "string",
						Description: "Assigned user",
					},
					"work_notes": {
						Type:        "string",
						Description: "Work notes to add",
					},
				},
				Required: []string{"change_id"},
			},
		}, func(args map[string]interface{}) (*mcp.CallToolResult, error) {
			return r.updateChangeRequest(args)
		})
		count++

		// Add Change Task
		server.RegisterTool(mcp.Tool{
			Name:        "add_change_task",
			Description: "Add a task to a change request",
			InputSchema: mcp.JSONSchema{
				Type: "object",
				Properties: map[string]mcp.Property{
					"change_id": {
						Type:        "string",
						Description: "Change request number or sys_id",
					},
					"short_description": {
						Type:        "string",
						Description: "Task description",
					},
					"assigned_to": {
						Type:        "string",
						Description: "Assigned user",
					},
					"planned_start_date": {
						Type:        "string",
						Description: "Planned start date",
					},
					"planned_end_date": {
						Type:        "string",
						Description: "Planned end date",
					},
				},
				Required: []string{"change_id", "short_description"},
			},
		}, func(args map[string]interface{}) (*mcp.CallToolResult, error) {
			return r.addChangeTask(args)
		})
		count++

		// Submit Change for Approval
		server.RegisterTool(mcp.Tool{
			Name:        "submit_change_for_approval",
			Description: "Submit a change request for approval",
			InputSchema: mcp.JSONSchema{
				Type: "object",
				Properties: map[string]mcp.Property{
					"change_id": {
						Type:        "string",
						Description: "Change request number or sys_id",
					},
				},
				Required: []string{"change_id"},
			},
		}, func(args map[string]interface{}) (*mcp.CallToolResult, error) {
			return r.submitChangeForApproval(args)
		})
		count++

		// Approve Change
		server.RegisterTool(mcp.Tool{
			Name:        "approve_change",
			Description: "Approve a change request",
			InputSchema: mcp.JSONSchema{
				Type: "object",
				Properties: map[string]mcp.Property{
					"change_id": {
						Type:        "string",
						Description: "Change request number or sys_id",
					},
					"comments": {
						Type:        "string",
						Description: "Approval comments",
					},
				},
				Required: []string{"change_id"},
			},
		}, func(args map[string]interface{}) (*mcp.CallToolResult, error) {
			return r.approveChange(args)
		})
		count++

		// Reject Change
		server.RegisterTool(mcp.Tool{
			Name:        "reject_change",
			Description: "Reject a change request",
			InputSchema: mcp.JSONSchema{
				Type: "object",
				Properties: map[string]mcp.Property{
					"change_id": {
						Type:        "string",
						Description: "Change request number or sys_id",
					},
					"reason": {
						Type:        "string",
						Description: "Rejection reason",
					},
				},
				Required: []string{"change_id", "reason"},
			},
		}, func(args map[string]interface{}) (*mcp.CallToolResult, error) {
			return r.rejectChange(args)
		})
		count++
	}

	return count
}

func (r *Registry) listChangeRequests(args map[string]interface{}) (*mcp.CallToolResult, error) {
	limit := GetIntArg(args, "limit", 10)
	offset := GetIntArg(args, "offset", 0)
	state := GetStringArg(args, "state", "")
	changeType := GetStringArg(args, "type", "")
	assignedTo := GetStringArg(args, "assigned_to", "")

	params := map[string]string{
		"sysparm_limit":                  fmt.Sprintf("%d", limit),
		"sysparm_offset":                 fmt.Sprintf("%d", offset),
		"sysparm_display_value":          "true",
		"sysparm_exclude_reference_link": "true",
	}

	var filters []string
	if state != "" {
		filters = append(filters, fmt.Sprintf("state=%s", state))
	}
	if changeType != "" {
		filters = append(filters, fmt.Sprintf("type=%s", changeType))
	}
	if assignedTo != "" {
		filters = append(filters, fmt.Sprintf("assigned_to=%s", assignedTo))
	}

	if len(filters) > 0 {
		params["sysparm_query"] = strings.Join(filters, "^")
	}

	result, err := r.client.Get("/table/change_request", params)
	if err != nil {
		return JSONResult(NewErrorResponse("Failed to list change requests", err)), nil
	}

	changes := []map[string]interface{}{}
	if resultList, ok := result["result"].([]interface{}); ok {
		for _, item := range resultList {
			if data, ok := item.(map[string]interface{}); ok {
				changes = append(changes, map[string]interface{}{
					"sys_id":            data["sys_id"],
					"number":            data["number"],
					"short_description": data["short_description"],
					"type":              data["type"],
					"state":             data["state"],
					"priority":          data["priority"],
					"risk":              data["risk"],
					"start_date":        data["start_date"],
					"end_date":          data["end_date"],
				})
			}
		}
	}

	return JSONResult(map[string]interface{}{
		"success":         true,
		"message":         fmt.Sprintf("Found %d change requests", len(changes)),
		"change_requests": changes,
	}), nil
}

func (r *Registry) getChangeRequest(args map[string]interface{}) (*mcp.CallToolResult, error) {
	changeID := GetStringArg(args, "change_id", "")
	if changeID == "" {
		return JSONResult(NewErrorResponse("change_id is required", nil)), nil
	}

	sysID, err := r.resolveChangeID(changeID)
	if err != nil {
		return JSONResult(NewErrorResponse("Failed to find change request", err)), nil
	}

	params := map[string]string{
		"sysparm_display_value":          "true",
		"sysparm_exclude_reference_link": "true",
	}

	result, err := r.client.Get(fmt.Sprintf("/table/change_request/%s", sysID), params)
	if err != nil {
		return JSONResult(NewErrorResponse("Failed to get change request", err)), nil
	}

	if data, ok := result["result"].(map[string]interface{}); ok {
		return JSONResult(map[string]interface{}{
			"success":        true,
			"message":        "Change request found",
			"change_request": data,
		}), nil
	}

	return JSONResult(map[string]interface{}{
		"success": false,
		"message": fmt.Sprintf("Change request not found: %s", changeID),
	}), nil
}

func (r *Registry) createChangeRequest(args map[string]interface{}) (*mcp.CallToolResult, error) {
	if r.readOnlyMode {
		return WriteBlockedResult(), nil
	}

	shortDesc := GetStringArg(args, "short_description", "")
	changeType := GetStringArg(args, "type", "")

	if shortDesc == "" || changeType == "" {
		return JSONResult(NewErrorResponse("short_description and type are required", nil)), nil
	}

	data := map[string]interface{}{
		"short_description": shortDesc,
		"type":              changeType,
	}

	if v := GetStringArg(args, "description", ""); v != "" {
		data["description"] = v
	}
	if v := GetStringArg(args, "category", ""); v != "" {
		data["category"] = v
	}
	if v := GetStringArg(args, "priority", ""); v != "" {
		data["priority"] = v
	}
	if v := GetStringArg(args, "risk", ""); v != "" {
		data["risk"] = v
	}
	if v := GetStringArg(args, "impact", ""); v != "" {
		data["impact"] = v
	}
	if v := GetStringArg(args, "assigned_to", ""); v != "" {
		data["assigned_to"] = v
	}
	if v := GetStringArg(args, "assignment_group", ""); v != "" {
		data["assignment_group"] = v
	}
	if v := GetStringArg(args, "start_date", ""); v != "" {
		data["start_date"] = v
	}
	if v := GetStringArg(args, "end_date", ""); v != "" {
		data["end_date"] = v
	}

	result, err := r.client.Post("/table/change_request", data)
	if err != nil {
		return JSONResult(NewErrorResponse("Failed to create change request", err)), nil
	}

	if resultData, ok := result["result"].(map[string]interface{}); ok {
		return JSONResult(map[string]interface{}{
			"success":       true,
			"message":       "Change request created successfully",
			"change_id":     resultData["sys_id"],
			"change_number": resultData["number"],
		}), nil
	}

	return JSONResult(NewErrorResponse("Unexpected response from ServiceNow", nil)), nil
}

func (r *Registry) updateChangeRequest(args map[string]interface{}) (*mcp.CallToolResult, error) {
	if r.readOnlyMode {
		return WriteBlockedResult(), nil
	}

	changeID := GetStringArg(args, "change_id", "")
	if changeID == "" {
		return JSONResult(NewErrorResponse("change_id is required", nil)), nil
	}

	sysID, err := r.resolveChangeID(changeID)
	if err != nil {
		return JSONResult(NewErrorResponse("Failed to find change request", err)), nil
	}

	data := map[string]interface{}{}

	if v := GetStringArg(args, "short_description", ""); v != "" {
		data["short_description"] = v
	}
	if v := GetStringArg(args, "description", ""); v != "" {
		data["description"] = v
	}
	if v := GetStringArg(args, "state", ""); v != "" {
		data["state"] = v
	}
	if v := GetStringArg(args, "priority", ""); v != "" {
		data["priority"] = v
	}
	if v := GetStringArg(args, "assigned_to", ""); v != "" {
		data["assigned_to"] = v
	}
	if v := GetStringArg(args, "work_notes", ""); v != "" {
		data["work_notes"] = v
	}

	result, err := r.client.Put(fmt.Sprintf("/table/change_request/%s", sysID), data)
	if err != nil {
		return JSONResult(NewErrorResponse("Failed to update change request", err)), nil
	}

	if resultData, ok := result["result"].(map[string]interface{}); ok {
		return JSONResult(map[string]interface{}{
			"success":       true,
			"message":       "Change request updated successfully",
			"change_id":     resultData["sys_id"],
			"change_number": resultData["number"],
		}), nil
	}

	return JSONResult(NewErrorResponse("Unexpected response from ServiceNow", nil)), nil
}

func (r *Registry) addChangeTask(args map[string]interface{}) (*mcp.CallToolResult, error) {
	if r.readOnlyMode {
		return WriteBlockedResult(), nil
	}

	changeID := GetStringArg(args, "change_id", "")
	shortDesc := GetStringArg(args, "short_description", "")

	if changeID == "" || shortDesc == "" {
		return JSONResult(NewErrorResponse("change_id and short_description are required", nil)), nil
	}

	sysID, err := r.resolveChangeID(changeID)
	if err != nil {
		return JSONResult(NewErrorResponse("Failed to find change request", err)), nil
	}

	data := map[string]interface{}{
		"change_request":    sysID,
		"short_description": shortDesc,
	}

	if v := GetStringArg(args, "assigned_to", ""); v != "" {
		data["assigned_to"] = v
	}
	if v := GetStringArg(args, "planned_start_date", ""); v != "" {
		data["planned_start_date"] = v
	}
	if v := GetStringArg(args, "planned_end_date", ""); v != "" {
		data["planned_end_date"] = v
	}

	result, err := r.client.Post("/table/change_task", data)
	if err != nil {
		return JSONResult(NewErrorResponse("Failed to add change task", err)), nil
	}

	if resultData, ok := result["result"].(map[string]interface{}); ok {
		return JSONResult(map[string]interface{}{
			"success":     true,
			"message":     "Change task added successfully",
			"task_id":     resultData["sys_id"],
			"task_number": resultData["number"],
		}), nil
	}

	return JSONResult(NewErrorResponse("Unexpected response from ServiceNow", nil)), nil
}

func (r *Registry) submitChangeForApproval(args map[string]interface{}) (*mcp.CallToolResult, error) {
	if r.readOnlyMode {
		return WriteBlockedResult(), nil
	}

	changeID := GetStringArg(args, "change_id", "")
	if changeID == "" {
		return JSONResult(NewErrorResponse("change_id is required", nil)), nil
	}

	sysID, err := r.resolveChangeID(changeID)
	if err != nil {
		return JSONResult(NewErrorResponse("Failed to find change request", err)), nil
	}

	// Update state to "Assess" (state -4) to trigger approval workflow
	data := map[string]interface{}{
		"state": "-4",
	}

	result, err := r.client.Put(fmt.Sprintf("/table/change_request/%s", sysID), data)
	if err != nil {
		return JSONResult(NewErrorResponse("Failed to submit change for approval", err)), nil
	}

	if resultData, ok := result["result"].(map[string]interface{}); ok {
		return JSONResult(map[string]interface{}{
			"success":       true,
			"message":       "Change request submitted for approval",
			"change_id":     resultData["sys_id"],
			"change_number": resultData["number"],
		}), nil
	}

	return JSONResult(NewErrorResponse("Unexpected response from ServiceNow", nil)), nil
}

func (r *Registry) approveChange(args map[string]interface{}) (*mcp.CallToolResult, error) {
	if r.readOnlyMode {
		return WriteBlockedResult(), nil
	}

	changeID := GetStringArg(args, "change_id", "")
	comments := GetStringArg(args, "comments", "")

	if changeID == "" {
		return JSONResult(NewErrorResponse("change_id is required", nil)), nil
	}

	sysID, err := r.resolveChangeID(changeID)
	if err != nil {
		return JSONResult(NewErrorResponse("Failed to find change request", err)), nil
	}

	// Find pending approval for this change
	params := map[string]string{
		"sysparm_query": fmt.Sprintf("sysapproval=%s^state=requested", sysID),
		"sysparm_limit": "1",
	}

	approvalResult, err := r.client.Get("/table/sysapproval_approver", params)
	if err != nil {
		return JSONResult(NewErrorResponse("Failed to find approval record", err)), nil
	}

	var approvalID string
	if resultList, ok := approvalResult["result"].([]interface{}); ok && len(resultList) > 0 {
		if approvalData, ok := resultList[0].(map[string]interface{}); ok {
			approvalID, _ = approvalData["sys_id"].(string)
		}
	}

	if approvalID == "" {
		return JSONResult(map[string]interface{}{
			"success": false,
			"message": "No pending approval found for this change request",
		}), nil
	}

	data := map[string]interface{}{
		"state": "approved",
	}
	if comments != "" {
		data["comments"] = comments
	}

	result, err := r.client.Put(fmt.Sprintf("/table/sysapproval_approver/%s", approvalID), data)
	if err != nil {
		return JSONResult(NewErrorResponse("Failed to approve change", err)), nil
	}

	if result["result"] != nil {
		return JSONResult(map[string]interface{}{
			"success": true,
			"message": "Change request approved",
		}), nil
	}

	return JSONResult(NewErrorResponse("Unexpected response from ServiceNow", nil)), nil
}

func (r *Registry) rejectChange(args map[string]interface{}) (*mcp.CallToolResult, error) {
	if r.readOnlyMode {
		return WriteBlockedResult(), nil
	}

	changeID := GetStringArg(args, "change_id", "")
	reason := GetStringArg(args, "reason", "")

	if changeID == "" || reason == "" {
		return JSONResult(NewErrorResponse("change_id and reason are required", nil)), nil
	}

	sysID, err := r.resolveChangeID(changeID)
	if err != nil {
		return JSONResult(NewErrorResponse("Failed to find change request", err)), nil
	}

	// Find pending approval for this change
	params := map[string]string{
		"sysparm_query": fmt.Sprintf("sysapproval=%s^state=requested", sysID),
		"sysparm_limit": "1",
	}

	approvalResult, err := r.client.Get("/table/sysapproval_approver", params)
	if err != nil {
		return JSONResult(NewErrorResponse("Failed to find approval record", err)), nil
	}

	var approvalID string
	if resultList, ok := approvalResult["result"].([]interface{}); ok && len(resultList) > 0 {
		if approvalData, ok := resultList[0].(map[string]interface{}); ok {
			approvalID, _ = approvalData["sys_id"].(string)
		}
	}

	if approvalID == "" {
		return JSONResult(map[string]interface{}{
			"success": false,
			"message": "No pending approval found for this change request",
		}), nil
	}

	data := map[string]interface{}{
		"state":    "rejected",
		"comments": reason,
	}

	result, err := r.client.Put(fmt.Sprintf("/table/sysapproval_approver/%s", approvalID), data)
	if err != nil {
		return JSONResult(NewErrorResponse("Failed to reject change", err)), nil
	}

	if result["result"] != nil {
		return JSONResult(map[string]interface{}{
			"success": true,
			"message": "Change request rejected",
		}), nil
	}

	return JSONResult(NewErrorResponse("Unexpected response from ServiceNow", nil)), nil
}

// resolveChangeID resolves a change number to sys_id
func (r *Registry) resolveChangeID(changeID string) (string, error) {
	if IsSysID(changeID) {
		return changeID, nil
	}

	params := map[string]string{
		"sysparm_query": fmt.Sprintf("number=%s", changeID),
		"sysparm_limit": "1",
	}

	result, err := r.client.Get("/table/change_request", params)
	if err != nil {
		return "", err
	}

	if resultList, ok := result["result"].([]interface{}); ok && len(resultList) > 0 {
		if data, ok := resultList[0].(map[string]interface{}); ok {
			if sysID, ok := data["sys_id"].(string); ok {
				return sysID, nil
			}
		}
	}

	return "", fmt.Errorf("change request not found: %s", changeID)
}
