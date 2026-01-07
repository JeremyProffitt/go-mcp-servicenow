package tools

import (
	"fmt"
	"strings"

	"github.com/elastiflow/go-mcp-servicenow/pkg/mcp"
)

// registerKnowledgeBaseTools registers all knowledge base tools
func (r *Registry) registerKnowledgeBaseTools(server *mcp.Server) int {
	count := 0

	// List Knowledge Bases
	server.RegisterTool(mcp.Tool{
		Name:        "list_knowledge_bases",
		Description: "List knowledge bases from ServiceNow",
		InputSchema: mcp.JSONSchema{
			Type: "object",
			Properties: map[string]mcp.Property{
				"limit": {
					Type:        "number",
					Description: "Maximum number of knowledge bases to return (default: 50)",
					Default:     50,
				},
				"active": {
					Type:        "boolean",
					Description: "Filter by active status",
				},
			},
		},
		Annotations: &mcp.ToolAnnotation{
			Title:        "List Knowledge Bases",
			ReadOnlyHint: true,
		},
	}, func(args map[string]interface{}) (*mcp.CallToolResult, error) {
		return r.listKnowledgeBases(args)
	})
	count++

	// List Articles
	server.RegisterTool(mcp.Tool{
		Name:        "list_knowledge_articles",
		Description: "List knowledge articles from ServiceNow",
		InputSchema: mcp.JSONSchema{
			Type: "object",
			Properties: map[string]mcp.Property{
				"limit": {
					Type:        "number",
					Description: "Maximum number of articles to return (default: 20)",
					Default:     20,
				},
				"offset": {
					Type:        "number",
					Description: "Offset for pagination",
					Default:     0,
				},
				"knowledge_base": {
					Type:        "string",
					Description: "Filter by knowledge base sys_id",
				},
				"category": {
					Type:        "string",
					Description: "Filter by category sys_id",
				},
				"query": {
					Type:        "string",
					Description: "Search query",
				},
			},
		},
		Annotations: &mcp.ToolAnnotation{
			Title:        "List Knowledge Articles",
			ReadOnlyHint: true,
		},
	}, func(args map[string]interface{}) (*mcp.CallToolResult, error) {
		return r.listKnowledgeArticles(args)
	})
	count++

	// Get Article
	server.RegisterTool(mcp.Tool{
		Name:        "get_knowledge_article",
		Description: "Get a specific knowledge article by ID or number",
		InputSchema: mcp.JSONSchema{
			Type: "object",
			Properties: map[string]mcp.Property{
				"article_id": {
					Type:        "string",
					Description: "Article sys_id or number",
				},
			},
			Required: []string{"article_id"},
		},
		Annotations: &mcp.ToolAnnotation{
			Title:        "Get Knowledge Article",
			ReadOnlyHint: true,
		},
	}, func(args map[string]interface{}) (*mcp.CallToolResult, error) {
		return r.getKnowledgeArticle(args)
	})
	count++

	// List KB Categories
	server.RegisterTool(mcp.Tool{
		Name:        "list_kb_categories",
		Description: "List knowledge base categories",
		InputSchema: mcp.JSONSchema{
			Type: "object",
			Properties: map[string]mcp.Property{
				"knowledge_base": {
					Type:        "string",
					Description: "Filter by knowledge base sys_id",
				},
				"parent": {
					Type:        "string",
					Description: "Filter by parent category sys_id",
				},
				"limit": {
					Type:        "number",
					Description: "Maximum number of categories to return (default: 100)",
					Default:     100,
				},
			},
		},
		Annotations: &mcp.ToolAnnotation{
			Title:        "List KB Categories",
			ReadOnlyHint: true,
		},
	}, func(args map[string]interface{}) (*mcp.CallToolResult, error) {
		return r.listKBCategories(args)
	})
	count++

	// Write operations
	if !r.readOnlyMode {
		// Create Knowledge Base
		server.RegisterTool(mcp.Tool{
			Name:        "create_knowledge_base",
			Description: "Create a new knowledge base in ServiceNow",
			InputSchema: mcp.JSONSchema{
				Type: "object",
				Properties: map[string]mcp.Property{
					"title": {
						Type:        "string",
						Description: "Knowledge base title",
					},
					"description": {
						Type:        "string",
						Description: "Knowledge base description",
					},
					"owner": {
						Type:        "string",
						Description: "Owner user sys_id",
					},
				},
				Required: []string{"title"},
			},
		}, func(args map[string]interface{}) (*mcp.CallToolResult, error) {
			return r.createKnowledgeBase(args)
		})
		count++

		// Create KB Category
		server.RegisterTool(mcp.Tool{
			Name:        "create_kb_category",
			Description: "Create a new category in a knowledge base",
			InputSchema: mcp.JSONSchema{
				Type: "object",
				Properties: map[string]mcp.Property{
					"label": {
						Type:        "string",
						Description: "Category label/name",
					},
					"knowledge_base": {
						Type:        "string",
						Description: "Knowledge base sys_id",
					},
					"parent": {
						Type:        "string",
						Description: "Parent category sys_id (for subcategories)",
					},
				},
				Required: []string{"label", "knowledge_base"},
			},
		}, func(args map[string]interface{}) (*mcp.CallToolResult, error) {
			return r.createKBCategory(args)
		})
		count++

		// Create Article
		server.RegisterTool(mcp.Tool{
			Name:        "create_knowledge_article",
			Description: "Create a new knowledge article in ServiceNow",
			InputSchema: mcp.JSONSchema{
				Type: "object",
				Properties: map[string]mcp.Property{
					"short_description": {
						Type:        "string",
						Description: "Article title/short description",
					},
					"text": {
						Type:        "string",
						Description: "Article body/content",
					},
					"knowledge_base": {
						Type:        "string",
						Description: "Knowledge base sys_id",
					},
					"category": {
						Type:        "string",
						Description: "Category sys_id",
					},
				},
				Required: []string{"short_description", "text", "knowledge_base"},
			},
		}, func(args map[string]interface{}) (*mcp.CallToolResult, error) {
			return r.createKnowledgeArticle(args)
		})
		count++

		// Update Article
		server.RegisterTool(mcp.Tool{
			Name:        "update_knowledge_article",
			Description: "Update an existing knowledge article",
			InputSchema: mcp.JSONSchema{
				Type: "object",
				Properties: map[string]mcp.Property{
					"article_id": {
						Type:        "string",
						Description: "Article sys_id",
					},
					"short_description": {
						Type:        "string",
						Description: "Article title",
					},
					"text": {
						Type:        "string",
						Description: "Article body",
					},
					"category": {
						Type:        "string",
						Description: "Category sys_id",
					},
				},
				Required: []string{"article_id"},
			},
		}, func(args map[string]interface{}) (*mcp.CallToolResult, error) {
			return r.updateKnowledgeArticle(args)
		})
		count++

		// Publish Article
		server.RegisterTool(mcp.Tool{
			Name:        "publish_knowledge_article",
			Description: "Publish a knowledge article",
			InputSchema: mcp.JSONSchema{
				Type: "object",
				Properties: map[string]mcp.Property{
					"article_id": {
						Type:        "string",
						Description: "Article sys_id",
					},
				},
				Required: []string{"article_id"},
			},
		}, func(args map[string]interface{}) (*mcp.CallToolResult, error) {
			return r.publishKnowledgeArticle(args)
		})
		count++
	}

	return count
}

