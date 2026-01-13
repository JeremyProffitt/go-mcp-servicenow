package tools

import (
	"fmt"
	"strings"

	"github.com/elastiflow/go-mcp-servicenow/pkg/mcp"
)

// registerAgileTools registers all agile management tools (stories, epics, scrum tasks, projects)
func (r *Registry) registerAgileTools(server *mcp.Server) int {
	count := 0

	// Helper for limit constraints
	limitMin := float64(1)
	limitMax := float64(1000)

	// === Stories ===
	server.RegisterTool(mcp.Tool{
		Name:        "list_stories",
		Description: "List user stories with optional filtering by state, sprint, or assignee. Stories represent work items in Agile development.",
		InputSchema: mcp.JSONSchema{
			Type: "object",
			Properties: map[string]mcp.Property{
				"limit": {
					Type:        "number",
					Description: "Maximum number of stories to return (default: 50)",
					Default:     50,
					Minimum:     &limitMin,
					Maximum:     &limitMax,
				},
				"state": {
					Type:        "string",
					Description: "Filter by state (e.g., 'Draft', 'Ready', 'In Progress', 'Complete')",
				},
				"sprint": {
					Type:        "string",
					Description: "Filter by sprint sys_id (e.g., 'a1b2c3d4e5f6g7h8i9j0k1l2m3n4o5p6')",
				},
				"assigned_to": {
					Type:        "string",
					Description: "Filter by assigned user (sys_id, username, or email)",
				},
			},
		},
		Annotations: &mcp.ToolAnnotation{
			Title:        "List Stories",
			ReadOnlyHint: true,
		},
	}, func(args map[string]interface{}) (*mcp.CallToolResult, error) {
		return r.listStories(args)
	})
	count++

	// === Epics ===
	server.RegisterTool(mcp.Tool{
		Name:        "list_epics",
		Description: "List epics with optional filtering. Epics are large bodies of work that contain multiple stories.",
		InputSchema: mcp.JSONSchema{
			Type: "object",
			Properties: map[string]mcp.Property{
				"limit": {
					Type:        "number",
					Description: "Maximum number of epics to return (default: 50)",
					Default:     50,
					Minimum:     &limitMin,
					Maximum:     &limitMax,
				},
				"state": {
					Type:        "string",
					Description: "Filter by state (e.g., 'Draft', 'Analysis', 'Development', 'Complete')",
				},
				"product": {
					Type:        "string",
					Description: "Filter by product sys_id (e.g., 'a1b2c3d4e5f6g7h8i9j0k1l2m3n4o5p6')",
				},
			},
		},
		Annotations: &mcp.ToolAnnotation{
			Title:        "List Epics",
			ReadOnlyHint: true,
		},
	}, func(args map[string]interface{}) (*mcp.CallToolResult, error) {
		return r.listEpics(args)
	})
	count++

	// === Scrum Tasks ===
	server.RegisterTool(mcp.Tool{
		Name:        "list_scrum_tasks",
		Description: "List scrum tasks with optional filtering. Tasks are work items that implement a story.",
		InputSchema: mcp.JSONSchema{
			Type: "object",
			Properties: map[string]mcp.Property{
				"limit": {
					Type:        "number",
					Description: "Maximum number of tasks to return (default: 50)",
					Default:     50,
					Minimum:     &limitMin,
					Maximum:     &limitMax,
				},
				"story": {
					Type:        "string",
					Description: "Filter by parent story sys_id (e.g., 'a1b2c3d4e5f6g7h8i9j0k1l2m3n4o5p6')",
				},
				"state": {
					Type:        "string",
					Description: "Filter by state (e.g., 'Draft', 'Ready', 'Work in progress', 'Complete')",
				},
				"assigned_to": {
					Type:        "string",
					Description: "Filter by assigned user (sys_id, username, or email)",
				},
			},
		},
		Annotations: &mcp.ToolAnnotation{
			Title:        "List Scrum Tasks",
			ReadOnlyHint: true,
		},
	}, func(args map[string]interface{}) (*mcp.CallToolResult, error) {
		return r.listScrumTasks(args)
	})
	count++

	// === Projects ===
	server.RegisterTool(mcp.Tool{
		Name:        "list_projects",
		Description: "List projects with optional filtering by state or active status.",
		InputSchema: mcp.JSONSchema{
			Type: "object",
			Properties: map[string]mcp.Property{
				"limit": {
					Type:        "number",
					Description: "Maximum number of projects to return (default: 50)",
					Default:     50,
					Minimum:     &limitMin,
					Maximum:     &limitMax,
				},
				"state": {
					Type:        "string",
					Description: "Filter by state (e.g., 'Draft', 'Pending', 'Open', 'Work in progress', 'Closed')",
				},
				"active": {
					Type:        "boolean",
					Description: "Filter by active status (true = only active, false = only inactive)",
				},
			},
		},
		Annotations: &mcp.ToolAnnotation{
			Title:        "List Projects",
			ReadOnlyHint: true,
		},
	}, func(args map[string]interface{}) (*mcp.CallToolResult, error) {
		return r.listProjects(args)
	})
	count++

	// Write operations
	if !r.readOnlyMode {
		// Create Story
		server.RegisterTool(mcp.Tool{
			Name:        "create_story",
			Description: "Create a new user story. Stories represent work items in Agile development.",
			InputSchema: mcp.JSONSchema{
				Type: "object",
				Properties: map[string]mcp.Property{
					"short_description": {
						Type:        "string",
						Description: "Story title/summary",
					},
					"description": {
						Type:        "string",
						Description: "Story description including acceptance criteria",
					},
					"product": {
						Type:        "string",
						Description: "Product sys_id (e.g., 'a1b2c3d4e5f6g7h8i9j0k1l2m3n4o5p6')",
					},
					"epic": {
						Type:        "string",
						Description: "Parent epic sys_id (e.g., 'a1b2c3d4e5f6g7h8i9j0k1l2m3n4o5p6')",
					},
					"sprint": {
						Type:        "string",
						Description: "Sprint sys_id (e.g., 'a1b2c3d4e5f6g7h8i9j0k1l2m3n4o5p6')",
					},
					"story_points": {
						Type:        "number",
						Description: "Story points (effort estimate, typically Fibonacci sequence: 1, 2, 3, 5, 8, 13)",
					},
					"assigned_to": {
						Type:        "string",
						Description: "Assigned user (sys_id, username, or email)",
					},
				},
				Required: []string{"short_description"},
			},
			Annotations: &mcp.ToolAnnotation{
				Title: "Create Story",
			},
		}, func(args map[string]interface{}) (*mcp.CallToolResult, error) {
			return r.createStory(args)
		})
		count++

		// Update Story
		server.RegisterTool(mcp.Tool{
			Name:        "update_story",
			Description: "Update an existing user story. At least one field besides story_id must be provided.",
			InputSchema: mcp.JSONSchema{
				Type: "object",
				Properties: map[string]mcp.Property{
					"story_id": {
						Type:        "string",
						Description: "Story sys_id (e.g., 'a1b2c3d4e5f6g7h8i9j0k1l2m3n4o5p6')",
					},
					"short_description": {
						Type:        "string",
						Description: "Story title/summary",
					},
					"state": {
						Type:        "string",
						Description: "Story state (e.g., 'Draft', 'Ready', 'In Progress', 'Complete')",
					},
					"story_points": {
						Type:        "number",
						Description: "Story points (effort estimate)",
					},
					"blocked": {
						Type:        "boolean",
						Description: "Whether the story is blocked",
					},
				},
				Required: []string{"story_id"},
			},
			Annotations: &mcp.ToolAnnotation{
				Title: "Update Story",
			},
		}, func(args map[string]interface{}) (*mcp.CallToolResult, error) {
			return r.updateStory(args)
		})
		count++

		// Create Epic
		server.RegisterTool(mcp.Tool{
			Name:        "create_epic",
			Description: "Create a new epic. Epics are large bodies of work that contain multiple stories.",
			InputSchema: mcp.JSONSchema{
				Type: "object",
				Properties: map[string]mcp.Property{
					"short_description": {
						Type:        "string",
						Description: "Epic title/summary",
					},
					"description": {
						Type:        "string",
						Description: "Epic description",
					},
					"product": {
						Type:        "string",
						Description: "Product sys_id (e.g., 'a1b2c3d4e5f6g7h8i9j0k1l2m3n4o5p6')",
					},
				},
				Required: []string{"short_description"},
			},
			Annotations: &mcp.ToolAnnotation{
				Title: "Create Epic",
			},
		}, func(args map[string]interface{}) (*mcp.CallToolResult, error) {
			return r.createEpic(args)
		})
		count++

		// Update Epic
		server.RegisterTool(mcp.Tool{
			Name:        "update_epic",
			Description: "Update an existing epic. At least one field besides epic_id must be provided.",
			InputSchema: mcp.JSONSchema{
				Type: "object",
				Properties: map[string]mcp.Property{
					"epic_id": {
						Type:        "string",
						Description: "Epic sys_id (e.g., 'a1b2c3d4e5f6g7h8i9j0k1l2m3n4o5p6')",
					},
					"short_description": {
						Type:        "string",
						Description: "Epic title/summary",
					},
					"state": {
						Type:        "string",
						Description: "Epic state (e.g., 'Draft', 'Analysis', 'Development', 'Complete')",
					},
				},
				Required: []string{"epic_id"},
			},
			Annotations: &mcp.ToolAnnotation{
				Title: "Update Epic",
			},
		}, func(args map[string]interface{}) (*mcp.CallToolResult, error) {
			return r.updateEpic(args)
		})
		count++

		// Create Scrum Task
		server.RegisterTool(mcp.Tool{
			Name:        "create_scrum_task",
			Description: "Create a new scrum task. Tasks are work items that implement a story.",
			InputSchema: mcp.JSONSchema{
				Type: "object",
				Properties: map[string]mcp.Property{
					"short_description": {
						Type:        "string",
						Description: "Task title/summary",
					},
					"story": {
						Type:        "string",
						Description: "Parent story sys_id (e.g., 'a1b2c3d4e5f6g7h8i9j0k1l2m3n4o5p6')",
					},
					"type": {
						Type:        "string",
						Description: "Task type (e.g., 'Development', 'Testing', 'Documentation')",
					},
					"assigned_to": {
						Type:        "string",
						Description: "Assigned user (sys_id, username, or email)",
					},
					"time_remaining": {
						Type:        "number",
						Description: "Remaining hours of work",
					},
				},
				Required: []string{"short_description"},
			},
			Annotations: &mcp.ToolAnnotation{
				Title: "Create Scrum Task",
			},
		}, func(args map[string]interface{}) (*mcp.CallToolResult, error) {
			return r.createScrumTask(args)
		})
		count++

		// Update Scrum Task
		server.RegisterTool(mcp.Tool{
			Name:        "update_scrum_task",
			Description: "Update an existing scrum task. At least one field besides task_id must be provided.",
			InputSchema: mcp.JSONSchema{
				Type: "object",
				Properties: map[string]mcp.Property{
					"task_id": {
						Type:        "string",
						Description: "Task sys_id (e.g., 'a1b2c3d4e5f6g7h8i9j0k1l2m3n4o5p6')",
					},
					"state": {
						Type:        "string",
						Description: "Task state (e.g., 'Draft', 'Ready', 'Work in progress', 'Complete')",
					},
					"time_remaining": {
						Type:        "number",
						Description: "Remaining hours of work",
					},
				},
				Required: []string{"task_id"},
			},
			Annotations: &mcp.ToolAnnotation{
				Title: "Update Scrum Task",
			},
		}, func(args map[string]interface{}) (*mcp.CallToolResult, error) {
			return r.updateScrumTask(args)
		})
		count++

		// Create Project
		server.RegisterTool(mcp.Tool{
			Name:        "create_project",
			Description: "Create a new project. Projects are used for tracking larger initiatives.",
			InputSchema: mcp.JSONSchema{
				Type: "object",
				Properties: map[string]mcp.Property{
					"short_description": {
						Type:        "string",
						Description: "Project title/summary",
					},
					"description": {
						Type:        "string",
						Description: "Project description",
					},
					"start_date": {
						Type:        "string",
						Description: "Project start date (format: YYYY-MM-DD)",
					},
					"end_date": {
						Type:        "string",
						Description: "Project end date (format: YYYY-MM-DD)",
					},
				},
				Required: []string{"short_description"},
			},
			Annotations: &mcp.ToolAnnotation{
				Title: "Create Project",
			},
		}, func(args map[string]interface{}) (*mcp.CallToolResult, error) {
			return r.createProject(args)
		})
		count++

		// Update Project
		server.RegisterTool(mcp.Tool{
			Name:        "update_project",
			Description: "Update an existing project. At least one field besides project_id must be provided.",
			InputSchema: mcp.JSONSchema{
				Type: "object",
				Properties: map[string]mcp.Property{
					"project_id": {
						Type:        "string",
						Description: "Project sys_id (e.g., 'a1b2c3d4e5f6g7h8i9j0k1l2m3n4o5p6')",
					},
					"short_description": {
						Type:        "string",
						Description: "Project title/summary",
					},
					"state": {
						Type:        "string",
						Description: "Project state (e.g., 'Draft', 'Pending', 'Open', 'Work in progress', 'Closed')",
					},
				},
				Required: []string{"project_id"},
			},
			Annotations: &mcp.ToolAnnotation{
				Title: "Update Project",
			},
		}, func(args map[string]interface{}) (*mcp.CallToolResult, error) {
			return r.updateProject(args)
		})
		count++
	}

	return count
}

