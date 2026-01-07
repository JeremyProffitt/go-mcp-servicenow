package tools

import (
	"fmt"
	"strings"

	"github.com/elastiflow/go-mcp-servicenow/pkg/mcp"
)

// registerUserTools registers all user management tools
func (r *Registry) registerUserTools(server *mcp.Server) int {
	count := 0

	// List Users
	server.RegisterTool(mcp.Tool{
		Name:        "list_users",
		Description: "List users from ServiceNow with filtering options",
		InputSchema: mcp.JSONSchema{
			Type: "object",
			Properties: map[string]mcp.Property{
				"limit": {
					Type:        "number",
					Description: "Maximum number of users to return (default: 50)",
					Default:     50,
				},
				"offset": {
					Type:        "number",
					Description: "Offset for pagination",
					Default:     0,
				},
				"active": {
					Type:        "boolean",
					Description: "Filter by active status",
				},
				"department": {
					Type:        "string",
					Description: "Filter by department",
				},
				"query": {
					Type:        "string",
					Description: "Search query (name, email, etc.)",
				},
			},
		},
		Annotations: &mcp.ToolAnnotation{
			Title:        "List Users",
			ReadOnlyHint: true,
		},
	}, func(args map[string]interface{}) (*mcp.CallToolResult, error) {
		return r.listUsers(args)
	})
	count++

	// Get User
	server.RegisterTool(mcp.Tool{
		Name:        "get_user",
		Description: "Get a specific user by sys_id, username, or email",
		InputSchema: mcp.JSONSchema{
			Type: "object",
			Properties: map[string]mcp.Property{
				"user_id": {
					Type:        "string",
					Description: "User sys_id, username, or email",
				},
			},
			Required: []string{"user_id"},
		},
		Annotations: &mcp.ToolAnnotation{
			Title:        "Get User",
			ReadOnlyHint: true,
		},
	}, func(args map[string]interface{}) (*mcp.CallToolResult, error) {
		return r.getUser(args)
	})
	count++

	// List Groups
	server.RegisterTool(mcp.Tool{
		Name:        "list_groups",
		Description: "List groups from ServiceNow",
		InputSchema: mcp.JSONSchema{
			Type: "object",
			Properties: map[string]mcp.Property{
				"limit": {
					Type:        "number",
					Description: "Maximum number of groups to return (default: 50)",
					Default:     50,
				},
				"active": {
					Type:        "boolean",
					Description: "Filter by active status",
				},
				"query": {
					Type:        "string",
					Description: "Search query for group name",
				},
			},
		},
		Annotations: &mcp.ToolAnnotation{
			Title:        "List Groups",
			ReadOnlyHint: true,
		},
	}, func(args map[string]interface{}) (*mcp.CallToolResult, error) {
		return r.listGroups(args)
	})
	count++

	// Write operations
	if !r.readOnlyMode {
		// Create User
		server.RegisterTool(mcp.Tool{
			Name:        "create_user",
			Description: "Create a new user in ServiceNow",
			InputSchema: mcp.JSONSchema{
				Type: "object",
				Properties: map[string]mcp.Property{
					"user_name": {
						Type:        "string",
						Description: "Username",
					},
					"first_name": {
						Type:        "string",
						Description: "First name",
					},
					"last_name": {
						Type:        "string",
						Description: "Last name",
					},
					"email": {
						Type:        "string",
						Description: "Email address",
					},
					"title": {
						Type:        "string",
						Description: "Job title",
					},
					"department": {
						Type:        "string",
						Description: "Department",
					},
					"manager": {
						Type:        "string",
						Description: "Manager user sys_id",
					},
				},
				Required: []string{"user_name", "first_name", "last_name", "email"},
			},
		}, func(args map[string]interface{}) (*mcp.CallToolResult, error) {
			return r.createUser(args)
		})
		count++

		// Update User
		server.RegisterTool(mcp.Tool{
			Name:        "update_user",
			Description: "Update an existing user in ServiceNow",
			InputSchema: mcp.JSONSchema{
				Type: "object",
				Properties: map[string]mcp.Property{
					"user_id": {
						Type:        "string",
						Description: "User sys_id",
					},
					"first_name": {
						Type:        "string",
						Description: "First name",
					},
					"last_name": {
						Type:        "string",
						Description: "Last name",
					},
					"email": {
						Type:        "string",
						Description: "Email address",
					},
					"title": {
						Type:        "string",
						Description: "Job title",
					},
					"department": {
						Type:        "string",
						Description: "Department",
					},
					"manager": {
						Type:        "string",
						Description: "Manager user sys_id",
					},
					"active": {
						Type:        "boolean",
						Description: "Active status",
					},
				},
				Required: []string{"user_id"},
			},
		}, func(args map[string]interface{}) (*mcp.CallToolResult, error) {
			return r.updateUser(args)
		})
		count++

		// Create Group
		server.RegisterTool(mcp.Tool{
			Name:        "create_group",
			Description: "Create a new group in ServiceNow",
			InputSchema: mcp.JSONSchema{
				Type: "object",
				Properties: map[string]mcp.Property{
					"name": {
						Type:        "string",
						Description: "Group name",
					},
					"description": {
						Type:        "string",
						Description: "Group description",
					},
					"manager": {
						Type:        "string",
						Description: "Manager user sys_id",
					},
					"email": {
						Type:        "string",
						Description: "Group email",
					},
				},
				Required: []string{"name"},
			},
		}, func(args map[string]interface{}) (*mcp.CallToolResult, error) {
			return r.createGroup(args)
		})
		count++

		// Update Group
		server.RegisterTool(mcp.Tool{
			Name:        "update_group",
			Description: "Update an existing group in ServiceNow",
			InputSchema: mcp.JSONSchema{
				Type: "object",
				Properties: map[string]mcp.Property{
					"group_id": {
						Type:        "string",
						Description: "Group sys_id",
					},
					"name": {
						Type:        "string",
						Description: "Group name",
					},
					"description": {
						Type:        "string",
						Description: "Group description",
					},
					"manager": {
						Type:        "string",
						Description: "Manager user sys_id",
					},
					"active": {
						Type:        "boolean",
						Description: "Active status",
					},
				},
				Required: []string{"group_id"},
			},
		}, func(args map[string]interface{}) (*mcp.CallToolResult, error) {
			return r.updateGroup(args)
		})
		count++

		// Add Group Members
		server.RegisterTool(mcp.Tool{
			Name:        "add_group_members",
			Description: "Add members to a group in ServiceNow",
			InputSchema: mcp.JSONSchema{
				Type: "object",
				Properties: map[string]mcp.Property{
					"group_id": {
						Type:        "string",
						Description: "Group sys_id",
					},
					"user_ids": {
						Type:        "array",
						Description: "List of user sys_ids to add",
						Items:       &mcp.Property{Type: "string"},
					},
				},
				Required: []string{"group_id", "user_ids"},
			},
		}, func(args map[string]interface{}) (*mcp.CallToolResult, error) {
			return r.addGroupMembers(args)
		})
		count++

		// Remove Group Members
		server.RegisterTool(mcp.Tool{
			Name:        "remove_group_members",
			Description: "Remove members from a group in ServiceNow",
			InputSchema: mcp.JSONSchema{
				Type: "object",
				Properties: map[string]mcp.Property{
					"group_id": {
						Type:        "string",
						Description: "Group sys_id",
					},
					"user_ids": {
						Type:        "array",
						Description: "List of user sys_ids to remove",
						Items:       &mcp.Property{Type: "string"},
					},
				},
				Required: []string{"group_id", "user_ids"},
			},
		}, func(args map[string]interface{}) (*mcp.CallToolResult, error) {
			return r.removeGroupMembers(args)
		})
		count++
	}

	return count
}

