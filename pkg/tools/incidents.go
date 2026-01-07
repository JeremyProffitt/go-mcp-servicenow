package tools

import (
	"fmt"
	"strings"

	"github.com/elastiflow/go-mcp-servicenow/pkg/mcp"
)

// registerIncidentTools registers all incident management tools
func (r *Registry) registerIncidentTools(server *mcp.Server) int {
	count := 0

	// List Incidents (read-only)
	server.RegisterTool(mcp.Tool{
		Name:        "list_incidents",
		Description: "List incidents from ServiceNow with optional filtering",
		InputSchema: mcp.JSONSchema{
			Type: "object",
			Properties: map[string]mcp.Property{
				"limit": {
					Type:        "number",
					Description: "Maximum number of incidents to return (default: 10)",
					Default:     10,
				},
				"offset": {
					Type:        "number",
					Description: "Offset for pagination (default: 0)",
					Default:     0,
				},
				"state": {
					Type:        "string",
					Description: "Filter by incident state",
				},
				"assigned_to": {
					Type:        "string",
					Description: "Filter by assigned user",
				},
				"category": {
					Type:        "string",
					Description: "Filter by category",
				},
				"query": {
					Type:        "string",
					Description: "Search query for incidents",
				},
			},
		},
		Annotations: &mcp.ToolAnnotation{
			Title:        "List Incidents",
			ReadOnlyHint: true,
		},
	}, func(args map[string]interface{}) (*mcp.CallToolResult, error) {
		return r.listIncidents(args)
	})
	count++

	// Get Incident by Number (read-only)
	server.RegisterTool(mcp.Tool{
		Name:        "get_incident",
		Description: "Get a specific incident by number or sys_id",
		InputSchema: mcp.JSONSchema{
			Type: "object",
			Properties: map[string]mcp.Property{
				"incident_id": {
					Type:        "string",
					Description: "Incident number (e.g., INC0010001) or sys_id",
				},
			},
			Required: []string{"incident_id"},
		},
		Annotations: &mcp.ToolAnnotation{
			Title:        "Get Incident",
			ReadOnlyHint: true,
		},
	}, func(args map[string]interface{}) (*mcp.CallToolResult, error) {
		return r.getIncident(args)
	})
	count++

	// Write operations (only if not read-only mode)
	if !r.readOnlyMode {
		// Create Incident
		server.RegisterTool(mcp.Tool{
			Name:        "create_incident",
			Description: "Create a new incident in ServiceNow",
			InputSchema: mcp.JSONSchema{
				Type: "object",
				Properties: map[string]mcp.Property{
					"short_description": {
						Type:        "string",
						Description: "Short description of the incident",
					},
					"description": {
						Type:        "string",
						Description: "Detailed description of the incident",
					},
					"caller_id": {
						Type:        "string",
						Description: "User who reported the incident",
					},
					"category": {
						Type:        "string",
						Description: "Category of the incident",
					},
					"subcategory": {
						Type:        "string",
						Description: "Subcategory of the incident",
					},
					"priority": {
						Type:        "string",
						Description: "Priority of the incident (1-5)",
					},
					"impact": {
						Type:        "string",
						Description: "Impact of the incident (1-3)",
					},
					"urgency": {
						Type:        "string",
						Description: "Urgency of the incident (1-3)",
					},
					"assigned_to": {
						Type:        "string",
						Description: "User assigned to the incident",
					},
					"assignment_group": {
						Type:        "string",
						Description: "Group assigned to the incident",
					},
				},
				Required: []string{"short_description"},
			},
			Annotations: &mcp.ToolAnnotation{
				Title:           "Create Incident",
				DestructiveHint: false,
			},
		}, func(args map[string]interface{}) (*mcp.CallToolResult, error) {
			return r.createIncident(args)
		})
		count++

		// Update Incident
		server.RegisterTool(mcp.Tool{
			Name:        "update_incident",
			Description: "Update an existing incident in ServiceNow",
			InputSchema: mcp.JSONSchema{
				Type: "object",
				Properties: map[string]mcp.Property{
					"incident_id": {
						Type:        "string",
						Description: "Incident ID or sys_id",
					},
					"short_description": {
						Type:        "string",
						Description: "Short description of the incident",
					},
					"description": {
						Type:        "string",
						Description: "Detailed description of the incident",
					},
					"state": {
						Type:        "string",
						Description: "State of the incident",
					},
					"category": {
						Type:        "string",
						Description: "Category of the incident",
					},
					"priority": {
						Type:        "string",
						Description: "Priority of the incident",
					},
					"impact": {
						Type:        "string",
						Description: "Impact of the incident",
					},
					"urgency": {
						Type:        "string",
						Description: "Urgency of the incident",
					},
					"assigned_to": {
						Type:        "string",
						Description: "User assigned to the incident",
					},
					"assignment_group": {
						Type:        "string",
						Description: "Group assigned to the incident",
					},
					"work_notes": {
						Type:        "string",
						Description: "Work notes to add to the incident",
					},
				},
				Required: []string{"incident_id"},
			},
			Annotations: &mcp.ToolAnnotation{
				Title: "Update Incident",
			},
		}, func(args map[string]interface{}) (*mcp.CallToolResult, error) {
			return r.updateIncident(args)
		})
		count++

		// Add Comment
		server.RegisterTool(mcp.Tool{
			Name:        "add_incident_comment",
			Description: "Add a comment to an incident in ServiceNow",
			InputSchema: mcp.JSONSchema{
				Type: "object",
				Properties: map[string]mcp.Property{
					"incident_id": {
						Type:        "string",
						Description: "Incident ID or sys_id",
					},
					"comment": {
						Type:        "string",
						Description: "Comment to add to the incident",
					},
					"is_work_note": {
						Type:        "boolean",
						Description: "Whether the comment is a work note (default: false)",
						Default:     false,
					},
				},
				Required: []string{"incident_id", "comment"},
			},
			Annotations: &mcp.ToolAnnotation{
				Title: "Add Incident Comment",
			},
		}, func(args map[string]interface{}) (*mcp.CallToolResult, error) {
			return r.addIncidentComment(args)
		})
		count++

		// Resolve Incident
		server.RegisterTool(mcp.Tool{
			Name:        "resolve_incident",
			Description: "Resolve an incident in ServiceNow",
			InputSchema: mcp.JSONSchema{
				Type: "object",
				Properties: map[string]mcp.Property{
					"incident_id": {
						Type:        "string",
						Description: "Incident ID or sys_id",
					},
					"resolution_code": {
						Type:        "string",
						Description: "Resolution code for the incident",
					},
					"resolution_notes": {
						Type:        "string",
						Description: "Resolution notes for the incident",
					},
				},
				Required: []string{"incident_id", "resolution_code", "resolution_notes"},
			},
			Annotations: &mcp.ToolAnnotation{
				Title: "Resolve Incident",
			},
		}, func(args map[string]interface{}) (*mcp.CallToolResult, error) {
			return r.resolveIncident(args)
		})
		count++
	}

	return count
}

