package tools

import (
	"context"
	"fmt"

	mcpsdk "github.com/modelcontextprotocol/go-sdk/mcp"
)

var GetResourceTypeTool = &mcpsdk.Tool{
	Name: "get_resource_type",
	Description: "Gets a resource type — the contract behind Massdriver's connection system. Every dependency a bundle declares " +
		"and every resource a bundle produces references one. " +
		"Returns `schema` (the JSON Schema for the data the type exposes, and the shape create_resource expects in `payload`), " +
		"`uiSchema` (rendering hints for the import form), `instructions` (step-by-step import instructions in markdown, typically one entry per workflow such as CLI and cloud console), " +
		"`connectionOrientation` (LINK when the dependency is wired explicitly between instances, ENVIRONMENT_DEFAULT when it is satisfied by an environment-level default), " +
		"and `effectiveAttributes` (auto-injected md-* system attributes). " +
		"Call this before create_resource to learn what payload the type requires.",
}

type GetResourceTypeInput struct {
	ID string `json:"id" jsonschema:"The resource type identifier, optionally with a version suffix. Accepts a bare identifier ('aws-iam-role', resolving to the newest stable release), an exact version ('aws-iam-role@1.2.3'), a tilde range ('aws-iam-role@~1.2' for the latest 1.2.x, 'aws-iam-role@~1' for the latest 1.x.x), or a channel ('aws-iam-role@latest', 'aws-iam-role@latest+dev' to include dev builds). The returned id always carries the fully resolved version."`
}

func HandleGetResourceType(c *Client) func(context.Context, *mcpsdk.CallToolRequest, GetResourceTypeInput) (*mcpsdk.CallToolResult, any, error) {
	return func(ctx context.Context, _ *mcpsdk.CallToolRequest, args GetResourceTypeInput) (*mcpsdk.CallToolResult, any, error) {
		if args.ID == "" {
			return nil, nil, fmt.Errorf("get_resource_type: id is required")
		}

		resourceType, err := c.ResourceTypes.Get(ctx, args.ID)
		if err != nil {
			return nil, nil, fmt.Errorf("get_resource_type: %w", err)
		}

		return jsonResultStripping(resourceType, "icon")
	}
}

var ListResourceTypeDependentsTool = &mcpsdk.Tool{
	Name: "list_resource_type_dependents",
	Description: "Lists what depends on a resource type within one environment — the blast radius before you change or remove it, " +
		"or before you repoint an environment default of that type. " +
		"Returns one entry per (instance, dependency field) pair, so an instance whose bundle depends on the same type through two fields appears twice. " +
		"The instance and resource type in each entry are slim references (id and name only); use get_instance or get_resource_type for the full shape. " +
		"The result set is bounded by the environment, so it is returned whole rather than paginated.",
}

type ListResourceTypeDependentsInput struct {
	EnvironmentID  string `json:"environment_id"   jsonschema:"The environment to scope the search to (e.g. 'myproj-staging'). Dependents are always looked up within a single environment."`
	ResourceTypeID string `json:"resource_type_id" jsonschema:"The resource type to find dependents of. Accepts a bare identifier ('aws-vpc') or a versioned one ('aws-vpc@1.0.0'); a version suffix is accepted but matching resolves at the type level, since bundles reference resource types without a version."`
}

func HandleListResourceTypeDependents(c *Client) func(context.Context, *mcpsdk.CallToolRequest, ListResourceTypeDependentsInput) (*mcpsdk.CallToolResult, any, error) {
	return func(ctx context.Context, _ *mcpsdk.CallToolRequest, args ListResourceTypeDependentsInput) (*mcpsdk.CallToolResult, any, error) {
		if args.EnvironmentID == "" {
			return nil, nil, fmt.Errorf("list_resource_type_dependents: environment_id is required")
		}
		if args.ResourceTypeID == "" {
			return nil, nil, fmt.Errorf("list_resource_type_dependents: resource_type_id is required")
		}

		dependents, err := c.ResourceTypes.Dependents(ctx, args.EnvironmentID, args.ResourceTypeID)
		if err != nil {
			return nil, nil, fmt.Errorf("list_resource_type_dependents: %w", err)
		}

		out := listResult(dependents)
		return jsonResultStripping(out, "icon")
	}
}
