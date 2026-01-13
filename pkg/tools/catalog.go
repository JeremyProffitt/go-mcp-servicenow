package tools

import (
	"fmt"
	"strings"

	"github.com/elastiflow/go-mcp-servicenow/pkg/mcp"
)

// registerCatalogTools registers all service catalog tools
func (r *Registry) registerCatalogTools(server *mcp.Server) int {
	count := 0

	// Helper for limit/offset constraints
	limitMin := float64(1)
	limitMax := float64(1000)
	offsetMin := float64(0)

	// List Catalogs
	server.RegisterTool(mcp.Tool{
		Name:        "list_catalogs",
		Description: "List available service catalogs. Catalogs contain categories which contain orderable items.",
		InputSchema: mcp.JSONSchema{
			Type: "object",
			Properties: map[string]mcp.Property{
				"limit": {
					Type:        "number",
					Description: "Maximum number of catalogs to return (default: 50)",
					Default:     50,
					Minimum:     &limitMin,
					Maximum:     &limitMax,
				},
			},
		},
		Annotations: &mcp.ToolAnnotation{
			Title:        "List Catalogs",
			ReadOnlyHint: true,
		},
	}, func(args map[string]interface{}) (*mcp.CallToolResult, error) {
		return r.listCatalogs(args)
	})
	count++

	// List Catalog Items
	server.RegisterTool(mcp.Tool{
		Name:        "list_catalog_items",
		Description: "List service catalog items (orderable products/services) with optional filtering by category or search query.",
		InputSchema: mcp.JSONSchema{
			Type: "object",
			Properties: map[string]mcp.Property{
				"limit": {
					Type:        "number",
					Description: "Maximum number of items to return (default: 50)",
					Default:     50,
					Minimum:     &limitMin,
					Maximum:     &limitMax,
				},
				"offset": {
					Type:        "number",
					Description: "Offset for pagination (default: 0)",
					Default:     0,
					Minimum:     &offsetMin,
				},
				"category": {
					Type:        "string",
					Description: "Filter by category sys_id (e.g., 'a1b2c3d4e5f6g7h8i9j0k1l2m3n4o5p6')",
				},
				"query": {
					Type:        "string",
					Description: "Search query (searches name and short_description). For advanced filtering, use ServiceNow encoded query syntax (^ for AND, | for OR)",
				},
			},
		},
		Annotations: &mcp.ToolAnnotation{
			Title:        "List Catalog Items",
			ReadOnlyHint: true,
		},
	}, func(args map[string]interface{}) (*mcp.CallToolResult, error) {
		return r.listCatalogItems(args)
	})
	count++

	// Get Catalog Item
	server.RegisterTool(mcp.Tool{
		Name:        "get_catalog_item",
		Description: "Get detailed information about a specific catalog item including description, price, and configuration options.",
		InputSchema: mcp.JSONSchema{
			Type: "object",
			Properties: map[string]mcp.Property{
				"item_id": {
					Type:        "string",
					Description: "Catalog item sys_id (e.g., 'a1b2c3d4e5f6g7h8i9j0k1l2m3n4o5p6')",
				},
			},
			Required: []string{"item_id"},
		},
		Annotations: &mcp.ToolAnnotation{
			Title:        "Get Catalog Item",
			ReadOnlyHint: true,
		},
	}, func(args map[string]interface{}) (*mcp.CallToolResult, error) {
		return r.getCatalogItem(args)
	})
	count++

	// List Catalog Categories
	server.RegisterTool(mcp.Tool{
		Name:        "list_catalog_categories",
		Description: "List service catalog categories. Categories organize catalog items and can be nested (parent/child hierarchy).",
		InputSchema: mcp.JSONSchema{
			Type: "object",
			Properties: map[string]mcp.Property{
				"catalog_id": {
					Type:        "string",
					Description: "Filter by catalog sys_id (e.g., 'a1b2c3d4e5f6g7h8i9j0k1l2m3n4o5p6')",
				},
				"parent_id": {
					Type:        "string",
					Description: "Filter by parent category sys_id to get subcategories",
				},
				"limit": {
					Type:        "number",
					Description: "Maximum number of categories to return (default: 100)",
					Default:     100,
					Minimum:     &limitMin,
					Maximum:     &limitMax,
				},
			},
		},
		Annotations: &mcp.ToolAnnotation{
			Title:        "List Catalog Categories",
			ReadOnlyHint: true,
		},
	}, func(args map[string]interface{}) (*mcp.CallToolResult, error) {
		return r.listCatalogCategories(args)
	})
	count++

	// List Catalog Item Variables
	server.RegisterTool(mcp.Tool{
		Name:        "list_catalog_item_variables",
		Description: "List all form variables (input fields) for a catalog item. Variables define the questions/options shown when ordering.",
		InputSchema: mcp.JSONSchema{
			Type: "object",
			Properties: map[string]mcp.Property{
				"item_id": {
					Type:        "string",
					Description: "Catalog item sys_id (e.g., 'a1b2c3d4e5f6g7h8i9j0k1l2m3n4o5p6')",
				},
			},
			Required: []string{"item_id"},
		},
		Annotations: &mcp.ToolAnnotation{
			Title:        "List Catalog Item Variables",
			ReadOnlyHint: true,
		},
	}, func(args map[string]interface{}) (*mcp.CallToolResult, error) {
		return r.listCatalogItemVariables(args)
	})
	count++

	// Write operations
	if !r.readOnlyMode {
		// Create Catalog Category
		server.RegisterTool(mcp.Tool{
			Name:        "create_catalog_category",
			Description: "Create a new service catalog category. Categories organize catalog items and can be nested.",
			InputSchema: mcp.JSONSchema{
				Type: "object",
				Properties: map[string]mcp.Property{
					"title": {
						Type:        "string",
						Description: "Category title/name",
					},
					"description": {
						Type:        "string",
						Description: "Category description",
					},
					"catalog_id": {
						Type:        "string",
						Description: "Parent catalog sys_id (e.g., 'a1b2c3d4e5f6g7h8i9j0k1l2m3n4o5p6')",
					},
					"parent_id": {
						Type:        "string",
						Description: "Parent category sys_id for creating subcategories",
					},
				},
				Required: []string{"title"},
			},
			Annotations: &mcp.ToolAnnotation{
				Title: "Create Catalog Category",
			},
		}, func(args map[string]interface{}) (*mcp.CallToolResult, error) {
			return r.createCatalogCategory(args)
		})
		count++

		// Update Catalog Category
		server.RegisterTool(mcp.Tool{
			Name:        "update_catalog_category",
			Description: "Update an existing catalog category. At least one field besides category_id must be provided.",
			InputSchema: mcp.JSONSchema{
				Type: "object",
				Properties: map[string]mcp.Property{
					"category_id": {
						Type:        "string",
						Description: "Category sys_id (e.g., 'a1b2c3d4e5f6g7h8i9j0k1l2m3n4o5p6')",
					},
					"title": {
						Type:        "string",
						Description: "Category title/name",
					},
					"description": {
						Type:        "string",
						Description: "Category description",
					},
				},
				Required: []string{"category_id"},
			},
			Annotations: &mcp.ToolAnnotation{
				Title: "Update Catalog Category",
			},
		}, func(args map[string]interface{}) (*mcp.CallToolResult, error) {
			return r.updateCatalogCategory(args)
		})
		count++

		// Update Catalog Item
		server.RegisterTool(mcp.Tool{
			Name:        "update_catalog_item",
			Description: "Update a catalog item. At least one field besides item_id must be provided.",
			InputSchema: mcp.JSONSchema{
				Type: "object",
				Properties: map[string]mcp.Property{
					"item_id": {
						Type:        "string",
						Description: "Catalog item sys_id (e.g., 'a1b2c3d4e5f6g7h8i9j0k1l2m3n4o5p6')",
					},
					"name": {
						Type:        "string",
						Description: "Item name",
					},
					"short_description": {
						Type:        "string",
						Description: "Brief summary of the item",
					},
					"description": {
						Type:        "string",
						Description: "Full description with details",
					},
					"category": {
						Type:        "string",
						Description: "Category sys_id to move the item to",
					},
					"active": {
						Type:        "boolean",
						Description: "Whether the item is active and orderable",
					},
				},
				Required: []string{"item_id"},
			},
			Annotations: &mcp.ToolAnnotation{
				Title: "Update Catalog Item",
			},
		}, func(args map[string]interface{}) (*mcp.CallToolResult, error) {
			return r.updateCatalogItem(args)
		})
		count++

		// Create Catalog Item Variable
		server.RegisterTool(mcp.Tool{
			Name:        "create_catalog_item_variable",
			Description: "Create a new form variable (input field) for a catalog item. Variables define questions shown when ordering.",
			InputSchema: mcp.JSONSchema{
				Type: "object",
				Properties: map[string]mcp.Property{
					"item_id": {
						Type:        "string",
						Description: "Catalog item sys_id (e.g., 'a1b2c3d4e5f6g7h8i9j0k1l2m3n4o5p6')",
					},
					"name": {
						Type:        "string",
						Description: "Variable internal name (no spaces, used in scripts)",
					},
					"question_text": {
						Type:        "string",
						Description: "Label text shown to users",
					},
					"type": {
						Type:        "string",
						Description: "Variable input type",
						Enum:        []string{"string", "integer", "boolean", "reference", "select_box", "multi_line_text"},
					},
					"mandatory": {
						Type:        "boolean",
						Description: "Whether the variable is required",
					},
					"order": {
						Type:        "number",
						Description: "Display order (lower numbers appear first)",
					},
				},
				Required: []string{"item_id", "name", "question_text", "type"},
			},
			Annotations: &mcp.ToolAnnotation{
				Title: "Create Catalog Item Variable",
			},
		}, func(args map[string]interface{}) (*mcp.CallToolResult, error) {
			return r.createCatalogItemVariable(args)
		})
		count++

		// Move Catalog Items
		server.RegisterTool(mcp.Tool{
			Name:        "move_catalog_items",
			Description: "Move one or more catalog items to a different category.",
			InputSchema: mcp.JSONSchema{
				Type: "object",
				Properties: map[string]mcp.Property{
					"item_ids": {
						Type:        "array",
						Description: "List of catalog item sys_ids to move",
						Items:       &mcp.Property{Type: "string"},
					},
					"target_category_id": {
						Type:        "string",
						Description: "Target category sys_id (e.g., 'a1b2c3d4e5f6g7h8i9j0k1l2m3n4o5p6')",
					},
				},
				Required: []string{"item_ids", "target_category_id"},
			},
			Annotations: &mcp.ToolAnnotation{
				Title: "Move Catalog Items",
			},
		}, func(args map[string]interface{}) (*mcp.CallToolResult, error) {
			return r.moveCatalogItems(args)
		})
		count++
	}

	return count
}