func (r *Registry) listIncidents(args map[string]interface{}) (*mcp.CallToolResult, error) {
	limit := GetIntArg(args, "limit", 10)
	offset := GetIntArg(args, "offset", 0)
	state := GetStringArg(args, "state", "")
	assignedTo := GetStringArg(args, "assigned_to", "")
	category := GetStringArg(args, "category", "")
	query := GetStringArg(args, "query", "")

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
	if assignedTo != "" {
		filters = append(filters, fmt.Sprintf("assigned_to=%s", assignedTo))
	}
	if category != "" {
		filters = append(filters, fmt.Sprintf("category=%s", category))
	}
	if query != "" {
		filters = append(filters, fmt.Sprintf("short_descriptionLIKE%s^ORdescriptionLIKE%s", query, query))
	}

	if len(filters) > 0 {
		params["sysparm_query"] = strings.Join(filters, "^")
	}

	result, err := r.client.Get("/table/incident", params)
	if err != nil {
		return JSONResult(NewErrorResponse("Failed to list incidents", err)), nil
	}

	incidents := []map[string]interface{}{}
	if resultList, ok := result["result"].([]interface{}); ok {
		for _, item := range resultList {
			if incidentData, ok := item.(map[string]interface{}); ok {
				incident := map[string]interface{}{
					"sys_id":            incidentData["sys_id"],
					"number":            incidentData["number"],
					"short_description": incidentData["short_description"],
					"description":       incidentData["description"],
					"state":             incidentData["state"],
					"priority":          incidentData["priority"],
					"category":          incidentData["category"],
					"subcategory":       incidentData["subcategory"],
					"created_on":        incidentData["sys_created_on"],
					"updated_on":        incidentData["sys_updated_on"],
				}

				// Handle assigned_to which could be a string or object
				if assignedTo, ok := incidentData["assigned_to"].(map[string]interface{}); ok {
					incident["assigned_to"] = assignedTo["display_value"]
				} else {
					incident["assigned_to"] = incidentData["assigned_to"]
				}

				incidents = append(incidents, incident)
			}
		}
	}

	return JSONResult(map[string]interface{}{
		"success":   true,
		"message":   fmt.Sprintf("Found %d incidents", len(incidents)),
		"incidents": incidents,
	}), nil
}

