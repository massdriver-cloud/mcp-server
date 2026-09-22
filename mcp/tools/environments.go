package tools

import (
	"context"
	"fmt"

	"github.com/massdriver-cloud/massdriver-sdk-go/massdriver/platform/environments"
	mcpsdk "github.com/modelcontextprotocol/go-sdk/mcp"
)

var ListEnvironmentsTool = &mcpsdk.Tool{
	Name: "list_environments",
	Description: "Lists environments in the organization, one page at a time. " +
		"PREFER filtering by `project_id` — unfiltered lists span every project. " +
		"Returns up to `page_size` environments (default 25, max 100) plus a `next_cursor` for the following page. " +
		"To continue, call again with `cursor` set to the previous `next_cursor`. " +
		"Do NOT paginate to exhaustion unless the user explicitly asked for every environment.",
}

type ListEnvironmentsInput struct {
	ProjectID string `json:"project_id,omitempty" jsonschema:"Optional. Filter to environments belonging to this project ID. Leave empty to list across all projects."`
	Cursor    string `json:"cursor,omitempty"     jsonschema:"Optional. Opaque cursor from a prior call's next_cursor. Omit for the first page."`
	PageSize  int    `json:"page_size,omitempty"  jsonschema:"Optional. Page size (1-100, default 25)."`
}

func HandleListEnvironments(c *Client) func(context.Context, *mcpsdk.CallToolRequest, ListEnvironmentsInput) (*mcpsdk.CallToolResult, any, error) {
	return func(ctx context.Context, _ *mcpsdk.CallToolRequest, args ListEnvironmentsInput) (*mcpsdk.CallToolResult, any, error) {
		page, err := c.Environments.ListPage(ctx, environments.ListInput{
			ProjectID: args.ProjectID,
			PageSize:  clampPageSize(args.PageSize),
			After:     args.Cursor,
		})
		if err != nil {
			return nil, nil, fmt.Errorf("list_environments: %w", err)
		}

		out := pageResult(page)
		result, err := jsonResult(out)
		if err != nil {
			return nil, nil, err
		}
		return result, out, nil
	}
}

var GetEnvironmentTool = &mcpsdk.Tool{
	Name:        "get_environment",
	Description: "Gets a specific environment by its full identifier (e.g., 'myproject-staging'), including its default resource bindings (the `defaults` field — the default resource bound per type for instances in this environment).",
}

type GetEnvironmentInput struct {
	ID string `json:"id" jsonschema:"The environment identifier, typically in the format 'project-environment' (e.g., 'myproj-staging')."`
}

func HandleGetEnvironment(c *Client) func(context.Context, *mcpsdk.CallToolRequest, GetEnvironmentInput) (*mcpsdk.CallToolResult, any, error) {
	return func(ctx context.Context, _ *mcpsdk.CallToolRequest, args GetEnvironmentInput) (*mcpsdk.CallToolResult, any, error) {
		if args.ID == "" {
			return nil, nil, fmt.Errorf("get_environment: id is required")
		}

		env, err := c.Environments.Get(ctx, args.ID)
		if err != nil {
			return nil, nil, fmt.Errorf("get_environment: %w", err)
		}

		result, err := jsonResult(env)
		if err != nil {
			return nil, nil, err
		}
		return result, env, nil
	}
}

var CreateEnvironmentTool = &mcpsdk.Tool{
	Name:        "create_environment",
	Description: "Creates a new environment within a project.",
}

type CreateEnvironmentInput struct {
	ProjectID              string         `json:"project_id"   jsonschema:"The ID of the project to create the environment in."`
	ID                     string         `json:"id"           jsonschema:"Unique identifier for the environment within the project, max 20 lowercase alphanumeric characters. Cannot be changed after creation."`
	Name                   string         `json:"name"                  jsonschema:"Human-readable name shown in the UI."`
	Description            string         `json:"description,omitempty" jsonschema:"Optional description of the environment. Max 255 characters."`
	Attributes             map[string]any `json:"attributes,omitempty" jsonschema:"Optional. Custom attribute tags at the environment scope (e.g., {\"env\":\"prod\"}). Must conform to the organization's custom-attribute schema; some may be required."`
	DecommissionProtection bool           `json:"decommission_protection,omitempty" jsonschema:"Optional. When true, blocks decommission_environment and per-instance DECOMMISSION deployments until disabled via update_environment. Default false."`
	SeparationOfDuty       bool           `json:"separation_of_duty,omitempty"      jsonschema:"Optional. When true, deployment proposals in this environment must be approved by someone other than the proposer. Default false."`
}

