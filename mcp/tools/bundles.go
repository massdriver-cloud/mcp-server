package tools

import (
	"context"
	"fmt"

	"github.com/massdriver-cloud/massdriver-sdk-go/massdriver/platform/bundles"
	mcpsdk "github.com/modelcontextprotocol/go-sdk/mcp"
)

var ListBundlesTool = &mcpsdk.Tool{
	Name: "list_bundles",
	Description: "Lists bundles in the catalog, one page at a time. " +
		"PREFER filtering by `search`, `oci_repo_name`, `resource_type`, or `dependency_type` to narrow the catalog. " +
		"Returns up to `page_size` bundles (default 25, max 100) plus a `next_cursor` for the following page. " +
		"To continue, call again with `cursor` set to the previous `next_cursor`. " +
		"Do NOT paginate to exhaustion unless the user explicitly asked for every bundle.",
}

type ListBundlesInput struct {
	Search         string `json:"search,omitempty"          jsonschema:"Optional. Search term to filter bundles by name or description."`
	OciRepoName    string `json:"oci_repo_name,omitempty"   jsonschema:"Optional. Filter to bundles from this OCI repository."`
	ResourceType   string `json:"resource_type,omitempty"   jsonschema:"Optional. Filter by resource type (e.g., 'aws-rds-instance')."`
	DependencyType string `json:"dependency_type,omitempty" jsonschema:"Optional. Filter by dependency type."`
	Cursor         string `json:"cursor,omitempty"          jsonschema:"Optional. Opaque cursor from a prior call's next_cursor. Omit for the first page."`
	PageSize       int    `json:"page_size,omitempty"       jsonschema:"Optional. Page size (1-100, default 25)."`
}

func HandleListBundles(c *Client) func(context.Context, *mcpsdk.CallToolRequest, ListBundlesInput) (*mcpsdk.CallToolResult, any, error) {
	return func(ctx context.Context, _ *mcpsdk.CallToolRequest, args ListBundlesInput) (*mcpsdk.CallToolResult, any, error) {
		page, err := c.Bundles.ListPage(ctx, bundles.ListInput{
			Search:         args.Search,
			OciRepoName:    args.OciRepoName,
			ResourceType:   args.ResourceType,
			DependencyType: args.DependencyType,
			PageSize:       clampPageSize(args.PageSize),
			After:          args.Cursor,
		})
		if err != nil {
			return nil, nil, fmt.Errorf("list_bundles: %w", err)
		}

		out := pageResult(page)
		return jsonResultStripping(out, "icon")
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

		return jsonResultStripping(bundle, "icon")
	}
}
