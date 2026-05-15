package tools

import (
	"context"
	"fmt"

	"github.com/massdriver-cloud/massdriver-sdk-go/massdriver/platform/instances"
	mcpsdk "github.com/modelcontextprotocol/go-sdk/mcp"
)

var ListInstancesTool = &mcpsdk.Tool{
	Name:        "list_instances",
	Description: "Lists all instances in the organization. Optionally filter by project ID, environment ID, or status.",
}

type ListInstancesInput struct {
	ProjectID     string `json:"project_id"     jsonschema:"Optional. Filter to instances belonging to this project ID."`
	EnvironmentID string `json:"environment_id" jsonschema:"Optional. Filter to instances belonging to this environment ID."`
	Status        string `json:"status"         jsonschema:"Optional. Filter by status: INITIALIZED, PROVISIONED, DECOMMISSIONED, or FAILED."`
}

func HandleListInstances(c *Client) func(context.Context, *mcpsdk.CallToolRequest, ListInstancesInput) (*mcpsdk.CallToolResult, any, error) {
	return func(ctx context.Context, _ *mcpsdk.CallToolRequest, args ListInstancesInput) (*mcpsdk.CallToolResult, any, error) {
		input := instances.ListInput{
			ProjectID:     args.ProjectID,
			EnvironmentID: args.EnvironmentID,
			Status:        instances.Status(args.Status),
		}

		list, err := c.Instances.List(ctx, input)
		if err != nil {
			return nil, nil, fmt.Errorf("list_instances: %w", err)
		}

		result, err := jsonResult(list)
		if err != nil {
			return nil, nil, err
		}
		return result, list, nil
	}
}

var GetInstanceTool = &mcpsdk.Tool{
	Name:        "get_instance",
	Description: "Gets a specific instance by ID, including its environment, project, and current bundle release.",
}

type GetInstanceInput struct {
	ID string `json:"id" jsonschema:"The instance ID."`
}

