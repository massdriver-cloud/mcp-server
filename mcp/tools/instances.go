package tools

import (
	"context"
	"fmt"

	"github.com/massdriver-cloud/massdriver-sdk-go/massdriver/platform/instances"
	mcpsdk "github.com/modelcontextprotocol/go-sdk/mcp"
)

var ListInstancesTool = &mcpsdk.Tool{
	Name: "list_instances",
	Description: "Lists instances in the organization, one page at a time. " +
		"STRONGLY PREFER filtering by `project_id`, `environment_id`, or `status` — unfiltered lists can span thousands of instances. " +
		"Returns up to `page_size` instances (default 25, max 100) plus a `next_cursor` for the following page. " +
		"To continue, call again with `cursor` set to the previous `next_cursor`. " +
		"Do NOT paginate to exhaustion unless the user explicitly asked for every instance.",
}

type ListInstancesInput struct {
	ProjectID     string `json:"project_id,omitempty"     jsonschema:"Optional. Filter to instances belonging to this project ID."`
	EnvironmentID string `json:"environment_id,omitempty" jsonschema:"Optional. Filter to instances belonging to this environment ID."`
	Status        string `json:"status,omitempty"         jsonschema:"Optional. Filter by status: INITIALIZED, PROVISIONED, DECOMMISSIONED, or FAILED."`
	Cursor        string `json:"cursor,omitempty"         jsonschema:"Optional. Opaque cursor from a prior call's next_cursor. Omit for the first page."`
	PageSize      int    `json:"page_size,omitempty"      jsonschema:"Optional. Page size (1-100, default 25)."`
}

func HandleListInstances(c *Client) func(context.Context, *mcpsdk.CallToolRequest, ListInstancesInput) (*mcpsdk.CallToolResult, any, error) {
	return func(ctx context.Context, _ *mcpsdk.CallToolRequest, args ListInstancesInput) (*mcpsdk.CallToolResult, any, error) {
		page, err := c.Instances.ListPage(ctx, instances.ListInput{
			ProjectID:     args.ProjectID,
			EnvironmentID: args.EnvironmentID,
			Status:        instances.Status(args.Status),
			PageSize:      clampPageSize(args.PageSize),
			After:         args.Cursor,
		})
		if err != nil {
			return nil, nil, fmt.Errorf("list_instances: %w", err)
		}

		out := pageResult(page)
		result, err := jsonResult(out)
		if err != nil {
			return nil, nil, err
		}
		return result, out, nil
	}
}

var GetInstanceTool = &mcpsdk.Tool{
	Name: "get_instance",
	Description: "Gets a specific instance by ID, including its environment, project, and current bundle release. " +
		"Returns paramsSchema (the JSON Schema for this instance's deploy params, resolved for its pinned bundle version) and params (values from the most recent deployment — empty until first deployed).",
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
				return errorResult(fmt.Sprintf("set_instance_secret failed: %s", mutationErr(err))), nil, nil
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
				return errorResult(fmt.Sprintf("remove_instance_secret failed: %s", mutationErr(err))), nil, nil
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
	Name: "list_alarms",
	Description: "Lists alarms across instances, one page at a time. " +
		"PREFER filtering by `project_id`, `environment_id`, `component_id`, `instance_id`, or `oci_repo_name` to keep the result set focused. " +
		"Returns up to `page_size` alarms (default 25, max 100) plus a `next_cursor` for the following page. " +
		"To continue, call again with `cursor` set to the previous `next_cursor`. " +
		"Do NOT paginate to exhaustion unless the user explicitly asked for every alarm.",
}

type ListAlarmsInput struct {
	ProjectID     string `json:"project_id,omitempty"     jsonschema:"Optional. Filter alarms to this project."`
	EnvironmentID string `json:"environment_id,omitempty" jsonschema:"Optional. Filter alarms to this environment."`
	ComponentID   string `json:"component_id,omitempty"   jsonschema:"Optional. Filter alarms to this component."`
	InstanceID    string `json:"instance_id,omitempty"    jsonschema:"Optional. Filter alarms to this instance."`
	OciRepoName   string `json:"oci_repo_name,omitempty"  jsonschema:"Optional. Filter alarms to instances of this bundle."`
	Cursor        string `json:"cursor,omitempty"         jsonschema:"Optional. Opaque cursor from a prior call's next_cursor. Omit for the first page."`
	PageSize      int    `json:"page_size,omitempty"      jsonschema:"Optional. Page size (1-100, default 25)."`
}