func (r *Registry) getIncident(args map[string]interface{}) (*mcp.CallToolResult, error) {
	incidentID := GetStringArg(args, "incident_id", "")
	if incidentID == "" {
		return JSONResult(NewErrorResponse("incident_id is required", nil)), nil
	}

	var params map[string]string
	var endpoint string

	if IsSysID(incidentID) {
		endpoint = fmt.Sprintf("/table/incident/%s", incidentID)
		params = map[string]string{
			"sysparm_display_value":          "true",
			"sysparm_exclude_reference_link": "true",
		}
	} else {
		endpoint = "/table/incident"
		params = map[string]string{
			"sysparm_query":                  fmt.Sprintf("number=%s", incidentID),
			"sysparm_limit":                  "1",
			"sysparm_display_value":          "true",
			"sysparm_exclude_reference_link": "true",
		}
	}

	result, err := r.client.Get(endpoint, params)
	if err != nil {
		return JSONResult(NewErrorResponse("Failed to get incident", err)), nil
	}

	var incidentData map[string]interface{}
	if IsSysID(incidentID) {
		incidentData, _ = result["result"].(map[string]interface{})
	} else {
		if resultList, ok := result["result"].([]interface{}); ok && len(resultList) > 0 {
			incidentData, _ = resultList[0].(map[string]interface{})
		}
	}

	if incidentData == nil {
		return JSONResult(map[string]interface{}{
			"success": false,
			"message": fmt.Sprintf("Incident not found: %s", incidentID),
		}), nil
	}

	incident := map[string]interface{}{
		"sys_id":            incidentData["sys_id"],
		"number":            incidentData["number"],
		"short_description": incidentData["short_description"],
		"description":       incidentData["description"],
		"state":             incidentData["state"],
		"priority":          incidentData["priority"],
		"impact":            incidentData["impact"],
		"urgency":           incidentData["urgency"],
		"category":          incidentData["category"],
		"subcategory":       incidentData["subcategory"],
		"created_on":        incidentData["sys_created_on"],
		"updated_on":        incidentData["sys_updated_on"],
	}

	if assignedTo, ok := incidentData["assigned_to"].(map[string]interface{}); ok {
		incident["assigned_to"] = assignedTo["display_value"]
	} else {
		incident["assigned_to"] = incidentData["assigned_to"]
	}

	return JSONResult(map[string]interface{}{
		"success":  true,
		"message":  fmt.Sprintf("Incident %s found", incidentData["number"]),
		"incident": incident,
	}), nil
}

