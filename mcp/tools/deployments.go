package tools

import (
	"context"
	"fmt"

	"github.com/massdriver-cloud/massdriver-sdk-go/massdriver/platform/deployments"
	mcpsdk "github.com/modelcontextprotocol/go-sdk/mcp"
)

var ListDeploymentsTool = &mcpsdk.Tool{
	Name: "list_deployments",
	Description: "Lists deployments in the organization, newest first, one page at a time. " +
		"STRONGLY PREFER filtering by `instance_id`, `status`, or `action` — unfiltered lists span the entire org history. " +
		"Returns up to `page_size` deployments (default 25, max 100) plus a `next_cursor` for the following page. " +
		"To continue, call again with `cursor` set to the previous `next_cursor`. " +
		"Do NOT paginate to exhaustion unless the user explicitly asked for every deployment.",
}

type ListDeploymentsInput struct {
	InstanceID string `json:"instance_id,omitempty" jsonschema:"Optional. Filter to deployments for this instance ID."`
	Status     string `json:"status,omitempty"      jsonschema:"Optional. Filter by status: PROPOSED, APPROVED, PENDING, RUNNING, COMPLETED, FAILED, REJECTED, or ABORTED."`
	Action     string `json:"action,omitempty"      jsonschema:"Optional. Filter by action: PROVISION, DECOMMISSION, or PLAN."`
	Cursor     string `json:"cursor,omitempty"      jsonschema:"Optional. Opaque cursor from a prior call's next_cursor. Omit for the first page."`
	PageSize   int    `json:"page_size,omitempty"   jsonschema:"Optional. Page size (1-100, default 25)."`
}

func HandleListDeployments(c *Client) func(context.Context, *mcpsdk.CallToolRequest, ListDeploymentsInput) (*mcpsdk.CallToolResult, any, error) {
	return func(ctx context.Context, _ *mcpsdk.CallToolRequest, args ListDeploymentsInput) (*mcpsdk.CallToolResult, any, error) {
		page, err := c.Deployments.ListPage(ctx, deployments.ListInput{
			InstanceID: args.InstanceID,
			Status:     deployments.Status(args.Status),
			Action:     deployments.Action(args.Action),
			PageSize:   clampPageSize(args.PageSize),
			After:      args.Cursor,
		})
		if err != nil {
			return nil, nil, fmt.Errorf("list_deployments: %w", err)
		}

		out := pageResult(page)
		result, err := jsonResult(out)
		if err != nil {
			return nil, nil, err
		}
		return result, out, nil
	}
}

var GetDeploymentTool = &mcpsdk.Tool{
	Name:        "get_deployment",
	Description: "Gets a specific deployment by ID, including its status, action, elapsed time, and associated instance.",
}

type GetDeploymentInput struct {
	ID string `json:"id" jsonschema:"The deployment ID."`
}

func HandleGetDeployment(c *Client) func(context.Context, *mcpsdk.CallToolRequest, GetDeploymentInput) (*mcpsdk.CallToolResult, any, error) {
	return func(ctx context.Context, _ *mcpsdk.CallToolRequest, args GetDeploymentInput) (*mcpsdk.CallToolResult, any, error) {
		if args.ID == "" {
			return nil, nil, fmt.Errorf("get_deployment: id is required")
		}

		deployment, err := c.Deployments.Get(ctx, args.ID)
		if err != nil {
			return nil, nil, fmt.Errorf("get_deployment: %w", err)
		}

		result, err := jsonResult(deployment)
		if err != nil {
			return nil, nil, err
		}
		return result, deployment, nil
	}
}

var GetDeploymentLogsTool = &mcpsdk.Tool{
	Name:        "get_deployment_logs",
	Description: "Gets the concatenated logs for a specific deployment.",
}

type GetDeploymentLogsInput struct {
	ID string `json:"id" jsonschema:"The deployment ID to fetch logs for."`
}