func HandleGetInstance(c *Client) func(context.Context, *mcpsdk.CallToolRequest, GetInstanceInput) (*mcpsdk.CallToolResult, any, error) {
	return func(ctx context.Context, _ *mcpsdk.CallToolRequest, args GetInstanceInput) (*mcpsdk.CallToolResult, any, error) {
		if args.ID == "" {
			return nil, nil, fmt.Errorf("get_instance: id is required")
		}

		instance, err := c.Instances.Get(ctx, args.ID)
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

var SetInstanceSecretTool = &mcpsdk.Tool{
	Name:        "set_instance_secret",
	Description: "Sets or updates a secret on an instance. Secrets are injected into deployments as environment variables.",
}

type SetInstanceSecretInput struct {
	InstanceID string `json:"instance_id" jsonschema:"The instance ID to set the secret on."`
	Name       string `json:"name"        jsonschema:"Secret name (will be uppercased as an env var)."`
	Value      string `json:"value"       jsonschema:"Secret value."`
}

func HandleSetInstanceSecret(c *Client) func(context.Context, *mcpsdk.CallToolRequest, SetInstanceSecretInput) (*mcpsdk.CallToolResult, any, error) {
	return func(ctx context.Context, _ *mcpsdk.CallToolRequest, args SetInstanceSecretInput) (*mcpsdk.CallToolResult, any, error) {
		if args.InstanceID == "" {
			return nil, nil, fmt.Errorf("set_instance_secret: instance_id is required")
		}
		if args.Name == "" {
			return nil, nil, fmt.Errorf("set_instance_secret: name is required")
		}
		if args.Value == "" {
			return nil, nil, fmt.Errorf("set_instance_secret: value is required")
		}

		secret, err := c.Instances.SetSecret(ctx, args.InstanceID, args.Name, args.Value)
		if err != nil {
			if isMutationFailed(err) {
				return textResult(fmt.Sprintf("set_instance_secret failed: %s", mutationErr(err))), nil, nil
			}
			return nil, nil, fmt.Errorf("set_instance_secret: %w", err)
		}

		result, err := jsonResult(secret)
		if err != nil {
			return nil, nil, err
		}
		return result, secret, nil
	}
}

var RemoveInstanceSecretTool = &mcpsdk.Tool{
	Name:        "remove_instance_secret",
	Description: "Removes a secret from an instance.",
}

type RemoveInstanceSecretInput struct {
	InstanceID string `json:"instance_id" jsonschema:"The instance ID to remove the secret from."`
	Name       string `json:"name"        jsonschema:"Secret name to remove."`
}

func HandleRemoveInstanceSecret(c *Client) func(context.Context, *mcpsdk.CallToolRequest, RemoveInstanceSecretInput) (*mcpsdk.CallToolResult, any, error) {
	return func(ctx context.Context, _ *mcpsdk.CallToolRequest, args RemoveInstanceSecretInput) (*mcpsdk.CallToolResult, any, error) {
		if args.InstanceID == "" {
			return nil, nil, fmt.Errorf("remove_instance_secret: instance_id is required")
		}
		if args.Name == "" {
			return nil, nil, fmt.Errorf("remove_instance_secret: name is required")
		}

		secret, err := c.Instances.RemoveSecret(ctx, args.InstanceID, args.Name)
		if err != nil {
			if isMutationFailed(err) {
				return textResult(fmt.Sprintf("remove_instance_secret failed: %s", mutationErr(err))), nil, nil
			}
			return nil, nil, fmt.Errorf("remove_instance_secret: %w", err)
		}

		result, err := jsonResult(secret)
		if err != nil {
			return nil, nil, err
		}
		return result, secret, nil
	}
}

var ListAlarmsTool = &mcpsdk.Tool{
	Name:        "list_alarms",
	Description: "Lists alarms across instances. Optionally filter by project, environment, component, instance, or bundle.",
}

type ListAlarmsInput struct {
	ProjectID     string `json:"project_id"     jsonschema:"Optional. Filter alarms to this project."`
	EnvironmentID string `json:"environment_id" jsonschema:"Optional. Filter alarms to this environment."`
	ComponentID   string `json:"component_id"   jsonschema:"Optional. Filter alarms to this component."`
	InstanceID    string `json:"instance_id"    jsonschema:"Optional. Filter alarms to this instance."`
	OciRepoName   string `json:"oci_repo_name"  jsonschema:"Optional. Filter alarms to instances of this bundle."`
}

func HandleListAlarms(c *Client) func(context.Context, *mcpsdk.CallToolRequest, ListAlarmsInput) (*mcpsdk.CallToolResult, any, error) {
	return func(ctx context.Context, _ *mcpsdk.CallToolRequest, args ListAlarmsInput) (*mcpsdk.CallToolResult, any, error) {
		input := instances.ListAlarmsInput{
			ProjectID:     args.ProjectID,
			EnvironmentID: args.EnvironmentID,
			ComponentID:   args.ComponentID,
			InstanceID:    args.InstanceID,
			OciRepoName:   args.OciRepoName,
		}

		alarms, err := c.Instances.ListAlarms(ctx, input)
		if err != nil {
			return nil, nil, fmt.Errorf("list_alarms: %w", err)
		}

		result, err := jsonResult(alarms)
		if err != nil {
			return nil, nil, err
		}
		return result, alarms, nil
	}
}

var UpdateInstanceTool = &mcpsdk.Tool{
	Name:        "update_instance",
	Description: "Updates an instance's version pin.",
}

type UpdateInstanceInput struct {
	ID      string `json:"id"      jsonschema:"The instance ID to update."`
	Version string `json:"version" jsonschema:"The version constraint to pin (e.g., '~1.2', '1.2.3')."`
}

func HandleUpdateInstance(c *Client) func(context.Context, *mcpsdk.CallToolRequest, UpdateInstanceInput) (*mcpsdk.CallToolResult, any, error) {
	return func(ctx context.Context, _ *mcpsdk.CallToolRequest, args UpdateInstanceInput) (*mcpsdk.CallToolResult, any, error) {
		if args.ID == "" {
			return nil, nil, fmt.Errorf("update_instance: id is required")
		}
		if args.Version == "" {
			return nil, nil, fmt.Errorf("update_instance: version is required")
		}

		instance, err := c.Instances.Update(ctx, args.ID, instances.UpdateInput{
			Version: args.Version,
		})
		if err != nil {
			if isMutationFailed(err) {
				return textResult(fmt.Sprintf("update_instance failed: %s", mutationErr(err))), nil, nil
			}
			return nil, nil, fmt.Errorf("update_instance: %w", err)
		}

		result, err := jsonResult(instance)
		if err != nil {
			return nil, nil, err
		}
		return result, instance, nil
	}
}