func (r *Registry) listCatalogs(args map[string]interface{}) (*mcp.CallToolResult, error) {
	limit := GetIntArg(args, "limit", 50)

	params := map[string]string{
		"sysparm_limit":                  fmt.Sprintf("%d", limit),
		"sysparm_display_value":          "true",
		"sysparm_exclude_reference_link": "true",
	}

	result, err := r.client.Get("/table/sc_catalog", params)
	if err != nil {
		return JSONResult(NewErrorResponse("Failed to list catalogs", err)), nil
	}

	catalogs := []map[string]interface{}{}
	if resultList, ok := result["result"].([]interface{}); ok {
		for _, item := range resultList {
			if data, ok := item.(map[string]interface{}); ok {
				catalogs = append(catalogs, map[string]interface{}{
					"sys_id":      data["sys_id"],
					"title":       data["title"],
					"description": data["description"],
					"active":      data["active"],
				})
			}
		}
	}

	return JSONResult(map[string]interface{}{
		"success":  true,
		"message":  fmt.Sprintf("Found %d catalogs", len(catalogs)),
		"catalogs": catalogs,
	}), nil
}

func (r *Registry) listCatalogItems(args map[string]interface{}) (*mcp.CallToolResult, error) {
	limit := GetIntArg(args, "limit", 50)
	offset := GetIntArg(args, "offset", 0)
	category := GetStringArg(args, "category", "")
	query := GetStringArg(args, "query", "")

	params := map[string]string{
		"sysparm_limit":                  fmt.Sprintf("%d", limit),
		"sysparm_offset":                 fmt.Sprintf("%d", offset),
		"sysparm_display_value":          "true",
		"sysparm_exclude_reference_link": "true",
	}

	var filters []string
	if category != "" {
		filters = append(filters, fmt.Sprintf("category=%s", category))
	}
	if query != "" {
		filters = append(filters, fmt.Sprintf("nameLIKE%s^ORshort_descriptionLIKE%s", query, query))
	}

	if len(filters) > 0 {
		params["sysparm_query"] = strings.Join(filters, "^")
	}

	result, err := r.client.Get("/table/sc_cat_item", params)
	if err != nil {
		return JSONResult(NewErrorResponse("Failed to list catalog items", err)), nil
	}

	items := []map[string]interface{}{}
	if resultList, ok := result["result"].([]interface{}); ok {
		for _, item := range resultList {
			if data, ok := item.(map[string]interface{}); ok {
				items = append(items, map[string]interface{}{
					"sys_id":            data["sys_id"],
					"name":              data["name"],
					"short_description": data["short_description"],
					"category":          data["category"],
					"active":            data["active"],
					"price":             data["price"],
				})
			}
		}
	}

	return JSONResult(map[string]interface{}{
		"success": true,
		"message": fmt.Sprintf("Found %d catalog items", len(items)),
		"items":   items,
	}), nil
}