func (r *Registry) listUsers(args map[string]interface{}) (*mcp.CallToolResult, error) {
	limit := GetIntArg(args, "limit", 50)
	offset := GetIntArg(args, "offset", 0)
	department := GetStringArg(args, "department", "")
	query := GetStringArg(args, "query", "")

	params := map[string]string{
		"sysparm_limit":                  fmt.Sprintf("%d", limit),
		"sysparm_offset":                 fmt.Sprintf("%d", offset),
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
	if department != "" {
		filters = append(filters, fmt.Sprintf("department=%s", department))
	}
	if query != "" {
		filters = append(filters, fmt.Sprintf("nameLIKE%s^ORemailLIKE%s^ORuser_nameLIKE%s", query, query, query))
	}

	if len(filters) > 0 {
		params["sysparm_query"] = strings.Join(filters, "^")
	}

	result, err := r.client.Get("/table/sys_user", params)
	if err != nil {
		return JSONResult(NewErrorResponse("Failed to list users", err)), nil
	}

	users := []map[string]interface{}{}
	if resultList, ok := result["result"].([]interface{}); ok {
		for _, item := range resultList {
			if data, ok := item.(map[string]interface{}); ok {
				users = append(users, map[string]interface{}{
					"sys_id":     data["sys_id"],
					"user_name":  data["user_name"],
					"first_name": data["first_name"],
					"last_name":  data["last_name"],
					"email":      data["email"],
					"title":      data["title"],
					"department": data["department"],
					"active":     data["active"],
				})
			}
		}
	}

	return JSONResult(map[string]interface{}{
		"success": true,
		"message": fmt.Sprintf("Found %d users", len(users)),
		"users":   users,
	}), nil
}

func (r *Registry) getUser(args map[string]interface{}) (*mcp.CallToolResult, error) {
	userID := GetStringArg(args, "user_id", "")
	if userID == "" {
		return JSONResult(NewErrorResponse("user_id is required", nil)), nil
	}

	var params map[string]string
	var endpoint string

	if IsSysID(userID) {
		endpoint = fmt.Sprintf("/table/sys_user/%s", userID)
		params = map[string]string{
			"sysparm_display_value":          "true",
			"sysparm_exclude_reference_link": "true",
		}
	} else {
		endpoint = "/table/sys_user"
		params = map[string]string{
			"sysparm_query":                  fmt.Sprintf("user_name=%s^ORemail=%s", userID, userID),
			"sysparm_limit":                  "1",
			"sysparm_display_value":          "true",
			"sysparm_exclude_reference_link": "true",
		}
	}

	result, err := r.client.Get(endpoint, params)
	if err != nil {
		return JSONResult(NewErrorResponse("Failed to get user", err)), nil
	}

	var userData map[string]interface{}
	if IsSysID(userID) {
		userData, _ = result["result"].(map[string]interface{})
	} else {
		if resultList, ok := result["result"].([]interface{}); ok && len(resultList) > 0 {
			userData, _ = resultList[0].(map[string]interface{})
		}
	}

	if userData == nil {
		return JSONResult(map[string]interface{}{
			"success": false,
			"message": fmt.Sprintf("User not found: %s", userID),
		}), nil
	}

	return JSONResult(map[string]interface{}{
		"success": true,
		"message": "User found",
		"user":    userData,
	}), nil
}

func (r *Registry) listGroups(args map[string]interface{}) (*mcp.CallToolResult, error) {
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
		filters = append(filters, fmt.Sprintf("nameLIKE%s", query))
	}

	if len(filters) > 0 {
		params["sysparm_query"] = strings.Join(filters, "^")
	}

	result, err := r.client.Get("/table/sys_user_group", params)
	if err != nil {
		return JSONResult(NewErrorResponse("Failed to list groups", err)), nil
	}

	groups := []map[string]interface{}{}
	if resultList, ok := result["result"].([]interface{}); ok {
		for _, item := range resultList {
			if data, ok := item.(map[string]interface{}); ok {
				groups = append(groups, map[string]interface{}{
					"sys_id":      data["sys_id"],
					"name":        data["name"],
					"description": data["description"],
					"manager":     data["manager"],
					"email":       data["email"],
					"active":      data["active"],
				})
			}
		}
	}

	return JSONResult(map[string]interface{}{
		"success": true,
		"message": fmt.Sprintf("Found %d groups", len(groups)),
		"groups":  groups,
	}), nil
}

