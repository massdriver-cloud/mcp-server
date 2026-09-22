package tools

import (
	"context"
	"fmt"

	mcpsdk "github.com/modelcontextprotocol/go-sdk/mcp"
)

var GetResourceTypeTool = &mcpsdk.Tool{
	Name: "get_resource_type",
	Description: "Gets a resource type — the contract behind Massdriver's connection system — including its full JSON schema, " +
		"import instructions, and connection orientation (LINK = wired explicitly between instances; ENVIRONMENT_DEFAULT = " +
		"satisfied automatically by an environment default). The version portion of the ID may be an exact semver " +
		"('aws-iam-role@1.2.3'), a range ('aws-iam-role@~1'), a channel ('aws-iam-role@latest'), or omitted entirely " +
		"(resolves to latest); the returned ID always carries the fully resolved version.",
}

type GetResourceTypeInput struct {
	ID string `json:"id" jsonschema:"The resource type ID, optionally versioned (e.g., 'aws-iam-role', 'aws-iam-role@1.2.3', 'aws-iam-role@~1', 'aws-iam-role@latest')."`
}

func HandleGetResourceType(c *Client) func(context.Context, *mcpsdk.CallToolRequest, GetResourceTypeInput) (*mcpsdk.CallToolResult, any, error) {
	return func(ctx context.Context, _ *mcpsdk.CallToolRequest, args GetResourceTypeInput) (*mcpsdk.CallToolResult, any, error) {
		if args.ID == "" {
			return nil, nil, fmt.Errorf("get_resource_type: id is required")
		}

		rt, err := c.ResourceTypes.Get(ctx, args.ID)
		if err != nil {
			return nil, nil, fmt.Errorf("get_resource_type: %w", err)
		}

		return jsonResultStripping(rt, "icon")
	}
}

var ListResourceTypeDependentsTool = &mcpsdk.Tool{
	Name: "list_resource_type_dependents",
	Description: "Lists the instances in an environment that depend on a resource type, one entry per (instance, dependency field) " +
		"pair. Use it to see what a resource type is used by before changing or removing a resource of that type. Returns the " +
		"whole list (not paginated).",
}

type ListResourceTypeDependentsInput struct {
	EnvironmentID  string `json:"environment_id"   jsonschema:"The environment ID to inspect (e.g., 'myproj-staging')."`
	ResourceTypeID string `json:"resource_type_id" jsonschema:"The resource type ID (e.g., 'aws-vpc'). A version suffix is accepted but matching resolves at the type level."`
}

func HandleListResourceTypeDependents(c *Client) func(context.Context, *mcpsdk.CallToolRequest, ListResourceTypeDependentsInput) (*mcpsdk.CallToolResult, any, error) {
	return func(ctx context.Context, _ *mcpsdk.CallToolRequest, args ListResourceTypeDependentsInput) (*mcpsdk.CallToolResult, any, error) {
		if args.EnvironmentID == "" {
			return nil, nil, fmt.Errorf("list_resource_type_dependents: environment_id is required")
		}
		if args.ResourceTypeID == "" {
			return nil, nil, fmt.Errorf("list_resource_type_dependents: resource_type_id is required")
		}

		deps, err := c.ResourceTypes.Dependents(ctx, args.EnvironmentID, args.ResourceTypeID)
		if err != nil {
			return nil, nil, fmt.Errorf("list_resource_type_dependents: %w", err)
		}

		return jsonResultStripping(listResult(deps), "icon")
	}
}
