# Integration Guide

This document provides detailed instructions for integrating the ServiceNow MCP server with various AI assistants and deployment environments.

## Quick Start

The fastest way to get started depends on your use case:

| Use Case | Recommended Setup |
|----------|-------------------|
| Local Claude Desktop | Stdio mode with basic auth |
| Remote/Multi-user | HTTP mode with Docker |
| Production | HTTP mode on ECS/Kubernetes |
| Testing | Stdio mode with read-only flag |

## Claude Desktop

### Configuration

1. Locate your Claude Desktop configuration file:
   - **macOS**: `~/Library/Application Support/Claude/claude_desktop_config.json`
   - **Windows**: `%APPDATA%\Claude\claude_desktop_config.json`
   - **Linux**: `~/.config/Claude/claude_desktop_config.json`

2. Add the ServiceNow MCP server configuration:

```json
{
  "mcpServers": {
    "servicenow": {
      "command": "/path/to/go-mcp-servicenow",
      "env": {
        "SERVICENOW_INSTANCE_URL": "https://your-instance.service-now.com",
        "SERVICENOW_AUTH_TYPE": "basic",
        "SERVICENOW_USERNAME": "your-username",
        "SERVICENOW_PASSWORD": "your-password"
      }
    }
  }
}
```

3. Restart Claude Desktop

### Read-Only Mode

For safer operation, enable read-only mode:

```json
{
  "mcpServers": {
    "servicenow": {
      "command": "/path/to/go-mcp-servicenow",
      "args": ["--read-only"],
      "env": {
        "SERVICENOW_INSTANCE_URL": "https://your-instance.service-now.com",
        "SERVICENOW_AUTH_TYPE": "basic",
        "SERVICENOW_USERNAME": "your-username",
        "SERVICENOW_PASSWORD": "your-password"
      }
    }
  }
}
```

### Debugging Configuration

Enable debug logging to troubleshoot issues:

```json
{
  "mcpServers": {
    "servicenow": {
      "command": "/path/to/go-mcp-servicenow",
      "args": ["--log-level", "debug"],
      "env": {
        "SERVICENOW_INSTANCE_URL": "https://your-instance.service-now.com",
        "SERVICENOW_AUTH_TYPE": "basic",
        "SERVICENOW_USERNAME": "your-username",
        "SERVICENOW_PASSWORD": "your-password",
        "MCP_LOG_DIR": "/tmp/servicenow-mcp-logs"
      }
    }
  }
}
```

## HTTP Mode Integration

### Starting the Server

```bash
./go-mcp-servicenow --http --host 0.0.0.0 --port 3000
```

### Authentication

When `MCP_AUTH_TOKEN` is set, all requests must include the `X-MCP-Auth-Token` header:

```bash
curl -X POST http://localhost:3000/ \
  -H "Content-Type: application/json" \
  -H "X-MCP-Auth-Token: your-token" \
  -d '{"jsonrpc":"2.0","id":1,"method":"tools/list"}'
```

### Health Check

```bash
curl http://localhost:3000/health
```

Response:
```json
{"status":"healthy","server":"go-mcp-servicenow"}
```

### Example: List Incidents

```bash
curl -X POST http://localhost:3000/ \
  -H "Content-Type: application/json" \
  -H "X-MCP-Auth-Token: your-token" \
  -d '{
    "jsonrpc": "2.0",
    "id": 1,
    "method": "tools/call",
    "params": {
      "name": "list_incidents",
      "arguments": {
        "limit": 5,
        "state": "1"
      }
    }
  }'
```

### Example: Get Incident Details

```bash
curl -X POST http://localhost:3000/ \
  -H "Content-Type: application/json" \
  -H "X-MCP-Auth-Token: your-token" \
  -d '{
    "jsonrpc": "2.0",
    "id": 2,
    "method": "tools/call",
    "params": {
      "name": "get_incident",
      "arguments": {
        "incident_id": "INC0010001"
      }
    }
  }'
```

## Docker Deployment

### Build

```bash
docker build -t go-mcp-servicenow .
```

### Run

```bash
docker run -d \
  --name servicenow-mcp \
  -p 3000:3000 \
  -e SERVICENOW_INSTANCE_URL="https://your-instance.service-now.com" \
  -e SERVICENOW_AUTH_TYPE="basic" \
  -e SERVICENOW_USERNAME="admin" \
  -e SERVICENOW_PASSWORD="password" \
  -e MCP_AUTH_TOKEN="your-mcp-token" \
  go-mcp-servicenow
```

