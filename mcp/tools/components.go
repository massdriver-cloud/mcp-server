package tools

import (
	"context"
	"fmt"

	"github.com/massdriver-cloud/massdriver-sdk-go/massdriver/platform/components"
	mcpsdk "github.com/modelcontextprotocol/go-sdk/mcp"
)

var ListComponentsTool = &mcpsdk.Tool{
	Name:        "list_components",
	Description: "Lists all components in a project's blueprint.",
}

type ListComponentsInput struct {
	ProjectID string `json:"project_id" jsonschema:"The project ID to list components for."`
}

func HandleListComponents(c *Client) func(context.Context, *mcpsdk.CallToolRequest, ListComponentsInput) (*mcpsdk.CallToolResult, any, error) {
	return func(ctx context.Context, _ *mcpsdk.CallToolRequest, args ListComponentsInput) (*mcpsdk.CallToolResult, any, error) {
		if args.ProjectID == "" {
			return nil, nil, fmt.Errorf("list_components: project_id is required")
		}

		list, err := c.Components.List(ctx, components.ListInput{ProjectID: args.ProjectID})
		if err != nil {
			return nil, nil, fmt.Errorf("list_components: %w", err)
		}

		result, err := jsonResult(list)
		if err != nil {
			return nil, nil, err
		}
		return result, list, nil
	}
}

var GetComponentTool = &mcpsdk.Tool{
	Name:        "get_component",
	Description: "Gets a specific component by ID, including its configuration and links.",
}

type GetComponentInput struct {
	ID string `json:"id" jsonschema:"The component ID."`
}

func HandleGetComponent(c *Client) func(context.Context, *mcpsdk.CallToolRequest, GetComponentInput) (*mcpsdk.CallToolResult, any, error) {
	return func(ctx context.Context, _ *mcpsdk.CallToolRequest, args GetComponentInput) (*mcpsdk.CallToolResult, any, error) {
		if args.ID == "" {
			return nil, nil, fmt.Errorf("get_component: id is required")
		}

		component, err := c.Components.Get(ctx, args.ID)
		if err != nil {
			return nil, nil, fmt.Errorf("get_component: %w", err)
		}

		result, err := jsonResult(component)
		if err != nil {
			return nil, nil, err
		}
		return result, component, nil
	}
}

var AddComponentTool = &mcpsdk.Tool{
	Name:        "add_component",
	Description: "Adds an infrastructure component to a project's blueprint. Each component is a specific instance of a bundle (e.g., a Redis cache or PostgreSQL database).",
}

type AddComponentInput struct {
	ProjectID   string `json:"project_id"   jsonschema:"The project ID to add the component to."`
	BundleName  string `json:"bundle_name"  jsonschema:"Name of the bundle to add (e.g., 'aws-aurora-postgres')."`
	ID          string `json:"id"           jsonschema:"Short identifier for this component, max 20 lowercase alphanumeric characters. Immutable after creation."`
	Name        string `json:"name"         jsonschema:"Display name for this component (e.g., 'Billing Database')."`
	Description string `json:"description"  jsonschema:"Optional description of this component's purpose."`
}

func HandleAddComponent(c *Client) func(context.Context, *mcpsdk.CallToolRequest, AddComponentInput) (*mcpsdk.CallToolResult, any, error) {
	return func(ctx context.Context, _ *mcpsdk.CallToolRequest, args AddComponentInput) (*mcpsdk.CallToolResult, any, error) {
		if args.ProjectID == "" {
			return nil, nil, fmt.Errorf("add_component: project_id is required")
		}
		if args.BundleName == "" {
			return nil, nil, fmt.Errorf("add_component: bundle_name is required")
		}
		if args.ID == "" {
			return nil, nil, fmt.Errorf("add_component: id is required")
		}
		if args.Name == "" {
			return nil, nil, fmt.Errorf("add_component: name is required")
		}

		component, err := c.Components.Add(ctx, args.ProjectID, components.AddInput{
			OciRepoName: args.BundleName,
			ID:          args.ID,
			Name:        args.Name,
			Description: args.Description,
		})
		if err != nil {
			if isMutationFailed(err) {
				return textResult(fmt.Sprintf("add_component failed: %s", mutationErr(err))), nil, nil
			}
			return nil, nil, fmt.Errorf("add_component: %w", err)
		}

		result, err := jsonResult(component)
		if err != nil {
			return nil, nil, err
		}
		return result, component, nil
	}
}

var UpdateComponentTool = &mcpsdk.Tool{
	Name:        "update_component",
	Description: "Updates a component's name, description, or configuration attributes.",
}

type UpdateComponentInput struct {
	ID          string `json:"id"          jsonschema:"The component ID to update."`
	Name        string `json:"name"        jsonschema:"Optional. New display name."`
	Description string `json:"description" jsonschema:"Optional. New description."`
}