func HandleCreateEnvironment(c *Client) func(context.Context, *mcpsdk.CallToolRequest, CreateEnvironmentInput) (*mcpsdk.CallToolResult, any, error) {
	return func(ctx context.Context, _ *mcpsdk.CallToolRequest, args CreateEnvironmentInput) (*mcpsdk.CallToolResult, any, error) {
		if args.ProjectID == "" {
			return nil, nil, fmt.Errorf("create_environment: project_id is required")
		}
		if args.ID == "" {
			return nil, nil, fmt.Errorf("create_environment: id is required")
		}
		if args.Name == "" {
			return nil, nil, fmt.Errorf("create_environment: name is required")
		}

		env, err := c.Environments.Create(ctx, args.ProjectID, environments.CreateInput{
			ID:                     args.ID,
			Name:                   args.Name,
			Description:            args.Description,
			Attributes:             args.Attributes,
			DecommissionProtection: args.DecommissionProtection,
			SeparationOfDuty:       args.SeparationOfDuty,
		})
		if err != nil {
			if isMutationFailed(err) {
				return errorResult(fmt.Sprintf("create_environment failed: %s", mutationErr(err))), nil, nil
			}
			return nil, nil, fmt.Errorf("create_environment: %w", err)
		}

		result, err := jsonResult(env)
		if err != nil {
			return nil, nil, err
		}
		return result, env, nil
	}
}

var UpdateEnvironmentTool = &mcpsdk.Tool{
	Name:        "update_environment",
	Description: "Updates an environment's name, description, custom attributes, decommission protection, or separation of duty. Only the fields you provide are changed.",
}

type UpdateEnvironmentInput struct {
	ID                     string         `json:"id"                      jsonschema:"The environment identifier (e.g., 'myproj-staging')."`
	Name                   *string        `json:"name,omitempty"          jsonschema:"Optional. New human-readable name. Omit to leave unchanged; cannot be set to an empty string."`
	Description            *string        `json:"description,omitempty"   jsonschema:"Optional. New description, max 255 characters. Omit to leave unchanged; pass an empty string to clear it."`
	Attributes             map[string]any `json:"attributes,omitempty"    jsonschema:"Optional. Replacement custom attribute tags at the environment scope. Omit to leave unchanged; when provided, replaces the full attribute set. Must conform to the organization's custom-attribute schema."`
	DecommissionProtection *bool          `json:"decommission_protection,omitempty" jsonschema:"Optional. Toggles the guard that blocks decommission_environment and per-instance DECOMMISSION deployments. Omit to leave unchanged."`
	SeparationOfDuty       *bool          `json:"separation_of_duty,omitempty"      jsonschema:"Optional. Toggles whether deployment proposals in this environment must be approved by someone other than the proposer. Omit to leave unchanged."`
}

func HandleUpdateEnvironment(c *Client) func(context.Context, *mcpsdk.CallToolRequest, UpdateEnvironmentInput) (*mcpsdk.CallToolResult, any, error) {
	return func(ctx context.Context, _ *mcpsdk.CallToolRequest, args UpdateEnvironmentInput) (*mcpsdk.CallToolResult, any, error) {
		if args.ID == "" {
			return nil, nil, fmt.Errorf("update_environment: id is required")
		}

		env, err := c.Environments.Update(ctx, args.ID, environments.UpdateInput{
			Name:                   args.Name,
			Description:            args.Description,
			Attributes:             args.Attributes,
			DecommissionProtection: args.DecommissionProtection,
			SeparationOfDuty:       args.SeparationOfDuty,
		})
		if err != nil {
			if isMutationFailed(err) {
				return errorResult(fmt.Sprintf("update_environment failed: %s", mutationErr(err))), nil, nil
			}
			return nil, nil, fmt.Errorf("update_environment: %w", err)
		}

		result, err := jsonResult(env)
		if err != nil {
			return nil, nil, err
		}
		return result, env, nil
	}
}

