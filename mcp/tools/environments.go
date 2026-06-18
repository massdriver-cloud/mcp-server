package tools

import (
	"context"
	"fmt"

	"github.com/massdriver-cloud/massdriver-sdk-go/massdriver/platform/environments"
	mcpsdk "github.com/modelcontextprotocol/go-sdk/mcp"
)

var ListEnvironmentsTool = &mcpsdk.Tool{
	Name: "list_environments",
	Description: "Lists environments in the organization, one page at a time. " +
		"PREFER filtering by `project_id` — unfiltered lists span every project. " +
		"Returns up to `page_size` environments (default 25, max 100) plus a `next_cursor` for the following page. " +
		"To continue, call again with `cursor` set to the previous `next_cursor`. " +
		"Do NOT paginate to exhaustion unless the user explicitly asked for every environment.",
}

type ListEnvironmentsInput struct {
	ProjectID string `json:"project_id,omitempty" jsonschema:"Optional. Filter to environments belonging to this project ID. Leave empty to list across all projects."`
	Cursor    string `json:"cursor,omitempty"     jsonschema:"Optional. Opaque cursor from a prior call's next_cursor. Omit for the first page."`
	PageSize  int    `json:"page_size,omitempty"  jsonschema:"Optional. Page size (1-100, default 25)."`
}

func HandleListEnvironments(c *Client) func(context.Context, *mcpsdk.CallToolRequest, ListEnvironmentsInput) (*mcpsdk.CallToolResult, any, error) {
	return func(ctx context.Context, _ *mcpsdk.CallToolRequest, args ListEnvironmentsInput) (*mcpsdk.CallToolResult, any, error) {
		page, err := c.Environments.ListPage(ctx, environments.ListInput{
			ProjectID: args.ProjectID,
			PageSize:  clampPageSize(args.PageSize),
			After:     args.Cursor,
		})
		if err != nil {
			return nil, nil, fmt.Errorf("list_environments: %w", err)
		}

		out := pageResult(page)
		result, err := jsonResult(out)
		if err != nil {
			return nil, nil, err
		}
		return result, out, nil
	}
}

var GetEnvironmentTool = &mcpsdk.Tool{
	Name:        "get_environment",
	Description: "Gets a specific environment by its full identifier (e.g., 'myproject-staging').",
}

type GetEnvironmentInput struct {
	ID string `json:"id" jsonschema:"The environment identifier, typically in the format 'project-environment' (e.g., 'myproj-staging')."`
}

func HandleGetEnvironment(c *Client) func(context.Context, *mcpsdk.CallToolRequest, GetEnvironmentInput) (*mcpsdk.CallToolResult, any, error) {
	return func(ctx context.Context, _ *mcpsdk.CallToolRequest, args GetEnvironmentInput) (*mcpsdk.CallToolResult, any, error) {
		if args.ID == "" {
			return nil, nil, fmt.Errorf("get_environment: id is required")
		}

		env, err := c.Environments.Get(ctx, args.ID)
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
	ProjectID   string         `json:"project_id"   jsonschema:"The ID of the project to create the environment in."`
	ID          string         `json:"id"           jsonschema:"Unique identifier for the environment within the project, max 20 lowercase alphanumeric characters. Cannot be changed after creation."`
	Name        string         `json:"name"                  jsonschema:"Human-readable name shown in the UI."`
	Description string         `json:"description,omitempty" jsonschema:"Optional description of the environment."`
	Attributes  map[string]any `json:"attributes,omitempty" jsonschema:"Optional. Custom attribute tags at the environment scope (e.g., {\"env\":\"prod\"}). Must conform to the organization's custom-attribute schema; some may be required."`
}

func HandleCreateEnvironment(c *Client) func(context.Context, *mcpsdk.CallToolRequest, CreateEnvironmentInput) (*mcpsdk.CallToolResult, any, error) {
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

		env, err := c.Environments.Create(ctx, args.ProjectID, environments.CreateInput{
			ID:          args.ID,
			Name:        args.Name,
			Description: args.Description,
			Attributes:  args.Attributes,
		})
		if err != nil {
			if isMutationFailed(err) {
				return errorResult(fmt.Sprintf("create_environment failed: %s", mutationErr(err))), nil, nil
			}
			return nil, nil, fmt.Errorf("create_environment: %w", err)
		}

		result, err := jsonResult(env)
		if err != nil {
			return nil, nil, err
		}
		return result, env, nil
	}
}

var UpdateEnvironmentTool = &mcpsdk.Tool{
	Name:        "update_environment",
	Description: "Updates an environment's name or description.",
}

type UpdateEnvironmentInput struct {
	ID          string         `json:"id"                    jsonschema:"The environment identifier (e.g., 'myproj-staging')."`
	Name        string         `json:"name,omitempty"        jsonschema:"Optional. New human-readable name for the environment."`
	Description string         `json:"description,omitempty" jsonschema:"Optional. New description for the environment."`
	Attributes  map[string]any `json:"attributes,omitempty" jsonschema:"Optional. Replacement custom attribute tags at the environment scope. Must conform to the organization's custom-attribute schema."`
}

