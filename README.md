# go-mcp-servicenow

A Go implementation of the Model Context Protocol (MCP) server for ServiceNow. This server enables AI assistants to interact with ServiceNow instances through a standardized protocol.

## Features

- **Full MCP Protocol Support**: JSON-RPC 2.0 based communication
- **Multiple Authentication Methods**: Basic Auth, OAuth 2.0, API Key
- **Comprehensive ServiceNow Coverage**: 70+ tools across multiple domains
- **Dual Server Modes**: Stdio (for local) and HTTP (for remote/containerized)
- **Read-Only Mode**: Optional restriction of write operations
- **Rate Limiting**: Built-in protection (5 calls per 20 seconds)

## Tool Categories

| Category | Tools |
|----------|-------|
| Incidents | list_incidents, get_incident, create_incident, update_incident, add_incident_comment, resolve_incident |
| Catalog | list_catalogs, list_catalog_items, get_catalog_item, list_catalog_categories, list_catalog_item_variables, create_catalog_category, update_catalog_category, update_catalog_item, create_catalog_item_variable, move_catalog_items |
| Change Management | list_change_requests, get_change_request, create_change_request, update_change_request, add_change_task, submit_change_for_approval, approve_change, reject_change |
| Knowledge Base | list_knowledge_bases, list_knowledge_articles, get_knowledge_article, list_kb_categories, create_knowledge_base, create_kb_category, create_knowledge_article, update_knowledge_article, publish_knowledge_article |
| Users & Groups | list_users, get_user, list_groups, create_user, update_user, create_group, update_group, add_group_members, remove_group_members |
| Workflows | list_workflows, get_workflow, create_workflow, update_workflow, delete_workflow |
| Script Includes | list_script_includes, get_script_include, create_script_include, update_script_include, delete_script_include |
| Changesets | list_changesets, get_changeset, create_changeset, update_changeset, commit_changeset |
| Agile | list_stories, list_epics, list_scrum_tasks, list_projects, create_story, update_story, create_epic, update_epic, create_scrum_task, update_scrum_task, create_project, update_project |

## Installation

### Build from Source

```bash
git clone https://github.com/elastiflow/go-mcp-servicenow.git
cd go-mcp-servicenow
go build -o go-mcp-servicenow .
```

### Docker

```bash
docker build -t go-mcp-servicenow .
```

## Configuration

### Environment Variables