var DeleteEnvironmentTool = &mcpsdk.Tool{
	Name:        "delete_environment",
	Description: "Deletes an environment. All instances in the environment must be decommissioned before deletion.",
}

type DeleteEnvironmentInput struct {
	ID string `json:"id" jsonschema:"The environment identifier to delete (e.g., 'myproj-staging')."`
}

func HandleDeleteEnvironment(c *Client) func(context.Context, *mcpsdk.CallToolRequest, DeleteEnvironmentInput) (*mcpsdk.CallToolResult, any, error) {
	return func(ctx context.Context, _ *mcpsdk.CallToolRequest, args DeleteEnvironmentInput) (*mcpsdk.CallToolResult, any, error) {
		if args.ID == "" {
			return nil, nil, fmt.Errorf("delete_environment: id is required")
		}

		_, err := c.Environments.Delete(ctx, args.ID)
		if err != nil {
			if isMutationFailed(err) {
				return errorResult(fmt.Sprintf("delete_environment failed: %s", mutationErr(err))), nil, nil
			}
			return nil, nil, fmt.Errorf("delete_environment: %w", err)
		}

		return textResult(fmt.Sprintf("environment %q deleted successfully", args.ID)), nil, nil
	}
}

var SetEnvironmentDefaultTool = &mcpsdk.Tool{
	Name: "set_environment_default",
	Description: "Sets a resource as the default of its type for an environment — new instances bind to it for that dependency type. " +
		"The resource must first be shared to the environment with create_resource_grant (whose recipient_conditions match the environment, or are omitted for all environments), otherwise this fails with \"not granted to this environment\". Only one default per resource type.",
}

type SetEnvironmentDefaultInput struct {
	EnvironmentID string `json:"environment_id" jsonschema:"The environment ID to set the default on."`
	ResourceID    string `json:"resource_id"    jsonschema:"The resource ID to bind as the default."`
}

func HandleSetEnvironmentDefault(c *Client) func(context.Context, *mcpsdk.CallToolRequest, SetEnvironmentDefaultInput) (*mcpsdk.CallToolResult, any, error) {
	return func(ctx context.Context, _ *mcpsdk.CallToolRequest, args SetEnvironmentDefaultInput) (*mcpsdk.CallToolResult, any, error) {
		if args.EnvironmentID == "" {
			return nil, nil, fmt.Errorf("set_environment_default: environment_id is required")
		}
		if args.ResourceID == "" {
			return nil, nil, fmt.Errorf("set_environment_default: resource_id is required")
		}

		envDefault, err := c.Environments.SetDefault(ctx, args.EnvironmentID, args.ResourceID)
		if err != nil {
			if isMutationFailed(err) {
				return errorResult(fmt.Sprintf("set_environment_default failed: %s", mutationErr(err))), nil, nil
			}
			return nil, nil, fmt.Errorf("set_environment_default: %w", err)
		}

		result, err := jsonResult(envDefault)
		if err != nil {
			return nil, nil, err
		}
		return result, envDefault, nil
	}
}

var RemoveEnvironmentDefaultTool = &mcpsdk.Tool{
	Name:        "remove_environment_default",
	Description: "Removes a default resource binding from an environment.",
}

type RemoveEnvironmentDefaultInput struct {
	ID string `json:"id" jsonschema:"The environment default ID to remove."`
}

