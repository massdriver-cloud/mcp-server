# Massdriver MCP Server

A Model Context Protocol (MCP) server for Massdriver, providing AI assistants with the ability to interact with your Massdriver organization's projects, environments, and infrastructure.

## Features

This MCP server provides the following tools:

### `list_projects`
Lists all projects in your Massdriver organization, including:
- Project metadata (ID, name, slug, description)
- Environment information
- Cost data (when available)

## Setup

### Prerequisites

1. **Massdriver Account**: You need access to a Massdriver organization
2. **API Credentials**: You'll need:
   - A Massdriver API key
   - Your organization ID

### Environment Variables

Set the following environment variables:

```bash
export MASSDRIVER_API_KEY="your-api-key-here"
export MASSDRIVER_ORGANIZATION_ID="your-org-id-here"
export MASSDRIVER_URL="https://api.massdriver.cloud"  # Optional, defaults to production
```

### Building

```bash
go build -o mcp-server
```

### Usage with MCP Clients

#### Claude Desktop

Add to your Claude Desktop configuration file (`~/Library/Application Support/Claude/claude_desktop_config.json` on macOS):

```json
{
  "mcpServers": {
    "massdriver": {
      "command": "/path/to/mcp-server",
      "env": {
        "MASSDRIVER_API_KEY": "your-api-key-here",
        "MASSDRIVER_ORGANIZATION_ID": "your-org-id-here"
      }
    }
  }
}
```

#### Other MCP Clients

The server communicates over stdin/stdout using the standard MCP protocol. You can integrate it with any MCP-compatible client by running:

```bash
./mcp-server
```

## Available Tools

### list_projects

**Description**: Retrieve all projects in your Massdriver organization

**Parameters**: None (uses organization from configuration)

**Example Usage**:
```
Please list all projects in my Massdriver organization
```

**Output**: Returns structured data containing:
- Project details (ID, name, slug, description)
- Associated environments
- Cost information (daily/monthly averages when available)

## Architecture

The MCP server is organized as follows:

- `main.go` - Entry point and server initialization
- `mcp/server.go` - Main MCP server wrapper
- `mcp/config.go` - Configuration management
- `mcp/tools/` - Individual MCP tool implementations
- `src/api/` - Existing Massdriver GraphQL API integration

## Development

### Adding New Tools

To add a new MCP tool:

1. Create a new file in `mcp/tools/`
2. Implement the tool following the pattern in `list_projects.go`
3. Register the tool in `mcp/server.go` in the `registerTools()` method

### Example Tool Structure

```go
func NewTool(ctx context.Context, req *mcpsdk.CallToolRequest, args ToolInput, mdClient *client.Client) (*mcpsdk.CallToolResult, *ToolOutput, error) {
    // 1. Validate input
    // 2. Call Massdriver API via existing api package
    // 3. Transform data to MCP format
    // 4. Return structured result
}
```

## Troubleshooting

### Authentication Issues

- Verify your `MASSDRIVER_API_KEY` is correct and not expired
- Ensure `MASSDRIVER_ORGANIZATION_ID` matches your organization
- Check that you have appropriate permissions in Massdriver

### Connection Issues

- Verify the `MASSDRIVER_URL` (defaults to production if not set)
- Check your network connectivity to Massdriver's API

### MCP Client Issues

- Ensure the server binary is executable and in the correct path
- Check that environment variables are properly set in your MCP client configuration
- Review MCP client logs for detailed error messages

## License

This project follows the same license as the broader Massdriver ecosystem.