func (r *Registry) listKnowledgeBases(args map[string]interface{}) (*mcp.CallToolResult, error) {
	limit := GetIntArg(args, "limit", 50)

	params := map[string]string{
		"sysparm_limit":                  fmt.Sprintf("%d", limit),
		"sysparm_display_value":          "true",
		"sysparm_exclude_reference_link": "true",
	}

	if active, exists := args["active"]; exists {
		if active.(bool) {
			params["sysparm_query"] = "active=true"
		} else {
			params["sysparm_query"] = "active=false"
		}
	}

	result, err := r.client.Get("/table/kb_knowledge_base", params)
	if err != nil {
		return JSONResult(NewErrorResponse("Failed to list knowledge bases", err)), nil
	}

	kbs := []map[string]interface{}{}
	if resultList, ok := result["result"].([]interface{}); ok {
		for _, item := range resultList {
			if data, ok := item.(map[string]interface{}); ok {
				kbs = append(kbs, map[string]interface{}{
					"sys_id":      data["sys_id"],
					"title":       data["title"],
					"description": data["description"],
					"owner":       data["owner"],
					"active":      data["active"],
				})
			}
		}
	}

	return JSONResult(map[string]interface{}{
		"success":         true,
		"message":         fmt.Sprintf("Found %d knowledge bases", len(kbs)),
		"knowledge_bases": kbs,
	}), nil
}