func HandleGetDeploymentLogs(c *Client) func(context.Context, *mcpsdk.CallToolRequest, GetDeploymentLogsInput) (*mcpsdk.CallToolResult, any, error) {
	return func(ctx context.Context, _ *mcpsdk.CallToolRequest, args GetDeploymentLogsInput) (*mcpsdk.CallToolResult, any, error) {
		if args.ID == "" {
			return nil, nil, fmt.Errorf("get_deployment_logs: id is required")
		}

		logs, err := c.Deployments.GetLogs(ctx, args.ID)
		if err != nil {
			return nil, nil, fmt.Errorf("get_deployment_logs: %w", err)
		}

		if logs == "" {
			return textResult("no logs available"), nil, nil
		}
		return textResult(logs), nil, nil
	}
}

var CreateDeploymentTool = &mcpsdk.Tool{
	Name:        "create_deployment",
	Description: "Creates and starts a deployment for an instance. Use action PROVISION to deploy, DECOMMISSION to tear down, or PLAN to preview changes.",
}

type CreateDeploymentInput struct {
	InstanceID string         `json:"instance_id" jsonschema:"The instance ID to deploy."`
	Action     string         `json:"action"      jsonschema:"Deployment action: PROVISION, DECOMMISSION, or PLAN."`
	Params     map[string]any `json:"params"      jsonschema:"Optional. Parameter overrides for the deployment."`
	Message    string         `json:"message"     jsonschema:"Optional. Deployment message or reason."`
}

func HandleCreateDeployment(c *Client) func(context.Context, *mcpsdk.CallToolRequest, CreateDeploymentInput) (*mcpsdk.CallToolResult, any, error) {
	return func(ctx context.Context, _ *mcpsdk.CallToolRequest, args CreateDeploymentInput) (*mcpsdk.CallToolResult, any, error) {
		if args.InstanceID == "" {
			return nil, nil, fmt.Errorf("create_deployment: instance_id is required")
		}
		if args.Action == "" {
			return nil, nil, fmt.Errorf("create_deployment: action is required")
		}

		deployment, err := c.Deployments.Create(ctx, args.InstanceID, deployments.CreateInput{
			Action:  deployments.Action(args.Action),
			Params:  args.Params,
			Message: args.Message,
		})
		if err != nil {
			if isMutationFailed(err) {
				return textResult(fmt.Sprintf("create_deployment failed: %s", mutationErr(err))), nil, nil
			}
			return nil, nil, fmt.Errorf("create_deployment: %w", err)
		}

		result, err := jsonResult(deployment)
		if err != nil {
			return nil, nil, err
		}
		return result, deployment, nil
	}
}

var AbortDeploymentTool = &mcpsdk.Tool{
	Name:        "abort_deployment",
	Description: "Aborts a running deployment.",
}

type AbortDeploymentInput struct {
	ID string `json:"id" jsonschema:"The deployment ID to abort."`
}

func HandleAbortDeployment(c *Client) func(context.Context, *mcpsdk.CallToolRequest, AbortDeploymentInput) (*mcpsdk.CallToolResult, any, error) {
	return func(ctx context.Context, _ *mcpsdk.CallToolRequest, args AbortDeploymentInput) (*mcpsdk.CallToolResult, any, error) {
		if args.ID == "" {
			return nil, nil, fmt.Errorf("abort_deployment: id is required")
		}

		deployment, err := c.Deployments.Abort(ctx, args.ID)
		if err != nil {
			if isMutationFailed(err) {
				return textResult(fmt.Sprintf("abort_deployment failed: %s", mutationErr(err))), nil, nil
			}
			return nil, nil, fmt.Errorf("abort_deployment: %w", err)
		}

		result, err := jsonResult(deployment)
		if err != nil {
			return nil, nil, err
		}
		return result, deployment, nil
	}
}

var ProposeDeploymentTool = &mcpsdk.Tool{
	Name:        "propose_deployment",
	Description: "Proposes a deployment for approval. Only supports PROVISION and DECOMMISSION actions. The deployment enters PROPOSED status and must be approved or rejected.",
}

