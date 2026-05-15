# Massdriver MCP Server

A [Model Context Protocol](https://modelcontextprotocol.io) (MCP) server for [Massdriver](https://massdriver.cloud), providing AI assistants with tools to manage your infrastructure platform — projects, environments, deployments, policies, and more.

## Quick Start

### Prerequisites

- Go 1.24+
- A Massdriver account with API credentials

### Configuration

The server reads credentials from environment variables or `~/.config/massdriver/config.yaml`:

```bash
export MASSDRIVER_API_KEY="your-api-key"
export MASSDRIVER_ORGANIZATION_ID="your-org-id"
# Optional:
export MASSDRIVER_URL="https://api.massdriver.cloud"
```

### Build & Run

```bash
make build
./bin/mcp-server
```

### Docker

```bash
make docker.build
docker run -e MASSDRIVER_API_KEY -e MASSDRIVER_ORGANIZATION_ID massdrivercloud/mcp-server
```

## MCP Client Configuration

### Claude Desktop

Add to `~/Library/Application Support/Claude/claude_desktop_config.json` (macOS) or `%APPDATA%\Claude\claude_desktop_config.json` (Windows):

```json
{
  "mcpServers": {
    "massdriver": {
      "command": "/path/to/mcp-server",
      "env": {
        "MASSDRIVER_API_KEY": "your-api-key",
        "MASSDRIVER_ORGANIZATION_ID": "your-org-id"
      }
    }
  }
}
```

### Other Clients

The server uses stdio transport, compatible with any MCP client.

## Available Tools (83)

### Projects (5)
`list_projects` `get_project` `create_project` `update_project` `delete_project`

### Environments (7)
`list_environments` `get_environment` `create_environment` `update_environment` `delete_environment` `set_environment_default` `remove_environment_default`

### Instances (6)
`list_instances` `get_instance` `update_instance` `set_instance_secret` `remove_instance_secret` `list_alarms`

### Deployments (8)
`list_deployments` `get_deployment` `get_deployment_logs` `create_deployment` `propose_deployment` `approve_deployment` `reject_deployment` `abort_deployment`

### Components (7)
`list_components` `get_component` `add_component` `update_component` `remove_component` `link_components` `unlink_components`

### Bundles (2)
`list_bundles` `get_bundle`

### Resources (8)
`list_resources` `get_resource` `create_resource` `update_resource` `delete_resource` `export_resource` `create_resource_grant` `delete_resource_grant`

### Organization (4)
`get_organization` `create_custom_attribute` `update_custom_attribute` `delete_custom_attribute`

### Viewer (1)
`get_viewer`

### Audit Logs (3)
`get_audit_log` `list_audit_logs` `list_audit_log_event_types`

### Groups (10)
`list_groups` `get_group` `create_group` `update_group` `delete_group` `add_group_user` `remove_group_user` `revoke_group_invitation` `add_group_service_account` `remove_group_service_account`

### Service Accounts (5)
`list_service_accounts` `get_service_account` `create_service_account` `update_service_account` `delete_service_account`

### OCI Repos (4)
`list_oci_repos` `get_oci_repo` `create_oci_repo` `update_oci_repo`

### Policies (11)
`get_policy` `create_policy` `update_policy` `delete_policy` `list_policy_actions` `list_policy_entities` `evaluate_policy` `evaluate_policies_batch` `explain_policy` `get_policy_attribute_schema` `list_policy_attribute_values`

### Server (1)
`get_server`

### URLs (1)
`get_url`

## Development

```bash
make test     # Run tests
make lint     # Run linters (go vet + golangci-lint)
make build    # Build binary
make tidy     # go mod tidy
```

### Architecture

```
main.go              # Entrypoint — initializes SDK client, starts stdio server
mcp/
  server.go          # MCP server setup, tool registration, SDK wiring
  tools/
    services.go      # Service interfaces + Client struct (DI)
    helpers.go       # textResult, jsonResult, mutationErr helpers
    projects.go      # One file per service domain
    ...
```

Tool handlers depend on service interfaces defined in `services.go`. In production, `server.go` wires real SDK services. In tests, stub structs with func fields are injected directly.

### Adding a Tool

1. Add the method to the appropriate interface in `mcp/tools/services.go`
2. Add the tool definition and handler in the corresponding `mcp/tools/<service>.go`
3. Register the tool in `mcp/server.go`
4. Add stub method and tests in the `*_test.go` file
5. Add the tool name to the expected list in `main_test.go`

## License

This project follows the same license as the broader Massdriver ecosystem.