func (r *Registry) listKnowledgeArticles(args map[string]interface{}) (*mcp.CallToolResult, error) {
	limit := GetIntArg(args, "limit", 20)
	offset := GetIntArg(args, "offset", 0)
	kb := GetStringArg(args, "knowledge_base", "")
	category := GetStringArg(args, "category", "")
	query := GetStringArg(args, "query", "")

	params := map[string]string{
		"sysparm_limit":                  fmt.Sprintf("%d", limit),
		"sysparm_offset":                 fmt.Sprintf("%d", offset),
		"sysparm_display_value":          "true",
		"sysparm_exclude_reference_link": "true",
	}

	var filters []string
	if kb != "" {
		filters = append(filters, fmt.Sprintf("kb_knowledge_base=%s", kb))
	}
	if category != "" {
		filters = append(filters, fmt.Sprintf("kb_category=%s", category))
	}
	if query != "" {
		filters = append(filters, fmt.Sprintf("short_descriptionLIKE%s^ORtextLIKE%s", query, query))
	}

	if len(filters) > 0 {
		params["sysparm_query"] = strings.Join(filters, "^")
	}

	result, err := r.client.Get("/table/kb_knowledge", params)
	if err != nil {
		return JSONResult(NewErrorResponse("Failed to list knowledge articles", err)), nil
	}

	articles := []map[string]interface{}{}
	if resultList, ok := result["result"].([]interface{}); ok {
		for _, item := range resultList {
			if data, ok := item.(map[string]interface{}); ok {
				articles = append(articles, map[string]interface{}{
					"sys_id":             data["sys_id"],
					"number":             data["number"],
					"short_description":  data["short_description"],
					"kb_knowledge_base":  data["kb_knowledge_base"],
					"kb_category":        data["kb_category"],
					"workflow_state":     data["workflow_state"],
					"sys_view_count":     data["sys_view_count"],
					"sys_created_on":     data["sys_created_on"],
				})
			}
		}
	}

	return JSONResult(map[string]interface{}{
		"success":  true,
		"message":  fmt.Sprintf("Found %d articles", len(articles)),
		"articles": articles,
	}), nil
}