func (r *Registry) getCatalogItem(args map[string]interface{}) (*mcp.CallToolResult, error) {
	itemID := GetStringArg(args, "item_id", "")
	if itemID == "" {
		return JSONResult(NewErrorResponse("item_id is required", nil)), nil
	}

	params := map[string]string{
		"sysparm_display_value":          "true",
		"sysparm_exclude_reference_link": "true",
	}

	result, err := r.client.Get(fmt.Sprintf("/table/sc_cat_item/%s", itemID), params)
	if err != nil {
		return JSONResult(NewErrorResponse("Failed to get catalog item", err)), nil
	}

	if data, ok := result["result"].(map[string]interface{}); ok {
		return JSONResult(map[string]interface{}{
			"success": true,
			"message": "Catalog item found",
			"item":    data,
		}), nil
	}

	return JSONResult(map[string]interface{}{
		"success": false,
		"message": fmt.Sprintf("Catalog item not found: %s", itemID),
	}), nil
}

func (r *Registry) listCatalogCategories(args map[string]interface{}) (*mcp.CallToolResult, error) {
	limit := GetIntArg(args, "limit", 100)
	catalogID := GetStringArg(args, "catalog_id", "")
	parentID := GetStringArg(args, "parent_id", "")

	params := map[string]string{
		"sysparm_limit":                  fmt.Sprintf("%d", limit),
		"sysparm_display_value":          "true",
		"sysparm_exclude_reference_link": "true",
	}

	var filters []string
	if catalogID != "" {
		filters = append(filters, fmt.Sprintf("sc_catalog=%s", catalogID))
	}
	if parentID != "" {
		filters = append(filters, fmt.Sprintf("parent=%s", parentID))
	}

	if len(filters) > 0 {
		params["sysparm_query"] = strings.Join(filters, "^")
	}

	result, err := r.client.Get("/table/sc_category", params)
	if err != nil {
		return JSONResult(NewErrorResponse("Failed to list catalog categories", err)), nil
	}

	categories := []map[string]interface{}{}
	if resultList, ok := result["result"].([]interface{}); ok {
		for _, item := range resultList {
			if data, ok := item.(map[string]interface{}); ok {
				categories = append(categories, map[string]interface{}{
					"sys_id":      data["sys_id"],
					"title":       data["title"],
					"description": data["description"],
					"parent":      data["parent"],
					"sc_catalog":  data["sc_catalog"],
					"active":      data["active"],
				})
			}
		}
	}

	return JSONResult(map[string]interface{}{
		"success":    true,
		"message":    fmt.Sprintf("Found %d categories", len(categories)),
		"categories": categories,
	}), nil
}