func (r *Registry) listStories(args map[string]interface{}) (*mcp.CallToolResult, error) {
	limit := GetIntArg(args, "limit", 50)
	state := GetStringArg(args, "state", "")
	sprint := GetStringArg(args, "sprint", "")
	assignedTo := GetStringArg(args, "assigned_to", "")

	params := map[string]string{
		"sysparm_limit":                  fmt.Sprintf("%d", limit),
		"sysparm_display_value":          "true",
		"sysparm_exclude_reference_link": "true",
	}

	var filters []string
	if state != "" {
		filters = append(filters, fmt.Sprintf("state=%s", state))
	}
	if sprint != "" {
		filters = append(filters, fmt.Sprintf("sprint=%s", sprint))
	}
	if assignedTo != "" {
		filters = append(filters, fmt.Sprintf("assigned_to=%s", assignedTo))
	}

	if len(filters) > 0 {
		params["sysparm_query"] = strings.Join(filters, "^")
	}

	result, err := r.client.Get("/table/rm_story", params)
	if err != nil {
		return JSONResult(NewErrorResponse("Failed to list stories", err)), nil
	}

	stories := []map[string]interface{}{}
	if resultList, ok := result["result"].([]interface{}); ok {
		for _, item := range resultList {
			if data, ok := item.(map[string]interface{}); ok {
				stories = append(stories, map[string]interface{}{
					"sys_id":            data["sys_id"],
					"number":            data["number"],
					"short_description": data["short_description"],
					"state":             data["state"],
					"story_points":      data["story_points"],
					"sprint":            data["sprint"],
					"epic":              data["epic"],
					"blocked":           data["blocked"],
				})
			}
		}
	}

	return JSONResult(map[string]interface{}{
		"success": true,
		"message": fmt.Sprintf("Found %d stories", len(stories)),
		"stories": stories,
	}), nil
}

