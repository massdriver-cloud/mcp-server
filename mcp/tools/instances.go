package tools

import (
	"context"
	"fmt"

	mdclient "github.com/massdriver-cloud/massdriver-sdk-go/massdriver/client"
	"github.com/massdriver-cloud/mcp-server/internal/api"
	mcpsdk "github.com/modelcontextprotocol/go-sdk/mcp"
)

var ListInstancesTool = &mcpsdk.Tool{
	Name:        "list_instances",
	Description: "Lists all instances in the organization. Optionally filter by project ID, environment ID, or status.",
}

type ListInstancesInput struct {
	ProjectID     string `json:"project_id"     jsonschema:"Optional. Filter to instances belonging to this project ID."`
	EnvironmentID string `json:"environment_id" jsonschema:"Optional. Filter to instances belonging to this environment ID."`
	Status        string `json:"status"         jsonschema:"Optional. Filter by status: INITIALIZED, PROVISIONED, DECOMMISSIONED, FAILED, or EXTERNAL."`
}

func HandleListInstances(c *mdclient.Client) func(context.Context, *mcpsdk.CallToolRequest, ListInstancesInput) (*mcpsdk.CallToolResult, any, error) {
	return func(ctx context.Context, _ *mcpsdk.CallToolRequest, args ListInstancesInput) (*mcpsdk.CallToolResult, any, error) {
		filter := api.InstancesFilter{}
		if args.ProjectID != "" {
			filter.ProjectId = api.IdFilter{Eq: args.ProjectID}
		}
		if args.EnvironmentID != "" {
			filter.EnvironmentId = api.IdFilter{Eq: args.EnvironmentID}
		}
		if args.Status != "" {
			filter.Status = api.InstanceStatusFilter{Eq: api.InstanceStatus(args.Status)}
		}

		instances, err := api.ListInstances(ctx, c, filter)
		if err != nil {
			return nil, nil, fmt.Errorf("list_instances: %w", err)
		}

		result, err := jsonResult(instances)
		if err != nil {
			return nil, nil, err
		}
		return result, instances, nil
	}
}

var GetInstanceTool = &mcpsdk.Tool{
	Name:        "get_instance",
	Description: "Gets a specific instance by ID, including its environment, project, and current bundle release.",
}

type GetInstanceInput struct {
	ID string `json:"id" jsonschema:"The instance ID."`
}

func HandleGetInstance(c *mdclient.Client) func(context.Context, *mcpsdk.CallToolRequest, GetInstanceInput) (*mcpsdk.CallToolResult, any, error) {
	return func(ctx context.Context, _ *mcpsdk.CallToolRequest, args GetInstanceInput) (*mcpsdk.CallToolResult, any, error) {
		if args.ID == "" {
			return nil, nil, fmt.Errorf("get_instance: id is required")
		}

		instance, err := api.GetInstance(ctx, c, args.ID)
		if err != nil {
			return nil, nil, fmt.Errorf("get_instance: %w", err)
		}

		result, err := jsonResult(instance)
		if err != nil {
			return nil, nil, err
		}
		return result, instance, nil
	}
}
