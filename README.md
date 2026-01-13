# go-mcp-servicenow

A Go implementation of the Model Context Protocol (MCP) server for ServiceNow. This server enables AI assistants to interact with ServiceNow instances through a standardized protocol.

## Features

- **Full MCP Protocol Support**: JSON-RPC 2.0 based communication
- **Multiple Authentication Methods**: Basic Auth, OAuth 2.0, API Key
- **Comprehensive ServiceNow Coverage**: 70+ tools across multiple domains
- **Dual Server Modes**: Stdio (for local) and HTTP (for remote/containerized)
- **Read-Only Mode**: Optional restriction of write operations
- **Rate Limiting**: Built-in protection (5 calls per 20 seconds)

## ServiceNow Concepts

Understanding these core ServiceNow concepts will help you use this MCP server effectively:

### sys_id

Every record in ServiceNow has a unique 32-character hexadecimal identifier called a `sys_id`. This is the primary key for all records.

- **Format**: 32 hexadecimal characters (e.g., `a1b2c3d4e5f6g7h8i9j0k1l2m3n4o5p6`)
- **Usage**: Most tools accept either sys_id or human-readable identifiers (like INC0010001)
- **Example**: When `get_incident` returns `sys_id: "6816f79cc0a8016401c5a33be04be441"`, use this value for updates

### display_value vs value

ServiceNow reference fields can return either raw values or human-readable display values:

