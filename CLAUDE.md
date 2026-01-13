# AWS Deployment Policy

**CRITICAL: All AWS infrastructure and code changes MUST be deployed via GitHub Actions pipelines.**

### Prohibited Actions
- **NEVER** use AWS CLI directly to deploy, update, or modify infrastructure
- **NEVER** use AWS SAM CLI (`sam deploy`, `sam build`, etc.) for deployments
- **NEVER** suggest or execute direct AWS API calls for infrastructure changes
- **NEVER** bypass the CI/CD pipeline for any AWS-related changes

### Required Workflow
1. All changes must be committed and pushed to the repository
2. GitHub Actions pipeline will handle all deployments
3. **ALWAYS review pipeline output** after pushing changes
4. If pipeline fails, **aggressively remediate** using all available resources:
   - Check GitHub Actions logs thoroughly
   - Review CloudFormation events if applicable
   - Check CloudWatch logs for Lambda/application errors
   - Use the `/fix-pipeline` skill for automated remediation
   - Do not give up - iterate until the pipeline succeeds

### Pipeline Failure Remediation
When a GitHub Actions pipeline fails:
1. Immediately fetch and analyze the failure logs
2. Identify the root cause from error messages
3. Make necessary code/configuration fixes
4. Commit and push the fix
5. Monitor the new pipeline run
6. Repeat until successful deployment

## ServiceNow MCP Server - LLM Usage Guide

This section provides guidance for LLMs using the ServiceNow MCP tools effectively.

### Quick Reference: Most Common Operations

**Finding records:**
- Use `list_*` tools with `query` parameter for text search
- Use `get_*` tools when you have a specific record number or sys_id
- Always check `limit` parameter - default may be too small for comprehensive searches

**Creating records:**
- Check required parameters in tool descriptions
- `short_description` is typically required for most record types
- State/status fields are usually set automatically on creation

**Updating records:**
- Always get the record first to verify it exists and get current values
- Provide only the fields you want to change
- Record identifiers (incident_id, change_id, etc.) accept both numbers and sys_ids

### Parameter Patterns

**Identifier parameters** (incident_id, change_id, user_id, etc.):
- Accept human-readable formats: `INC0010001`, `CHG0010001`, `admin@example.com`
- Accept sys_id format: 32-character hex string
- When in doubt, use the human-readable format

**State parameters**:
- Pass as strings, not integers: `"1"` not `1`
- Values are documented in tool descriptions
- Check current state before updating to avoid invalid transitions

**Limit/offset parameters**:
- Default limits are conservative (10-50 records)
- Maximum is typically 1000
- Use offset for pagination through large result sets

**Query parameters**:
- Simple text search: `query: "network issue"`
- Encoded query syntax for complex filters: `query: "priority=1^state!=7"`
- Combine operators: `^` (AND), `^OR` (OR)

### Best Practices

**When searching for records:**
```
1. Start with list_* tool with appropriate filters
2. If too many results, add more specific filters
3. If no results, broaden the search or check spelling
4. Use get_* once you have a specific record ID
```

**When modifying records:**
```
1. Retrieve current record state with get_* tool
2. Verify the record exists and is in expected state
3. Make minimal changes - only fields that need updating
4. Verify success from tool response
```

**When creating related records:**
```
1. Create parent record first (e.g., change request)
2. Capture sys_id from response
3. Use sys_id when creating child records (e.g., change tasks)
```

### Error Recovery

**"Record not found" errors:**
- Verify the record number/sys_id is correct
- Check if searching the right table (incident vs change)
- Try listing records to find similar numbers

**"Write operation blocked" errors:**
- Server is in read-only mode
- Inform user that modification requires write access

**"Rate limit exceeded" errors:**
- Wait before making additional requests
- Reduce frequency of calls
- Batch operations where possible

**"Access denied" errors:**
- User may lack required ServiceNow roles
- Check if operation requires special permissions

### ServiceNow Domain Knowledge

**Understanding record relationships:**
- Incidents are standalone support tickets
- Change requests contain change tasks (child records)
- Knowledge articles belong to knowledge bases and categories
- Users belong to groups via membership records
- Stories belong to sprints and epics

**Common ServiceNow tables:**
- `incident` - Support incidents
- `change_request` - Change management
- `change_task` - Tasks within changes
- `kb_knowledge` - Knowledge articles
- `sys_user` - Users
- `sys_user_group` - Groups
- `sc_cat_item` - Catalog items
- `rm_story` - Agile stories