func (r *Registry) listEpics(args map[string]interface{}) (*mcp.CallToolResult, error) {
	limit := GetIntArg(args, "limit", 50)
	state := GetStringArg(args, "state", "")
	product := GetStringArg(args, "product", "")

	params := map[string]string{
		"sysparm_limit":                  fmt.Sprintf("%d", limit),
		"sysparm_display_value":          "true",
		"sysparm_exclude_reference_link": "true",
	}

	var filters []string
	if state != "" {
		filters = append(filters, fmt.Sprintf("state=%s", state))
	}
	if product != "" {
		filters = append(filters, fmt.Sprintf("product=%s", product))
	}

	if len(filters) > 0 {
		params["sysparm_query"] = strings.Join(filters, "^")
	}

	result, err := r.client.Get("/table/rm_epic", params)
	if err != nil {
		return JSONResult(NewErrorResponse("Failed to list epics", err)), nil
	}

	epics := []map[string]interface{}{}
	if resultList, ok := result["result"].([]interface{}); ok {
		for _, item := range resultList {
			if data, ok := item.(map[string]interface{}); ok {
				epics = append(epics, map[string]interface{}{
					"sys_id":            data["sys_id"],
					"number":            data["number"],
					"short_description": data["short_description"],
					"state":             data["state"],
					"product":           data["product"],
				})
			}
		}
	}

	return JSONResult(map[string]interface{}{
		"success": true,
		"message": fmt.Sprintf("Found %d epics", len(epics)),
		"epics":   epics,
	}), nil
}

