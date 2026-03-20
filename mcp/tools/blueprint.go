package tools

import (
	"context"
	"fmt"

	mdclient "github.com/massdriver-cloud/massdriver-sdk-go/massdriver/client"
	"github.com/massdriver-cloud/mcp-server/internal/api"
	mcpsdk "github.com/modelcontextprotocol/go-sdk/mcp"
)

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

func HandleAddComponent(c *mdclient.Client) func(context.Context, *mcpsdk.CallToolRequest, AddComponentInput) (*mcpsdk.CallToolResult, any, error) {
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

		input := api.AddComponentInput{
			BundleName:  args.BundleName,
			Id:          args.ID,
			Name:        args.Name,
			Description: args.Description,
		}

		payload, err := api.AddComponent(ctx, c, args.ProjectID, input)
		if err != nil {
			return nil, nil, fmt.Errorf("add_component: %w", err)
		}

		if !payload.Successful {
			msgs := payloadMessages(payload.Messages)
			return textResult(fmt.Sprintf("add_component failed: %s", msgs)), payload, nil
		}

		result, err := jsonResult(payload.Result)
		if err != nil {
			return nil, nil, err
		}
		return result, payload, nil
	}
}

var RemoveComponentTool = &mcpsdk.Tool{
	Name:        "remove_component",
	Description: "Removes a component from a project's blueprint.",
}

type RemoveComponentInput struct {
	ProjectID string `json:"project_id" jsonschema:"The project ID the component belongs to."`
	ID        string `json:"id"         jsonschema:"The component ID to remove."`
}

func HandleRemoveComponent(c *mdclient.Client) func(context.Context, *mcpsdk.CallToolRequest, RemoveComponentInput) (*mcpsdk.CallToolResult, any, error) {
	return func(ctx context.Context, _ *mcpsdk.CallToolRequest, args RemoveComponentInput) (*mcpsdk.CallToolResult, any, error) {
		if args.ProjectID == "" {
			return nil, nil, fmt.Errorf("remove_component: project_id is required")
		}
		if args.ID == "" {
			return nil, nil, fmt.Errorf("remove_component: id is required")
		}

		payload, err := api.RemoveComponent(ctx, c, args.ProjectID, args.ID)
		if err != nil {
			return nil, nil, fmt.Errorf("remove_component: %w", err)
		}

		if !payload.Successful {
			msgs := payloadMessages(payload.Messages)
			return textResult(fmt.Sprintf("remove_component failed: %s", msgs)), payload, nil
		}

		return textResult(fmt.Sprintf("component %q removed successfully", args.ID)), payload, nil
	}
}

var LinkComponentsTool = &mcpsdk.Tool{
	Name:        "link_components",
	Description: "Creates a link between two components in a project's blueprint, connecting an output field on the source to an input field on the destination.",
}

type LinkComponentsInput struct {
	ProjectID   string `json:"project_id"   jsonschema:"The project ID containing both components."`
	From        string `json:"from"         jsonschema:"ID of the source component that produces the artifact."`
	FromField   string `json:"from_field"   jsonschema:"Output field name on the source component."`
	FromVersion string `json:"from_version" jsonschema:"Version constraint for the source component (e.g., '~1.0', 'latest')."`
	To          string `json:"to"           jsonschema:"ID of the destination component that consumes the artifact."`
	ToField     string `json:"to_field"     jsonschema:"Input field name on the destination component."`
	ToVersion   string `json:"to_version"   jsonschema:"Version constraint for the destination component (e.g., '~1.0', 'latest')."`
}

func HandleLinkComponents(c *mdclient.Client) func(context.Context, *mcpsdk.CallToolRequest, LinkComponentsInput) (*mcpsdk.CallToolResult, any, error) {
	return func(ctx context.Context, _ *mcpsdk.CallToolRequest, args LinkComponentsInput) (*mcpsdk.CallToolResult, any, error) {
		if args.ProjectID == "" {
			return nil, nil, fmt.Errorf("link_components: project_id is required")
		}
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

		input := api.LinkComponentsInput{
			From:        args.From,
			FromField:   args.FromField,
			FromVersion: args.FromVersion,
			To:          args.To,
			ToField:     args.ToField,
			ToVersion:   args.ToVersion,
		}

		payload, err := api.LinkComponents(ctx, c, args.ProjectID, input)
		if err != nil {
			return nil, nil, fmt.Errorf("link_components: %w", err)
		}

		if !payload.Successful {
			msgs := payloadMessages(payload.Messages)
			return textResult(fmt.Sprintf("link_components failed: %s", msgs)), payload, nil
		}

		result, err := jsonResult(payload.Result)
		if err != nil {
			return nil, nil, err
		}
		return result, payload, nil
	}
}

var UnlinkComponentsTool = &mcpsdk.Tool{
	Name:        "unlink_components",
	Description: "Removes a link between two components in a project's blueprint.",
}

type UnlinkComponentsInput struct {
	ID string `json:"id" jsonschema:"The link ID to remove."`
}

func HandleUnlinkComponents(c *mdclient.Client) func(context.Context, *mcpsdk.CallToolRequest, UnlinkComponentsInput) (*mcpsdk.CallToolResult, any, error) {
	return func(ctx context.Context, _ *mcpsdk.CallToolRequest, args UnlinkComponentsInput) (*mcpsdk.CallToolResult, any, error) {
		if args.ID == "" {
			return nil, nil, fmt.Errorf("unlink_components: id is required")
		}

		payload, err := api.UnlinkComponents(ctx, c, args.ID)
		if err != nil {
			return nil, nil, fmt.Errorf("unlink_components: %w", err)
		}

		if !payload.Successful {
			msgs := payloadMessages(payload.Messages)
			return textResult(fmt.Sprintf("unlink_components failed: %s", msgs)), payload, nil
		}

		return textResult(fmt.Sprintf("link %q removed successfully", args.ID)), payload, nil
	}
}
