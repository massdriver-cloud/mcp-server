package tools

import (
	"context"
	"fmt"

	"github.com/massdriver-cloud/massdriver-sdk-go/massdriver/platform/organizations"
	mcpsdk "github.com/modelcontextprotocol/go-sdk/mcp"
)

var GetOrganizationTool = &mcpsdk.Tool{
	Name:        "get_organization",
	Description: "Gets the current organization's details (id, name, subscription status, timestamps). Custom attributes and members are not included here — use list_custom_attributes and list_organization_members for those.",
}

type GetOrganizationInput struct{}

func HandleGetOrganization(c *Client) func(context.Context, *mcpsdk.CallToolRequest, GetOrganizationInput) (*mcpsdk.CallToolResult, any, error) {
	return func(ctx context.Context, _ *mcpsdk.CallToolRequest, _ GetOrganizationInput) (*mcpsdk.CallToolResult, any, error) {
		org, err := c.Organizations.Get(ctx)
		if err != nil {
			return nil, nil, fmt.Errorf("get_organization: %w", err)
		}

		result, err := jsonResult(org)
		if err != nil {
			return nil, nil, err
		}
		return result, org, nil
	}
}

// OrganizationSettings is the tool-facing shape of the organization's
// behavior settings. The SDK's organizations.Settings carries no JSON tags, so
// it would serialize with Go field names; wrapping it keeps the wire format
// consistent with every other tool result.
type OrganizationSettings struct {
	DefaultBundleAccess string `json:"defaultBundleAccess"`
}

func toOrganizationSettings(s *organizations.Settings) OrganizationSettings {
	return OrganizationSettings{DefaultBundleAccess: string(s.DefaultBundleAccess)}
}

var GetOrganizationSettingsTool = &mcpsdk.Tool{
	Name: "get_organization_settings",
	Description: "Gets the organization's behavior settings. Currently returns `defaultBundleAccess`: the access new bundle " +
		"repositories receive when created — NONE (each new repository stays restricted until a grant is authored) or " +
		"ALL_PROJECTS (each new bundle repository gets an org-wide repo:pull grant, making its bundles usable by every project). " +
		"Every setting has a default, so an organization created before a setting existed reads it as the default. " +
		"Requires the organization:manageSettings action (organization admins); other callers get a forbidden error.",
}

type GetOrganizationSettingsInput struct{}

func HandleGetOrganizationSettings(c *Client) func(context.Context, *mcpsdk.CallToolRequest, GetOrganizationSettingsInput) (*mcpsdk.CallToolResult, any, error) {
	return func(ctx context.Context, _ *mcpsdk.CallToolRequest, _ GetOrganizationSettingsInput) (*mcpsdk.CallToolResult, any, error) {
		settings, err := c.Organizations.GetSettings(ctx)
		if err != nil {
			return nil, nil, fmt.Errorf("get_organization_settings: %w", err)
		}

		out := toOrganizationSettings(settings)
		result, err := jsonResult(out)
		if err != nil {
			return nil, nil, err
		}
		return result, out, nil
	}
}

var UpdateOrganizationSettingsTool = &mcpsdk.Tool{
	Name: "update_organization_settings",
	Description: "Updates the organization's behavior settings and returns the resulting settings. Only the settings you provide " +
		"are changed. Changing `default_bundle_access` affects only repositories created afterwards — access to existing " +
		"repositories is managed through their grants (see create_oci_repo_grant). " +
		"Requires the organization:manageSettings action (organization admins); other callers get a forbidden error.",
}

type UpdateOrganizationSettingsInput struct {
	DefaultBundleAccess string `json:"default_bundle_access,omitempty" jsonschema:"Optional. Access granted to new bundle repositories at creation: NONE (each new repository stays restricted until a grant is authored) or ALL_PROJECTS (each new bundle repository gets an org-wide repo:pull grant, revocable like any other grant). Omit to leave unchanged."`
}

func HandleUpdateOrganizationSettings(c *Client) func(context.Context, *mcpsdk.CallToolRequest, UpdateOrganizationSettingsInput) (*mcpsdk.CallToolResult, any, error) {
	return func(ctx context.Context, _ *mcpsdk.CallToolRequest, args UpdateOrganizationSettingsInput) (*mcpsdk.CallToolResult, any, error) {
		settings, err := c.Organizations.UpdateSettings(ctx, organizations.UpdateSettingsInput{
			DefaultBundleAccess: organizations.DefaultBundleAccess(args.DefaultBundleAccess),
		})
		if err != nil {
			if isMutationFailed(err) {
				return errorResult(fmt.Sprintf("update_organization_settings failed: %s", mutationErr(err))), nil, nil
			}
			return nil, nil, fmt.Errorf("update_organization_settings: %w", err)
		}

		out := toOrganizationSettings(settings)
		result, err := jsonResult(out)
		if err != nil {
			return nil, nil, err
		}
		return result, out, nil
	}
}

