package tools

import (
	"context"
	"fmt"

	"github.com/massdriver-cloud/massdriver-sdk-go/massdriver/platform/organizations"
	"github.com/massdriver-cloud/massdriver-sdk-go/massdriver/platform/policies"
	"github.com/massdriver-cloud/massdriver-sdk-go/massdriver/platform/types"
	mcpsdk "github.com/modelcontextprotocol/go-sdk/mcp"
)

var GetPolicyTool = &mcpsdk.Tool{
	Name:        "get_policy",
	Description: "Gets a specific ABAC policy by ID.",
}

type GetPolicyInput struct {
	ID string `json:"id" jsonschema:"The policy ID."`
}

func HandleGetPolicy(c *Client) func(context.Context, *mcpsdk.CallToolRequest, GetPolicyInput) (*mcpsdk.CallToolResult, any, error) {
	return func(ctx context.Context, _ *mcpsdk.CallToolRequest, args GetPolicyInput) (*mcpsdk.CallToolResult, any, error) {
		if args.ID == "" {
			return nil, nil, fmt.Errorf("get_policy: id is required")
		}

		policy, err := c.Policies.Get(ctx, args.ID)
		if err != nil {
			return nil, nil, fmt.Errorf("get_policy: %w", err)
		}

		result, err := jsonResult(policy)
		if err != nil {
			return nil, nil, err
		}
		return result, policy, nil
	}
}

var CreatePolicyTool = &mcpsdk.Tool{
	Name:        "create_policy",
	Description: "Creates a new ABAC policy on a group. Policies control what actions group members can perform, optionally scoped by attribute conditions.",
}

type CreatePolicyInput struct {
	GroupID    string                 `json:"group_id"   jsonschema:"The group ID to attach the policy to."`
	Effect     string                 `json:"effect"               jsonschema:"Policy effect: ALLOW or DENY."`
	Actions    []string               `json:"actions,omitempty"    jsonschema:"Optional. Actions the policy applies to (e.g., 'deployment:create'). Empty means all actions."`
	Conditions types.PolicyConditions `json:"conditions,omitempty" jsonschema:"Optional. Attribute conditions scoping the policy (map of attribute key to allowed values)."`
}

func HandleCreatePolicy(c *Client) func(context.Context, *mcpsdk.CallToolRequest, CreatePolicyInput) (*mcpsdk.CallToolResult, any, error) {
	return func(ctx context.Context, _ *mcpsdk.CallToolRequest, args CreatePolicyInput) (*mcpsdk.CallToolResult, any, error) {
		if args.GroupID == "" {
			return nil, nil, fmt.Errorf("create_policy: group_id is required")
		}
		if args.Effect == "" {
			return nil, nil, fmt.Errorf("create_policy: effect is required")
		}

		policy, err := c.Policies.Create(ctx, args.GroupID, policies.CreatePolicyInput{
			Effect:     policies.Effect(args.Effect),
			Actions:    args.Actions,
			Conditions: args.Conditions,
		})
		if err != nil {
			if isMutationFailed(err) {
				return errorResult(fmt.Sprintf("create_policy failed: %s", mutationErr(err))), nil, nil
			}
			return nil, nil, fmt.Errorf("create_policy: %w", err)
		}

		result, err := jsonResult(policy)
		if err != nil {
			return nil, nil, err
		}
		return result, policy, nil
	}
}

var UpdatePolicyTool = &mcpsdk.Tool{
	Name:        "update_policy",
	Description: "Updates an existing ABAC policy's effect, actions, or conditions.",
}

type UpdatePolicyInput struct {
	ID         string                  `json:"id"                   jsonschema:"The policy ID to update."`
	Effect     string                  `json:"effect"               jsonschema:"Policy effect: ALLOW or DENY."`
	Actions    []string                `json:"actions,omitempty"    jsonschema:"Optional. Actions the policy applies to."`
	Conditions *types.PolicyConditions `json:"conditions,omitempty" jsonschema:"Optional. Attribute conditions scoping the policy. Pass null to leave unchanged."`
}

func HandleUpdatePolicy(c *Client) func(context.Context, *mcpsdk.CallToolRequest, UpdatePolicyInput) (*mcpsdk.CallToolResult, any, error) {
	return func(ctx context.Context, _ *mcpsdk.CallToolRequest, args UpdatePolicyInput) (*mcpsdk.CallToolResult, any, error) {
		if args.ID == "" {
			return nil, nil, fmt.Errorf("update_policy: id is required")
		}

		policy, err := c.Policies.Update(ctx, args.ID, policies.UpdatePolicyInput{
			Effect:     policies.Effect(args.Effect),
			Actions:    args.Actions,
			Conditions: args.Conditions,
		})
		if err != nil {
			if isMutationFailed(err) {
				return errorResult(fmt.Sprintf("update_policy failed: %s", mutationErr(err))), nil, nil
			}
			return nil, nil, fmt.Errorf("update_policy: %w", err)
		}

		result, err := jsonResult(policy)
		if err != nil {
			return nil, nil, err
		}
		return result, policy, nil
	}
}

