package tools

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"time"

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
	Name: "get_deployment_logs",
	Description: "Gets the logs for a specific deployment. By default returns a snapshot of the logs so far. " +
		"Set follow=true to block until the deployment reaches a terminal status (COMPLETED, FAILED, ABORTED, or REJECTED), then return the final status plus the complete logs — " +
		"use this after create_deployment or approve_deployment to deploy and see the result in a single call.",
}

type GetDeploymentLogsInput struct {
	ID             string `json:"id"                        jsonschema:"The deployment ID to fetch logs for."`
	Follow         bool   `json:"follow,omitempty"          jsonschema:"Optional. If true, wait until the deployment finishes and return the final status plus complete logs. Default false (snapshot of logs so far)."`
	TimeoutSeconds int    `json:"timeout_seconds,omitempty" jsonschema:"Optional. When follow is true, maximum seconds to wait (default 300, max 600)."`
}

func HandleGetDeploymentLogs(c *Client) func(context.Context, *mcpsdk.CallToolRequest, GetDeploymentLogsInput) (*mcpsdk.CallToolResult, any, error) {
	return func(ctx context.Context, _ *mcpsdk.CallToolRequest, args GetDeploymentLogsInput) (*mcpsdk.CallToolResult, any, error) {
		if args.ID == "" {
			return nil, nil, fmt.Errorf("get_deployment_logs: id is required")
		}

		// Snapshot mode: return whatever logs exist right now.
		if !args.Follow {
			logs, err := c.Deployments.GetLogs(ctx, args.ID)
			if err != nil {
				return nil, nil, fmt.Errorf("get_deployment_logs: %w", err)
			}
			if logs == "" {
				return textResult("no logs available"), nil, nil
			}
			return textResult(logs), nil, nil
		}

		// Follow mode: tail until the deployment terminates (or we time out),
		// then return the final status with the aggregated logs.
		timeout := args.TimeoutSeconds
		if timeout <= 0 {
			timeout = 300
		}
		if timeout > 600 {
			timeout = 600
		}
		waitCtx, cancel := context.WithTimeout(ctx, time.Duration(timeout)*time.Second)
		defer cancel()

		var buf bytes.Buffer
		err := c.Deployments.TailLogs(waitCtx, args.ID, &buf)
		switch {
		case err == nil, errors.Is(err, context.DeadlineExceeded), errors.Is(err, context.Canceled):
			// Streamed to terminal, or timed out with partial logs — both fine.
		case errors.Is(err, deployments.ErrStreamingRequiresPAT):
			// Live streaming needs bearer auth we don't have; fall back to polling
			// the status to terminal, then take a one-shot log snapshot.
			if _, perr := pollUntilTerminal(waitCtx, c, args.ID); perr != nil {
				return nil, nil, fmt.Errorf("get_deployment_logs: %w", perr)
			}
			logs, lerr := c.Deployments.GetLogs(ctx, args.ID)
			if lerr != nil {
				return nil, nil, fmt.Errorf("get_deployment_logs: %w", lerr)
			}
			buf.Reset()
			buf.WriteString(logs)
		default:
			return nil, nil, fmt.Errorf("get_deployment_logs: %w", err)
		}

		status := "UNKNOWN"
		if dep, derr := c.Deployments.Get(ctx, args.ID); derr == nil {
			status = dep.Status
		}

		var header string
		if isTerminalDeploymentStatus(status) {
			header = fmt.Sprintf("deployment %s finished with status: %s\n\n", args.ID, status)
		} else {
			header = fmt.Sprintf("deployment %s did not finish within %ds (current status: %s)\n\n", args.ID, timeout, status)
		}
		logs := buf.String()
		if logs == "" {
			logs = "(no logs)"
		}
		return textResult(header + logs), nil, nil
	}
}

var CreateDeploymentTool = &mcpsdk.Tool{
	Name: "create_deployment",
	Description: "Creates and starts a deployment for an instance. Use action PROVISION to deploy, DECOMMISSION to tear down, or PLAN to preview changes. " +
		"The params map must conform to the instance's params schema (see get_instance.paramsSchema). Use get_deployment_logs with follow=true to block until it finishes and see the result.",
}

type CreateDeploymentInput struct {
	InstanceID string         `json:"instance_id" jsonschema:"The instance ID to deploy."`
	Action     string         `json:"action"            jsonschema:"Deployment action: PROVISION, DECOMMISSION, or PLAN."`
	Params     map[string]any `json:"params,omitempty"  jsonschema:"Optional. Parameter overrides for the deployment."`
	Message    string         `json:"message,omitempty" jsonschema:"Optional. Deployment message or reason."`
}