func (r *Registry) listScrumTasks(args map[string]interface{}) (*mcp.CallToolResult, error) {
	limit := GetIntArg(args, "limit", 50)
	story := GetStringArg(args, "story", "")
	state := GetStringArg(args, "state", "")
	assignedTo := GetStringArg(args, "assigned_to", "")

	params := map[string]string{
		"sysparm_limit":                  fmt.Sprintf("%d", limit),
		"sysparm_display_value":          "true",
		"sysparm_exclude_reference_link": "true",
	}

	var filters []string
	if story != "" {
		filters = append(filters, fmt.Sprintf("story=%s", story))
	}
	if state != "" {
		filters = append(filters, fmt.Sprintf("state=%s", state))
	}
	if assignedTo != "" {
		filters = append(filters, fmt.Sprintf("assigned_to=%s", assignedTo))
	}

	if len(filters) > 0 {
		params["sysparm_query"] = strings.Join(filters, "^")
	}

	result, err := r.client.Get("/table/rm_scrum_task", params)
	if err != nil {
		return JSONResult(NewErrorResponse("Failed to list scrum tasks", err)), nil
	}

	tasks := []map[string]interface{}{}
	if resultList, ok := result["result"].([]interface{}); ok {
		for _, item := range resultList {
			if data, ok := item.(map[string]interface{}); ok {
				tasks = append(tasks, map[string]interface{}{
					"sys_id":            data["sys_id"],
					"number":            data["number"],
					"short_description": data["short_description"],
					"state":             data["state"],
					"story":             data["story"],
					"type":              data["type"],
					"time_remaining":    data["time_remaining"],
				})
			}
		}
	}

	return JSONResult(map[string]interface{}{
		"success":     true,
		"message":     fmt.Sprintf("Found %d scrum tasks", len(tasks)),
		"scrum_tasks": tasks,
	}), nil
}