var CreateCustomAttributeTool = &mcpsdk.Tool{
	Name: "create_custom_attribute",
	Description: "Creates a custom attribute definition for the organization. " +
		"Attribute keys are unique across the ENTIRE organization and ALL scopes (case-insensitively) — the same key cannot be declared at two different scopes, so pick per-scope key names up front (e.g. `team` at PROJECT scope blocks `team` at ENVIRONMENT scope). " +
		"When `required` is true, the attribute becomes MANDATORY org-wide for every resource at its scope — set it deliberately. " +
		"This tool defaults `required` to false when you omit it.",
}

type CreateCustomAttributeInput struct {
	Key      string   `json:"key"      jsonschema:"Attribute key name: 1-64 characters, identifier-like (starts with a letter or underscore; letters, digits, and underscores only). Case-insensitive (TEAM and team are the same key); the md- prefix is reserved. Keys are unique org-wide across ALL scopes."`
	Scope    string   `json:"scope"    jsonschema:"Attribute scope: PROJECT, ENVIRONMENT, COMPONENT, or REPO."`
	Required *bool    `json:"required,omitempty" jsonschema:"Optional. Whether the attribute is mandatory org-wide at its scope. Defaults to false when omitted. Set true only when you intend to require it on every resource at that scope."`
	Values   []string `json:"values,omitempty"   jsonschema:"Optional. Allowed values for the attribute."`
}

func HandleCreateCustomAttribute(c *Client) func(context.Context, *mcpsdk.CallToolRequest, CreateCustomAttributeInput) (*mcpsdk.CallToolResult, any, error) {
	return func(ctx context.Context, _ *mcpsdk.CallToolRequest, args CreateCustomAttributeInput) (*mcpsdk.CallToolResult, any, error) {
		if args.Key == "" {
			return nil, nil, fmt.Errorf("create_custom_attribute: key is required")
		}
		if args.Scope == "" {
			return nil, nil, fmt.Errorf("create_custom_attribute: scope is required")
		}

		// The API defaults an omitted `required` to TRUE, which would silently
		// make the attribute mandatory org-wide. Send an explicit false when the
		// caller doesn't specify, so omitting the field is the safe default.
		required := args.Required
		if required == nil {
			f := false
			required = &f
		}

		attr, err := c.Organizations.CreateCustomAttribute(ctx, organizations.CreateCustomAttributeInput{
			Key:      args.Key,
			Scope:    organizations.AttributeScope(args.Scope),
			Required: required,
			Values:   args.Values,
		})
		if err != nil {
			if isMutationFailed(err) {
				return errorResult(fmt.Sprintf("create_custom_attribute failed: %s", mutationErr(err))), nil, nil
			}
			return nil, nil, fmt.Errorf("create_custom_attribute: %w", err)
		}

		result, err := jsonResult(attr)
		if err != nil {
			return nil, nil, err
		}
		return result, attr, nil
	}
}

var UpdateCustomAttributeTool = &mcpsdk.Tool{
	Name:        "update_custom_attribute",
	Description: "Updates a custom attribute definition for the organization.",
}

type UpdateCustomAttributeInput struct {
	ID       string   `json:"id"                 jsonschema:"The custom attribute ID to update."`
	Required *bool    `json:"required,omitempty" jsonschema:"Optional. Whether the attribute is required."`
	Values   []string `json:"values,omitempty"   jsonschema:"Optional. Allowed values for the attribute."`
}

func HandleUpdateCustomAttribute(c *Client) func(context.Context, *mcpsdk.CallToolRequest, UpdateCustomAttributeInput) (*mcpsdk.CallToolResult, any, error) {
	return func(ctx context.Context, _ *mcpsdk.CallToolRequest, args UpdateCustomAttributeInput) (*mcpsdk.CallToolResult, any, error) {
		if args.ID == "" {
			return nil, nil, fmt.Errorf("update_custom_attribute: id is required")
		}

		attr, err := c.Organizations.UpdateCustomAttribute(ctx, args.ID, organizations.UpdateCustomAttributeInput{
			Required: args.Required,
			Values:   args.Values,
		})
		if err != nil {
			if isMutationFailed(err) {
				return errorResult(fmt.Sprintf("update_custom_attribute failed: %s", mutationErr(err))), nil, nil
			}
			return nil, nil, fmt.Errorf("update_custom_attribute: %w", err)
		}

		result, err := jsonResult(attr)
		if err != nil {
			return nil, nil, err
		}
		return result, attr, nil
	}
}