func (r *Registry) createUser(args map[string]interface{}) (*mcp.CallToolResult, error) {
	if r.readOnlyMode {
		return WriteBlockedResult(), nil
	}

	userName := GetStringArg(args, "user_name", "")
	firstName := GetStringArg(args, "first_name", "")
	lastName := GetStringArg(args, "last_name", "")
	email := GetStringArg(args, "email", "")

	if userName == "" || firstName == "" || lastName == "" || email == "" {
		return JSONResult(NewErrorResponse("user_name, first_name, last_name, and email are required", nil)), nil
	}

	data := map[string]interface{}{
		"user_name":  userName,
		"first_name": firstName,
		"last_name":  lastName,
		"email":      email,
	}

	if v := GetStringArg(args, "title", ""); v != "" {
		data["title"] = v
	}
	if v := GetStringArg(args, "department", ""); v != "" {
		data["department"] = v
	}
	if v := GetStringArg(args, "manager", ""); v != "" {
		data["manager"] = v
	}

	result, err := r.client.Post("/table/sys_user", data)
	if err != nil {
		return JSONResult(NewErrorResponse("Failed to create user", err)), nil
	}

	if resultData, ok := result["result"].(map[string]interface{}); ok {
		return JSONResult(map[string]interface{}{
			"success": true,
			"message": "User created successfully",
			"user_id": resultData["sys_id"],
		}), nil
	}

	return JSONResult(NewErrorResponse("Unexpected response from ServiceNow", nil)), nil
}