func (r *Registry) listProjects(args map[string]interface{}) (*mcp.CallToolResult, error) {
	limit := GetIntArg(args, "limit", 50)
	state := GetStringArg(args, "state", "")

	params := map[string]string{
		"sysparm_limit":                  fmt.Sprintf("%d", limit),
		"sysparm_display_value":          "true",
		"sysparm_exclude_reference_link": "true",
	}

	var filters []string
	if state != "" {
		filters = append(filters, fmt.Sprintf("state=%s", state))
	}
	if active, exists := args["active"]; exists {
		if active.(bool) {
			filters = append(filters, "active=true")
		} else {
			filters = append(filters, "active=false")
		}
	}

	if len(filters) > 0 {
		params["sysparm_query"] = strings.Join(filters, "^")
	}

	result, err := r.client.Get("/table/pm_project", params)
	if err != nil {
		return JSONResult(NewErrorResponse("Failed to list projects", err)), nil
	}

	projects := []map[string]interface{}{}
	if resultList, ok := result["result"].([]interface{}); ok {
		for _, item := range resultList {
			if data, ok := item.(map[string]interface{}); ok {
				projects = append(projects, map[string]interface{}{
					"sys_id":            data["sys_id"],
					"number":            data["number"],
					"short_description": data["short_description"],
					"state":             data["state"],
					"start_date":        data["start_date"],
					"end_date":          data["end_date"],
					"active":            data["active"],
				})
			}
		}
	}

	return JSONResult(map[string]interface{}{
		"success":  true,
		"message":  fmt.Sprintf("Found %d projects", len(projects)),
		"projects": projects,
	}), nil
}

