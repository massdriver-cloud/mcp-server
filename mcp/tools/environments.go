package tools

import (
	"context"
	"fmt"

	mdclient "github.com/massdriver-cloud/massdriver-sdk-go/massdriver/client"
	"github.com/massdriver-cloud/mcp-server/internal/api"
	mcpsdk "github.com/modelcontextprotocol/go-sdk/mcp"
)

var ListEnvironmentsTool = &mcpsdk.Tool{
	Name:        "list_environments",
	Description: "Lists all environments in the organization. Optionally filter by project ID.",
}

type ListEnvironmentsInput struct {
	ProjectID string `json:"project_id" jsonschema:"Optional. Filter to environments belonging to this project ID. Leave empty to list all environments across all projects."`
}

func HandleListEnvironments(c *mdclient.Client) func(context.Context, *mcpsdk.CallToolRequest, ListEnvironmentsInput) (*mcpsdk.CallToolResult, any, error) {
	return func(ctx context.Context, _ *mcpsdk.CallToolRequest, args ListEnvironmentsInput) (*mcpsdk.CallToolResult, any, error) {
		filter := api.EnvironmentsFilter{}
		if args.ProjectID != "" {
			filter.ProjectId = api.IdFilter{Eq: args.ProjectID}
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

var GetEnvironmentTool = &mcpsdk.Tool{
	Name:        "get_environment",
	Description: "Gets a specific environment by its full identifier (e.g., 'myproject-staging').",
}

type GetEnvironmentInput struct {
	ID string `json:"id" jsonschema:"The environment identifier, typically in the format 'project-environment' (e.g., 'myproj-staging')."`
}

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

var CreateEnvironmentTool = &mcpsdk.Tool{
	Name:        "create_environment",
	Description: "Creates a new environment within a project.",
}

type CreateEnvironmentInput struct {
	ProjectID   string `json:"project_id"   jsonschema:"The ID of the project to create the environment in."`
	ID          string `json:"id"           jsonschema:"Unique identifier for the environment within the project, max 20 lowercase alphanumeric characters. Cannot be changed after creation."`
	Name        string `json:"name"         jsonschema:"Human-readable name shown in the UI."`
	Description string `json:"description"  jsonschema:"Optional description of the environment."`
}

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

var UpdateEnvironmentTool = &mcpsdk.Tool{
	Name:        "update_environment",
	Description: "Updates an environment's name or description.",
}

type UpdateEnvironmentInput struct {
	ID          string `json:"id"          jsonschema:"The environment identifier (e.g., 'myproj-staging')."`
	Name        string `json:"name"        jsonschema:"Optional. New human-readable name for the environment."`
	Description string `json:"description" jsonschema:"Optional. New description for the environment."`
}

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

var DeleteEnvironmentTool = &mcpsdk.Tool{
	Name:        "delete_environment",
	Description: "Deletes an environment. All instances in the environment must be decommissioned before deletion.",
}

type DeleteEnvironmentInput struct {
	ID string `json:"id" jsonschema:"The environment identifier to delete (e.g., 'myproj-staging')."`
}

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
