package tools

import (
	"context"
	"fmt"

	"github.com/massdriver-cloud/massdriver-sdk-go/massdriver/platform/organizations"
	mcpsdk "github.com/modelcontextprotocol/go-sdk/mcp"
)

var GetOrganizationTool = &mcpsdk.Tool{
	Name:        "get_organization",
	Description: "Gets the current organization's details, including subscription status, custom attributes, and member counts.",
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

var CreateCustomAttributeTool = &mcpsdk.Tool{
	Name:        "create_custom_attribute",
	Description: "Creates a custom attribute definition for the organization.",
}

type CreateCustomAttributeInput struct {
	Key      string   `json:"key"      jsonschema:"Attribute key name."`
	Scope    string   `json:"scope"    jsonschema:"Attribute scope: PROJECT, ENVIRONMENT, COMPONENT, or REPO."`
	Required *bool    `json:"required,omitempty" jsonschema:"Optional. Whether the attribute is required."`
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

		attr, err := c.Organizations.CreateCustomAttribute(ctx, organizations.CreateCustomAttributeInput{
			Key:      args.Key,
			Scope:    organizations.AttributeScope(args.Scope),
			Required: args.Required,
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