func (r *Registry) updateUser(args map[string]interface{}) (*mcp.CallToolResult, error) {
	if r.readOnlyMode {
		return WriteBlockedResult(), nil
	}

	userID := GetStringArg(args, "user_id", "")
	if userID == "" {
		return JSONResult(NewErrorResponse("user_id is required", nil)), nil
	}

	data := map[string]interface{}{}

	if v := GetStringArg(args, "first_name", ""); v != "" {
		data["first_name"] = v
	}
	if v := GetStringArg(args, "last_name", ""); v != "" {
		data["last_name"] = v
	}
	if v := GetStringArg(args, "email", ""); v != "" {
		data["email"] = v
	}
	if v := GetStringArg(args, "title", ""); v != "" {
		data["title"] = v
	}
	if v := GetStringArg(args, "department", ""); v != "" {
		data["department"] = v
	}
	if v := GetStringArg(args, "manager", ""); v != "" {
		data["manager"] = v
	}
	if v, exists := args["active"]; exists {
		data["active"] = v
	}

	result, err := r.client.Put(fmt.Sprintf("/table/sys_user/%s", userID), data)
	if err != nil {
		return JSONResult(NewErrorResponse("Failed to update user", err)), nil
	}

	if resultData, ok := result["result"].(map[string]interface{}); ok {
		return JSONResult(map[string]interface{}{
			"success": true,
			"message": "User updated successfully",
			"user_id": resultData["sys_id"],
		}), nil
	}

	return JSONResult(NewErrorResponse("Unexpected response from ServiceNow", nil)), nil
}

func (r *Registry) createGroup(args map[string]interface{}) (*mcp.CallToolResult, error) {
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
	if v := GetStringArg(args, "manager", ""); v != "" {
		data["manager"] = v
	}
	if v := GetStringArg(args, "email", ""); v != "" {
		data["email"] = v
	}

	result, err := r.client.Post("/table/sys_user_group", data)
	if err != nil {
		return JSONResult(NewErrorResponse("Failed to create group", err)), nil
	}

	if resultData, ok := result["result"].(map[string]interface{}); ok {
		return JSONResult(map[string]interface{}{
			"success":  true,
			"message":  "Group created successfully",
			"group_id": resultData["sys_id"],
		}), nil
	}

	return JSONResult(NewErrorResponse("Unexpected response from ServiceNow", nil)), nil
}