type ProposeDeploymentInput struct {
	InstanceID string         `json:"instance_id" jsonschema:"The instance ID to deploy."`
	Action     string         `json:"action"      jsonschema:"Deployment action: PROVISION or DECOMMISSION."`
	Params     map[string]any `json:"params"      jsonschema:"Optional. Parameter overrides for the deployment."`
	Message    string         `json:"message"     jsonschema:"Optional. Deployment message or reason."`
}

func HandleProposeDeployment(c *Client) func(context.Context, *mcpsdk.CallToolRequest, ProposeDeploymentInput) (*mcpsdk.CallToolResult, any, error) {
	return func(ctx context.Context, _ *mcpsdk.CallToolRequest, args ProposeDeploymentInput) (*mcpsdk.CallToolResult, any, error) {
		if args.InstanceID == "" {
			return nil, nil, fmt.Errorf("propose_deployment: instance_id is required")
		}
		if args.Action == "" {
			return nil, nil, fmt.Errorf("propose_deployment: action is required")
		}

		deployment, err := c.Deployments.Propose(ctx, args.InstanceID, deployments.ProposeInput{
			Action:  deployments.Action(args.Action),
			Params:  args.Params,
			Message: args.Message,
		})
		if err != nil {
			if isMutationFailed(err) {
				return textResult(fmt.Sprintf("propose_deployment failed: %s", mutationErr(err))), nil, nil
			}
			return nil, nil, fmt.Errorf("propose_deployment: %w", err)
		}

		result, err := jsonResult(deployment)
		if err != nil {
			return nil, nil, err
		}
		return result, deployment, nil
	}
}

var ApproveDeploymentTool = &mcpsdk.Tool{
	Name:        "approve_deployment",
	Description: "Approves a proposed deployment, allowing it to proceed.",
}

type ApproveDeploymentInput struct {
	ID string `json:"id" jsonschema:"The deployment ID to approve."`
}

func HandleApproveDeployment(c *Client) func(context.Context, *mcpsdk.CallToolRequest, ApproveDeploymentInput) (*mcpsdk.CallToolResult, any, error) {
	return func(ctx context.Context, _ *mcpsdk.CallToolRequest, args ApproveDeploymentInput) (*mcpsdk.CallToolResult, any, error) {
		if args.ID == "" {
			return nil, nil, fmt.Errorf("approve_deployment: id is required")
		}

		deployment, err := c.Deployments.Approve(ctx, args.ID)
		if err != nil {
			if isMutationFailed(err) {
				return textResult(fmt.Sprintf("approve_deployment failed: %s", mutationErr(err))), nil, nil
			}
			return nil, nil, fmt.Errorf("approve_deployment: %w", err)
		}

		result, err := jsonResult(deployment)
		if err != nil {
			return nil, nil, err
		}
		return result, deployment, nil
	}
}

var RejectDeploymentTool = &mcpsdk.Tool{
	Name:        "reject_deployment",
	Description: "Rejects a proposed deployment, preventing it from proceeding.",
}

type RejectDeploymentInput struct {
	ID string `json:"id" jsonschema:"The deployment ID to reject."`
}

func HandleRejectDeployment(c *Client) func(context.Context, *mcpsdk.CallToolRequest, RejectDeploymentInput) (*mcpsdk.CallToolResult, any, error) {
	return func(ctx context.Context, _ *mcpsdk.CallToolRequest, args RejectDeploymentInput) (*mcpsdk.CallToolResult, any, error) {
		if args.ID == "" {
			return nil, nil, fmt.Errorf("reject_deployment: id is required")
		}

		deployment, err := c.Deployments.Reject(ctx, args.ID)
		if err != nil {
			if isMutationFailed(err) {
				return textResult(fmt.Sprintf("reject_deployment failed: %s", mutationErr(err))), nil, nil
			}
			return nil, nil, fmt.Errorf("reject_deployment: %w", err)
		}

		result, err := jsonResult(deployment)
		if err != nil {
			return nil, nil, err
		}
		return result, deployment, nil
	}
}
