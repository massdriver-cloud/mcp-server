package tools

import (
	"context"
	"fmt"

	"github.com/massdriver-cloud/massdriver-sdk-go/massdriver/platform/ocirepos"
	mcpsdk "github.com/modelcontextprotocol/go-sdk/mcp"
)

var ListOciReposTool = &mcpsdk.Tool{
	Name: "list_oci_repos",
	Description: "Lists OCI repositories in the organization, one page at a time. " +
		"PREFER filtering by `search` or `artifact_type` to focus the catalog. " +
		"Returns up to `page_size` repositories (default 25, max 100) plus a `next_cursor` for the following page. " +
		"To continue, call again with `cursor` set to the previous `next_cursor`. " +
		"Do NOT paginate to exhaustion unless the user explicitly asked for every repository.",
}

type ListOciReposInput struct {
	Search       string `json:"search,omitempty"        jsonschema:"Optional. Search term to filter repositories."`
	ArtifactType string `json:"artifact_type,omitempty" jsonschema:"Optional. Filter by artifact type (e.g., 'BUNDLE')."`
	Cursor       string `json:"cursor,omitempty"        jsonschema:"Optional. Opaque cursor from a prior call's next_cursor. Omit for the first page."`
	PageSize     int    `json:"page_size,omitempty"     jsonschema:"Optional. Page size (1-100, default 25)."`
}

func HandleListOciRepos(c *Client) func(context.Context, *mcpsdk.CallToolRequest, ListOciReposInput) (*mcpsdk.CallToolResult, any, error) {
	return func(ctx context.Context, _ *mcpsdk.CallToolRequest, args ListOciReposInput) (*mcpsdk.CallToolResult, any, error) {
		page, err := c.OciRepos.ListPage(ctx, ocirepos.ListInput{
			Search:       args.Search,
			ArtifactType: ocirepos.ArtifactType(args.ArtifactType),
			PageSize:     clampPageSize(args.PageSize),
			After:        args.Cursor,
		})
		if err != nil {
			return nil, nil, fmt.Errorf("list_oci_repos: %w", err)
		}

		out := pageResult(page)
		return jsonResultStripping(out, "icon")
	}
}

var GetOciRepoTool = &mcpsdk.Tool{
	Name:        "get_oci_repo",
	Description: "Gets a specific OCI repository by ID.",
}

type GetOciRepoInput struct {
	ID string `json:"id" jsonschema:"The OCI repository ID."`
}

func HandleGetOciRepo(c *Client) func(context.Context, *mcpsdk.CallToolRequest, GetOciRepoInput) (*mcpsdk.CallToolResult, any, error) {
	return func(ctx context.Context, _ *mcpsdk.CallToolRequest, args GetOciRepoInput) (*mcpsdk.CallToolResult, any, error) {
		if args.ID == "" {
			return nil, nil, fmt.Errorf("get_oci_repo: id is required")
		}

		repo, err := c.OciRepos.Get(ctx, args.ID)
		if err != nil {
			return nil, nil, fmt.Errorf("get_oci_repo: %w", err)
		}

		return jsonResultStripping(repo, "icon")
	}
}

var CreateOciRepoTool = &mcpsdk.Tool{
	Name:        "create_oci_repo",
	Description: "Creates a new OCI repository.",
}

type CreateOciRepoInput struct {
	ID           string         `json:"id"            jsonschema:"Repository name (immutable after creation)."`
	ArtifactType string         `json:"artifact_type"        jsonschema:"Artifact type (e.g., 'BUNDLE')."`
	Attributes   map[string]any `json:"attributes,omitempty" jsonschema:"Optional. Custom attributes for the repository."`
}

func HandleCreateOciRepo(c *Client) func(context.Context, *mcpsdk.CallToolRequest, CreateOciRepoInput) (*mcpsdk.CallToolResult, any, error) {
	return func(ctx context.Context, _ *mcpsdk.CallToolRequest, args CreateOciRepoInput) (*mcpsdk.CallToolResult, any, error) {
		if args.ID == "" {
			return nil, nil, fmt.Errorf("create_oci_repo: id is required")
		}
		if args.ArtifactType == "" {
			return nil, nil, fmt.Errorf("create_oci_repo: artifact_type is required")
		}

		repo, err := c.OciRepos.Create(ctx, ocirepos.CreateInput{
			ID:           args.ID,
			ArtifactType: ocirepos.ArtifactType(args.ArtifactType),
			Attributes:   args.Attributes,
		})
		if err != nil {
			if isMutationFailed(err) {
				return errorResult(fmt.Sprintf("create_oci_repo failed: %s", mutationErr(err))), nil, nil
			}
			return nil, nil, fmt.Errorf("create_oci_repo: %w", err)
		}

		return jsonResultStripping(repo, "icon")
	}
}

var UpdateOciRepoTool = &mcpsdk.Tool{
	Name:        "update_oci_repo",
	Description: "Updates an OCI repository's attributes.",
}

type UpdateOciRepoInput struct {
	ID         string         `json:"id"                   jsonschema:"The OCI repository ID to update."`
	Attributes map[string]any `json:"attributes,omitempty" jsonschema:"Optional. New attributes for the repository."`
}

func HandleUpdateOciRepo(c *Client) func(context.Context, *mcpsdk.CallToolRequest, UpdateOciRepoInput) (*mcpsdk.CallToolResult, any, error) {
	return func(ctx context.Context, _ *mcpsdk.CallToolRequest, args UpdateOciRepoInput) (*mcpsdk.CallToolResult, any, error) {
		if args.ID == "" {
			return nil, nil, fmt.Errorf("update_oci_repo: id is required")
		}

		repo, err := c.OciRepos.Update(ctx, args.ID, ocirepos.UpdateInput{
			Attributes: args.Attributes,
		})
		if err != nil {
			if isMutationFailed(err) {
				return errorResult(fmt.Sprintf("update_oci_repo failed: %s", mutationErr(err))), nil, nil
			}
			return nil, nil, fmt.Errorf("update_oci_repo: %w", err)
		}

		return jsonResultStripping(repo, "icon")
	}
}

var DeleteOciRepoTool = &mcpsdk.Tool{
	Name:        "delete_oci_repo",
	Description: "Deletes an OCI repository.",
}

type DeleteOciRepoInput struct {
	ID string `json:"id" jsonschema:"The OCI repository ID to delete."`
}

func HandleDeleteOciRepo(c *Client) func(context.Context, *mcpsdk.CallToolRequest, DeleteOciRepoInput) (*mcpsdk.CallToolResult, any, error) {
	return func(ctx context.Context, _ *mcpsdk.CallToolRequest, args DeleteOciRepoInput) (*mcpsdk.CallToolResult, any, error) {
		if args.ID == "" {
			return nil, nil, fmt.Errorf("delete_oci_repo: id is required")
		}

		_, err := c.OciRepos.Delete(ctx, args.ID)
		if err != nil {
			if isMutationFailed(err) {
				return errorResult(fmt.Sprintf("delete_oci_repo failed: %s", mutationErr(err))), nil, nil
			}
			return nil, nil, fmt.Errorf("delete_oci_repo: %w", err)
		}

		return textResult(fmt.Sprintf("OCI repository %q deleted successfully", args.ID)), nil, nil
	}
}