- **value**: The sys_id of the referenced record
- **display_value**: Human-readable text (e.g., user's full name instead of sys_id)
- This server returns `display_value` by default for readability

### Encoded Query Syntax

Many list tools support ServiceNow's encoded query syntax for advanced filtering:

| Operator | Meaning | Example |
|----------|---------|---------|
| `=` | Equals | `state=1` |
| `!=` | Not equals | `state!=7` |
| `^` | AND | `state=1^priority=2` |
| `^OR` | OR | `state=1^ORstate=2` |
| `LIKE` | Contains | `short_descriptionLIKEnetwork` |
| `STARTSWITH` | Starts with | `numberSTARTSWITHINC` |
| `>` | Greater than | `priority>2` |
| `<` | Less than | `sys_created_on<2024-01-01` |
| `ORDERBY` | Sort ascending | `ORDERBYsys_created_on` |
| `ORDERBYDESC` | Sort descending | `ORDERBYDESCpriority` |

**Example**: Find high-priority open incidents: `state=1^priority<=2^ORDERBYDESCsys_created_on`

## Parameter Formats

### Record Identifiers

Most tools accept multiple identifier formats:

| Type | Number Format | sys_id Format |
|------|---------------|---------------|
| Incident | `INC0010001` | `a1b2c3d4e5f6g7h8i9j0k1l2m3n4o5p6` |
| Change Request | `CHG0010001` | 32-char hex string |
| Knowledge Article | `KB0010001` | 32-char hex string |
| User | `admin` (username) or `admin@example.com` (email) | 32-char hex string |

### State Values

Different record types use different state codes:

**Incidents**:
- `1` = New
- `2` = In Progress
- `3` = On Hold
- `6` = Resolved
- `7` = Closed
- `8` = Canceled

**Change Requests**:
- `-5` = New
- `-4` = Assess
- `-3` = Authorize
- `-2` = Scheduled
- `-1` = Implement
- `0` = Review
- `3` = Closed
- `4` = Canceled

### Priority and Impact Values

| Value | Priority | Impact/Urgency |
|-------|----------|----------------|
| `1` | Critical | High |
| `2` | High | Medium |
| `3` | Moderate | Low |
| `4` | Low | - |
| `5` | Planning | - |

### Date/Time Format

Use ISO 8601 format: `YYYY-MM-DD HH:MM:SS`

Example: `2024-12-15 14:30:00`

## Tool Reference

### Incident Management

| Tool | Description | Key Parameters |
|------|-------------|----------------|
| `list_incidents` | List incidents with filtering | `limit`, `state`, `assigned_to`, `category`, `query` |
| `get_incident` | Get incident details | `incident_id` (number or sys_id) |
| `create_incident` | Create new incident | `short_description` (required), `priority`, `category` |
| `update_incident` | Update existing incident | `incident_id`, fields to update |
| `add_incident_comment` | Add comment/work note | `incident_id`, `comment`, `is_work_note` |
| `resolve_incident` | Resolve an incident | `incident_id`, `resolution_code`, `resolution_notes` |

### Change Management

| Tool | Description | Key Parameters |
|------|-------------|----------------|
| `list_change_requests` | List changes with filtering | `limit`, `state`, `type`, `assigned_to` |
| `get_change_request` | Get change details | `change_id` (number or sys_id) |
| `create_change_request` | Create new change | `short_description`, `type` (normal/standard/emergency) |
| `update_change_request` | Update existing change | `change_id`, fields to update |
| `add_change_task` | Add task to change | `change_id`, `short_description` |
| `submit_change_for_approval` | Submit for approval | `change_id` |
| `approve_change` | Approve pending change | `change_id`, `comments` |
| `reject_change` | Reject pending change | `change_id`, `reason` |

### Service Catalog

| Tool | Description | Key Parameters |
|------|-------------|----------------|
| `list_catalogs` | List service catalogs | `limit` |
| `list_catalog_items` | List orderable items | `limit`, `category`, `query` |
| `get_catalog_item` | Get item details | `item_id` |
| `list_catalog_categories` | List categories | `catalog_id`, `parent_id` |
| `list_catalog_item_variables` | List form variables | `item_id` |
| `create_catalog_category` | Create category | `title`, `catalog_id` |
| `update_catalog_category` | Update category | `category_id`, fields to update |
| `update_catalog_item` | Update item | `item_id`, fields to update |
| `create_catalog_item_variable` | Create form field | `item_id`, `name`, `question_text`, `type` |
| `move_catalog_items` | Move items to category | `item_ids`, `target_category_id` |

### Knowledge Base

| Tool | Description | Key Parameters |
|------|-------------|----------------|
| `list_knowledge_bases` | List knowledge bases | `limit`, `active` |
| `list_knowledge_articles` | List articles | `limit`, `knowledge_base`, `category`, `query` |
| `get_knowledge_article` | Get article details | `article_id` (number or sys_id) |
| `list_kb_categories` | List KB categories | `knowledge_base`, `parent` |
| `create_knowledge_base` | Create KB | `title`, `description` |
| `create_kb_category` | Create category | `label`, `knowledge_base` |
| `create_knowledge_article` | Create article | `short_description`, `text`, `knowledge_base` |
| `update_knowledge_article` | Update article | `article_id`, fields to update |
| `publish_knowledge_article` | Publish article | `article_id` |

### Users and Groups

| Tool | Description | Key Parameters |
|------|-------------|----------------|
| `list_users` | List users with filtering | `limit`, `active`, `department`, `query` |
| `get_user` | Get user details | `user_id` (sys_id, username, or email) |
| `list_groups` | List groups | `limit`, `active`, `query` |
| `create_user` | Create user | `user_name`, `first_name`, `last_name`, `email` |
| `update_user` | Update user | `user_id`, fields to update |
| `create_group` | Create group | `name`, `description` |
| `update_group` | Update group | `group_id`, fields to update |
| `add_group_members` | Add users to group | `group_id`, `user_ids` |
| `remove_group_members` | Remove users from group | `group_id`, `user_ids` |

### Workflows

| Tool | Description | Key Parameters |
|------|-------------|----------------|
| `list_workflows` | List workflows | `limit`, `active`, `table` |
| `get_workflow` | Get workflow details | `workflow_id` (sys_id or name) |
| `create_workflow` | Create workflow | `name`, `table` |
| `update_workflow` | Update workflow | `workflow_id`, fields to update |
| `delete_workflow` | Delete workflow | `workflow_id` |

### Script Includes

| Tool | Description | Key Parameters |
|------|-------------|----------------|
| `list_script_includes` | List script includes | `limit`, `active`, `query` |
| `get_script_include` | Get script details | `script_id` (sys_id or name) |
| `create_script_include` | Create script | `name`, `api_name`, `script` |
| `update_script_include` | Update script | `script_id`, fields to update |
| `delete_script_include` | Delete script | `script_id` |

### Changesets (Update Sets)

| Tool | Description | Key Parameters |
|------|-------------|----------------|
| `list_changesets` | List update sets | `limit`, `state`, `created_by` |
| `get_changeset` | Get changeset details | `changeset_id` (sys_id or name) |
| `create_changeset` | Create update set | `name`, `description` |
| `update_changeset` | Update changeset | `changeset_id`, fields to update |
| `commit_changeset` | Mark as complete | `changeset_id` |

### Agile Development

| Tool | Description | Key Parameters |
|------|-------------|----------------|
| `list_stories` | List user stories | `limit`, `state`, `sprint`, `assigned_to` |
| `list_epics` | List epics | `limit`, `state`, `product` |
| `list_scrum_tasks` | List scrum tasks | `limit`, `story`, `state`, `assigned_to` |
| `list_projects` | List projects | `limit`, `state`, `active` |
| `create_story` | Create user story | `short_description`, `story_points`, `sprint` |
| `update_story` | Update story | `story_id`, fields to update |
| `create_epic` | Create epic | `short_description`, `product` |
| `update_epic` | Update epic | `epic_id`, fields to update |
| `create_scrum_task` | Create task | `short_description`, `story`, `type` |
| `update_scrum_task` | Update task | `task_id`, fields to update |
| `create_project` | Create project | `short_description`, `start_date`, `end_date` |
| `update_project` | Update project | `project_id`, fields to update |

## Common Workflows

### Incident Lifecycle

1. **Create incident**: `create_incident` with `short_description` and `category`
2. **Assign to group/user**: `update_incident` with `assignment_group` or `assigned_to`
3. **Add work notes**: `add_incident_comment` with `is_work_note: true`
4. **Update progress**: `update_incident` with `state: "2"` (In Progress)
5. **Resolve**: `resolve_incident` with `resolution_code` and `resolution_notes`

### Change Request Process

1. **Create change**: `create_change_request` with `type` (normal/standard/emergency)
2. **Add tasks**: `add_change_task` for each implementation step
3. **Submit for approval**: `submit_change_for_approval`
4. **Approve/Reject**: `approve_change` or `reject_change`
5. **Track progress**: `update_change_request` with state updates

### Knowledge Article Publishing

1. **Create article**: `create_knowledge_article` (created in draft state)
2. **Update content**: `update_knowledge_article` to refine
3. **Publish**: `publish_knowledge_article` to make visible

### User Onboarding

1. **Create user**: `create_user` with required fields
2. **Find groups**: `list_groups` to identify relevant groups
3. **Add to groups**: `add_group_members` for each group

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

## Error Handling

Common errors and solutions:

| Error | Cause | Solution |
|-------|-------|----------|
| "Rate limit exceeded" | Too many requests | Wait 20 seconds, reduce request frequency |
| "Record not found" | Invalid ID | Verify the record number or sys_id exists |
| "Write operation blocked" | Read-only mode enabled | Remove `--read-only` flag or `READ_ONLY_MODE=true` |
| "Authentication failed" | Invalid credentials | Check username/password or token validity |
| "Access denied" | Insufficient permissions | Ensure user has required ServiceNow roles |

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