func HandleListAlarms(c *Client) func(context.Context, *mcpsdk.CallToolRequest, ListAlarmsInput) (*mcpsdk.CallToolResult, any, error) {
	return func(ctx context.Context, _ *mcpsdk.CallToolRequest, args ListAlarmsInput) (*mcpsdk.CallToolResult, any, error) {
		page, err := c.Instances.ListAlarmsPage(ctx, instances.ListAlarmsInput{
			ProjectID:     args.ProjectID,
			EnvironmentID: args.EnvironmentID,
			ComponentID:   args.ComponentID,
			InstanceID:    args.InstanceID,
			OciRepoName:   args.OciRepoName,
			PageSize:      clampPageSize(args.PageSize),
			After:         args.Cursor,
		})
		if err != nil {
			return nil, nil, fmt.Errorf("list_alarms: %w", err)
		}

		out := pageResult(page)
		result, err := jsonResult(out)
		if err != nil {
			return nil, nil, err
		}
		return result, out, nil
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
				return errorResult(fmt.Sprintf("update_instance failed: %s", mutationErr(err))), nil, nil
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

var SetRemoteReferenceTool = &mcpsdk.Tool{
	Name: "set_remote_reference",
	Description: "Overrides one of an instance's connection slots with a resource from another project (or an " +
		"imported resource). `field` names the connection slot to bind — a key in the instance bundle's " +
		"connectionsSchema. `resource_id` is either a UUID (for imported resources) or \"instance.field\" (for a " +
		"resource provisioned by another instance). This override takes priority over any blueprint-level link " +
		"wired into the same slot, and reverts to that link (or the environment default) when removed with " +
		"remove_remote_reference. The instance must not be in PROVISIONED or FAILED status.",
}

type SetRemoteReferenceInput struct {
	InstanceID string `json:"instance_id" jsonschema:"The instance ID whose connection slot is being overridden."`
	ResourceID string `json:"resource_id" jsonschema:"The resource to bind: a UUID for an imported resource, or \"instance.field\" for a resource provisioned by another instance."`
	Field      string `json:"field"       jsonschema:"The connection slot to bind — a key in the instance bundle's connectionsSchema."`
}

func HandleSetRemoteReference(c *Client) func(context.Context, *mcpsdk.CallToolRequest, SetRemoteReferenceInput) (*mcpsdk.CallToolResult, any, error) {
	return func(ctx context.Context, _ *mcpsdk.CallToolRequest, args SetRemoteReferenceInput) (*mcpsdk.CallToolResult, any, error) {
		if args.InstanceID == "" {
			return nil, nil, fmt.Errorf("set_remote_reference: instance_id is required")
		}
		if args.ResourceID == "" {
			return nil, nil, fmt.Errorf("set_remote_reference: resource_id is required")
		}
		if args.Field == "" {
			return nil, nil, fmt.Errorf("set_remote_reference: field is required")
		}

		ref, err := c.Instances.SetRemoteReference(ctx, args.InstanceID, args.ResourceID, args.Field)
		if err != nil {
			if isMutationFailed(err) {
				return errorResult(fmt.Sprintf("set_remote_reference failed: %s", mutationErr(err))), nil, nil
			}
			return nil, nil, fmt.Errorf("set_remote_reference: %w", err)
		}

		result, err := jsonResult(ref)
		if err != nil {
			return nil, nil, err
		}
		return result, ref, nil
	}
}

var RemoveRemoteReferenceTool = &mcpsdk.Tool{
	Name: "remove_remote_reference",
	Description: "Removes the remote-reference override from the named connection slot on an instance. The slot " +
		"reverts to its blueprint link (if any) or the environment default at the next deploy. `field` is a key in " +
		"the instance bundle's connectionsSchema. The instance must not be in PROVISIONED or FAILED status.",
}

type RemoveRemoteReferenceInput struct {
	InstanceID string `json:"instance_id" jsonschema:"The instance ID whose connection-slot override is being removed."`
	Field      string `json:"field"       jsonschema:"The connection slot to clear — a key in the instance bundle's connectionsSchema."`
}

func HandleRemoveRemoteReference(c *Client) func(context.Context, *mcpsdk.CallToolRequest, RemoveRemoteReferenceInput) (*mcpsdk.CallToolResult, any, error) {
	return func(ctx context.Context, _ *mcpsdk.CallToolRequest, args RemoveRemoteReferenceInput) (*mcpsdk.CallToolResult, any, error) {
		if args.InstanceID == "" {
			return nil, nil, fmt.Errorf("remove_remote_reference: instance_id is required")
		}
		if args.Field == "" {
			return nil, nil, fmt.Errorf("remove_remote_reference: field is required")
		}

		ref, err := c.Instances.RemoveRemoteReference(ctx, args.InstanceID, args.Field)
		if err != nil {
			if isMutationFailed(err) {
				return errorResult(fmt.Sprintf("remove_remote_reference failed: %s", mutationErr(err))), nil, nil
			}
			return nil, nil, fmt.Errorf("remove_remote_reference: %w", err)
		}

		result, err := jsonResult(ref)
		if err != nil {
			return nil, nil, err
		}
		return result, ref, nil
	}
}
