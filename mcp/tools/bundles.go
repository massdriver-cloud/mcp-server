package tools

import (
	"context"
	"fmt"

	mcpsdk "github.com/modelcontextprotocol/go-sdk/mcp"
)

var GetBundleTool = &mcpsdk.Tool{
	Name: "get_bundle",
	Description: "Gets a specific bundle by ID. Supports version constraints like 'aws-aurora-postgres@1.2.3', 'aws-aurora-postgres@~1', or 'aws-aurora-postgres@latest'. " +
		"To discover which bundles exist, use `list_oci_repos` with `artifact_type` set to 'BUNDLE'. " +
		"To list the available versions of a bundle, use `get_oci_repo` — the published version tags live on the bundle's OCI repository. " +
		"Returns resources (outputs the bundle produces) and dependencies (inputs it requires), each tagged with a resourceType — components connect when a producer's resource type matches a consumer's dependency type.",
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