func HandleRemoveEnvironmentDefault(c *Client) func(context.Context, *mcpsdk.CallToolRequest, RemoveEnvironmentDefaultInput) (*mcpsdk.CallToolResult, any, error) {
	return func(ctx context.Context, _ *mcpsdk.CallToolRequest, args RemoveEnvironmentDefaultInput) (*mcpsdk.CallToolResult, any, error) {
		if args.ID == "" {
			return nil, nil, fmt.Errorf("remove_environment_default: id is required")
		}

		envDefault, err := c.Environments.RemoveDefault(ctx, args.ID)
		if err != nil {
			if isMutationFailed(err) {
				return errorResult(fmt.Sprintf("remove_environment_default failed: %s", mutationErr(err))), nil, nil
			}
			return nil, nil, fmt.Errorf("remove_environment_default: %w", err)
		}

		result, err := jsonResult(envDefault)
		if err != nil {
			return nil, nil, err
		}
		return result, envDefault, nil
	}
}

var CompareEnvironmentsTool = &mcpsdk.Tool{
	Name: "compare_environments",
	Description: "Compares two environments in the same project, instance-by-instance. Instances are paired by " +
		"component; for each component the result reports the resolved bundle version on each side and a flat, " +
		"leaf-level diff of the configured params. When a component is deployed on only one side, the other side's " +
		"entry is null. Environment-level attributes and default resource wiring are not part of the comparison. " +
		"Both environments must belong to the same project.",
}

type CompareEnvironmentsInput struct {
	SourceID string `json:"source_id" jsonschema:"The source (baseline) environment ID."`
	TargetID string `json:"target_id" jsonschema:"The target environment ID to compare against the source. Must be in the same project as the source."`
}

func HandleCompareEnvironments(c *Client) func(context.Context, *mcpsdk.CallToolRequest, CompareEnvironmentsInput) (*mcpsdk.CallToolResult, any, error) {
	return func(ctx context.Context, _ *mcpsdk.CallToolRequest, args CompareEnvironmentsInput) (*mcpsdk.CallToolResult, any, error) {
		if args.SourceID == "" {
			return nil, nil, fmt.Errorf("compare_environments: source_id is required")
		}
		if args.TargetID == "" {
			return nil, nil, fmt.Errorf("compare_environments: target_id is required")
		}

		comparison, err := c.Environments.Compare(ctx, args.SourceID, args.TargetID)
		if err != nil {
			return nil, nil, fmt.Errorf("compare_environments: %w", err)
		}

		result, err := jsonResult(comparison)
		if err != nil {
			return nil, nil, err
		}
		return result, comparison, nil
	}
}

var ForkEnvironmentTool = &mcpsdk.Tool{
	Name: "fork_environment",
	Description: "Forks a new environment from an existing parent environment within the same project, copying the " +
		"parent's component configuration into the new environment. Requires `parent_id`, `id`, and `name`. " +
		"The copy toggles control what carries over from the parent: `copy_secrets` (each component's secret " +
		"values), `copy_remote_references` (each component's remote resource references), and " +
		"`copy_environment_defaults` (the parent's default resource connections). All default to false.",
}

type ForkEnvironmentInput struct {
	ParentID                string         `json:"parent_id"                          jsonschema:"The ID of the parent environment to fork from."`
	ID                      string         `json:"id"                                 jsonschema:"Unique identifier for the new environment within the project, max 20 lowercase alphanumeric characters. Cannot be changed after creation."`
	Name                    string         `json:"name"                               jsonschema:"Human-readable name shown in the UI."`
	Description             string         `json:"description,omitempty"              jsonschema:"Optional description of the fork's purpose. Max 255 characters."`
	Attributes              map[string]any `json:"attributes,omitempty"               jsonschema:"Optional. Custom attribute tags at the environment scope. Must conform to the organization's custom-attribute schema; some may be required."`
	CopySecrets             bool           `json:"copy_secrets,omitempty"             jsonschema:"Optional. When true, copies every component's secret values from the parent into the fork. Default false."`
	CopyRemoteReferences    bool           `json:"copy_remote_references,omitempty"   jsonschema:"Optional. When true, copies every component's remote resource references from the parent into the fork. Default false."`
	CopyEnvironmentDefaults bool           `json:"copy_environment_defaults,omitempty" jsonschema:"Optional. When true, copies the parent's default resource connections into the fork. Default false."`
	DecommissionProtection  bool           `json:"decommission_protection,omitempty"  jsonschema:"Optional. When true, blocks decommission_environment and per-instance DECOMMISSION deployments on the fork until disabled via update_environment. Default false."`
	SeparationOfDuty        bool           `json:"separation_of_duty,omitempty"       jsonschema:"Optional. When true, deployment proposals in the fork must be approved by someone other than the proposer. Default false."`
}