var DeletePolicyTool = &mcpsdk.Tool{
	Name:        "delete_policy",
	Description: "Deletes an ABAC policy.",
}

type DeletePolicyInput struct {
	ID string `json:"id" jsonschema:"The policy ID to delete."`
}

func HandleDeletePolicy(c *Client) func(context.Context, *mcpsdk.CallToolRequest, DeletePolicyInput) (*mcpsdk.CallToolResult, any, error) {
	return func(ctx context.Context, _ *mcpsdk.CallToolRequest, args DeletePolicyInput) (*mcpsdk.CallToolResult, any, error) {
		if args.ID == "" {
			return nil, nil, fmt.Errorf("delete_policy: id is required")
		}

		_, err := c.Policies.Delete(ctx, args.ID)
		if err != nil {
			if isMutationFailed(err) {
				return errorResult(fmt.Sprintf("delete_policy failed: %s", mutationErr(err))), nil, nil
			}
			return nil, nil, fmt.Errorf("delete_policy: %w", err)
		}

		return textResult(fmt.Sprintf("policy %q deleted successfully", args.ID)), nil, nil
	}
}

var ListPolicyActionsTool = &mcpsdk.Tool{
	Name:        "list_policy_actions",
	Description: "Lists all available policy actions that can be used in ABAC policies.",
}

type ListPolicyActionsInput struct{}

func HandleListPolicyActions(c *Client) func(context.Context, *mcpsdk.CallToolRequest, ListPolicyActionsInput) (*mcpsdk.CallToolResult, any, error) {
	return func(ctx context.Context, _ *mcpsdk.CallToolRequest, _ ListPolicyActionsInput) (*mcpsdk.CallToolResult, any, error) {
		actions, err := c.Policies.ListActions(ctx)
		if err != nil {
			return nil, nil, fmt.Errorf("list_policy_actions: %w", err)
		}

		result, err := jsonResult(actions)
		if err != nil {
			return nil, nil, err
		}
		return result, actions, nil
	}
}

var ListPolicyEntitiesTool = &mcpsdk.Tool{
	Name:        "list_policy_entities",
	Description: "Lists all available entity kinds that policies can target.",
}

type ListPolicyEntitiesInput struct{}

func HandleListPolicyEntities(c *Client) func(context.Context, *mcpsdk.CallToolRequest, ListPolicyEntitiesInput) (*mcpsdk.CallToolResult, any, error) {
	return func(ctx context.Context, _ *mcpsdk.CallToolRequest, _ ListPolicyEntitiesInput) (*mcpsdk.CallToolResult, any, error) {
		entities, err := c.Policies.ListEntities(ctx)
		if err != nil {
			return nil, nil, fmt.Errorf("list_policy_entities: %w", err)
		}

		result, err := jsonResult(entities)
		if err != nil {
			return nil, nil, err
		}
		return result, entities, nil
	}
}

var EvaluatePolicyTool = &mcpsdk.Tool{
	Name:        "evaluate_policy",
	Description: "Evaluates whether the current caller is allowed to perform an action on a specific entity.",
}

type EvaluatePolicyInput struct {
	Action   string `json:"action"    jsonschema:"The action to check (e.g., 'deployment:create')."`
	EntityID string `json:"entity_id" jsonschema:"The entity ID to check against."`
}

func HandleEvaluatePolicy(c *Client) func(context.Context, *mcpsdk.CallToolRequest, EvaluatePolicyInput) (*mcpsdk.CallToolResult, any, error) {
	return func(ctx context.Context, _ *mcpsdk.CallToolRequest, args EvaluatePolicyInput) (*mcpsdk.CallToolResult, any, error) {
		if args.Action == "" {
			return nil, nil, fmt.Errorf("evaluate_policy: action is required")
		}
		if args.EntityID == "" {
			return nil, nil, fmt.Errorf("evaluate_policy: entity_id is required")
		}

		decision, err := c.Policies.Evaluate(ctx, args.Action, args.EntityID)
		if err != nil {
			return nil, nil, fmt.Errorf("evaluate_policy: %w", err)
		}

		result, err := jsonResult(decision)
		if err != nil {
			return nil, nil, err
		}
		return result, decision, nil
	}
}

var EvaluatePoliciesBatchTool = &mcpsdk.Tool{
	Name:        "evaluate_policies_batch",
	Description: "Evaluates multiple permission checks in a single request (up to 10).",
}

type EvaluatePoliciesBatchInput struct {
	Checks []PolicyCheck `json:"checks" jsonschema:"List of action/entity pairs to check (max 10)."`
}

type PolicyCheck struct {
	Action   string `json:"action"    jsonschema:"The action to check."`
	EntityID string `json:"entity_id" jsonschema:"The entity ID to check against."`
}

