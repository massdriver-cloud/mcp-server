package tools

import (
	"context"
	"fmt"

	mdclient "github.com/massdriver-cloud/massdriver-sdk-go/massdriver/client"
	"github.com/massdriver-cloud/mcp-server/internal/api"
	mcpsdk "github.com/modelcontextprotocol/go-sdk/mcp"
)

// UpdateProjectTool is the MCP tool descriptor for update_project.
var UpdateProjectTool = &mcpsdk.Tool{
	Name:        "update_project",
	Description: "Updates a project's name or description.",
}

// UpdateProjectInput holds the input for update_project.
type UpdateProjectInput struct {
	ID          string `json:"id"          jsonschema:"The project ID to update."`
	Name        string `json:"name"        jsonschema:"Optional. New human-readable name for the project."`
	Description string `json:"description" jsonschema:"Optional. New description for the project."`
}

// HandleUpdateProject returns the handler for the update_project tool.
func HandleUpdateProject(c *mdclient.Client) func(context.Context, *mcpsdk.CallToolRequest, UpdateProjectInput) (*mcpsdk.CallToolResult, any, error) {
	return func(ctx context.Context, _ *mcpsdk.CallToolRequest, args UpdateProjectInput) (*mcpsdk.CallToolResult, any, error) {
		if args.ID == "" {
			return nil, nil, fmt.Errorf("update_project: id is required")
		}

		input := api.UpdateProjectInput{
			Name:        args.Name,
			Description: args.Description,
		}

		payload, err := api.UpdateProject(ctx, c, args.ID, input)
		if err != nil {
			return nil, nil, fmt.Errorf("update_project: %w", err)
		}

		if !payload.Successful {
			msgs := payloadMessages(payload.Messages)
			return textResult(fmt.Sprintf("update_project failed: %s", msgs)), payload, nil
		}

		result, err := jsonResult(payload.Result)
		if err != nil {
			return nil, nil, err
		}
		return result, payload, nil
	}
}
