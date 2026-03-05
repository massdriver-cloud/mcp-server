package tools

import (
	"context"
	"fmt"

	mdclient "github.com/massdriver-cloud/massdriver-sdk-go/massdriver/client"
	"github.com/massdriver-cloud/mcp-server/internal/api"
	mcpsdk "github.com/modelcontextprotocol/go-sdk/mcp"
)

// GetProjectTool is the MCP tool descriptor for get_project.
var GetProjectTool = &mcpsdk.Tool{
	Name:        "get_project",
	Description: "Gets a specific project by ID, including its environments.",
}

// GetProjectInput holds the input for get_project.
type GetProjectInput struct {
	ID string `json:"id" jsonschema:"The project ID (e.g., 'myproj')."`
}

// HandleGetProject returns the handler for the get_project tool.
func HandleGetProject(c *mdclient.Client) func(context.Context, *mcpsdk.CallToolRequest, GetProjectInput) (*mcpsdk.CallToolResult, any, error) {
	return func(ctx context.Context, _ *mcpsdk.CallToolRequest, args GetProjectInput) (*mcpsdk.CallToolResult, any, error) {
		if args.ID == "" {
			return nil, nil, fmt.Errorf("get_project: id is required")
		}

		project, err := api.GetProject(ctx, c, args.ID)
		if err != nil {
			return nil, nil, fmt.Errorf("get_project: %w", err)
		}

		result, err := jsonResult(project)
		if err != nil {
			return nil, nil, err
		}
		return result, project, nil
	}
}