func HandleUpdateComponent(c *Client) func(context.Context, *mcpsdk.CallToolRequest, UpdateComponentInput) (*mcpsdk.CallToolResult, any, error) {
	return func(ctx context.Context, _ *mcpsdk.CallToolRequest, args UpdateComponentInput) (*mcpsdk.CallToolResult, any, error) {
		if args.ID == "" {
			return nil, nil, fmt.Errorf("update_component: id is required")
		}

		component, err := c.Components.Update(ctx, args.ID, components.UpdateInput{
			Name:        args.Name,
			Description: args.Description,
		})
		if err != nil {
			if isMutationFailed(err) {
				return textResult(fmt.Sprintf("update_component failed: %s", mutationErr(err))), nil, nil
			}
			return nil, nil, fmt.Errorf("update_component: %w", err)
		}

		result, err := jsonResult(component)
		if err != nil {
			return nil, nil, err
		}
		return result, component, nil
	}
}

var RemoveComponentTool = &mcpsdk.Tool{
	Name:        "remove_component",
	Description: "Removes a component from a project's blueprint.",
}

type RemoveComponentInput struct {
	ID string `json:"id" jsonschema:"The component ID to remove."`
}

func HandleRemoveComponent(c *Client) func(context.Context, *mcpsdk.CallToolRequest, RemoveComponentInput) (*mcpsdk.CallToolResult, any, error) {
	return func(ctx context.Context, _ *mcpsdk.CallToolRequest, args RemoveComponentInput) (*mcpsdk.CallToolResult, any, error) {
		if args.ID == "" {
			return nil, nil, fmt.Errorf("remove_component: id is required")
		}

		_, err := c.Components.Remove(ctx, args.ID)
		if err != nil {
			if isMutationFailed(err) {
				return textResult(fmt.Sprintf("remove_component failed: %s", mutationErr(err))), nil, nil
			}
			return nil, nil, fmt.Errorf("remove_component: %w", err)
		}

		return textResult(fmt.Sprintf("component %q removed successfully", args.ID)), nil, nil
	}
}

var LinkComponentsTool = &mcpsdk.Tool{
	Name:        "link_components",
	Description: "Creates a link between two components in a project's blueprint, connecting an output field on the source to an input field on the destination.",
}

type LinkComponentsInput struct {
	From        string `json:"from"         jsonschema:"ID of the source component that produces the artifact."`
	FromField   string `json:"from_field"   jsonschema:"Output field name on the source component."`
	FromVersion string `json:"from_version" jsonschema:"Version constraint for the source component (e.g., '~1.0', 'latest')."`
	To          string `json:"to"           jsonschema:"ID of the destination component that consumes the artifact."`
	ToField     string `json:"to_field"     jsonschema:"Input field name on the destination component."`
	ToVersion   string `json:"to_version"   jsonschema:"Version constraint for the destination component (e.g., '~1.0', 'latest')."`
}

func HandleLinkComponents(c *Client) func(context.Context, *mcpsdk.CallToolRequest, LinkComponentsInput) (*mcpsdk.CallToolResult, any, error) {
	return func(ctx context.Context, _ *mcpsdk.CallToolRequest, args LinkComponentsInput) (*mcpsdk.CallToolResult, any, error) {
		if args.From == "" {
			return nil, nil, fmt.Errorf("link_components: from is required")
		}
		if args.FromField == "" {
			return nil, nil, fmt.Errorf("link_components: from_field is required")
		}
		if args.FromVersion == "" {
			return nil, nil, fmt.Errorf("link_components: from_version is required")
		}
		if args.To == "" {
			return nil, nil, fmt.Errorf("link_components: to is required")
		}
		if args.ToField == "" {
			return nil, nil, fmt.Errorf("link_components: to_field is required")
		}
		if args.ToVersion == "" {
			return nil, nil, fmt.Errorf("link_components: to_version is required")
		}

		link, err := c.Components.AddLink(ctx, components.AddLinkInput{
			FromComponentID: args.From,
			FromField:       args.FromField,
			FromVersion:     args.FromVersion,
			ToComponentID:   args.To,
			ToField:         args.ToField,
			ToVersion:       args.ToVersion,
		})
		if err != nil {
			if isMutationFailed(err) {
				return textResult(fmt.Sprintf("link_components failed: %s", mutationErr(err))), nil, nil
			}
			return nil, nil, fmt.Errorf("link_components: %w", err)
		}

		result, err := jsonResult(link)
		if err != nil {
			return nil, nil, err
		}
		return result, link, nil
	}
}

var UnlinkComponentsTool = &mcpsdk.Tool{
	Name:        "unlink_components",
	Description: "Removes a link between two components in a project's blueprint.",
}

type UnlinkComponentsInput struct {
	ID string `json:"id" jsonschema:"The link ID to remove."`
}

func HandleUnlinkComponents(c *Client) func(context.Context, *mcpsdk.CallToolRequest, UnlinkComponentsInput) (*mcpsdk.CallToolResult, any, error) {
	return func(ctx context.Context, _ *mcpsdk.CallToolRequest, args UnlinkComponentsInput) (*mcpsdk.CallToolResult, any, error) {
		if args.ID == "" {
			return nil, nil, fmt.Errorf("unlink_components: id is required")
		}

		_, err := c.Components.RemoveLink(ctx, args.ID)
		if err != nil {
			if isMutationFailed(err) {
				return textResult(fmt.Sprintf("unlink_components failed: %s", mutationErr(err))), nil, nil
			}
			return nil, nil, fmt.Errorf("unlink_components: %w", err)
		}

		return textResult(fmt.Sprintf("link %q removed successfully", args.ID)), nil, nil
	}
}
