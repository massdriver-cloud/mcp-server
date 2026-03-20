package tools

import (
	"context"
	"fmt"

	mdclient "github.com/massdriver-cloud/massdriver-sdk-go/massdriver/client"
	"github.com/massdriver-cloud/mcp-server/internal/api"
	mcpsdk "github.com/modelcontextprotocol/go-sdk/mcp"
)

var ListDeploymentsTool = &mcpsdk.Tool{
	Name:        "list_deployments",
	Description: "Lists deployments in the organization, newest first. Optionally filter by instance ID, status, or action.",
}

type ListDeploymentsInput struct {
	InstanceID string `json:"instance_id" jsonschema:"Optional. Filter to deployments for this instance ID."`
	Status     string `json:"status"      jsonschema:"Optional. Filter by status: PENDING, RUNNING, COMPLETED, FAILED, or ABORTED."`
	Action     string `json:"action"      jsonschema:"Optional. Filter by action: PROVISION, DECOMMISSION, or PLAN."`
}

func HandleListDeployments(c *mdclient.Client) func(context.Context, *mcpsdk.CallToolRequest, ListDeploymentsInput) (*mcpsdk.CallToolResult, any, error) {
	return func(ctx context.Context, _ *mcpsdk.CallToolRequest, args ListDeploymentsInput) (*mcpsdk.CallToolResult, any, error) {
		filter := api.DeploymentsFilter{}
		if args.InstanceID != "" {
			filter.InstanceId = api.IdFilter{Eq: args.InstanceID}
		}
		if args.Status != "" {
			filter.Status = api.DeploymentStatusFilter{Eq: api.DeploymentStatus(args.Status)}
		}
		if args.Action != "" {
			filter.Action = api.DeploymentActionFilter{Eq: api.DeploymentAction(args.Action)}
		}

		deployments, err := api.ListDeployments(ctx, c, filter)
		if err != nil {
			return nil, nil, fmt.Errorf("list_deployments: %w", err)
		}

		result, err := jsonResult(deployments)
		if err != nil {
			return nil, nil, err
		}
		return result, deployments, nil
	}
}

var GetDeploymentTool = &mcpsdk.Tool{
	Name:        "get_deployment",
	Description: "Gets a specific deployment by ID, including its status, action, elapsed time, and associated instance.",
}

type GetDeploymentInput struct {
	ID string `json:"id" jsonschema:"The deployment ID."`
}

func HandleGetDeployment(c *mdclient.Client) func(context.Context, *mcpsdk.CallToolRequest, GetDeploymentInput) (*mcpsdk.CallToolResult, any, error) {
	return func(ctx context.Context, _ *mcpsdk.CallToolRequest, args GetDeploymentInput) (*mcpsdk.CallToolResult, any, error) {
		if args.ID == "" {
			return nil, nil, fmt.Errorf("get_deployment: id is required")
		}

		deployment, err := api.GetDeployment(ctx, c, args.ID)
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
