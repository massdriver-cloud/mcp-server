package tools

import (
	"context"
	"fmt"

	mdclient "github.com/massdriver-cloud/massdriver-sdk-go/massdriver/client"
	"github.com/massdriver-cloud/mcp-server/internal/api"
	mcpsdk "github.com/modelcontextprotocol/go-sdk/mcp"
)

// DeleteEnvironmentTool is the MCP tool descriptor for delete_environment.
var DeleteEnvironmentTool = &mcpsdk.Tool{
	Name:        "delete_environment",
	Description: "Deletes an environment. All packages in the environment must be decommissioned before deletion.",
}

// DeleteEnvironmentInput holds the input for delete_environment.
type DeleteEnvironmentInput struct {
	ID string `json:"id" jsonschema:"The environment identifier to delete (e.g., 'myproj-staging')."`
}

// HandleDeleteEnvironment returns the handler for the delete_environment tool.
func HandleDeleteEnvironment(c *mdclient.Client) func(context.Context, *mcpsdk.CallToolRequest, DeleteEnvironmentInput) (*mcpsdk.CallToolResult, any, error) {
	return func(ctx context.Context, _ *mcpsdk.CallToolRequest, args DeleteEnvironmentInput) (*mcpsdk.CallToolResult, any, error) {
		if args.ID == "" {
			return nil, nil, fmt.Errorf("delete_environment: id is required")
		}

		payload, err := api.DeleteEnvironment(ctx, c, args.ID)
		if err != nil {
			return nil, nil, fmt.Errorf("delete_environment: %w", err)
		}

		if !payload.Successful {
			msgs := payloadMessages(payload.Messages)
			return textResult(fmt.Sprintf("delete_environment failed: %s", msgs)), payload, nil
		}

		return textResult(fmt.Sprintf("environment %q deleted successfully", args.ID)), payload, nil
	}
}