func (r *Registry) updateGroup(args map[string]interface{}) (*mcp.CallToolResult, error) {
	if r.readOnlyMode {
		return WriteBlockedResult(), nil
	}

	groupID := GetStringArg(args, "group_id", "")
	if groupID == "" {
		return JSONResult(NewErrorResponse("group_id is required", nil)), nil
	}

	data := map[string]interface{}{}

	if v := GetStringArg(args, "name", ""); v != "" {
		data["name"] = v
	}
	if v := GetStringArg(args, "description", ""); v != "" {
		data["description"] = v
	}
	if v := GetStringArg(args, "manager", ""); v != "" {
		data["manager"] = v
	}
	if v, exists := args["active"]; exists {
		data["active"] = v
	}

	result, err := r.client.Put(fmt.Sprintf("/table/sys_user_group/%s", groupID), data)
	if err != nil {
		return JSONResult(NewErrorResponse("Failed to update group", err)), nil
	}

	if resultData, ok := result["result"].(map[string]interface{}); ok {
		return JSONResult(map[string]interface{}{
			"success":  true,
			"message":  "Group updated successfully",
			"group_id": resultData["sys_id"],
		}), nil
	}

	return JSONResult(NewErrorResponse("Unexpected response from ServiceNow", nil)), nil
}

func (r *Registry) addGroupMembers(args map[string]interface{}) (*mcp.CallToolResult, error) {
	if r.readOnlyMode {
		return WriteBlockedResult(), nil
	}

	groupID := GetStringArg(args, "group_id", "")
	userIDs := GetStringArrayArg(args, "user_ids")

	if groupID == "" || len(userIDs) == 0 {
		return JSONResult(NewErrorResponse("group_id and user_ids are required", nil)), nil
	}

	addedCount := 0
	var lastErr error

	for _, userID := range userIDs {
		data := map[string]interface{}{
			"group": groupID,
			"user":  userID,
		}

		_, err := r.client.Post("/table/sys_user_grmember", data)
		if err != nil {
			lastErr = err
		} else {
			addedCount++
		}
	}

	if addedCount == len(userIDs) {
		return JSONResult(map[string]interface{}{
			"success": true,
			"message": fmt.Sprintf("Successfully added %d members to group", addedCount),
		}), nil
	}

	return JSONResult(map[string]interface{}{
		"success": addedCount > 0,
		"message": fmt.Sprintf("Added %d of %d members. Last error: %v", addedCount, len(userIDs), lastErr),
	}), nil
}

func (r *Registry) removeGroupMembers(args map[string]interface{}) (*mcp.CallToolResult, error) {
	if r.readOnlyMode {
		return WriteBlockedResult(), nil
	}

	groupID := GetStringArg(args, "group_id", "")
	userIDs := GetStringArrayArg(args, "user_ids")

	if groupID == "" || len(userIDs) == 0 {
		return JSONResult(NewErrorResponse("group_id and user_ids are required", nil)), nil
	}

	removedCount := 0
	var lastErr error

	for _, userID := range userIDs {
		// Find the membership record
		params := map[string]string{
			"sysparm_query": fmt.Sprintf("group=%s^user=%s", groupID, userID),
			"sysparm_limit": "1",
		}

		result, err := r.client.Get("/table/sys_user_grmember", params)
		if err != nil {
			lastErr = err
			continue
		}

		if resultList, ok := result["result"].([]interface{}); ok && len(resultList) > 0 {
			if memberData, ok := resultList[0].(map[string]interface{}); ok {
				if memberID, ok := memberData["sys_id"].(string); ok {
					_, err := r.client.Delete(fmt.Sprintf("/table/sys_user_grmember/%s", memberID))
					if err != nil {
						lastErr = err
					} else {
						removedCount++
					}
				}
			}
		}
	}

	if removedCount == len(userIDs) {
		return JSONResult(map[string]interface{}{
			"success": true,
			"message": fmt.Sprintf("Successfully removed %d members from group", removedCount),
		}), nil
	}

	return JSONResult(map[string]interface{}{
		"success": removedCount > 0,
		"message": fmt.Sprintf("Removed %d of %d members. Last error: %v", removedCount, len(userIDs), lastErr),
	}), nil
}
