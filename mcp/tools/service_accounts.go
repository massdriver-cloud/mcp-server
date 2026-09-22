package tools

import (
	"context"
	"fmt"

	"github.com/massdriver-cloud/massdriver-sdk-go/massdriver/platform/serviceaccounts"
	mcpsdk "github.com/modelcontextprotocol/go-sdk/mcp"
)

var ListServiceAccountsTool = &mcpsdk.Tool{
	Name: "list_service_accounts",
	Description: "Lists service accounts in the organization, one page at a time. " +
		"Returns up to `page_size` accounts (default 25, max 100) plus a `next_cursor` for the following page. " +
		"To continue, call again with `cursor` set to the previous `next_cursor`. " +
		"Use `search` to narrow by name when looking for a specific account. " +
		"Do NOT paginate to exhaustion unless the user explicitly asked for every account.",
}

type ListServiceAccountsInput struct {
	Search   string `json:"search,omitempty"    jsonschema:"Optional. Search term to filter service accounts by name."`
	Cursor   string `json:"cursor,omitempty"    jsonschema:"Optional. Opaque cursor from a prior call's next_cursor. Omit for the first page."`
	PageSize int    `json:"page_size,omitempty" jsonschema:"Optional. Page size (1-100, default 25)."`
}

func HandleListServiceAccounts(c *Client) func(context.Context, *mcpsdk.CallToolRequest, ListServiceAccountsInput) (*mcpsdk.CallToolResult, any, error) {
	return func(ctx context.Context, _ *mcpsdk.CallToolRequest, args ListServiceAccountsInput) (*mcpsdk.CallToolResult, any, error) {
		page, err := c.ServiceAccounts.ListPage(ctx, serviceaccounts.ListInput{
			Search:   args.Search,
			PageSize: clampPageSize(args.PageSize),
			After:    args.Cursor,
		})
		if err != nil {
			return nil, nil, fmt.Errorf("list_service_accounts: %w", err)
		}

		out := pageResult(page)
		result, err := jsonResult(out)
		if err != nil {
			return nil, nil, err
		}
		return result, out, nil
	}
}

var GetServiceAccountTool = &mcpsdk.Tool{
	Name:        "get_service_account",
	Description: "Gets a specific service account by ID.",
}

type GetServiceAccountInput struct {
	ID string `json:"id" jsonschema:"The service account ID."`
}

func HandleGetServiceAccount(c *Client) func(context.Context, *mcpsdk.CallToolRequest, GetServiceAccountInput) (*mcpsdk.CallToolResult, any, error) {
	return func(ctx context.Context, _ *mcpsdk.CallToolRequest, args GetServiceAccountInput) (*mcpsdk.CallToolResult, any, error) {
		if args.ID == "" {
			return nil, nil, fmt.Errorf("get_service_account: id is required")
		}

		sa, err := c.ServiceAccounts.Get(ctx, args.ID)
		if err != nil {
			return nil, nil, fmt.Errorf("get_service_account: %w", err)
		}

		result, err := jsonResult(sa)
		if err != nil {
			return nil, nil, err
		}
		return result, sa, nil
	}
}

// Service account creation is deliberately NOT exposed: the MCP server must
// never be able to mint new credentials (a created service account returns a
// live bearer token). Create service accounts through the Massdriver UI/CLI.

var UpdateServiceAccountTool = &mcpsdk.Tool{
	Name:        "update_service_account",
	Description: "Updates a service account's name or description.",
}

type UpdateServiceAccountInput struct {
	ID          string `json:"id"                    jsonschema:"The service account ID to update."`
	Name        string `json:"name,omitempty"        jsonschema:"Optional. New name for the service account."`
	Description string `json:"description,omitempty" jsonschema:"Optional. New description for the service account."`
}

func HandleUpdateServiceAccount(c *Client) func(context.Context, *mcpsdk.CallToolRequest, UpdateServiceAccountInput) (*mcpsdk.CallToolResult, any, error) {
	return func(ctx context.Context, _ *mcpsdk.CallToolRequest, args UpdateServiceAccountInput) (*mcpsdk.CallToolResult, any, error) {
		if args.ID == "" {
			return nil, nil, fmt.Errorf("update_service_account: id is required")
		}

		sa, err := c.ServiceAccounts.Update(ctx, args.ID, serviceaccounts.UpdateInput{
			Name:        args.Name,
			Description: args.Description,
		})
		if err != nil {
			if isMutationFailed(err) {
				return errorResult(fmt.Sprintf("update_service_account failed: %s", mutationErr(err))), nil, nil
			}
			return nil, nil, fmt.Errorf("update_service_account: %w", err)
		}

		result, err := jsonResult(sa)
		if err != nil {
			return nil, nil, err
		}
		return result, sa, nil
	}
}

var DeleteServiceAccountTool = &mcpsdk.Tool{
	Name:        "delete_service_account",
	Description: "Deletes a service account.",
}

type DeleteServiceAccountInput struct {
	ID string `json:"id" jsonschema:"The service account ID to delete."`
}

func HandleDeleteServiceAccount(c *Client) func(context.Context, *mcpsdk.CallToolRequest, DeleteServiceAccountInput) (*mcpsdk.CallToolResult, any, error) {
	return func(ctx context.Context, _ *mcpsdk.CallToolRequest, args DeleteServiceAccountInput) (*mcpsdk.CallToolResult, any, error) {
		if args.ID == "" {
			return nil, nil, fmt.Errorf("delete_service_account: id is required")
		}

		_, err := c.ServiceAccounts.Delete(ctx, args.ID)
		if err != nil {
			if isMutationFailed(err) {
				return errorResult(fmt.Sprintf("delete_service_account failed: %s", mutationErr(err))), nil, nil
			}
			return nil, nil, fmt.Errorf("delete_service_account: %w", err)
		}

		return textResult(fmt.Sprintf("service account %q deleted successfully", args.ID)), nil, nil
	}
}
