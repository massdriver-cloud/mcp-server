package tools

import (
	"context"
	"fmt"

	"github.com/massdriver-cloud/massdriver-sdk-go/massdriver/platform/groups"
	mcpsdk "github.com/modelcontextprotocol/go-sdk/mcp"
)

var ListGroupsTool = &mcpsdk.Tool{
	Name: "list_groups",
	Description: "Lists access control groups in the organization, one page at a time. " +
		"Returns up to `page_size` groups (default 25, max 100) plus a `next_cursor` for the following page. " +
		"To continue, call again with `cursor` set to the previous `next_cursor`. " +
		"Stop once you have what you need — do NOT paginate to exhaustion unless the user explicitly asked for every group.",
}

type ListGroupsInput struct {
	Cursor   string `json:"cursor,omitempty"    jsonschema:"Optional. Opaque cursor from a prior call's next_cursor. Omit for the first page."`
	PageSize int    `json:"page_size,omitempty" jsonschema:"Optional. Page size (1-100, default 25)."`
}

func HandleListGroups(c *Client) func(context.Context, *mcpsdk.CallToolRequest, ListGroupsInput) (*mcpsdk.CallToolResult, any, error) {
	return func(ctx context.Context, _ *mcpsdk.CallToolRequest, args ListGroupsInput) (*mcpsdk.CallToolResult, any, error) {
		page, err := c.Groups.ListPage(ctx, groups.ListInput{
			PageSize: clampPageSize(args.PageSize),
			After:    args.Cursor,
		})
		if err != nil {
			return nil, nil, fmt.Errorf("list_groups: %w", err)
		}

		out := pageResult(page)
		result, err := jsonResult(out)
		if err != nil {
			return nil, nil, err
		}
		return result, out, nil
	}
}

var GetGroupTool = &mcpsdk.Tool{
	Name:        "get_group",
	Description: "Gets a specific access control group by ID, including its members, service accounts, and policies.",
}

type GetGroupInput struct {
	ID string `json:"id" jsonschema:"The group ID."`
}

func HandleGetGroup(c *Client) func(context.Context, *mcpsdk.CallToolRequest, GetGroupInput) (*mcpsdk.CallToolResult, any, error) {
	return func(ctx context.Context, _ *mcpsdk.CallToolRequest, args GetGroupInput) (*mcpsdk.CallToolResult, any, error) {
		if args.ID == "" {
			return nil, nil, fmt.Errorf("get_group: id is required")
		}

		group, err := c.Groups.Get(ctx, args.ID)
		if err != nil {
			return nil, nil, fmt.Errorf("get_group: %w", err)
		}

		result, err := jsonResult(group)
		if err != nil {
			return nil, nil, err
		}
		return result, group, nil
	}
}

var CreateGroupTool = &mcpsdk.Tool{
	Name:        "create_group",
	Description: "Creates a new access control group in the organization.",
}

type CreateGroupInput struct {
	Name        string `json:"name"                  jsonschema:"The name of the group."`
	Description string `json:"description,omitempty" jsonschema:"Optional. A description of the group."`
}

func HandleCreateGroup(c *Client) func(context.Context, *mcpsdk.CallToolRequest, CreateGroupInput) (*mcpsdk.CallToolResult, any, error) {
	return func(ctx context.Context, _ *mcpsdk.CallToolRequest, args CreateGroupInput) (*mcpsdk.CallToolResult, any, error) {
		if args.Name == "" {
			return nil, nil, fmt.Errorf("create_group: name is required")
		}

		group, err := c.Groups.Create(ctx, groups.CreateInput{
			Name:        args.Name,
			Description: args.Description,
		})
		if err != nil {
			if isMutationFailed(err) {
				return textResult(fmt.Sprintf("create_group failed: %s", mutationErr(err))), nil, nil
			}
			return nil, nil, fmt.Errorf("create_group: %w", err)
		}

		result, err := jsonResult(group)
		if err != nil {
			return nil, nil, err
		}
		return result, group, nil
	}
}

var UpdateGroupTool = &mcpsdk.Tool{
	Name:        "update_group",
	Description: "Updates an existing access control group's name or description.",
}

type UpdateGroupInput struct {
	ID          string `json:"id"                    jsonschema:"The group ID to update."`
	Name        string `json:"name,omitempty"        jsonschema:"Optional. New name for the group."`
	Description string `json:"description,omitempty" jsonschema:"Optional. New description for the group."`
}

