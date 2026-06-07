package tools

import (
	"context"
	"fmt"

	"github.com/massdriver-cloud/massdriver-sdk-go/massdriver/platform/auditlogs"
	mcpsdk "github.com/modelcontextprotocol/go-sdk/mcp"
)

var ListAuditLogsTool = &mcpsdk.Tool{
	Name: "list_audit_logs",
	Description: "Lists audit log entries for the organization, newest first, one page at a time. " +
		"STRONGLY PREFER filtering by `type`, `actor_id`, or `actor_search` — audit logs grow unboundedly. " +
		"Returns up to `page_size` entries (default 25, max 100) plus a `next_cursor` for the following page. " +
		"To continue, call again with `cursor` set to the previous `next_cursor`. " +
		"Do NOT paginate to exhaustion unless the user explicitly asked for the full history.",
}

type ListAuditLogsInput struct {
	Type        string `json:"type,omitempty"         jsonschema:"Optional. Filter by event type (e.g., 'deployment.created')."`
	ActorID     string `json:"actor_id,omitempty"     jsonschema:"Optional. Filter by actor ID."`
	ActorSearch string `json:"actor_search,omitempty" jsonschema:"Optional. Search by actor name or email."`
	Cursor      string `json:"cursor,omitempty"       jsonschema:"Optional. Opaque cursor from a prior call's next_cursor. Omit for the first page."`
	PageSize    int    `json:"page_size,omitempty"    jsonschema:"Optional. Page size (1-100, default 25)."`
}

func HandleListAuditLogs(c *Client) func(context.Context, *mcpsdk.CallToolRequest, ListAuditLogsInput) (*mcpsdk.CallToolResult, any, error) {
	return func(ctx context.Context, _ *mcpsdk.CallToolRequest, args ListAuditLogsInput) (*mcpsdk.CallToolResult, any, error) {
		page, err := c.AuditLogs.ListPage(ctx, auditlogs.ListInput{
			Type:        args.Type,
			ActorID:     args.ActorID,
			ActorSearch: args.ActorSearch,
			PageSize:    clampPageSize(args.PageSize),
			After:       args.Cursor,
		})
		if err != nil {
			return nil, nil, fmt.Errorf("list_audit_logs: %w", err)
		}

		out := pageResult(page)
		result, err := jsonResult(out)
		if err != nil {
			return nil, nil, err
		}
		return result, out, nil
	}
}

var ListAuditLogEventTypesTool = &mcpsdk.Tool{
	Name:        "list_audit_log_event_types",
	Description: "Lists all available audit log event types that can be used to filter audit logs.",
}

type ListAuditLogEventTypesInput struct{}

func HandleListAuditLogEventTypes(c *Client) func(context.Context, *mcpsdk.CallToolRequest, ListAuditLogEventTypesInput) (*mcpsdk.CallToolResult, any, error) {
	return func(ctx context.Context, _ *mcpsdk.CallToolRequest, _ ListAuditLogEventTypesInput) (*mcpsdk.CallToolResult, any, error) {
		types, err := c.AuditLogs.ListEventTypes(ctx)
		if err != nil {
			return nil, nil, fmt.Errorf("list_audit_log_event_types: %w", err)
		}

		result, err := jsonResult(types)
		if err != nil {
			return nil, nil, err
		}
		return result, types, nil
	}
}

var GetAuditLogTool = &mcpsdk.Tool{
	Name:        "get_audit_log",
	Description: "Gets a specific audit log entry by ID.",
}

type GetAuditLogInput struct {
	ID string `json:"id" jsonschema:"The audit log entry ID."`
}

func HandleGetAuditLog(c *Client) func(context.Context, *mcpsdk.CallToolRequest, GetAuditLogInput) (*mcpsdk.CallToolResult, any, error) {
	return func(ctx context.Context, _ *mcpsdk.CallToolRequest, args GetAuditLogInput) (*mcpsdk.CallToolResult, any, error) {
		if args.ID == "" {
			return nil, nil, fmt.Errorf("get_audit_log: id is required")
		}

		log, err := c.AuditLogs.Get(ctx, args.ID)
		if err != nil {
			return nil, nil, fmt.Errorf("get_audit_log: %w", err)
		}

		result, err := jsonResult(log)
		if err != nil {
			return nil, nil, err
		}
		return result, log, nil
	}
}
