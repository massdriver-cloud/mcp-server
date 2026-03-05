package tools

import (
	"context"
	"fmt"

	mdclient "github.com/massdriver-cloud/massdriver-sdk-go/massdriver/client"
	"github.com/massdriver-cloud/mcp-server/internal/api"
	mcpsdk "github.com/modelcontextprotocol/go-sdk/mcp"
)

// DeleteProjectTool is the MCP tool descriptor for delete_project.
var DeleteProjectTool = &mcpsdk.Tool{
	Name:        "delete_project",
	Description: "Deletes a project. All environments in the project must be empty before deletion.",
}

// DeleteProjectInput holds the input for delete_project.
type DeleteProjectInput struct {
	ID string `json:"id" jsonschema:"The project ID to delete."`
}

// HandleDeleteProject returns the handler for the delete_project tool.
func HandleDeleteProject(c *mdclient.Client) func(context.Context, *mcpsdk.CallToolRequest, DeleteProjectInput) (*mcpsdk.CallToolResult, any, error) {
	return func(ctx context.Context, _ *mcpsdk.CallToolRequest, args DeleteProjectInput) (*mcpsdk.CallToolResult, any, error) {
		if args.ID == "" {
			return nil, nil, fmt.Errorf("delete_project: id is required")
		}

		payload, err := api.DeleteProject(ctx, c, args.ID)
		if err != nil {
			return nil, nil, fmt.Errorf("delete_project: %w", err)
		}

		if !payload.Successful {
			msgs := payloadMessages(payload.Messages)
			return textResult(fmt.Sprintf("delete_project failed: %s", msgs)), payload, nil
		}

		return textResult(fmt.Sprintf("project %q deleted successfully", args.ID)), payload, nil
	}
}