func HandleForkEnvironment(c *Client) func(context.Context, *mcpsdk.CallToolRequest, ForkEnvironmentInput) (*mcpsdk.CallToolResult, any, error) {
	return func(ctx context.Context, _ *mcpsdk.CallToolRequest, args ForkEnvironmentInput) (*mcpsdk.CallToolResult, any, error) {
		if args.ParentID == "" {
			return nil, nil, fmt.Errorf("fork_environment: parent_id is required")
		}
		if args.ID == "" {
			return nil, nil, fmt.Errorf("fork_environment: id is required")
		}
		if args.Name == "" {
			return nil, nil, fmt.Errorf("fork_environment: name is required")
		}

		env, err := c.Environments.Fork(ctx, args.ParentID, environments.ForkInput{
			ID:                      args.ID,
			Name:                    args.Name,
			Description:             args.Description,
			Attributes:              args.Attributes,
			CopySecrets:             args.CopySecrets,
			CopyRemoteReferences:    args.CopyRemoteReferences,
			CopyEnvironmentDefaults: args.CopyEnvironmentDefaults,
			DecommissionProtection:  args.DecommissionProtection,
			SeparationOfDuty:        args.SeparationOfDuty,
		})
		if err != nil {
			if isMutationFailed(err) {
				return errorResult(fmt.Sprintf("fork_environment failed: %s", mutationErr(err))), nil, nil
			}
			return nil, nil, fmt.Errorf("fork_environment: %w", err)
		}

		result, err := jsonResult(env)
		if err != nil {
			return nil, nil, err
		}
		return result, env, nil
	}
}

var DeployEnvironmentTool = &mcpsdk.Tool{
	Name: "deploy_environment",
	Description: "Schedules a deployment of every instance in the environment in dependency order. Cancels any " +
		"in-flight environment deployment and enqueues a fresh provision wave. Returns as soon as the wave is " +
		"enqueued — the infrastructure changes happen asynchronously.",
}

type DeployEnvironmentInput struct {
	ID string `json:"id" jsonschema:"The environment ID to deploy (e.g., 'myproj-staging')."`
}

func HandleDeployEnvironment(c *Client) func(context.Context, *mcpsdk.CallToolRequest, DeployEnvironmentInput) (*mcpsdk.CallToolResult, any, error) {
	return func(ctx context.Context, _ *mcpsdk.CallToolRequest, args DeployEnvironmentInput) (*mcpsdk.CallToolResult, any, error) {
		if args.ID == "" {
			return nil, nil, fmt.Errorf("deploy_environment: id is required")
		}

		env, err := c.Environments.Deploy(ctx, args.ID)
		if err != nil {
			if isMutationFailed(err) {
				return errorResult(fmt.Sprintf("deploy_environment failed: %s", mutationErr(err))), nil, nil
			}
			return nil, nil, fmt.Errorf("deploy_environment: %w", err)
		}

		result, err := jsonResult(env)
		if err != nil {
			return nil, nil, err
		}
		return result, env, nil
	}
}

var DecommissionEnvironmentTool = &mcpsdk.Tool{
	Name: "decommission_environment",
	Description: "Schedules a teardown of every instance in the environment in reverse dependency order. The " +
		"environment shell stays in place so it can be redeployed; use delete_environment to remove the empty " +
		"environment afterwards. Cancels any in-flight environment deployment and enqueues a fresh decommission " +
		"wave. Returns as soon as the wave is enqueued — the infrastructure changes happen asynchronously. Blocked " +
		"when the environment has decommission protection enabled; disable it via update_environment first.",
}

type DecommissionEnvironmentInput struct {
	ID string `json:"id" jsonschema:"The environment ID to decommission (e.g., 'myproj-staging')."`
}

