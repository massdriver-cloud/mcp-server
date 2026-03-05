package tools

import (
	"context"
	"fmt"

	mdclient "github.com/massdriver-cloud/massdriver-sdk-go/massdriver/client"
	"github.com/massdriver-cloud/mcp-server/internal/api"
	mcpsdk "github.com/modelcontextprotocol/go-sdk/mcp"
)

// CreateProjectTool is the MCP tool descriptor for create_project.
var CreateProjectTool = &mcpsdk.Tool{
	Name:        "create_project",
	Description: "Creates a new project in the Massdriver organization.",
}

// CreateProjectInput holds the input for create_project.
type CreateProjectInput struct {
	ID          string `json:"id"          jsonschema:"Unique identifier for the project, max 12 lowercase alphanumeric characters. Cannot be changed after creation."`
	Name        string `json:"name"        jsonschema:"Human-readable name shown in the UI."`
	Description string `json:"description" jsonschema:"Optional description of the project."`
}

// HandleCreateProject returns the handler for the create_project tool.
func HandleCreateProject(c *mdclient.Client) func(context.Context, *mcpsdk.CallToolRequest, CreateProjectInput) (*mcpsdk.CallToolResult, any, error) {
	return func(ctx context.Context, _ *mcpsdk.CallToolRequest, args CreateProjectInput) (*mcpsdk.CallToolResult, any, error) {
		if args.ID == "" {
			return nil, nil, fmt.Errorf("create_project: id is required")
		}
		if args.Name == "" {
			return nil, nil, fmt.Errorf("create_project: name is required")
		}

		input := api.CreateProjectInput{
			Id:          args.ID,
			Name:        args.Name,
			Description: args.Description,
		}

		payload, err := api.CreateProject(ctx, c, input)
		if err != nil {
			return nil, nil, fmt.Errorf("create_project: %w", err)
		}

		if !payload.Successful {
			msgs := payloadMessages(payload.Messages)
			return textResult(fmt.Sprintf("create_project failed: %s", msgs)), payload, nil
		}

		result, err := jsonResult(payload.Result)
		if err != nil {
			return nil, nil, err
		}
		return result, payload, nil
	}
}