func HandleUpdateEnvironment(c *Client) func(context.Context, *mcpsdk.CallToolRequest, UpdateEnvironmentInput) (*mcpsdk.CallToolResult, any, error) {
	return func(ctx context.Context, _ *mcpsdk.CallToolRequest, args UpdateEnvironmentInput) (*mcpsdk.CallToolResult, any, error) {
		if args.ID == "" {
			return nil, nil, fmt.Errorf("update_environment: id is required")
		}

		env, err := c.Environments.Update(ctx, args.ID, environments.UpdateInput{
			Name:        args.Name,
			Description: args.Description,
			Attributes:  args.Attributes,
		})
		if err != nil {
			if isMutationFailed(err) {
				return errorResult(fmt.Sprintf("update_environment failed: %s", mutationErr(err))), nil, nil
			}
			return nil, nil, fmt.Errorf("update_environment: %w", err)
		}

		result, err := jsonResult(env)
		if err != nil {
			return nil, nil, err
		}
		return result, env, nil
	}
}

var DeleteEnvironmentTool = &mcpsdk.Tool{
	Name:        "delete_environment",
	Description: "Deletes an environment. All instances in the environment must be decommissioned before deletion.",
}

type DeleteEnvironmentInput struct {
	ID string `json:"id" jsonschema:"The environment identifier to delete (e.g., 'myproj-staging')."`
}

func HandleDeleteEnvironment(c *Client) func(context.Context, *mcpsdk.CallToolRequest, DeleteEnvironmentInput) (*mcpsdk.CallToolResult, any, error) {
	return func(ctx context.Context, _ *mcpsdk.CallToolRequest, args DeleteEnvironmentInput) (*mcpsdk.CallToolResult, any, error) {
		if args.ID == "" {
			return nil, nil, fmt.Errorf("delete_environment: id is required")
		}

		_, err := c.Environments.Delete(ctx, args.ID)
		if err != nil {
			if isMutationFailed(err) {
				return errorResult(fmt.Sprintf("delete_environment failed: %s", mutationErr(err))), nil, nil
			}
			return nil, nil, fmt.Errorf("delete_environment: %w", err)
		}

		return textResult(fmt.Sprintf("environment %q deleted successfully", args.ID)), nil, nil
	}
}

var SetEnvironmentDefaultTool = &mcpsdk.Tool{
	Name:        "set_environment_default",
	Description: "Sets a default resource binding for an environment.",
}

type SetEnvironmentDefaultInput struct {
	EnvironmentID string `json:"environment_id" jsonschema:"The environment ID to set the default on."`
	ResourceID    string `json:"resource_id"    jsonschema:"The resource ID to bind as the default."`
}

func HandleSetEnvironmentDefault(c *Client) func(context.Context, *mcpsdk.CallToolRequest, SetEnvironmentDefaultInput) (*mcpsdk.CallToolResult, any, error) {
	return func(ctx context.Context, _ *mcpsdk.CallToolRequest, args SetEnvironmentDefaultInput) (*mcpsdk.CallToolResult, any, error) {
		if args.EnvironmentID == "" {
			return nil, nil, fmt.Errorf("set_environment_default: environment_id is required")
		}
		if args.ResourceID == "" {
			return nil, nil, fmt.Errorf("set_environment_default: resource_id is required")
		}

		envDefault, err := c.Environments.SetDefault(ctx, args.EnvironmentID, args.ResourceID)
		if err != nil {
			if isMutationFailed(err) {
				return errorResult(fmt.Sprintf("set_environment_default failed: %s", mutationErr(err))), nil, nil
			}
			return nil, nil, fmt.Errorf("set_environment_default: %w", err)
		}

		result, err := jsonResult(envDefault)
		if err != nil {
			return nil, nil, err
		}
		return result, envDefault, nil
	}
}

var RemoveEnvironmentDefaultTool = &mcpsdk.Tool{
	Name:        "remove_environment_default",
	Description: "Removes a default resource binding from an environment.",
}

type RemoveEnvironmentDefaultInput struct {
	ID string `json:"id" jsonschema:"The environment default ID to remove."`
}

func HandleRemoveEnvironmentDefault(c *Client) func(context.Context, *mcpsdk.CallToolRequest, RemoveEnvironmentDefaultInput) (*mcpsdk.CallToolResult, any, error) {
	return func(ctx context.Context, _ *mcpsdk.CallToolRequest, args RemoveEnvironmentDefaultInput) (*mcpsdk.CallToolResult, any, error) {
		if args.ID == "" {
			return nil, nil, fmt.Errorf("remove_environment_default: id is required")
		}

		envDefault, err := c.Environments.RemoveDefault(ctx, args.ID)
		if err != nil {
			if isMutationFailed(err) {
				return errorResult(fmt.Sprintf("remove_environment_default failed: %s", mutationErr(err))), nil, nil
			}
			return nil, nil, fmt.Errorf("remove_environment_default: %w", err)
		}

		result, err := jsonResult(envDefault)
		if err != nil {
			return nil, nil, err
		}
		return result, envDefault, nil
	}
}
