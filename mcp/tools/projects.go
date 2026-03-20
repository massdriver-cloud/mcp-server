package tools

import (
	"context"
	"fmt"

	mdclient "github.com/massdriver-cloud/massdriver-sdk-go/massdriver/client"
	"github.com/massdriver-cloud/mcp-server/internal/api"
	mcpsdk "github.com/modelcontextprotocol/go-sdk/mcp"
)

var ListProjectsTool = &mcpsdk.Tool{
	Name:        "list_projects",
	Description: "Lists all projects in the Massdriver organization.",
}

type ListProjectsInput struct{}

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

var GetProjectTool = &mcpsdk.Tool{
	Name:        "get_project",
	Description: "Gets a specific project by ID, including its environments.",
}

type GetProjectInput struct {
	ID string `json:"id" jsonschema:"The project ID (e.g., 'myproj')."`
}

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

var CreateProjectTool = &mcpsdk.Tool{
	Name:        "create_project",
	Description: "Creates a new project in the Massdriver organization.",
}

type CreateProjectInput struct {
	ID          string `json:"id"          jsonschema:"Unique identifier for the project, max 12 lowercase alphanumeric characters. Cannot be changed after creation."`
	Name        string `json:"name"        jsonschema:"Human-readable name shown in the UI."`
	Description string `json:"description" jsonschema:"Optional description of the project."`
}

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

var UpdateProjectTool = &mcpsdk.Tool{
	Name:        "update_project",
	Description: "Updates a project's name or description.",
}

type UpdateProjectInput struct {
	ID          string `json:"id"          jsonschema:"The project ID to update."`
	Name        string `json:"name"        jsonschema:"Optional. New human-readable name for the project."`
	Description string `json:"description" jsonschema:"Optional. New description for the project."`
}

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

var DeleteProjectTool = &mcpsdk.Tool{
	Name:        "delete_project",
	Description: "Deletes a project. All environments in the project must be empty before deletion.",
}

type DeleteProjectInput struct {
	ID string `json:"id" jsonschema:"The project ID to delete."`
}

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