**Workflow states typically progress forward:**
- Incidents: New -> In Progress -> Resolved -> Closed
- Changes: New -> Assess -> Authorize -> Scheduled -> Implement -> Review -> Closed
- Moving backwards may require special permissions

## MCP Server LLM Usability Checklist

**IMPORTANT**: This checklist must be reviewed and all items verified on every update to this repository. Any issues found must be resolved before merging changes.

### Tool Definitions

- [ ] **Clear Purpose**: Each tool has a description that clearly explains what it does and when to use it
- [ ] **No Redundant Platform Names**: Descriptions don't include unnecessary "from [Platform]" text
- [ ] **Parameter Hints**: Tool descriptions mention key parameters or capabilities
- [ ] **Use Case Guidance**: Complex tools include when-to-use hints vs similar tools
- [ ] **Consistent Naming**: All tools use snake_case naming convention
- [ ] **Action Verbs**: Tool names start with action verbs (get_, list_, create_, update_, delete_, search_)

### Parameter Documentation

- [ ] **Examples Provided**: All string parameters include format examples in descriptions
- [ ] **Format Hints**: Date/time, ID, and structured parameters document expected formats
- [ ] **Valid Values Listed**: Parameters with fixed options list valid values (e.g., "Status: 'open', 'closed', 'all'")
- [ ] **No Redundant Defaults**: Default values are in the Default field, not repeated in description text
- [ ] **Array Format Clear**: Array parameters explain expected item format
- [ ] **Object Structure Documented**: Object parameters describe expected properties

### Schema Constraints

- [ ] **Numeric Bounds**: All limit/offset/count parameters have Minimum and Maximum constraints
- [ ] **Integer Types**: Pagination and count parameters use "integer" not "number"
- [ ] **Enum Values**: Categorical parameters have Enum arrays defined in schema
- [ ] **Array Items Typed**: All array parameters have Items property with type defined
- [ ] **Object Properties**: Complex object parameters have Properties defined where structure is known
- [ ] **Pattern Validation**: ID fields have Pattern regex where format is standardized (optional)

### Tool Annotations

- [ ] **Title Set**: All tools have a human-readable Title annotation
- [ ] **ReadOnlyHint**: All get_*, list_*, search_*, describe_* tools have ReadOnlyHint: true
- [ ] **DestructiveHint**: All delete_* tools have DestructiveHint: true
- [ ] **IdempotentHint**: Safe-to-retry operations have IdempotentHint: true
- [ ] **OpenWorldHint**: Tools interacting with external systems have OpenWorldHint: true (optional)

### Token Efficiency

- [ ] **Concise Descriptions**: Tool descriptions are under 200 characters where possible
- [ ] **No Duplicate Info**: Information isn't repeated between tool and parameter descriptions
- [ ] **Abbreviated Common Terms**: Use "Max results" instead of "Maximum number of results to return"
- [ ] **Consistent Parameter Docs**: Common parameters (limit, offset, page) use identical descriptions

### Documentation

- [ ] **README Tool Reference**: README includes descriptions of what each tool does
- [ ] **Workflow Examples**: Common multi-tool workflows are documented
- [ ] **Error Handling Guide**: Common errors and recovery strategies documented
- [ ] **Parameter Patterns**: Common parameter formats (IDs, dates, queries) documented once

### Code Quality

- [ ] **Compiles Successfully**: `go build ./...` completes without errors
- [ ] **Tests Pass**: `go test ./...` completes without failures
- [ ] **No Unused Code**: No commented-out code or unused variables
- [ ] **Consistent Formatting**: Code follows Go formatting standards (`go fmt`)

### Pre-Commit Verification

Before committing changes to this repository, run:

```bash
# Verify compilation
go build ./...

# Run all tests
go test ./...

# Check formatting
go fmt ./...

# Verify tool definitions (manual review)
# Review any new or modified tools against this checklist
```

### Issue Resolution Process

If any checklist item fails:

1. **Document the Issue**: Note which item failed and in which file
2. **Fix the Issue**: Make the necessary code changes
3. **Verify the Fix**: Re-run the relevant checks
4. **Update Tests**: Add tests for new functionality if applicable
5. **Re-verify Checklist**: Ensure fix didn't break other items