func (r *Registry) listCatalogItemVariables(args map[string]interface{}) (*mcp.CallToolResult, error) {
	itemID := GetStringArg(args, "item_id", "")
	if itemID == "" {
		return JSONResult(NewErrorResponse("item_id is required", nil)), nil
	}

	params := map[string]string{
		"sysparm_query":                  fmt.Sprintf("cat_item=%s", itemID),
		"sysparm_display_value":          "true",
		"sysparm_exclude_reference_link": "true",
	}

	result, err := r.client.Get("/table/item_option_new", params)
	if err != nil {
		return JSONResult(NewErrorResponse("Failed to list catalog item variables", err)), nil
	}

	variables := []map[string]interface{}{}
	if resultList, ok := result["result"].([]interface{}); ok {
		for _, item := range resultList {
			if data, ok := item.(map[string]interface{}); ok {
				variables = append(variables, map[string]interface{}{
					"sys_id":        data["sys_id"],
					"name":          data["name"],
					"question_text": data["question_text"],
					"type":          data["type"],
					"mandatory":     data["mandatory"],
					"order":         data["order"],
				})
			}
		}
	}

	return JSONResult(map[string]interface{}{
		"success":   true,
		"message":   fmt.Sprintf("Found %d variables", len(variables)),
		"variables": variables,
	}), nil
}