### Docker Compose

```yaml
version: '3.8'
services:
  servicenow-mcp:
    build: .
    ports:
      - "3000:3000"
    environment:
      - SERVICENOW_INSTANCE_URL=https://your-instance.service-now.com
      - SERVICENOW_AUTH_TYPE=basic
      - SERVICENOW_USERNAME=admin
      - SERVICENOW_PASSWORD=password
      - MCP_AUTH_TOKEN=your-mcp-token
      - MCP_LOG_LEVEL=info
    restart: unless-stopped
    healthcheck:
      test: ["CMD", "wget", "--no-verbose", "--tries=1", "--spider", "http://localhost:3000/health"]
      interval: 30s
      timeout: 10s
      retries: 3
```

### Docker with Read-Only Mode

```bash
docker run -d \
  --name servicenow-mcp-readonly \
  -p 3000:3000 \
  -e SERVICENOW_INSTANCE_URL="https://your-instance.service-now.com" \
  -e SERVICENOW_AUTH_TYPE="basic" \
  -e SERVICENOW_USERNAME="admin" \
  -e SERVICENOW_PASSWORD="password" \
  -e READ_ONLY_MODE="true" \
  go-mcp-servicenow
```

## AWS ECS Deployment

### Prerequisites

1. ECR repository created
2. ECS cluster ready
3. Secrets stored in AWS Secrets Manager

### Push to ECR

```bash
# Authenticate
aws ecr get-login-password --region us-east-1 | docker login --username AWS --password-stdin ACCOUNT_ID.dkr.ecr.us-east-1.amazonaws.com

# Tag and push
docker tag go-mcp-servicenow:latest ACCOUNT_ID.dkr.ecr.us-east-1.amazonaws.com/go-mcp-servicenow:latest
docker push ACCOUNT_ID.dkr.ecr.us-east-1.amazonaws.com/go-mcp-servicenow:latest
```

### Create Secrets

```bash
aws secretsmanager create-secret \
  --name go-mcp-servicenow/SERVICENOW_INSTANCE_URL \
  --secret-string "https://your-instance.service-now.com"

aws secretsmanager create-secret \
  --name go-mcp-servicenow/SERVICENOW_USERNAME \
  --secret-string "admin"

aws secretsmanager create-secret \
  --name go-mcp-servicenow/SERVICENOW_PASSWORD \
  --secret-string "password"
```

### Deploy Task Definition

1. Update `ecs-task-definition.json` with your account ID and region
2. Register the task definition:

```bash
aws ecs register-task-definition --cli-input-json file://ecs-task-definition.json
```

3. Create or update the service:

```bash
aws ecs create-service \
  --cluster your-cluster \
  --service-name go-mcp-servicenow \
  --task-definition go-mcp-servicenow \
  --desired-count 1 \
  --launch-type FARGATE \
  --network-configuration "awsvpcConfiguration={subnets=[subnet-xxx],securityGroups=[sg-xxx],assignPublicIp=ENABLED}"
```

## Kubernetes Deployment

### ConfigMap

```yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: servicenow-mcp-config
data:
  MCP_LOG_LEVEL: "info"
  READ_ONLY_MODE: "false"
```

### Secret

```yaml
apiVersion: v1
kind: Secret
metadata:
  name: servicenow-mcp-secrets
type: Opaque
stringData:
  SERVICENOW_INSTANCE_URL: "https://your-instance.service-now.com"
  SERVICENOW_AUTH_TYPE: "basic"
  SERVICENOW_USERNAME: "admin"
  SERVICENOW_PASSWORD: "password"
  MCP_AUTH_TOKEN: "your-mcp-token"
```

### Deployment

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: servicenow-mcp
spec:
  replicas: 1
  selector:
    matchLabels:
      app: servicenow-mcp
  template:
    metadata:
      labels:
        app: servicenow-mcp
    spec:
      containers:
      - name: servicenow-mcp
        image: your-registry/go-mcp-servicenow:latest
        ports:
        - containerPort: 3000
        envFrom:
        - configMapRef:
            name: servicenow-mcp-config
        - secretRef:
            name: servicenow-mcp-secrets
        livenessProbe:
          httpGet:
            path: /health
            port: 3000
          initialDelaySeconds: 5
          periodSeconds: 30
        readinessProbe:
          httpGet:
            path: /health
            port: 3000
          initialDelaySeconds: 5
          periodSeconds: 10
        resources:
          requests:
            memory: "64Mi"
            cpu: "100m"
          limits:
            memory: "256Mi"
            cpu: "500m"