func (r *Registry) getKnowledgeArticle(args map[string]interface{}) (*mcp.CallToolResult, error) {
	articleID := GetStringArg(args, "article_id", "")
	if articleID == "" {
		return JSONResult(NewErrorResponse("article_id is required", nil)), nil
	}

	var params map[string]string
	var endpoint string

	if IsSysID(articleID) {
		endpoint = fmt.Sprintf("/table/kb_knowledge/%s", articleID)
		params = map[string]string{
			"sysparm_display_value":          "true",
			"sysparm_exclude_reference_link": "true",
		}
	} else {
		endpoint = "/table/kb_knowledge"
		params = map[string]string{
			"sysparm_query":                  fmt.Sprintf("number=%s", articleID),
			"sysparm_limit":                  "1",
			"sysparm_display_value":          "true",
			"sysparm_exclude_reference_link": "true",
		}
	}

	result, err := r.client.Get(endpoint, params)
	if err != nil {
		return JSONResult(NewErrorResponse("Failed to get article", err)), nil
	}

	var articleData map[string]interface{}
	if IsSysID(articleID) {
		articleData, _ = result["result"].(map[string]interface{})
	} else {
		if resultList, ok := result["result"].([]interface{}); ok && len(resultList) > 0 {
			articleData, _ = resultList[0].(map[string]interface{})
		}
	}

	if articleData == nil {
		return JSONResult(map[string]interface{}{
			"success": false,
			"message": fmt.Sprintf("Article not found: %s", articleID),
		}), nil
	}

	return JSONResult(map[string]interface{}{
		"success": true,
		"message": "Article found",
		"article": articleData,
	}), nil
}

