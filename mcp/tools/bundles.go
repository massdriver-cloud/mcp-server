package tools

import (
	"context"
	"fmt"

	"github.com/massdriver-cloud/massdriver-sdk-go/massdriver/platform/bundles"
	mcpsdk "github.com/modelcontextprotocol/go-sdk/mcp"
)

var ListBundlesTool = &mcpsdk.Tool{
	Name:        "list_bundles",
	Description: "Lists available bundles in the catalog. Optionally filter by OCI repo name, resource type, or search term.",
}

type ListBundlesInput struct {
	Search         string `json:"search"          jsonschema:"Optional. Search term to filter bundles by name or description."`
	OciRepoName    string `json:"oci_repo_name"   jsonschema:"Optional. Filter to bundles from this OCI repository."`
	ResourceType   string `json:"resource_type"   jsonschema:"Optional. Filter by resource type (e.g., 'aws-rds-instance')."`
	DependencyType string `json:"dependency_type" jsonschema:"Optional. Filter by dependency type."`
}

func HandleListBundles(c *Client) func(context.Context, *mcpsdk.CallToolRequest, ListBundlesInput) (*mcpsdk.CallToolResult, any, error) {
	return func(ctx context.Context, _ *mcpsdk.CallToolRequest, args ListBundlesInput) (*mcpsdk.CallToolResult, any, error) {
		input := bundles.ListInput{
			Search:         args.Search,
			OciRepoName:    args.OciRepoName,
			ResourceType:   args.ResourceType,
			DependencyType: args.DependencyType,
		}

		list, err := c.Bundles.List(ctx, input)
		if err != nil {
			return nil, nil, fmt.Errorf("list_bundles: %w", err)
		}

		result, err := jsonResult(list)
		if err != nil {
			return nil, nil, err
		}
		return result, list, nil
	}
}

var GetBundleTool = &mcpsdk.Tool{
	Name:        "get_bundle",
	Description: "Gets a specific bundle by ID. Supports version constraints like 'aws-aurora-postgres@1.2.3', 'aws-aurora-postgres@~1', or 'aws-aurora-postgres@latest'.",
}

type GetBundleInput struct {
	ID string `json:"id" jsonschema:"The bundle ID, optionally with a version constraint (e.g., 'aws-aurora-postgres', 'aws-aurora-postgres@latest', 'aws-aurora-postgres@~1.2')."`
}

func HandleGetBundle(c *Client) func(context.Context, *mcpsdk.CallToolRequest, GetBundleInput) (*mcpsdk.CallToolResult, any, error) {
	return func(ctx context.Context, _ *mcpsdk.CallToolRequest, args GetBundleInput) (*mcpsdk.CallToolResult, any, error) {
		if args.ID == "" {
			return nil, nil, fmt.Errorf("get_bundle: id is required")
		}

		bundle, err := c.Bundles.Get(ctx, args.ID)
		if err != nil {
			return nil, nil, fmt.Errorf("get_bundle: %w", err)
		}

		result, err := jsonResult(bundle)
		if err != nil {
			return nil, nil, err
		}
		return result, bundle, nil
	}
}