func (r *Registry) createCatalogCategory(args map[string]interface{}) (*mcp.CallToolResult, error) {
	if r.readOnlyMode {
		return WriteBlockedResult(), nil
	}

	title := GetStringArg(args, "title", "")
	if title == "" {
		return JSONResult(NewErrorResponse("title is required", nil)), nil
	}

	data := map[string]interface{}{
		"title": title,
	}

	if v := GetStringArg(args, "description", ""); v != "" {
		data["description"] = v
	}
	if v := GetStringArg(args, "catalog_id", ""); v != "" {
		data["sc_catalog"] = v
	}
	if v := GetStringArg(args, "parent_id", ""); v != "" {
		data["parent"] = v
	}

	result, err := r.client.Post("/table/sc_category", data)
	if err != nil {
		return JSONResult(NewErrorResponse("Failed to create catalog category", err)), nil
	}

	if resultData, ok := result["result"].(map[string]interface{}); ok {
		return JSONResult(map[string]interface{}{
			"success":     true,
			"message":     "Catalog category created successfully",
			"category_id": resultData["sys_id"],
		}), nil
	}

	return JSONResult(NewErrorResponse("Unexpected response from ServiceNow", nil)), nil
}

func (r *Registry) updateCatalogCategory(args map[string]interface{}) (*mcp.CallToolResult, error) {
	if r.readOnlyMode {
		return WriteBlockedResult(), nil
	}

	categoryID := GetStringArg(args, "category_id", "")
	if categoryID == "" {
		return JSONResult(NewErrorResponse("category_id is required", nil)), nil
	}

	data := map[string]interface{}{}

	if v := GetStringArg(args, "title", ""); v != "" {
		data["title"] = v
	}
	if v := GetStringArg(args, "description", ""); v != "" {
		data["description"] = v
	}

	result, err := r.client.Put(fmt.Sprintf("/table/sc_category/%s", categoryID), data)
	if err != nil {
		return JSONResult(NewErrorResponse("Failed to update catalog category", err)), nil
	}

	if resultData, ok := result["result"].(map[string]interface{}); ok {
		return JSONResult(map[string]interface{}{
			"success":     true,
			"message":     "Catalog category updated successfully",
			"category_id": resultData["sys_id"],
		}), nil
	}

	return JSONResult(NewErrorResponse("Unexpected response from ServiceNow", nil)), nil
}