func HandleEvaluatePoliciesBatch(c *Client) func(context.Context, *mcpsdk.CallToolRequest, EvaluatePoliciesBatchInput) (*mcpsdk.CallToolResult, any, error) {
	return func(ctx context.Context, _ *mcpsdk.CallToolRequest, args EvaluatePoliciesBatchInput) (*mcpsdk.CallToolResult, any, error) {
		if len(args.Checks) == 0 {
			return nil, nil, fmt.Errorf("evaluate_policies_batch: checks is required")
		}

		checks := make([]policies.Check, len(args.Checks))
		for i, c := range args.Checks {
			checks[i] = policies.Check{Action: c.Action, EntityID: c.EntityID}
		}

		decisions, err := c.Policies.EvaluateBatch(ctx, checks)
		if err != nil {
			return nil, nil, fmt.Errorf("evaluate_policies_batch: %w", err)
		}

		result, err := jsonResult(decisions)
		if err != nil {
			return nil, nil, err
		}
		return result, decisions, nil
	}
}

var ExplainPolicyTool = &mcpsdk.Tool{
	Name:        "explain_policy",
	Description: "Returns a human-readable explanation of what a policy configuration would permit or deny.",
}

type ExplainPolicyInput struct {
	Effect     string                 `json:"effect"               jsonschema:"Policy effect: ALLOW or DENY."`
	Actions    []string               `json:"actions,omitempty"    jsonschema:"Optional. Actions the policy applies to."`
	Conditions types.PolicyConditions `json:"conditions,omitempty" jsonschema:"Optional. Attribute conditions scoping the policy."`
}

func HandleExplainPolicy(c *Client) func(context.Context, *mcpsdk.CallToolRequest, ExplainPolicyInput) (*mcpsdk.CallToolResult, any, error) {
	return func(ctx context.Context, _ *mcpsdk.CallToolRequest, args ExplainPolicyInput) (*mcpsdk.CallToolResult, any, error) {
		if args.Effect == "" {
			return nil, nil, fmt.Errorf("explain_policy: effect is required")
		}

		lines, err := c.Policies.Explain(ctx, policies.ExplainInput{
			Effect:     policies.Effect(args.Effect),
			Actions:    args.Actions,
			Conditions: args.Conditions,
		})
		if err != nil {
			return nil, nil, fmt.Errorf("explain_policy: %w", err)
		}

		result, err := jsonResult(lines)
		if err != nil {
			return nil, nil, err
		}
		return result, lines, nil
	}
}

var GetPolicyAttributeSchemaTool = &mcpsdk.Tool{
	Name:        "get_policy_attribute_schema",
	Description: "Gets the JSON Schema describing valid condition attributes for a given policy action.",
}

type GetPolicyAttributeSchemaInput struct {
	Action string `json:"action" jsonschema:"The policy action to get the attribute schema for."`
}

func HandleGetPolicyAttributeSchema(c *Client) func(context.Context, *mcpsdk.CallToolRequest, GetPolicyAttributeSchemaInput) (*mcpsdk.CallToolResult, any, error) {
	return func(ctx context.Context, _ *mcpsdk.CallToolRequest, args GetPolicyAttributeSchemaInput) (*mcpsdk.CallToolResult, any, error) {
		if args.Action == "" {
			return nil, nil, fmt.Errorf("get_policy_attribute_schema: action is required")
		}

		schema, err := c.Policies.CustomAttributeSchema(ctx, args.Action)
		if err != nil {
			return nil, nil, fmt.Errorf("get_policy_attribute_schema: %w", err)
		}

		return textResult(string(schema)), nil, nil
	}
}

var ListPolicyAttributeValuesTool = &mcpsdk.Tool{
	Name:        "list_policy_attribute_values",
	Description: "Lists the permitted values for a custom attribute key at a given scope.",
}

type ListPolicyAttributeValuesInput struct {
	Scope string `json:"scope" jsonschema:"Attribute scope: PROJECT, ENVIRONMENT, COMPONENT, or REPO."`
	Key   string `json:"key"   jsonschema:"The attribute key to list values for."`
}

func HandleListPolicyAttributeValues(c *Client) func(context.Context, *mcpsdk.CallToolRequest, ListPolicyAttributeValuesInput) (*mcpsdk.CallToolResult, any, error) {
	return func(ctx context.Context, _ *mcpsdk.CallToolRequest, args ListPolicyAttributeValuesInput) (*mcpsdk.CallToolResult, any, error) {
		if args.Scope == "" {
			return nil, nil, fmt.Errorf("list_policy_attribute_values: scope is required")
		}
		if args.Key == "" {
			return nil, nil, fmt.Errorf("list_policy_attribute_values: key is required")
		}

		values, err := c.Policies.CustomAttributeValues(ctx, organizations.AttributeScope(args.Scope), args.Key)
		if err != nil {
			return nil, nil, fmt.Errorf("list_policy_attribute_values: %w", err)
		}

		result, err := jsonResult(values)
		if err != nil {
			return nil, nil, err
		}
		return result, values, nil
	}
}