```

### Service

```yaml
apiVersion: v1
kind: Service
metadata:
  name: servicenow-mcp
spec:
  selector:
    app: servicenow-mcp
  ports:
  - port: 3000
    targetPort: 3000
  type: ClusterIP
```

## MCP Protocol Reference

### Initialize

Required first call to establish the session:

```json
{
  "jsonrpc": "2.0",
  "id": 1,
  "method": "initialize",
  "params": {
    "protocolVersion": "2024-11-05",
    "capabilities": {},
    "clientInfo": {
      "name": "test-client",
      "version": "1.0.0"
    }
  }
}
```

### List Tools

Discover available tools:

```json
{
  "jsonrpc": "2.0",
  "id": 2,
  "method": "tools/list"
}
```

### Call Tool

Execute a specific tool:

```json
{
  "jsonrpc": "2.0",
  "id": 3,
  "method": "tools/call",
  "params": {
    "name": "list_incidents",
    "arguments": {
      "limit": 10,
      "state": "1"
    }
  }
}
```

### Response Format

Successful tool calls return:

```json
{
  "jsonrpc": "2.0",
  "id": 3,
  "result": {
    "content": [
      {
        "type": "text",
        "text": "{\"success\":true,\"message\":\"Found 5 incidents\",\"incidents\":[...]}"
      }
    ]
  }
}
```

Error responses include:

```json
{
  "jsonrpc": "2.0",
  "id": 3,
  "result": {
    "content": [
      {
        "type": "text",
        "text": "{\"success\":false,\"message\":\"Incident not found: INC9999999\"}"
      }
    ],
    "isError": true
  }
}
```

## Troubleshooting

### Connection Issues

1. Verify ServiceNow instance URL is correct and accessible
2. Check authentication credentials
3. Ensure network connectivity to ServiceNow
4. Verify firewall rules allow outbound HTTPS

### Authentication Failures

1. Verify `SERVICENOW_AUTH_TYPE` matches your credential type
2. For OAuth, ensure client ID/secret are valid and not expired
3. Check user has appropriate ServiceNow roles
4. Verify the ServiceNow instance allows API access

### Rate Limiting

The server implements a 5 calls per 20 seconds rate limit. If exceeded, requests return:

```json
{
  "content": [{"type": "text", "text": "Rate limit exceeded: Maximum 5 tool calls per 20 seconds. Please try again later."}],
  "isError": true
}
```

**Solutions:**
- Space out requests over time
- Batch operations where possible
- Use more specific queries to reduce total calls

### Permission Errors

ServiceNow operations require appropriate roles:

| Operation | Minimum Role |
|-----------|--------------|
| Read incidents | itil |
| Create/update incidents | itil |
| Read change requests | itil |
| Create/update changes | change_manager |
| Manage users/groups | user_admin |
| Manage knowledge | knowledge_admin |

### Logs

Check logs in the configured log directory (default: OS temp directory):

```bash
# Find logs
ls /tmp/go-mcp-servicenow/

# View latest log
tail -f /tmp/go-mcp-servicenow/go-mcp-servicenow.log

# Search for errors
grep -i error /tmp/go-mcp-servicenow/go-mcp-servicenow.log
```

### Common Error Messages

| Error | Cause | Solution |
|-------|-------|----------|
| "connection refused" | Server not running or wrong port | Verify server is running and port is correct |
| "unauthorized" | Invalid MCP auth token | Check MCP_AUTH_TOKEN value |
| "failed to authenticate" | Invalid ServiceNow credentials | Verify username/password or OAuth tokens |
| "record not found" | Invalid record ID | Use list tools to find valid IDs |
| "write operation blocked" | Read-only mode enabled | Remove --read-only flag |

### Debug Mode

Enable debug logging for detailed troubleshooting:

```bash
./go-mcp-servicenow --http --log-level debug
```

This will log:
- All incoming requests
- ServiceNow API calls
- Response details
- Error stack traces