func (r *Registry) createIncident(args map[string]interface{}) (*mcp.CallToolResult, error) {
	if r.readOnlyMode {
		return WriteBlockedResult(), nil
	}

	shortDesc := GetStringArg(args, "short_description", "")
	if shortDesc == "" {
		return JSONResult(NewErrorResponse("short_description is required", nil)), nil
	}

	data := map[string]interface{}{
		"short_description": shortDesc,
	}

	if v := GetStringArg(args, "description", ""); v != "" {
		data["description"] = v
	}
	if v := GetStringArg(args, "caller_id", ""); v != "" {
		data["caller_id"] = v
	}
	if v := GetStringArg(args, "category", ""); v != "" {
		data["category"] = v
	}
	if v := GetStringArg(args, "subcategory", ""); v != "" {
		data["subcategory"] = v
	}
	if v := GetStringArg(args, "priority", ""); v != "" {
		data["priority"] = v
	}
	if v := GetStringArg(args, "impact", ""); v != "" {
		data["impact"] = v
	}
	if v := GetStringArg(args, "urgency", ""); v != "" {
		data["urgency"] = v
	}
	if v := GetStringArg(args, "assigned_to", ""); v != "" {
		data["assigned_to"] = v
	}
	if v := GetStringArg(args, "assignment_group", ""); v != "" {
		data["assignment_group"] = v
	}

	result, err := r.client.Post("/table/incident", data)
	if err != nil {
		return JSONResult(NewErrorResponse("Failed to create incident", err)), nil
	}

	if resultData, ok := result["result"].(map[string]interface{}); ok {
		return JSONResult(map[string]interface{}{
			"success":         true,
			"message":         "Incident created successfully",
			"incident_id":     resultData["sys_id"],
			"incident_number": resultData["number"],
		}), nil
	}

	return JSONResult(NewErrorResponse("Unexpected response from ServiceNow", nil)), nil
}

