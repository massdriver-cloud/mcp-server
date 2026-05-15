package tools

import (
	"context"
	"fmt"

	"github.com/massdriver-cloud/massdriver-sdk-go/massdriver/platform/resources"
	"github.com/massdriver-cloud/massdriver-sdk-go/massdriver/platform/types"
	mcpsdk "github.com/modelcontextprotocol/go-sdk/mcp"
)

var ListResourcesTool = &mcpsdk.Tool{
	Name:        "list_resources",
	Description: "Lists resources (provisioned or imported). Optionally filter by origin, resource type, environment, or search term.",
}

type ListResourcesInput struct {
	Origin        string `json:"origin"         jsonschema:"Optional. Filter by origin: IMPORTED or PROVISIONED."`
	ResourceType  string `json:"resource_type"  jsonschema:"Optional. Filter by resource type."`
	EnvironmentID string `json:"environment_id" jsonschema:"Optional. Filter to resources in this environment."`
	Search        string `json:"search"         jsonschema:"Optional. Search term to filter resources."`
}

func HandleListResources(c *Client) func(context.Context, *mcpsdk.CallToolRequest, ListResourcesInput) (*mcpsdk.CallToolResult, any, error) {
	return func(ctx context.Context, _ *mcpsdk.CallToolRequest, args ListResourcesInput) (*mcpsdk.CallToolResult, any, error) {
		input := resources.ListInput{
			Origin:        resources.Origin(args.Origin),
			ResourceType:  args.ResourceType,
			EnvironmentID: args.EnvironmentID,
			Search:        args.Search,
		}

		list, err := c.Resources.List(ctx, input)
		if err != nil {
			return nil, nil, fmt.Errorf("list_resources: %w", err)
		}

		result, err := jsonResult(list)
		if err != nil {
			return nil, nil, err
		}
		return result, list, nil
	}
}

var GetResourceTool = &mcpsdk.Tool{
	Name:        "get_resource",
	Description: "Gets a specific resource by ID. Payload values are masked; use export_resource for unmasked data.",
}

type GetResourceInput struct {
	ID string `json:"id" jsonschema:"The resource ID."`
}

func HandleGetResource(c *Client) func(context.Context, *mcpsdk.CallToolRequest, GetResourceInput) (*mcpsdk.CallToolResult, any, error) {
	return func(ctx context.Context, _ *mcpsdk.CallToolRequest, args GetResourceInput) (*mcpsdk.CallToolResult, any, error) {
		if args.ID == "" {
			return nil, nil, fmt.Errorf("get_resource: id is required")
		}

		resource, err := c.Resources.Get(ctx, args.ID)
		if err != nil {
			return nil, nil, fmt.Errorf("get_resource: %w", err)
		}

		result, err := jsonResult(resource)
		if err != nil {
			return nil, nil, err
		}
		return result, resource, nil
	}
}

var ExportResourceTool = &mcpsdk.Tool{
	Name:        "export_resource",
	Description: "Exports a resource with unmasked payload data. This action is audit-logged.",
}

type ExportResourceInput struct {
	ID     string `json:"id"     jsonschema:"The resource ID to export."`
	Format string `json:"format" jsonschema:"Optional. Export format, defaults to 'json'."`
}

func HandleExportResource(c *Client) func(context.Context, *mcpsdk.CallToolRequest, ExportResourceInput) (*mcpsdk.CallToolResult, any, error) {
	return func(ctx context.Context, _ *mcpsdk.CallToolRequest, args ExportResourceInput) (*mcpsdk.CallToolResult, any, error) {
		if args.ID == "" {
			return nil, nil, fmt.Errorf("export_resource: id is required")
		}

		format := args.Format
		if format == "" {
			format = resources.FormatJSON
		}

		exported, err := c.Resources.Export(ctx, args.ID, format)
		if err != nil {
			return nil, nil, fmt.Errorf("export_resource: %w", err)
		}

		result, err := jsonResult(exported)
		if err != nil {
			return nil, nil, err
		}
		return result, exported, nil
	}
}

var CreateResourceTool = &mcpsdk.Tool{
	Name:        "create_resource",
	Description: "Imports (creates) a resource by providing its type and payload data.",
}

type CreateResourceInput struct {
	ResourceTypeID string         `json:"resource_type_id" jsonschema:"The resource type ID."`
	Name           string         `json:"name"             jsonschema:"Display name for the resource."`
	Payload        map[string]any `json:"payload"          jsonschema:"Resource payload data."`
}

func HandleCreateResource(c *Client) func(context.Context, *mcpsdk.CallToolRequest, CreateResourceInput) (*mcpsdk.CallToolResult, any, error) {
	return func(ctx context.Context, _ *mcpsdk.CallToolRequest, args CreateResourceInput) (*mcpsdk.CallToolResult, any, error) {
		if args.ResourceTypeID == "" {
			return nil, nil, fmt.Errorf("create_resource: resource_type_id is required")
		}
		if args.Name == "" {
			return nil, nil, fmt.Errorf("create_resource: name is required")
		}

		resource, err := c.Resources.Create(ctx, args.ResourceTypeID, resources.CreateInput{
			Name:    args.Name,
			Payload: args.Payload,
		})
		if err != nil {
			if isMutationFailed(err) {
				return textResult(fmt.Sprintf("create_resource failed: %s", mutationErr(err))), nil, nil
			}
			return nil, nil, fmt.Errorf("create_resource: %w", err)
		}

		result, err := jsonResult(resource)
		if err != nil {
			return nil, nil, err
		}
		return result, resource, nil
	}
}