func (r *Registry) listKBCategories(args map[string]interface{}) (*mcp.CallToolResult, error) {
	limit := GetIntArg(args, "limit", 100)
	kb := GetStringArg(args, "knowledge_base", "")
	parent := GetStringArg(args, "parent", "")

	params := map[string]string{
		"sysparm_limit":                  fmt.Sprintf("%d", limit),
		"sysparm_display_value":          "true",
		"sysparm_exclude_reference_link": "true",
	}

	var filters []string
	if kb != "" {
		filters = append(filters, fmt.Sprintf("kb_knowledge_base=%s", kb))
	}
	if parent != "" {
		filters = append(filters, fmt.Sprintf("parent_id=%s", parent))
	}

	if len(filters) > 0 {
		params["sysparm_query"] = strings.Join(filters, "^")
	}

	result, err := r.client.Get("/table/kb_category", params)
	if err != nil {
		return JSONResult(NewErrorResponse("Failed to list KB categories", err)), nil
	}

	categories := []map[string]interface{}{}
	if resultList, ok := result["result"].([]interface{}); ok {
		for _, item := range resultList {
			if data, ok := item.(map[string]interface{}); ok {
				categories = append(categories, map[string]interface{}{
					"sys_id":            data["sys_id"],
					"label":             data["label"],
					"kb_knowledge_base": data["kb_knowledge_base"],
					"parent_id":         data["parent_id"],
					"active":            data["active"],
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

func (r *Registry) createKnowledgeBase(args map[string]interface{}) (*mcp.CallToolResult, error) {
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
	if v := GetStringArg(args, "owner", ""); v != "" {
		data["owner"] = v
	}

	result, err := r.client.Post("/table/kb_knowledge_base", data)
	if err != nil {
		return JSONResult(NewErrorResponse("Failed to create knowledge base", err)), nil
	}

	if resultData, ok := result["result"].(map[string]interface{}); ok {
		return JSONResult(map[string]interface{}{
			"success":           true,
			"message":           "Knowledge base created successfully",
			"knowledge_base_id": resultData["sys_id"],
		}), nil
	}

	return JSONResult(NewErrorResponse("Unexpected response from ServiceNow", nil)), nil
}

func (r *Registry) createKBCategory(args map[string]interface{}) (*mcp.CallToolResult, error) {
	if r.readOnlyMode {
		return WriteBlockedResult(), nil
	}

	label := GetStringArg(args, "label", "")
	kb := GetStringArg(args, "knowledge_base", "")

	if label == "" || kb == "" {
		return JSONResult(NewErrorResponse("label and knowledge_base are required", nil)), nil
	}

	data := map[string]interface{}{
		"label":             label,
		"kb_knowledge_base": kb,
	}

	if v := GetStringArg(args, "parent", ""); v != "" {
		data["parent_id"] = v
	}

	result, err := r.client.Post("/table/kb_category", data)
	if err != nil {
		return JSONResult(NewErrorResponse("Failed to create KB category", err)), nil
	}

	if resultData, ok := result["result"].(map[string]interface{}); ok {
		return JSONResult(map[string]interface{}{
			"success":     true,
			"message":     "KB category created successfully",
			"category_id": resultData["sys_id"],
		}), nil
	}

	return JSONResult(NewErrorResponse("Unexpected response from ServiceNow", nil)), nil
}

func (r *Registry) createKnowledgeArticle(args map[string]interface{}) (*mcp.CallToolResult, error) {
	if r.readOnlyMode {
		return WriteBlockedResult(), nil
	}

	shortDesc := GetStringArg(args, "short_description", "")
	text := GetStringArg(args, "text", "")
	kb := GetStringArg(args, "knowledge_base", "")

	if shortDesc == "" || text == "" || kb == "" {
		return JSONResult(NewErrorResponse("short_description, text, and knowledge_base are required", nil)), nil
	}

	data := map[string]interface{}{
		"short_description": shortDesc,
		"text":              text,
		"kb_knowledge_base": kb,
	}

	if v := GetStringArg(args, "category", ""); v != "" {
		data["kb_category"] = v
	}

	result, err := r.client.Post("/table/kb_knowledge", data)
	if err != nil {
		return JSONResult(NewErrorResponse("Failed to create knowledge article", err)), nil
	}

	if resultData, ok := result["result"].(map[string]interface{}); ok {
		return JSONResult(map[string]interface{}{
			"success":        true,
			"message":        "Knowledge article created successfully",
			"article_id":     resultData["sys_id"],
			"article_number": resultData["number"],
		}), nil
	}

	return JSONResult(NewErrorResponse("Unexpected response from ServiceNow", nil)), nil
}

func (r *Registry) updateKnowledgeArticle(args map[string]interface{}) (*mcp.CallToolResult, error) {
	if r.readOnlyMode {
		return WriteBlockedResult(), nil
	}

	articleID := GetStringArg(args, "article_id", "")
	if articleID == "" {
		return JSONResult(NewErrorResponse("article_id is required", nil)), nil
	}

	data := map[string]interface{}{}

	if v := GetStringArg(args, "short_description", ""); v != "" {
		data["short_description"] = v
	}
	if v := GetStringArg(args, "text", ""); v != "" {
		data["text"] = v
	}
	if v := GetStringArg(args, "category", ""); v != "" {
		data["kb_category"] = v
	}

	result, err := r.client.Put(fmt.Sprintf("/table/kb_knowledge/%s", articleID), data)
	if err != nil {
		return JSONResult(NewErrorResponse("Failed to update knowledge article", err)), nil
	}

	if resultData, ok := result["result"].(map[string]interface{}); ok {
		return JSONResult(map[string]interface{}{
			"success":        true,
			"message":        "Knowledge article updated successfully",
			"article_id":     resultData["sys_id"],
			"article_number": resultData["number"],
		}), nil
	}

	return JSONResult(NewErrorResponse("Unexpected response from ServiceNow", nil)), nil
}

func (r *Registry) publishKnowledgeArticle(args map[string]interface{}) (*mcp.CallToolResult, error) {
	if r.readOnlyMode {
		return WriteBlockedResult(), nil
	}

	articleID := GetStringArg(args, "article_id", "")
	if articleID == "" {
		return JSONResult(NewErrorResponse("article_id is required", nil)), nil
	}

	data := map[string]interface{}{
		"workflow_state": "published",
	}

	result, err := r.client.Put(fmt.Sprintf("/table/kb_knowledge/%s", articleID), data)
	if err != nil {
		return JSONResult(NewErrorResponse("Failed to publish knowledge article", err)), nil
	}

	if resultData, ok := result["result"].(map[string]interface{}); ok {
		return JSONResult(map[string]interface{}{
			"success":        true,
			"message":        "Knowledge article published successfully",
			"article_id":     resultData["sys_id"],
			"article_number": resultData["number"],
		}), nil
	}

	return JSONResult(NewErrorResponse("Unexpected response from ServiceNow", nil)), nil
}
