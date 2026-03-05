package tools

import (
	"context"
	"fmt"

	mdclient "github.com/massdriver-cloud/massdriver-sdk-go/massdriver/client"
	"github.com/massdriver-cloud/mcp-server/internal/api"
	mcpsdk "github.com/modelcontextprotocol/go-sdk/mcp"
)

// ListProjectsTool is the MCP tool descriptor for list_projects.
var ListProjectsTool = &mcpsdk.Tool{
	Name:        "list_projects",
	Description: "Lists all projects in the Massdriver organization.",
}

// ListProjectsInput holds the (empty) input for list_projects.
type ListProjectsInput struct{}

// HandleListProjects returns the handler for the list_projects tool.
func HandleListProjects(c *mdclient.Client) func(context.Context, *mcpsdk.CallToolRequest, ListProjectsInput) (*mcpsdk.CallToolResult, any, error) {
	return func(ctx context.Context, _ *mcpsdk.CallToolRequest, _ ListProjectsInput) (*mcpsdk.CallToolResult, any, error) {
		projects, err := api.ListProjects(ctx, c)
		if err != nil {
			return nil, nil, fmt.Errorf("list_projects: %w", err)
		}

		result, err := jsonResult(projects)
		if err != nil {
			return nil, nil, err
		}
		return result, projects, nil
	}
}
