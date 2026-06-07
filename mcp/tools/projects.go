package tools

import (
	"context"
	"fmt"

	"github.com/massdriver-cloud/massdriver-sdk-go/massdriver/platform/projects"
	mcpsdk "github.com/modelcontextprotocol/go-sdk/mcp"
)

var ListProjectsTool = &mcpsdk.Tool{
	Name: "list_projects",
	Description: "Lists projects in the Massdriver organization, one page at a time. " +
		"Returns up to `page_size` projects (default 25, max 100) plus a `next_cursor` for the following page. " +
		"To continue, call again with `cursor` set to the previous `next_cursor`. " +
		"Stop once you have what you need — do NOT paginate to exhaustion unless the user explicitly asked for every project.",
}

type ListProjectsInput struct {
	Cursor   string `json:"cursor,omitempty"    jsonschema:"Optional. Opaque cursor from a prior call's next_cursor. Omit for the first page."`
	PageSize int    `json:"page_size,omitempty" jsonschema:"Optional. Page size (1-100, default 25)."`
}

func HandleListProjects(c *Client) func(context.Context, *mcpsdk.CallToolRequest, ListProjectsInput) (*mcpsdk.CallToolResult, any, error) {
	return func(ctx context.Context, _ *mcpsdk.CallToolRequest, args ListProjectsInput) (*mcpsdk.CallToolResult, any, error) {
		page, err := c.Projects.ListPage(ctx, projects.ListInput{
			PageSize: clampPageSize(args.PageSize),
			After:    args.Cursor,
		})
		if err != nil {
			return nil, nil, fmt.Errorf("list_projects: %w", err)
		}

		out := pageResult(page)
		result, err := jsonResult(out)
		if err != nil {
			return nil, nil, err
		}
		return result, out, nil
	}
}

var GetProjectTool = &mcpsdk.Tool{
	Name:        "get_project",
	Description: "Gets a specific project by ID, including its environments.",
}

type GetProjectInput struct {
	ID string `json:"id" jsonschema:"The project ID (e.g., 'myproj')."`
}

func HandleGetProject(c *Client) func(context.Context, *mcpsdk.CallToolRequest, GetProjectInput) (*mcpsdk.CallToolResult, any, error) {
	return func(ctx context.Context, _ *mcpsdk.CallToolRequest, args GetProjectInput) (*mcpsdk.CallToolResult, any, error) {
		if args.ID == "" {
			return nil, nil, fmt.Errorf("get_project: id is required")
		}

		project, err := c.Projects.Get(ctx, args.ID)
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

func HandleCreateProject(c *Client) func(context.Context, *mcpsdk.CallToolRequest, CreateProjectInput) (*mcpsdk.CallToolResult, any, error) {
	return func(ctx context.Context, _ *mcpsdk.CallToolRequest, args CreateProjectInput) (*mcpsdk.CallToolResult, any, error) {
		if args.ID == "" {
			return nil, nil, fmt.Errorf("create_project: id is required")
		}
		if args.Name == "" {
			return nil, nil, fmt.Errorf("create_project: name is required")
		}

		project, err := c.Projects.Create(ctx, projects.CreateInput{
			ID:          args.ID,
			Name:        args.Name,
			Description: args.Description,
		})
		if err != nil {
			if isMutationFailed(err) {
				return textResult(fmt.Sprintf("create_project failed: %s", mutationErr(err))), nil, nil
			}
			return nil, nil, fmt.Errorf("create_project: %w", err)
		}

		result, err := jsonResult(project)
		if err != nil {
			return nil, nil, err
		}
		return result, project, nil
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

func HandleUpdateProject(c *Client) func(context.Context, *mcpsdk.CallToolRequest, UpdateProjectInput) (*mcpsdk.CallToolResult, any, error) {
	return func(ctx context.Context, _ *mcpsdk.CallToolRequest, args UpdateProjectInput) (*mcpsdk.CallToolResult, any, error) {
		if args.ID == "" {
			return nil, nil, fmt.Errorf("update_project: id is required")
		}

		project, err := c.Projects.Update(ctx, args.ID, projects.UpdateInput{
			Name:        args.Name,
			Description: args.Description,
		})
		if err != nil {
			if isMutationFailed(err) {
				return textResult(fmt.Sprintf("update_project failed: %s", mutationErr(err))), nil, nil
			}
			return nil, nil, fmt.Errorf("update_project: %w", err)
		}

		result, err := jsonResult(project)
		if err != nil {
			return nil, nil, err
		}
		return result, project, nil
	}
}

var DeleteProjectTool = &mcpsdk.Tool{
	Name:        "delete_project",
	Description: "Deletes a project. All environments in the project must be empty before deletion.",
}

type DeleteProjectInput struct {
	ID string `json:"id" jsonschema:"The project ID to delete."`
}

func HandleDeleteProject(c *Client) func(context.Context, *mcpsdk.CallToolRequest, DeleteProjectInput) (*mcpsdk.CallToolResult, any, error) {
	return func(ctx context.Context, _ *mcpsdk.CallToolRequest, args DeleteProjectInput) (*mcpsdk.CallToolResult, any, error) {
		if args.ID == "" {
			return nil, nil, fmt.Errorf("delete_project: id is required")
		}

		_, err := c.Projects.Delete(ctx, args.ID)
		if err != nil {
			if isMutationFailed(err) {
				return textResult(fmt.Sprintf("delete_project failed: %s", mutationErr(err))), nil, nil
			}
			return nil, nil, fmt.Errorf("delete_project: %w", err)
		}

		return textResult(fmt.Sprintf("project %q deleted successfully", args.ID)), nil, nil
	}
}