var DeleteCustomAttributeTool = &mcpsdk.Tool{
	Name:        "delete_custom_attribute",
	Description: "Deletes a custom attribute definition from the organization.",
}

type DeleteCustomAttributeInput struct {
	ID string `json:"id" jsonschema:"The custom attribute ID to delete."`
}

func HandleDeleteCustomAttribute(c *Client) func(context.Context, *mcpsdk.CallToolRequest, DeleteCustomAttributeInput) (*mcpsdk.CallToolResult, any, error) {
	return func(ctx context.Context, _ *mcpsdk.CallToolRequest, args DeleteCustomAttributeInput) (*mcpsdk.CallToolResult, any, error) {
		if args.ID == "" {
			return nil, nil, fmt.Errorf("delete_custom_attribute: id is required")
		}

		_, err := c.Organizations.DeleteCustomAttribute(ctx, args.ID)
		if err != nil {
			if isMutationFailed(err) {
				return errorResult(fmt.Sprintf("delete_custom_attribute failed: %s", mutationErr(err))), nil, nil
			}
			return nil, nil, fmt.Errorf("delete_custom_attribute: %w", err)
		}

		return textResult(fmt.Sprintf("custom attribute %q deleted successfully", args.ID)), nil, nil
	}
}

var ListOrganizationMembersTool = &mcpsdk.Tool{
	Name: "list_organization_members",
	Description: "Lists the members (user accounts) of the current organization, one page at a time. " +
		"Returns up to `page_size` members (default 25, max 100) plus a `next_cursor` for the following page. " +
		"To continue, call again with `cursor` set to the previous `next_cursor`. " +
		"Requires the organization:manageProfile permission — non-admin tokens will get a forbidden error.",
}

type ListOrganizationMembersInput struct {
	Cursor   string `json:"cursor,omitempty"    jsonschema:"Optional. Opaque cursor from a prior call's next_cursor. Omit for the first page."`
	PageSize int    `json:"page_size,omitempty" jsonschema:"Optional. Page size (1-100, default 25)."`
}

func HandleListOrganizationMembers(c *Client) func(context.Context, *mcpsdk.CallToolRequest, ListOrganizationMembersInput) (*mcpsdk.CallToolResult, any, error) {
	return func(ctx context.Context, _ *mcpsdk.CallToolRequest, args ListOrganizationMembersInput) (*mcpsdk.CallToolResult, any, error) {
		page, err := c.Organizations.ListMembersPage(ctx, organizations.ListMembersInput{
			PageSize: clampPageSize(args.PageSize),
			After:    args.Cursor,
		})
		if err != nil {
			return nil, nil, fmt.Errorf("list_organization_members: %w", err)
		}

		out := pageResult(page)
		result, err := jsonResult(out)
		if err != nil {
			return nil, nil, err
		}
		return result, out, nil
	}
}

var ListCustomAttributesTool = &mcpsdk.Tool{
	Name: "list_custom_attributes",
	Description: "Lists the organization's custom attribute definitions, one page at a time. " +
		"Use this to discover which attribute keys are declared (and their scopes and allowed values) before setting attributes on projects/environments/components or writing policy conditions. " +
		"Returns up to `page_size` attributes (default 25, max 100) plus a `next_cursor` for the following page. " +
		"To continue, call again with `cursor` set to the previous `next_cursor`.",
}

type ListCustomAttributesInput struct {
	Cursor   string `json:"cursor,omitempty"    jsonschema:"Optional. Opaque cursor from a prior call's next_cursor. Omit for the first page."`
	PageSize int    `json:"page_size,omitempty" jsonschema:"Optional. Page size (1-100, default 25)."`
}

func HandleListCustomAttributes(c *Client) func(context.Context, *mcpsdk.CallToolRequest, ListCustomAttributesInput) (*mcpsdk.CallToolResult, any, error) {
	return func(ctx context.Context, _ *mcpsdk.CallToolRequest, args ListCustomAttributesInput) (*mcpsdk.CallToolResult, any, error) {
		page, err := c.Organizations.ListCustomAttributesPage(ctx, organizations.ListCustomAttributesInput{
			PageSize: clampPageSize(args.PageSize),
			After:    args.Cursor,
		})
		if err != nil {
			return nil, nil, fmt.Errorf("list_custom_attributes: %w", err)
		}

		out := pageResult(page)
		result, err := jsonResult(out)
		if err != nil {
			return nil, nil, err
		}
		return result, out, nil
	}
}