func (r *Registry) updateCatalogItem(args map[string]interface{}) (*mcp.CallToolResult, error) {
	if r.readOnlyMode {
		return WriteBlockedResult(), nil
	}

	itemID := GetStringArg(args, "item_id", "")
	if itemID == "" {
		return JSONResult(NewErrorResponse("item_id is required", nil)), nil
	}

	data := map[string]interface{}{}

	if v := GetStringArg(args, "name", ""); v != "" {
		data["name"] = v
	}
	if v := GetStringArg(args, "short_description", ""); v != "" {
		data["short_description"] = v
	}
	if v := GetStringArg(args, "description", ""); v != "" {
		data["description"] = v
	}
	if v := GetStringArg(args, "category", ""); v != "" {
		data["category"] = v
	}
	if v, exists := args["active"]; exists {
		data["active"] = v
	}

	result, err := r.client.Put(fmt.Sprintf("/table/sc_cat_item/%s", itemID), data)
	if err != nil {
		return JSONResult(NewErrorResponse("Failed to update catalog item", err)), nil
	}

	if resultData, ok := result["result"].(map[string]interface{}); ok {
		return JSONResult(map[string]interface{}{
			"success": true,
			"message": "Catalog item updated successfully",
			"item_id": resultData["sys_id"],
		}), nil
	}

	return JSONResult(NewErrorResponse("Unexpected response from ServiceNow", nil)), nil
}

func (r *Registry) createCatalogItemVariable(args map[string]interface{}) (*mcp.CallToolResult, error) {
	if r.readOnlyMode {
		return WriteBlockedResult(), nil
	}

	itemID := GetStringArg(args, "item_id", "")
	name := GetStringArg(args, "name", "")
	questionText := GetStringArg(args, "question_text", "")
	varType := GetStringArg(args, "type", "")

	if itemID == "" || name == "" || questionText == "" || varType == "" {
		return JSONResult(NewErrorResponse("item_id, name, question_text, and type are required", nil)), nil
	}

	data := map[string]interface{}{
		"cat_item":      itemID,
		"name":          name,
		"question_text": questionText,
		"type":          varType,
	}

	if v, exists := args["mandatory"]; exists {
		data["mandatory"] = v
	}
	if v := GetIntArg(args, "order", 0); v > 0 {
		data["order"] = v
	}

	result, err := r.client.Post("/table/item_option_new", data)
	if err != nil {
		return JSONResult(NewErrorResponse("Failed to create catalog item variable", err)), nil
	}

	if resultData, ok := result["result"].(map[string]interface{}); ok {
		return JSONResult(map[string]interface{}{
			"success":     true,
			"message":     "Catalog item variable created successfully",
			"variable_id": resultData["sys_id"],
		}), nil
	}

	return JSONResult(NewErrorResponse("Unexpected response from ServiceNow", nil)), nil
}

func (r *Registry) moveCatalogItems(args map[string]interface{}) (*mcp.CallToolResult, error) {
	if r.readOnlyMode {
		return WriteBlockedResult(), nil
	}

	itemIDs := GetStringArrayArg(args, "item_ids")
	targetCategoryID := GetStringArg(args, "target_category_id", "")

	if len(itemIDs) == 0 || targetCategoryID == "" {
		return JSONResult(NewErrorResponse("item_ids and target_category_id are required", nil)), nil
	}

	movedCount := 0
	var lastErr error

	for _, itemID := range itemIDs {
		data := map[string]interface{}{
			"category": targetCategoryID,
		}

		_, err := r.client.Put(fmt.Sprintf("/table/sc_cat_item/%s", itemID), data)
		if err != nil {
			lastErr = err
		} else {
			movedCount++
		}
	}

	if movedCount == len(itemIDs) {
		return JSONResult(map[string]interface{}{
			"success": true,
			"message": fmt.Sprintf("Successfully moved %d catalog items", movedCount),
		}), nil
	}

	return JSONResult(map[string]interface{}{
		"success": movedCount > 0,
		"message": fmt.Sprintf("Moved %d of %d items. Last error: %v", movedCount, len(itemIDs), lastErr),
	}), nil
}