| Variable | Description | Required |
|----------|-------------|----------|
| `SERVICENOW_INSTANCE_URL` | ServiceNow instance URL (e.g., https://dev12345.service-now.com) | Yes |
| `SERVICENOW_AUTH_TYPE` | Authentication type: `basic`, `oauth`, or `api_key` | Yes |
| `SERVICENOW_USERNAME` | Username for basic/oauth auth | For basic/oauth |
| `SERVICENOW_PASSWORD` | Password for basic/oauth auth | For basic/oauth |
| `SERVICENOW_CLIENT_ID` | OAuth client ID | For oauth |
| `SERVICENOW_CLIENT_SECRET` | OAuth client secret | For oauth |
| `SERVICENOW_API_KEY` | API key for api_key auth | For api_key |
| `READ_ONLY_MODE` | Set to `true` to disable write operations | No |
| `MCP_AUTH_TOKEN` | Token for HTTP mode authentication | No |
| `MCP_LOG_DIR` | Directory for log files | No |
| `MCP_LOG_LEVEL` | Log level: debug, info, warn, error | No |

### Authentication Types

#### Basic Authentication
```bash
export SERVICENOW_INSTANCE_URL="https://dev12345.service-now.com"
export SERVICENOW_AUTH_TYPE="basic"
export SERVICENOW_USERNAME="admin"
export SERVICENOW_PASSWORD="password"
```

#### OAuth 2.0 (Client Credentials)
```bash
export SERVICENOW_INSTANCE_URL="https://dev12345.service-now.com"
export SERVICENOW_AUTH_TYPE="oauth"
export SERVICENOW_CLIENT_ID="your_client_id"
export SERVICENOW_CLIENT_SECRET="your_client_secret"
```

#### OAuth 2.0 (Password Grant)
```bash
export SERVICENOW_INSTANCE_URL="https://dev12345.service-now.com"
export SERVICENOW_AUTH_TYPE="oauth"
export SERVICENOW_CLIENT_ID="your_client_id"
export SERVICENOW_CLIENT_SECRET="your_client_secret"
export SERVICENOW_USERNAME="admin"
export SERVICENOW_PASSWORD="password"
```

#### API Key
```bash
export SERVICENOW_INSTANCE_URL="https://dev12345.service-now.com"
export SERVICENOW_AUTH_TYPE="api_key"
export SERVICENOW_API_KEY="your_api_key"
```

## Usage

### Stdio Mode (Default)

```bash
./go-mcp-servicenow
```

### HTTP Mode

```bash
./go-mcp-servicenow --http --host 0.0.0.0 --port 3000
```

### Command Line Options

| Flag | Description | Default |
|------|-------------|---------|
| `--http` | Run in HTTP mode | false |
| `--host` | HTTP host | 127.0.0.1 |
| `--port` | HTTP port | 3000 |
| `--read-only` | Enable read-only mode | false |
| `--log-dir` | Log directory | OS temp dir |
| `--log-level` | Log level | info |
| `--version` | Show version | - |

### Docker

```bash
docker run -p 3000:3000 \
  -e SERVICENOW_INSTANCE_URL="https://dev12345.service-now.com" \
  -e SERVICENOW_AUTH_TYPE="basic" \
  -e SERVICENOW_USERNAME="admin" \
  -e SERVICENOW_PASSWORD="password" \
  go-mcp-servicenow
```

## Claude Desktop Integration

Add to your Claude Desktop configuration (`claude_desktop_config.json`):

```json
{
  "mcpServers": {
    "servicenow": {
      "command": "/path/to/go-mcp-servicenow",
      "env": {
        "SERVICENOW_INSTANCE_URL": "https://dev12345.service-now.com",
        "SERVICENOW_AUTH_TYPE": "basic",
        "SERVICENOW_USERNAME": "admin",
        "SERVICENOW_PASSWORD": "password"
      }
    }
  }
}
```

## API Endpoints (HTTP Mode)

| Endpoint | Method | Description |
|----------|--------|-------------|
| `/` | POST | MCP JSON-RPC endpoint |
| `/health` | GET | Health check |

## Development

### Project Structure

```
go-mcp-servicenow/
├── main.go
├── go.mod
├── Dockerfile
├── ecs-task-definition.json
├── README.md
└── pkg/
    ├── mcp/
    │   ├── server.go      # MCP server implementation
    │   └── types.go       # MCP protocol types
    ├── auth/
    │   └── auth.go        # MCP authentication
    ├── logging/
    │   └── logging.go     # Structured logging
    ├── servicenow/
    │   ├── client.go      # ServiceNow API client
    │   └── config.go      # Configuration handling
    └── tools/
        ├── registry.go    # Tool registration
        ├── helpers.go     # Utility functions
        ├── incidents.go   # Incident tools
        ├── catalog.go     # Catalog tools
        ├── change.go      # Change management tools
        ├── knowledge.go   # Knowledge base tools
        ├── users.go       # User/group tools
        ├── workflow.go    # Workflow tools
        ├── script_include.go  # Script include tools
        ├── changeset.go   # Changeset tools
        └── agile.go       # Agile tools
```

### Building

```bash
go build -o go-mcp-servicenow .
```

### Testing

```bash
go test ./...
```

## License

MIT License

## Credits

Ported from [echelon-ai-labs/servicenow-mcp](https://github.com/echelon-ai-labs/servicenow-mcp) (Python).