func HandleCreateDeployment(c *Client) func(context.Context, *mcpsdk.CallToolRequest, CreateDeploymentInput) (*mcpsdk.CallToolResult, any, error) {
	return func(ctx context.Context, _ *mcpsdk.CallToolRequest, args CreateDeploymentInput) (*mcpsdk.CallToolResult, any, error) {
		if args.InstanceID == "" {
			return nil, nil, fmt.Errorf("create_deployment: instance_id is required")
		}
		if args.Action == "" {
			return nil, nil, fmt.Errorf("create_deployment: action is required")
		}

		// The API requires a non-null params map; default an omitted value to an
		// empty map so callers get a clear "required property" validation error
		// rather than a cryptic GraphQL "Expected type Map!, found null".
		params := args.Params
		if params == nil {
			params = map[string]any{}
		}

		deployment, err := c.Deployments.Create(ctx, args.InstanceID, deployments.CreateInput{
			Action:  deployments.Action(args.Action),
			Params:  params,
			Message: args.Message,
		})
		if err != nil {
			if isMutationFailed(err) {
				return errorResult(fmt.Sprintf("create_deployment failed: %s", mutationErr(err))), nil, nil
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
				return errorResult(fmt.Sprintf("abort_deployment failed: %s", mutationErr(err))), nil, nil
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
	Name: "propose_deployment",
	Description: "Proposes a deployment for approval. Only supports PROVISION and DECOMMISSION actions. The deployment enters PROPOSED status and must be approved or rejected. " +
		"The params map must conform to the instance's params schema (see get_instance.paramsSchema).",
}

type ProposeDeploymentInput struct {
	InstanceID string         `json:"instance_id" jsonschema:"The instance ID to deploy."`
	Action     string         `json:"action"            jsonschema:"Deployment action: PROVISION or DECOMMISSION."`
	Params     map[string]any `json:"params,omitempty"  jsonschema:"Optional. Parameter overrides for the deployment."`
	Message    string         `json:"message,omitempty" jsonschema:"Optional. Deployment message or reason."`
}

func HandleProposeDeployment(c *Client) func(context.Context, *mcpsdk.CallToolRequest, ProposeDeploymentInput) (*mcpsdk.CallToolResult, any, error) {
	return func(ctx context.Context, _ *mcpsdk.CallToolRequest, args ProposeDeploymentInput) (*mcpsdk.CallToolResult, any, error) {
		if args.InstanceID == "" {
			return nil, nil, fmt.Errorf("propose_deployment: instance_id is required")
		}
		if args.Action == "" {
			return nil, nil, fmt.Errorf("propose_deployment: action is required")
		}

		// See HandleCreateDeployment: the API requires a non-null params map.
		params := args.Params
		if params == nil {
			params = map[string]any{}
		}

		deployment, err := c.Deployments.Propose(ctx, args.InstanceID, deployments.ProposeInput{
			Action:  deployments.Action(args.Action),
			Params:  params,
			Message: args.Message,
		})
		if err != nil {
			if isMutationFailed(err) {
				return errorResult(fmt.Sprintf("propose_deployment failed: %s", mutationErr(err))), nil, nil
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
				return errorResult(fmt.Sprintf("approve_deployment failed: %s", mutationErr(err))), nil, nil
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
				return errorResult(fmt.Sprintf("reject_deployment failed: %s", mutationErr(err))), nil, nil
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

var PlanDeploymentTool = &mcpsdk.Tool{
	Name: "plan_deployment",
	Description: "Runs a fresh PLAN (dry-run preview) against an existing deployment's parameters. " +
		"This is a read-only preview: it copies the source deployment's params onto a new PLAN deployment and " +
		"changes nothing on the source deployment, the instance's saved configuration, or any other deployment. " +
		"The source deployment can be in any status — use it to preview a proposal before approving, replay a " +
		"completed deployment, or scope out a rollback against an older snapshot. Returns the new PLAN deployment.",
}

type PlanDeploymentInput struct {
	ID string `json:"id" jsonschema:"The ID of the source deployment whose params should be planned."`
}

func HandlePlanDeployment(c *Client) func(context.Context, *mcpsdk.CallToolRequest, PlanDeploymentInput) (*mcpsdk.CallToolResult, any, error) {
	return func(ctx context.Context, _ *mcpsdk.CallToolRequest, args PlanDeploymentInput) (*mcpsdk.CallToolResult, any, error) {
		if args.ID == "" {
			return nil, nil, fmt.Errorf("plan_deployment: id is required")
		}

		deployment, err := c.Deployments.Plan(ctx, args.ID)
		if err != nil {
			if isMutationFailed(err) {
				return errorResult(fmt.Sprintf("plan_deployment failed: %s", mutationErr(err))), nil, nil
			}
			return nil, nil, fmt.Errorf("plan_deployment: %w", err)
		}

		result, err := jsonResult(deployment)
		if err != nil {
			return nil, nil, err
		}
		return result, deployment, nil
	}
}

var RollbackDeploymentTool = &mcpsdk.Tool{
	Name: "rollback_deployment",
	Description: "Proposes a rollback to a past deployment's exact state. The source deployment is the historical " +
		"run to return to and must be a COMPLETED PROVISION. This creates a new PROPOSED PROVISION deployment " +
		"that snapshots the source's params, connection wiring, bundle version, and release — it does NOT apply " +
		"anything on its own. The proposal goes through the normal review flow: preview it with plan_deployment, " +
		"then approve_deployment to apply (pinning the instance to the source's exact configuration) or " +
		"reject_deployment to discard.",
}

type RollbackDeploymentInput struct {
	ID string `json:"id" jsonschema:"The ID of the source deployment (a COMPLETED PROVISION) to roll back to."`
}

func HandleRollbackDeployment(c *Client) func(context.Context, *mcpsdk.CallToolRequest, RollbackDeploymentInput) (*mcpsdk.CallToolResult, any, error) {
	return func(ctx context.Context, _ *mcpsdk.CallToolRequest, args RollbackDeploymentInput) (*mcpsdk.CallToolResult, any, error) {
		if args.ID == "" {
			return nil, nil, fmt.Errorf("rollback_deployment: id is required")
		}

		deployment, err := c.Deployments.Rollback(ctx, args.ID)
		if err != nil {
			if isMutationFailed(err) {
				return errorResult(fmt.Sprintf("rollback_deployment failed: %s", mutationErr(err))), nil, nil
			}
			return nil, nil, fmt.Errorf("rollback_deployment: %w", err)
		}

		result, err := jsonResult(deployment)
		if err != nil {
			return nil, nil, err
		}
		return result, deployment, nil
	}
}

var CompareDeploymentsTool = &mcpsdk.Tool{
	Name: "compare_deployments",
	Description: "Compares two deployments' snapshotted configuration: the bundle version on each side plus a flat, " +
		"leaf-level diff of their params. Use it to audit what a deployment changed or to contrast deployments from " +
		"different points in time. Runtime state, logs, and produced artifacts are out of scope. The two deployments " +
		"need not target the same instance, though comparing unrelated instances reports every param as present on " +
		"one side only.",
}

type CompareDeploymentsInput struct {
	SourceID string `json:"source_id" jsonschema:"The source (baseline) deployment ID."`
	TargetID string `json:"target_id" jsonschema:"The target deployment ID to compare against the source."`
}

func HandleCompareDeployments(c *Client) func(context.Context, *mcpsdk.CallToolRequest, CompareDeploymentsInput) (*mcpsdk.CallToolResult, any, error) {
	return func(ctx context.Context, _ *mcpsdk.CallToolRequest, args CompareDeploymentsInput) (*mcpsdk.CallToolResult, any, error) {
		if args.SourceID == "" {
			return nil, nil, fmt.Errorf("compare_deployments: source_id is required")
		}
		if args.TargetID == "" {
			return nil, nil, fmt.Errorf("compare_deployments: target_id is required")
		}

		comparison, err := c.Deployments.Compare(ctx, args.SourceID, args.TargetID)
		if err != nil {
			return nil, nil, fmt.Errorf("compare_deployments: %w", err)
		}

		result, err := jsonResult(comparison)
		if err != nil {
			return nil, nil, err
		}
		return result, comparison, nil
	}
}

// isTerminalDeploymentStatus reports whether a deployment is done and its
// status will not change further.
func isTerminalDeploymentStatus(s string) bool {
	switch deployments.Status(s) {
	case deployments.StatusCompleted, deployments.StatusFailed, deployments.StatusAborted, deployments.StatusRejected:
		return true
	default:
		return false
	}
}

// pollUntilTerminal polls a deployment until it reaches a terminal status or
// ctx is done, returning the last observed deployment. It is the fallback for
// follow mode when live log streaming is unavailable.
func pollUntilTerminal(ctx context.Context, c *Client, id string) (*deployments.Deployment, error) {
	const pollInterval = 3 * time.Second
	for {
		dep, err := c.Deployments.Get(ctx, id)
		if err != nil {
			return nil, err
		}
		if isTerminalDeploymentStatus(dep.Status) {
			return dep, nil
		}
		select {
		case <-ctx.Done():
			return dep, nil
		case <-time.After(pollInterval):
		}
	}
}