func HandleUpdateGroup(c *Client) func(context.Context, *mcpsdk.CallToolRequest, UpdateGroupInput) (*mcpsdk.CallToolResult, any, error) {
	return func(ctx context.Context, _ *mcpsdk.CallToolRequest, args UpdateGroupInput) (*mcpsdk.CallToolResult, any, error) {
		if args.ID == "" {
			return nil, nil, fmt.Errorf("update_group: id is required")
		}

		group, err := c.Groups.Update(ctx, args.ID, groups.UpdateInput{
			Name:        args.Name,
			Description: args.Description,
		})
		if err != nil {
			if isMutationFailed(err) {
				return textResult(fmt.Sprintf("update_group failed: %s", mutationErr(err))), nil, nil
			}
			return nil, nil, fmt.Errorf("update_group: %w", err)
		}

		result, err := jsonResult(group)
		if err != nil {
			return nil, nil, err
		}
		return result, group, nil
	}
}

var DeleteGroupTool = &mcpsdk.Tool{
	Name:        "delete_group",
	Description: "Deletes an access control group.",
}

type DeleteGroupInput struct {
	ID string `json:"id" jsonschema:"The group ID to delete."`
}

func HandleDeleteGroup(c *Client) func(context.Context, *mcpsdk.CallToolRequest, DeleteGroupInput) (*mcpsdk.CallToolResult, any, error) {
	return func(ctx context.Context, _ *mcpsdk.CallToolRequest, args DeleteGroupInput) (*mcpsdk.CallToolResult, any, error) {
		if args.ID == "" {
			return nil, nil, fmt.Errorf("delete_group: id is required")
		}

		_, err := c.Groups.Delete(ctx, args.ID)
		if err != nil {
			if isMutationFailed(err) {
				return textResult(fmt.Sprintf("delete_group failed: %s", mutationErr(err))), nil, nil
			}
			return nil, nil, fmt.Errorf("delete_group: %w", err)
		}

		return textResult(fmt.Sprintf("group %q deleted successfully", args.ID)), nil, nil
	}
}

var AddGroupUserTool = &mcpsdk.Tool{
	Name:        "add_group_user",
	Description: "Adds a user to an access control group by email. If the user is not yet a member of the organization, an invitation will be sent.",
}

type AddGroupUserInput struct {
	GroupID string `json:"group_id" jsonschema:"The group ID."`
	Email   string `json:"email"    jsonschema:"The email address of the user to add."`
}

func HandleAddGroupUser(c *Client) func(context.Context, *mcpsdk.CallToolRequest, AddGroupUserInput) (*mcpsdk.CallToolResult, any, error) {
	return func(ctx context.Context, _ *mcpsdk.CallToolRequest, args AddGroupUserInput) (*mcpsdk.CallToolResult, any, error) {
		if args.GroupID == "" {
			return nil, nil, fmt.Errorf("add_group_user: group_id is required")
		}
		if args.Email == "" {
			return nil, nil, fmt.Errorf("add_group_user: email is required")
		}

		addResult, err := c.Groups.AddUser(ctx, args.GroupID, args.Email)
		if err != nil {
			if isMutationFailed(err) {
				return textResult(fmt.Sprintf("add_group_user failed: %s", mutationErr(err))), nil, nil
			}
			return nil, nil, fmt.Errorf("add_group_user: %w", err)
		}

		result, err := jsonResult(addResult)
		if err != nil {
			return nil, nil, err
		}
		return result, addResult, nil
	}
}

var RemoveGroupUserTool = &mcpsdk.Tool{
	Name:        "remove_group_user",
	Description: "Removes a user from an access control group.",
}

type RemoveGroupUserInput struct {
	GroupID string `json:"group_id" jsonschema:"The group ID."`
	Email   string `json:"email"    jsonschema:"The email address of the user to remove."`
}

func HandleRemoveGroupUser(c *Client) func(context.Context, *mcpsdk.CallToolRequest, RemoveGroupUserInput) (*mcpsdk.CallToolResult, any, error) {
	return func(ctx context.Context, _ *mcpsdk.CallToolRequest, args RemoveGroupUserInput) (*mcpsdk.CallToolResult, any, error) {
		if args.GroupID == "" {
			return nil, nil, fmt.Errorf("remove_group_user: group_id is required")
		}
		if args.Email == "" {
			return nil, nil, fmt.Errorf("remove_group_user: email is required")
		}

		err := c.Groups.RemoveUser(ctx, args.GroupID, args.Email)
		if err != nil {
			if isMutationFailed(err) {
				return textResult(fmt.Sprintf("remove_group_user failed: %s", mutationErr(err))), nil, nil
			}
			return nil, nil, fmt.Errorf("remove_group_user: %w", err)
		}

		return textResult(fmt.Sprintf("user %q removed from group %q successfully", args.Email, args.GroupID)), nil, nil
	}
}

