package tools

import (
	"context"
	"fmt"

	mdclient "github.com/massdriver-cloud/massdriver-sdk-go/massdriver/client"
	"github.com/massdriver-cloud/mcp-server/internal/api"
	mcpsdk "github.com/modelcontextprotocol/go-sdk/mcp"
)

// ListEnvironmentsTool is the MCP tool descriptor for list_environments.
var ListEnvironmentsTool = &mcpsdk.Tool{
	Name:        "list_environments",
	Description: "Lists all environments in the organization. Optionally filter by project ID.",
}

// ListEnvironmentsInput holds the input for list_environments.
type ListEnvironmentsInput struct {
	ProjectID string `json:"project_id" jsonschema:"Optional. Filter to environments belonging to this project ID. Leave empty to list all environments across all projects."`
}

// HandleListEnvironments returns the handler for the list_environments tool.
func HandleListEnvironments(c *mdclient.Client) func(context.Context, *mcpsdk.CallToolRequest, ListEnvironmentsInput) (*mcpsdk.CallToolResult, any, error) {
	return func(ctx context.Context, _ *mcpsdk.CallToolRequest, args ListEnvironmentsInput) (*mcpsdk.CallToolResult, any, error) {
		filter := api.EnvironmentsFilter{}
		if args.ProjectID != "" {
			filter.ProjectIds = []string{args.ProjectID}
		}

		envs, err := api.ListEnvironments(ctx, c, filter)
		if err != nil {
			return nil, nil, fmt.Errorf("list_environments: %w", err)
		}

		result, err := jsonResult(envs)
		if err != nil {
			return nil, nil, err
		}
		return result, envs, nil
	}
}