func HandleDecommissionEnvironment(c *Client) func(context.Context, *mcpsdk.CallToolRequest, DecommissionEnvironmentInput) (*mcpsdk.CallToolResult, any, error) {
	return func(ctx context.Context, _ *mcpsdk.CallToolRequest, args DecommissionEnvironmentInput) (*mcpsdk.CallToolResult, any, error) {
		if args.ID == "" {
			return nil, nil, fmt.Errorf("decommission_environment: id is required")
		}

		env, err := c.Environments.Decommission(ctx, args.ID)
		if err != nil {
			if isMutationFailed(err) {
				return errorResult(fmt.Sprintf("decommission_environment failed: %s", mutationErr(err))), nil, nil
			}
			return nil, nil, fmt.Errorf("decommission_environment: %w", err)
		}

		result, err := jsonResult(env)
		if err != nil {
			return nil, nil, err
		}
		return result, env, nil
	}
}

var ListEnvironmentLinksTool = &mcpsdk.Tool{
	Name: "list_environment_links",
	Description: "Lists the blueprint links in effect in an environment given the component versions its instances actually run. " +
		"Where a project's links list every link in the architecture, this is the subset applying to this environment: a link " +
		"applies only where the versions at both ends fall inside its version range (from_version_constraint / " +
		"to_version_constraint, as tilde constraints like '~1'). A link whose source or destination has no instance in this " +
		"environment does not appear. Returns the whole list (not paginated).",
}

type ListEnvironmentLinksInput struct {
	ID string `json:"id" jsonschema:"The environment identifier (e.g., 'myproj-staging')."`
}

func HandleListEnvironmentLinks(c *Client) func(context.Context, *mcpsdk.CallToolRequest, ListEnvironmentLinksInput) (*mcpsdk.CallToolResult, any, error) {
	return func(ctx context.Context, _ *mcpsdk.CallToolRequest, args ListEnvironmentLinksInput) (*mcpsdk.CallToolResult, any, error) {
		if args.ID == "" {
			return nil, nil, fmt.Errorf("list_environment_links: id is required")
		}

		links, err := c.Environments.Links(ctx, args.ID)
		if err != nil {
			return nil, nil, fmt.Errorf("list_environment_links: %w", err)
		}

		out := listResult(links)
		result, err := jsonResult(out)
		if err != nil {
			return nil, nil, err
		}
		return result, out, nil
	}
}

var ListEnvironmentUnfulfilledDependenciesTool = &mcpsdk.Tool{
	Name: "list_environment_unfulfilled_dependencies",
	Description: "Lists required dependency inputs across an environment's instances that nothing fills — no blueprint link, no " +
		"per-instance remote reference, and no environment default of the matching resource type. Each entry is one input a " +
		"deploy would block on (optional inputs are never included), so use this to answer \"why won't this environment " +
		"deploy?\". To fix an entry: wire a resource into the slot (link_components or set_remote_reference) or bind an " +
		"environment default of the listed resource type (set_environment_default). Returns the whole list (not paginated), " +
		"sorted by instance identifier then input name.",
}

type ListEnvironmentUnfulfilledDependenciesInput struct {
	ID string `json:"id" jsonschema:"The environment identifier (e.g., 'myproj-staging')."`
}

func HandleListEnvironmentUnfulfilledDependencies(c *Client) func(context.Context, *mcpsdk.CallToolRequest, ListEnvironmentUnfulfilledDependenciesInput) (*mcpsdk.CallToolResult, any, error) {
	return func(ctx context.Context, _ *mcpsdk.CallToolRequest, args ListEnvironmentUnfulfilledDependenciesInput) (*mcpsdk.CallToolResult, any, error) {
		if args.ID == "" {
			return nil, nil, fmt.Errorf("list_environment_unfulfilled_dependencies: id is required")
		}

		deps, err := c.Environments.UnfulfilledDependencies(ctx, args.ID)
		if err != nil {
			return nil, nil, fmt.Errorf("list_environment_unfulfilled_dependencies: %w", err)
		}

		out := listResult(deps)
		result, err := jsonResult(out)
		if err != nil {
			return nil, nil, err
		}
		return result, out, nil
	}
}
