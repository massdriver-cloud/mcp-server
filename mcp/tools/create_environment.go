package tools

import (
	"context"
	"fmt"

	mdclient "github.com/massdriver-cloud/massdriver-sdk-go/massdriver/client"
	"github.com/massdriver-cloud/mcp-server/internal/api"
	mcpsdk "github.com/modelcontextprotocol/go-sdk/mcp"
)

// CreateEnvironmentTool is the MCP tool descriptor for create_environment.
var CreateEnvironmentTool = &mcpsdk.Tool{
	Name:        "create_environment",
	Description: "Creates a new environment within a project.",
}

// CreateEnvironmentInput holds the input for create_environment.
type CreateEnvironmentInput struct {
	ProjectID   string `json:"project_id"   jsonschema:"The ID of the project to create the environment in."`
	ID          string `json:"id"           jsonschema:"Unique identifier for the environment within the project, max 20 lowercase alphanumeric characters. Cannot be changed after creation."`
	Name        string `json:"name"         jsonschema:"Human-readable name shown in the UI."`
	Description string `json:"description"  jsonschema:"Optional description of the environment."`
}

// HandleCreateEnvironment returns the handler for the create_environment tool.
func HandleCreateEnvironment(c *mdclient.Client) func(context.Context, *mcpsdk.CallToolRequest, CreateEnvironmentInput) (*mcpsdk.CallToolResult, any, error) {
	return func(ctx context.Context, _ *mcpsdk.CallToolRequest, args CreateEnvironmentInput) (*mcpsdk.CallToolResult, any, error) {
		if args.ProjectID == "" {
			return nil, nil, fmt.Errorf("create_environment: project_id is required")
		}
		if args.ID == "" {
			return nil, nil, fmt.Errorf("create_environment: id is required")
		}
		if args.Name == "" {
			return nil, nil, fmt.Errorf("create_environment: name is required")
		}

		input := api.CreateEnvironmentInput{
			Id:          args.ID,
			Name:        args.Name,
			Description: args.Description,
		}

		payload, err := api.CreateEnvironment(ctx, c, args.ProjectID, input)
		if err != nil {
			return nil, nil, fmt.Errorf("create_environment: %w", err)
		}

		if !payload.Successful {
			msgs := payloadMessages(payload.Messages)
			return textResult(fmt.Sprintf("create_environment failed: %s", msgs)), payload, nil
		}

		result, err := jsonResult(payload.Result)
		if err != nil {
			return nil, nil, err
		}
		return result, payload, nil
	}
}