func (r *Registry) createStory(args map[string]interface{}) (*mcp.CallToolResult, error) {
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
	if v := GetStringArg(args, "product", ""); v != "" {
		data["product"] = v
	}
	if v := GetStringArg(args, "epic", ""); v != "" {
		data["epic"] = v
	}
	if v := GetStringArg(args, "sprint", ""); v != "" {
		data["sprint"] = v
	}
	if v := GetIntArg(args, "story_points", 0); v > 0 {
		data["story_points"] = v
	}
	if v := GetStringArg(args, "assigned_to", ""); v != "" {
		data["assigned_to"] = v
	}

	result, err := r.client.Post("/table/rm_story", data)
	if err != nil {
		return JSONResult(NewErrorResponse("Failed to create story", err)), nil
	}

	if resultData, ok := result["result"].(map[string]interface{}); ok {
		return JSONResult(map[string]interface{}{
			"success":  true,
			"message":  "Story created successfully",
			"story_id": resultData["sys_id"],
			"number":   resultData["number"],
		}), nil
	}

	return JSONResult(NewErrorResponse("Unexpected response from ServiceNow", nil)), nil
}

func (r *Registry) updateStory(args map[string]interface{}) (*mcp.CallToolResult, error) {
	if r.readOnlyMode {
		return WriteBlockedResult(), nil
	}

	storyID := GetStringArg(args, "story_id", "")
	if storyID == "" {
		return JSONResult(NewErrorResponse("story_id is required", nil)), nil
	}

	data := map[string]interface{}{}

	if v := GetStringArg(args, "short_description", ""); v != "" {
		data["short_description"] = v
	}
	if v := GetStringArg(args, "state", ""); v != "" {
		data["state"] = v
	}
	if v := GetIntArg(args, "story_points", 0); v > 0 {
		data["story_points"] = v
	}
	if v, exists := args["blocked"]; exists {
		data["blocked"] = v
	}

	result, err := r.client.Put(fmt.Sprintf("/table/rm_story/%s", storyID), data)
	if err != nil {
		return JSONResult(NewErrorResponse("Failed to update story", err)), nil
	}

	if resultData, ok := result["result"].(map[string]interface{}); ok {
		return JSONResult(map[string]interface{}{
			"success":  true,
			"message":  "Story updated successfully",
			"story_id": resultData["sys_id"],
		}), nil
	}

	return JSONResult(NewErrorResponse("Unexpected response from ServiceNow", nil)), nil
}

func (r *Registry) createEpic(args map[string]interface{}) (*mcp.CallToolResult, error) {
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
	if v := GetStringArg(args, "product", ""); v != "" {
		data["product"] = v
	}

	result, err := r.client.Post("/table/rm_epic", data)
	if err != nil {
		return JSONResult(NewErrorResponse("Failed to create epic", err)), nil
	}

	if resultData, ok := result["result"].(map[string]interface{}); ok {
		return JSONResult(map[string]interface{}{
			"success": true,
			"message": "Epic created successfully",
			"epic_id": resultData["sys_id"],
			"number":  resultData["number"],
		}), nil
	}

	return JSONResult(NewErrorResponse("Unexpected response from ServiceNow", nil)), nil
}

func (r *Registry) updateEpic(args map[string]interface{}) (*mcp.CallToolResult, error) {
	if r.readOnlyMode {
		return WriteBlockedResult(), nil
	}

	epicID := GetStringArg(args, "epic_id", "")
	if epicID == "" {
		return JSONResult(NewErrorResponse("epic_id is required", nil)), nil
	}

	data := map[string]interface{}{}

	if v := GetStringArg(args, "short_description", ""); v != "" {
		data["short_description"] = v
	}
	if v := GetStringArg(args, "state", ""); v != "" {
		data["state"] = v
	}

	result, err := r.client.Put(fmt.Sprintf("/table/rm_epic/%s", epicID), data)
	if err != nil {
		return JSONResult(NewErrorResponse("Failed to update epic", err)), nil
	}

	if resultData, ok := result["result"].(map[string]interface{}); ok {
		return JSONResult(map[string]interface{}{
			"success": true,
			"message": "Epic updated successfully",
			"epic_id": resultData["sys_id"],
		}), nil
	}

	return JSONResult(NewErrorResponse("Unexpected response from ServiceNow", nil)), nil
}