var UpdateResourceTool = &mcpsdk.Tool{
	Name:        "update_resource",
	Description: "Updates an existing resource by ID.",
}

type UpdateResourceInput struct {
	ID      string         `json:"id"      jsonschema:"The resource ID."`
	Name    string         `json:"name"    jsonschema:"Display name for the resource."`
	Payload map[string]any `json:"payload" jsonschema:"Resource payload data."`
}

func HandleUpdateResource(c *Client) func(context.Context, *mcpsdk.CallToolRequest, UpdateResourceInput) (*mcpsdk.CallToolResult, any, error) {
	return func(ctx context.Context, _ *mcpsdk.CallToolRequest, args UpdateResourceInput) (*mcpsdk.CallToolResult, any, error) {
		if args.ID == "" {
			return nil, nil, fmt.Errorf("update_resource: id is required")
		}
		if args.Name == "" {
			return nil, nil, fmt.Errorf("update_resource: name is required")
		}

		resource, err := c.Resources.Update(ctx, args.ID, resources.UpdateInput{
			Name:    args.Name,
			Payload: args.Payload,
		})
		if err != nil {
			if isMutationFailed(err) {
				return textResult(fmt.Sprintf("update_resource failed: %s", mutationErr(err))), nil, nil
			}
			return nil, nil, fmt.Errorf("update_resource: %w", err)
		}

		result, err := jsonResult(resource)
		if err != nil {
			return nil, nil, err
		}
		return result, resource, nil
	}
}

var DeleteResourceTool = &mcpsdk.Tool{
	Name:        "delete_resource",
	Description: "Deletes a resource by ID.",
}

type DeleteResourceInput struct {
	ID string `json:"id" jsonschema:"The resource ID."`
}

func HandleDeleteResource(c *Client) func(context.Context, *mcpsdk.CallToolRequest, DeleteResourceInput) (*mcpsdk.CallToolResult, any, error) {
	return func(ctx context.Context, _ *mcpsdk.CallToolRequest, args DeleteResourceInput) (*mcpsdk.CallToolResult, any, error) {
		if args.ID == "" {
			return nil, nil, fmt.Errorf("delete_resource: id is required")
		}

		_, err := c.Resources.Delete(ctx, args.ID)
		if err != nil {
			if isMutationFailed(err) {
				return textResult(fmt.Sprintf("delete_resource failed: %s", mutationErr(err))), nil, nil
			}
			return nil, nil, fmt.Errorf("delete_resource: %w", err)
		}

		return textResult(fmt.Sprintf("resource %q deleted successfully", args.ID)), nil, nil
	}
}

var CreateResourceGrantTool = &mcpsdk.Tool{
	Name:        "create_resource_grant",
	Description: "Creates a sharing grant on a resource, optionally restricting recipients by attribute conditions.",
}

type CreateResourceGrantInput struct {
	ResourceID          string                 `json:"resource_id"          jsonschema:"The resource ID to grant access to."`
	Action              string                 `json:"action"               jsonschema:"The action to grant, e.g. resource:export."`
	RecipientConditions types.PolicyConditions `json:"recipient_conditions" jsonschema:"Optional. Attribute conditions restricting grant recipients."`
}

func HandleCreateResourceGrant(c *Client) func(context.Context, *mcpsdk.CallToolRequest, CreateResourceGrantInput) (*mcpsdk.CallToolResult, any, error) {
	return func(ctx context.Context, _ *mcpsdk.CallToolRequest, args CreateResourceGrantInput) (*mcpsdk.CallToolResult, any, error) {
		if args.ResourceID == "" {
			return nil, nil, fmt.Errorf("create_resource_grant: resource_id is required")
		}
		if args.Action == "" {
			return nil, nil, fmt.Errorf("create_resource_grant: action is required")
		}

		grant, err := c.Resources.CreateGrant(ctx, args.ResourceID, resources.CreateGrantInput{
			Action:              args.Action,
			RecipientConditions: args.RecipientConditions,
		})
		if err != nil {
			if isMutationFailed(err) {
				return textResult(fmt.Sprintf("create_resource_grant failed: %s", mutationErr(err))), nil, nil
			}
			return nil, nil, fmt.Errorf("create_resource_grant: %w", err)
		}

		result, err := jsonResult(grant)
		if err != nil {
			return nil, nil, err
		}
		return result, grant, nil
	}
}

var DeleteResourceGrantTool = &mcpsdk.Tool{
	Name:        "delete_resource_grant",
	Description: "Deletes a sharing grant by ID.",
}

type DeleteResourceGrantInput struct {
	GrantID string `json:"grant_id" jsonschema:"The grant ID to delete."`
}

func HandleDeleteResourceGrant(c *Client) func(context.Context, *mcpsdk.CallToolRequest, DeleteResourceGrantInput) (*mcpsdk.CallToolResult, any, error) {
	return func(ctx context.Context, _ *mcpsdk.CallToolRequest, args DeleteResourceGrantInput) (*mcpsdk.CallToolResult, any, error) {
		if args.GrantID == "" {
			return nil, nil, fmt.Errorf("delete_resource_grant: grant_id is required")
		}

		err := c.Resources.DeleteGrant(ctx, args.GrantID)
		if err != nil {
			if isMutationFailed(err) {
				return textResult(fmt.Sprintf("delete_resource_grant failed: %s", mutationErr(err))), nil, nil
			}
			return nil, nil, fmt.Errorf("delete_resource_grant: %w", err)
		}

		return textResult(fmt.Sprintf("grant %q deleted successfully", args.GrantID)), nil, nil
	}
}
