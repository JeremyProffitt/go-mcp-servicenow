package tools

import (
	"github.com/elastiflow/go-mcp-servicenow/pkg/logging"
	"github.com/elastiflow/go-mcp-servicenow/pkg/mcp"
	"github.com/elastiflow/go-mcp-servicenow/pkg/servicenow"
)

// Registry manages tool registration
type Registry struct {
	client       *servicenow.Client
	logger       *logging.Logger
	readOnlyMode bool
}

// NewRegistry creates a new tool registry
func NewRegistry(client *servicenow.Client, logger *logging.Logger, readOnlyMode bool) *Registry {
	return &Registry{
		client:       client,
		logger:       logger,
		readOnlyMode: readOnlyMode,
	}
}

// RegisterAll registers all tools with the MCP server
func (r *Registry) RegisterAll(server *mcp.Server) int {
	count := 0

	// Incident Management Tools (read-only always registered)
	count += r.registerIncidentTools(server)

	// Catalog Tools
	count += r.registerCatalogTools(server)

	// Change Management Tools
	count += r.registerChangeTools(server)

	// Knowledge Base Tools
	count += r.registerKnowledgeBaseTools(server)

	// User Management Tools
	count += r.registerUserTools(server)

	// Workflow Tools
	count += r.registerWorkflowTools(server)

	// Script Include Tools
	count += r.registerScriptIncludeTools(server)

	// Changeset Tools
	count += r.registerChangesetTools(server)

	// Agile Tools (Story, Epic, Scrum Task, Project)
	count += r.registerAgileTools(server)

	// Meta tool: list_tool_packages
	r.registerMetaTools(server)
	count++

	return count
}

// registerMetaTools registers metadata/introspection tools
func (r *Registry) registerMetaTools(server *mcp.Server) {
	server.RegisterTool(mcp.Tool{
		Name:        "list_tool_packages",
		Description: "Lists available tool packages and the currently loaded one.",
		InputSchema: mcp.JSONSchema{
			Type:       "object",
			Properties: map[string]mcp.Property{},
		},
		Annotations: &mcp.ToolAnnotation{
			Title:        "List Tool Packages",
			ReadOnlyHint: true,
		},
	}, func(args map[string]interface{}) (*mcp.CallToolResult, error) {
		result := map[string]interface{}{
			"current_package": "full",
			"available_packages": []string{
				"full",
				"service_desk",
				"catalog_builder",
				"change_coordinator",
				"knowledge_author",
				"platform_developer",
				"system_administrator",
				"agile_management",
				"none",
			},
			"message": "Currently loaded package: 'full'. Set MCP_TOOL_PACKAGE env var to switch.",
		}
		return JSONResult(result), nil
	})
}