func (r *Registry) createScrumTask(args map[string]interface{}) (*mcp.CallToolResult, error) {
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

	if v := GetStringArg(args, "story", ""); v != "" {
		data["story"] = v
	}
	if v := GetStringArg(args, "type", ""); v != "" {
		data["type"] = v
	}
	if v := GetStringArg(args, "assigned_to", ""); v != "" {
		data["assigned_to"] = v
	}
	if v := GetIntArg(args, "time_remaining", 0); v > 0 {
		data["time_remaining"] = v
	}

	result, err := r.client.Post("/table/rm_scrum_task", data)
	if err != nil {
		return JSONResult(NewErrorResponse("Failed to create scrum task", err)), nil
	}

	if resultData, ok := result["result"].(map[string]interface{}); ok {
		return JSONResult(map[string]interface{}{
			"success": true,
			"message": "Scrum task created successfully",
			"task_id": resultData["sys_id"],
			"number":  resultData["number"],
		}), nil
	}

	return JSONResult(NewErrorResponse("Unexpected response from ServiceNow", nil)), nil
}

func (r *Registry) updateScrumTask(args map[string]interface{}) (*mcp.CallToolResult, error) {
	if r.readOnlyMode {
		return WriteBlockedResult(), nil
	}

	taskID := GetStringArg(args, "task_id", "")
	if taskID == "" {
		return JSONResult(NewErrorResponse("task_id is required", nil)), nil
	}

	data := map[string]interface{}{}

	if v := GetStringArg(args, "state", ""); v != "" {
		data["state"] = v
	}
	if v := GetIntArg(args, "time_remaining", 0); v >= 0 {
		data["time_remaining"] = v
	}

	result, err := r.client.Put(fmt.Sprintf("/table/rm_scrum_task/%s", taskID), data)
	if err != nil {
		return JSONResult(NewErrorResponse("Failed to update scrum task", err)), nil
	}

	if resultData, ok := result["result"].(map[string]interface{}); ok {
		return JSONResult(map[string]interface{}{
			"success": true,
			"message": "Scrum task updated successfully",
			"task_id": resultData["sys_id"],
		}), nil
	}

	return JSONResult(NewErrorResponse("Unexpected response from ServiceNow", nil)), nil
}

func (r *Registry) createProject(args map[string]interface{}) (*mcp.CallToolResult, error) {
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
	if v := GetStringArg(args, "start_date", ""); v != "" {
		data["start_date"] = v
	}
	if v := GetStringArg(args, "end_date", ""); v != "" {
		data["end_date"] = v
	}

	result, err := r.client.Post("/table/pm_project", data)
	if err != nil {
		return JSONResult(NewErrorResponse("Failed to create project", err)), nil
	}

	if resultData, ok := result["result"].(map[string]interface{}); ok {
		return JSONResult(map[string]interface{}{
			"success":    true,
			"message":    "Project created successfully",
			"project_id": resultData["sys_id"],
			"number":     resultData["number"],
		}), nil
	}

	return JSONResult(NewErrorResponse("Unexpected response from ServiceNow", nil)), nil
}

func (r *Registry) updateProject(args map[string]interface{}) (*mcp.CallToolResult, error) {
	if r.readOnlyMode {
		return WriteBlockedResult(), nil
	}

	projectID := GetStringArg(args, "project_id", "")
	if projectID == "" {
		return JSONResult(NewErrorResponse("project_id is required", nil)), nil
	}

	data := map[string]interface{}{}

	if v := GetStringArg(args, "short_description", ""); v != "" {
		data["short_description"] = v
	}
	if v := GetStringArg(args, "state", ""); v != "" {
		data["state"] = v
	}

	result, err := r.client.Put(fmt.Sprintf("/table/pm_project/%s", projectID), data)
	if err != nil {
		return JSONResult(NewErrorResponse("Failed to update project", err)), nil
	}

	if resultData, ok := result["result"].(map[string]interface{}); ok {
		return JSONResult(map[string]interface{}{
			"success":    true,
			"message":    "Project updated successfully",
			"project_id": resultData["sys_id"],
		}), nil
	}

	return JSONResult(NewErrorResponse("Unexpected response from ServiceNow", nil)), nil
}