var RevokeGroupInvitationTool = &mcpsdk.Tool{
	Name:        "revoke_group_invitation",
	Description: "Revokes a pending invitation for a user to join an access control group.",
}

type RevokeGroupInvitationInput struct {
	GroupID string `json:"group_id" jsonschema:"The group ID."`
	Email   string `json:"email"    jsonschema:"The email address of the invitation to revoke."`
}

func HandleRevokeGroupInvitation(c *Client) func(context.Context, *mcpsdk.CallToolRequest, RevokeGroupInvitationInput) (*mcpsdk.CallToolResult, any, error) {
	return func(ctx context.Context, _ *mcpsdk.CallToolRequest, args RevokeGroupInvitationInput) (*mcpsdk.CallToolResult, any, error) {
		if args.GroupID == "" {
			return nil, nil, fmt.Errorf("revoke_group_invitation: group_id is required")
		}
		if args.Email == "" {
			return nil, nil, fmt.Errorf("revoke_group_invitation: email is required")
		}

		err := c.Groups.RevokeInvitation(ctx, args.GroupID, args.Email)
		if err != nil {
			if isMutationFailed(err) {
				return textResult(fmt.Sprintf("revoke_group_invitation failed: %s", mutationErr(err))), nil, nil
			}
			return nil, nil, fmt.Errorf("revoke_group_invitation: %w", err)
		}

		return textResult(fmt.Sprintf("invitation for %q revoked from group %q successfully", args.Email, args.GroupID)), nil, nil
	}
}

var AddGroupServiceAccountTool = &mcpsdk.Tool{
	Name:        "add_group_service_account",
	Description: "Adds a service account to an access control group.",
}

type AddGroupServiceAccountInput struct {
	GroupID          string `json:"group_id"           jsonschema:"The group ID."`
	ServiceAccountID string `json:"service_account_id" jsonschema:"The service account ID to add."`
}

func HandleAddGroupServiceAccount(c *Client) func(context.Context, *mcpsdk.CallToolRequest, AddGroupServiceAccountInput) (*mcpsdk.CallToolResult, any, error) {
	return func(ctx context.Context, _ *mcpsdk.CallToolRequest, args AddGroupServiceAccountInput) (*mcpsdk.CallToolResult, any, error) {
		if args.GroupID == "" {
			return nil, nil, fmt.Errorf("add_group_service_account: group_id is required")
		}
		if args.ServiceAccountID == "" {
			return nil, nil, fmt.Errorf("add_group_service_account: service_account_id is required")
		}

		err := c.Groups.AddServiceAccount(ctx, args.GroupID, args.ServiceAccountID)
		if err != nil {
			if isMutationFailed(err) {
				return textResult(fmt.Sprintf("add_group_service_account failed: %s", mutationErr(err))), nil, nil
			}
			return nil, nil, fmt.Errorf("add_group_service_account: %w", err)
		}

		return textResult(fmt.Sprintf("service account %q added to group %q successfully", args.ServiceAccountID, args.GroupID)), nil, nil
	}
}

var RemoveGroupServiceAccountTool = &mcpsdk.Tool{
	Name:        "remove_group_service_account",
	Description: "Removes a service account from an access control group.",
}

type RemoveGroupServiceAccountInput struct {
	GroupID          string `json:"group_id"           jsonschema:"The group ID."`
	ServiceAccountID string `json:"service_account_id" jsonschema:"The service account ID to remove."`
}

func HandleRemoveGroupServiceAccount(c *Client) func(context.Context, *mcpsdk.CallToolRequest, RemoveGroupServiceAccountInput) (*mcpsdk.CallToolResult, any, error) {
	return func(ctx context.Context, _ *mcpsdk.CallToolRequest, args RemoveGroupServiceAccountInput) (*mcpsdk.CallToolResult, any, error) {
		if args.GroupID == "" {
			return nil, nil, fmt.Errorf("remove_group_service_account: group_id is required")
		}
		if args.ServiceAccountID == "" {
			return nil, nil, fmt.Errorf("remove_group_service_account: service_account_id is required")
		}

		err := c.Groups.RemoveServiceAccount(ctx, args.GroupID, args.ServiceAccountID)
		if err != nil {
			if isMutationFailed(err) {
				return textResult(fmt.Sprintf("remove_group_service_account failed: %s", mutationErr(err))), nil, nil
			}
			return nil, nil, fmt.Errorf("remove_group_service_account: %w", err)
		}

		return textResult(fmt.Sprintf("service account %q removed from group %q successfully", args.ServiceAccountID, args.GroupID)), nil, nil
	}
}