func (r *Registry) updateIncident(args map[string]interface{}) (*mcp.CallToolResult, error) {
	if r.readOnlyMode {
		return WriteBlockedResult(), nil
	}

	incidentID := GetStringArg(args, "incident_id", "")
	if incidentID == "" {
		return JSONResult(NewErrorResponse("incident_id is required", nil)), nil
	}

	// Get sys_id if incident number was provided
	sysID := incidentID
	if !IsSysID(incidentID) {
		params := map[string]string{
			"sysparm_query": fmt.Sprintf("number=%s", incidentID),
			"sysparm_limit": "1",
		}
		result, err := r.client.Get("/table/incident", params)
		if err != nil {
			return JSONResult(NewErrorResponse("Failed to find incident", err)), nil
		}

		if resultList, ok := result["result"].([]interface{}); ok && len(resultList) > 0 {
			if incidentData, ok := resultList[0].(map[string]interface{}); ok {
				sysID, _ = incidentData["sys_id"].(string)
			}
		}

		if sysID == "" || sysID == incidentID {
			return JSONResult(map[string]interface{}{
				"success": false,
				"message": fmt.Sprintf("Incident not found: %s", incidentID),
			}), nil
		}
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
	if v := GetStringArg(args, "category", ""); v != "" {
		data["category"] = v
	}
	if v := GetStringArg(args, "priority", ""); v != "" {
		data["priority"] = v
	}
	if v := GetStringArg(args, "impact", ""); v != "" {
		data["impact"] = v
	}
	if v := GetStringArg(args, "urgency", ""); v != "" {
		data["urgency"] = v
	}
	if v := GetStringArg(args, "assigned_to", ""); v != "" {
		data["assigned_to"] = v
	}
	if v := GetStringArg(args, "assignment_group", ""); v != "" {
		data["assignment_group"] = v
	}
	if v := GetStringArg(args, "work_notes", ""); v != "" {
		data["work_notes"] = v
	}

	result, err := r.client.Put(fmt.Sprintf("/table/incident/%s", sysID), data)
	if err != nil {
		return JSONResult(NewErrorResponse("Failed to update incident", err)), nil
	}

	if resultData, ok := result["result"].(map[string]interface{}); ok {
		return JSONResult(map[string]interface{}{
			"success":         true,
			"message":         "Incident updated successfully",
			"incident_id":     resultData["sys_id"],
			"incident_number": resultData["number"],
		}), nil
	}

	return JSONResult(NewErrorResponse("Unexpected response from ServiceNow", nil)), nil
}

func (r *Registry) addIncidentComment(args map[string]interface{}) (*mcp.CallToolResult, error) {
	if r.readOnlyMode {
		return WriteBlockedResult(), nil
	}

	incidentID := GetStringArg(args, "incident_id", "")
	comment := GetStringArg(args, "comment", "")
	isWorkNote := GetBoolArg(args, "is_work_note", false)

	if incidentID == "" || comment == "" {
		return JSONResult(NewErrorResponse("incident_id and comment are required", nil)), nil
	}

	// Get sys_id if incident number was provided
	sysID := incidentID
	if !IsSysID(incidentID) {
		params := map[string]string{
			"sysparm_query": fmt.Sprintf("number=%s", incidentID),
			"sysparm_limit": "1",
		}
		result, err := r.client.Get("/table/incident", params)
		if err != nil {
			return JSONResult(NewErrorResponse("Failed to find incident", err)), nil
		}

		if resultList, ok := result["result"].([]interface{}); ok && len(resultList) > 0 {
			if incidentData, ok := resultList[0].(map[string]interface{}); ok {
				sysID, _ = incidentData["sys_id"].(string)
			}
		}

		if sysID == "" || sysID == incidentID {
			return JSONResult(map[string]interface{}{
				"success": false,
				"message": fmt.Sprintf("Incident not found: %s", incidentID),
			}), nil
		}
	}

	data := map[string]interface{}{}
	if isWorkNote {
		data["work_notes"] = comment
	} else {
		data["comments"] = comment
	}

	result, err := r.client.Put(fmt.Sprintf("/table/incident/%s", sysID), data)
	if err != nil {
		return JSONResult(NewErrorResponse("Failed to add comment", err)), nil
	}

	if resultData, ok := result["result"].(map[string]interface{}); ok {
		return JSONResult(map[string]interface{}{
			"success":         true,
			"message":         "Comment added successfully",
			"incident_id":     resultData["sys_id"],
			"incident_number": resultData["number"],
		}), nil
	}

	return JSONResult(NewErrorResponse("Unexpected response from ServiceNow", nil)), nil
}

func (r *Registry) resolveIncident(args map[string]interface{}) (*mcp.CallToolResult, error) {
	if r.readOnlyMode {
		return WriteBlockedResult(), nil
	}

	incidentID := GetStringArg(args, "incident_id", "")
	resolutionCode := GetStringArg(args, "resolution_code", "")
	resolutionNotes := GetStringArg(args, "resolution_notes", "")

	if incidentID == "" || resolutionCode == "" || resolutionNotes == "" {
		return JSONResult(NewErrorResponse("incident_id, resolution_code, and resolution_notes are required", nil)), nil
	}

	// Get sys_id if incident number was provided
	sysID := incidentID
	if !IsSysID(incidentID) {
		params := map[string]string{
			"sysparm_query": fmt.Sprintf("number=%s", incidentID),
			"sysparm_limit": "1",
		}
		result, err := r.client.Get("/table/incident", params)
		if err != nil {
			return JSONResult(NewErrorResponse("Failed to find incident", err)), nil
		}

		if resultList, ok := result["result"].([]interface{}); ok && len(resultList) > 0 {
			if incidentData, ok := resultList[0].(map[string]interface{}); ok {
				sysID, _ = incidentData["sys_id"].(string)
			}
		}

		if sysID == "" || sysID == incidentID {
			return JSONResult(map[string]interface{}{
				"success": false,
				"message": fmt.Sprintf("Incident not found: %s", incidentID),
			}), nil
		}
	}

	data := map[string]interface{}{
		"state":       "6", // Resolved
		"close_code":  resolutionCode,
		"close_notes": resolutionNotes,
		"resolved_at": "now",
	}

	result, err := r.client.Put(fmt.Sprintf("/table/incident/%s", sysID), data)
	if err != nil {
		return JSONResult(NewErrorResponse("Failed to resolve incident", err)), nil
	}

	if resultData, ok := result["result"].(map[string]interface{}); ok {
		return JSONResult(map[string]interface{}{
			"success":         true,
			"message":         "Incident resolved successfully",
			"incident_id":     resultData["sys_id"],
			"incident_number": resultData["number"],
		}), nil
	}

	return JSONResult(NewErrorResponse("Unexpected response from ServiceNow", nil)), nil
}
