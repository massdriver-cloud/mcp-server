package tools

import (
	"context"
	"fmt"

	mdclient "github.com/massdriver-cloud/massdriver-sdk-go/massdriver/client"
	"github.com/massdriver-cloud/mcp-server/internal/api"
	mcpsdk "github.com/modelcontextprotocol/go-sdk/mcp"
)

// GetEnvironmentTool is the MCP tool descriptor for get_environment.
var GetEnvironmentTool = &mcpsdk.Tool{
	Name:        "get_environment",
	Description: "Gets a specific environment by its full identifier (e.g., 'myproject-staging').",
}

// GetEnvironmentInput holds the input for get_environment.
type GetEnvironmentInput struct {
	ID string `json:"id" jsonschema:"The environment identifier, typically in the format 'project-environment' (e.g., 'myproj-staging')."`
}

// HandleGetEnvironment returns the handler for the get_environment tool.
func HandleGetEnvironment(c *mdclient.Client) func(context.Context, *mcpsdk.CallToolRequest, GetEnvironmentInput) (*mcpsdk.CallToolResult, any, error) {
	return func(ctx context.Context, _ *mcpsdk.CallToolRequest, args GetEnvironmentInput) (*mcpsdk.CallToolResult, any, error) {
		if args.ID == "" {
			return nil, nil, fmt.Errorf("get_environment: id is required")
		}

		env, err := api.GetEnvironment(ctx, c, args.ID)
		if err != nil {
			return nil, nil, fmt.Errorf("get_environment: %w", err)
		}

		result, err := jsonResult(env)
		if err != nil {
			return nil, nil, err
		}
		return result, env, nil
	}
}
