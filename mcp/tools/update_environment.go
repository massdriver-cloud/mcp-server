package tools

import (
	"context"
	"fmt"

	mdclient "github.com/massdriver-cloud/massdriver-sdk-go/massdriver/client"
	"github.com/massdriver-cloud/mcp-server/internal/api"
	mcpsdk "github.com/modelcontextprotocol/go-sdk/mcp"
)

// UpdateEnvironmentTool is the MCP tool descriptor for update_environment.
var UpdateEnvironmentTool = &mcpsdk.Tool{
	Name:        "update_environment",
	Description: "Updates an environment's name or description.",
}

// UpdateEnvironmentInput holds the input for update_environment.
type UpdateEnvironmentInput struct {
	ID          string `json:"id"          jsonschema:"The environment identifier (e.g., 'myproj-staging')."`
	Name        string `json:"name"        jsonschema:"Optional. New human-readable name for the environment."`
	Description string `json:"description" jsonschema:"Optional. New description for the environment."`
}

// HandleUpdateEnvironment returns the handler for the update_environment tool.
func HandleUpdateEnvironment(c *mdclient.Client) func(context.Context, *mcpsdk.CallToolRequest, UpdateEnvironmentInput) (*mcpsdk.CallToolResult, any, error) {
	return func(ctx context.Context, _ *mcpsdk.CallToolRequest, args UpdateEnvironmentInput) (*mcpsdk.CallToolResult, any, error) {
		if args.ID == "" {
			return nil, nil, fmt.Errorf("update_environment: id is required")
		}

		input := api.UpdateEnvironmentInput{
			Name:        args.Name,
			Description: args.Description,
		}

		payload, err := api.UpdateEnvironment(ctx, c, args.ID, input)
		if err != nil {
			return nil, nil, fmt.Errorf("update_environment: %w", err)
		}

		if !payload.Successful {
			msgs := payloadMessages(payload.Messages)
			return textResult(fmt.Sprintf("update_environment failed: %s", msgs)), payload, nil
		}

		result, err := jsonResult(payload.Result)
		if err != nil {
			return nil, nil, err
		}
		return result, payload, nil
	}
}
