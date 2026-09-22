package tools

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/google/jsonschema-go/jsonschema"
	mcpsdk "github.com/modelcontextprotocol/go-sdk/mcp"
)

// This file centralizes two pieces of per-tool metadata that are easier to audit
// in one place than scattered across every tool definition:
//
//   - Behavioral annotations (read-only / destructive / idempotent hints) that
//     let MCP clients reason about a tool's safety before calling it.
//   - JSON Schema enum constraints for fields whose valid values are a known,
//     closed set, so clients can validate inputs up front.
//
// It runs in init() so the metadata is attached to the package-level tool vars
// before mcp/server.go registers them.

func init() {
	applyAnnotations()
	applyEnums()
	applyMaxLengths()
}

// readOnly marks a tool that never modifies state.
func readOnly() *mcpsdk.ToolAnnotations {
	return &mcpsdk.ToolAnnotations{ReadOnlyHint: true}
}

// writeHints marks a mutating tool. destructive should be true when the tool can
// remove or overwrite existing state (deletes, decommissions, link removals);
// idempotent should be true when repeating the same call leaves the system in
// the same state.
func writeHints(destructive, idempotent bool) *mcpsdk.ToolAnnotations {
	d := destructive
	return &mcpsdk.ToolAnnotations{DestructiveHint: &d, IdempotentHint: idempotent}
}

// humanTitle derives a display title from a snake_case tool name, e.g.
// "list_oci_repos" -> "List OCI Repos". MCP clients (and the Claude
// connectors directory) use Title for human-facing tool listings.
func humanTitle(name string) string {
	words := strings.Split(name, "_")
	for i, w := range words {
		switch w {
		case "oci", "url", "id":
			words[i] = strings.ToUpper(w)
		default:
			words[i] = strings.ToUpper(w[:1]) + w[1:]
		}
	}
	return strings.Join(words, " ")
}

// annotate assigns the given annotations and a derived display title to each tool.
func annotate(tools []*mcpsdk.Tool, annotations func() *mcpsdk.ToolAnnotations) {
	for _, t := range tools {
		t.Title = humanTitle(t.Name)
		t.Annotations = annotations()
	}
}

func applyAnnotations() {
	readers := []*mcpsdk.Tool{
		GetProjectTool, ListProjectsTool,
		GetEnvironmentTool, ListEnvironmentsTool, ListEnvironmentLinksTool, ListEnvironmentUnfulfilledDependenciesTool,
		GetInstanceTool, ListInstancesTool, ListAlarmsTool,
		GetDeploymentTool, ListDeploymentsTool, GetDeploymentLogsTool, CompareDeploymentsTool,
		CompareEnvironmentsTool,
		GetComponentTool, ListComponentsTool,
		GetBundleTool,
		GetResourceTool, ListResourcesTool, ExportResourceTool, ListResourceGrantsTool,
		GetResourceTypeTool, ListResourceTypeDependentsTool,
		GetOrganizationTool, GetOrganizationSettingsTool,
		GetViewerTool,
		GetAuditLogTool, ListAuditLogsTool, ListAuditLogEventTypesTool,
		GetGroupTool, ListGroupsTool,
		ListGroupMembersTool, ListGroupServiceAccountsTool, ListGroupInvitationsTool, ListGroupPoliciesTool,
		ListOrganizationMembersTool, ListCustomAttributesTool,
		GetServiceAccountTool, ListServiceAccountsTool,
		GetOciRepoTool, ListOciReposTool, ListOciRepoGrantsTool,
		GetPolicyTool, ListPolicyActionsTool, ListPolicyEntitiesTool,
		EvaluatePolicyTool, EvaluatePoliciesBatchTool, ExplainPolicyTool,
		GetPolicyAttributeSchemaTool, ListPolicyAttributeValuesTool,
		GetServerTool, GetURLTool,
	}
	annotate(readers, readOnly)

	// Additive creators: not destructive, not idempotent (a second identical
	// call creates a duplicate or fails).
	additive := []*mcpsdk.Tool{
		CreateProjectTool, CloneProjectTool, CreateEnvironmentTool, ForkEnvironmentTool, AddComponentTool, LinkComponentsTool,
		CreateResourceTool, CreateResourceGrantTool, CreateOciRepoGrantTool, CreateCustomAttributeTool,
		CreateGroupTool, CreateServiceAccountTool, CreateOciRepoTool, CreatePolicyTool,
		ProposeDeploymentTool, RejectDeploymentTool, PlanDeploymentTool, RollbackDeploymentTool,
	}
	annotate(additive, func() *mcpsdk.ToolAnnotations { return writeHints(false, false) })

	// In-place mutations: not destructive in the data-loss sense, and idempotent
	// (re-applying the same values is a no-op).
	updates := []*mcpsdk.Tool{
		UpdateProjectTool, UpdateEnvironmentTool, SetEnvironmentDefaultTool,
		UpdateInstanceTool, SetInstanceSecretTool, SetRemoteReferenceTool,
		UpdateComponentTool, SetComponentPositionTool, UpdateResourceTool, UpdateCustomAttributeTool,
		UpdateGroupTool, AddGroupUserTool, AddGroupServiceAccountTool,
		UpdateServiceAccountTool, UpdateOciRepoTool, UpdatePolicyTool,
		UpdateOrganizationSettingsTool,
	}
	annotate(updates, func() *mcpsdk.ToolAnnotations { return writeHints(false, true) })

	// Destructive removals: idempotent (the target ends up absent either way).
	destructiveIdempotent := []*mcpsdk.Tool{
		DeleteProjectTool, DeleteEnvironmentTool, RemoveEnvironmentDefaultTool,
		RemoveInstanceSecretTool, RemoveRemoteReferenceTool, CopyInstanceTool, OrphanInstanceTool, RemoveComponentTool, UnlinkComponentsTool,
		DeleteResourceTool, DeleteResourceGrantTool, DeleteCustomAttributeTool,
		DeleteGroupTool, RemoveGroupUserTool, RevokeGroupInvitationTool,
		RemoveGroupServiceAccountTool, DeleteServiceAccountTool, DeletePolicyTool,
		DeleteOciRepoTool, DeleteOciRepoGrantTool,
	}
	annotate(destructiveIdempotent, func() *mcpsdk.ToolAnnotations { return writeHints(true, true) })

	// Deployment lifecycle actions that execute or interrupt infrastructure
	// changes: potentially destructive and not idempotent.
	destructiveNonIdempotent := []*mcpsdk.Tool{
		CreateDeploymentTool, ApproveDeploymentTool, AbortDeploymentTool,
		DeployEnvironmentTool, DecommissionEnvironmentTool,
	}
	annotate(destructiveNonIdempotent, func() *mcpsdk.ToolAnnotations { return writeHints(true, false) })
}

// applyEnums attaches JSON Schema enum constraints to fields whose valid values
// are a known closed set (mirroring the corresponding SDK enum types). Only
// fully-enumerable fields are constrained; open-ended fields (e.g. artifact_type
// or export format) are intentionally left unconstrained.
func applyEnums() {
	const (
		allow = "ALLOW"
		deny  = "DENY"
	)
	scopes := []string{"PROJECT", "ENVIRONMENT", "COMPONENT", "REPO"}
	effects := []string{allow, deny}
	deployActions := []string{"PROVISION", "DECOMMISSION", "PLAN"}
	deployStatuses := []string{"PROPOSED", "APPROVED", "PENDING", "RUNNING", "COMPLETED", "FAILED", "REJECTED", "ABORTED"}

	withEnums(CreateDeploymentTool, CreateDeploymentInput{}, map[string][]string{"action": deployActions})
	withEnums(ProposeDeploymentTool, ProposeDeploymentInput{}, map[string][]string{"action": {"PROVISION", "DECOMMISSION"}})
	withEnums(ListDeploymentsTool, ListDeploymentsInput{}, map[string][]string{"action": deployActions, "status": deployStatuses})
	withEnums(ListInstancesTool, ListInstancesInput{}, map[string][]string{"status": {"INITIALIZED", "PROVISIONED", "DECOMMISSIONED", "FAILED"}})
	withEnums(ListResourcesTool, ListResourcesInput{}, map[string][]string{"origin": {"IMPORTED", "PROVISIONED"}})
	withEnums(CreateResourceGrantTool, CreateResourceGrantInput{}, map[string][]string{"action": {"resource:export"}})
	withEnums(CreateCustomAttributeTool, CreateCustomAttributeInput{}, map[string][]string{"scope": scopes})
	withEnums(UpdateOrganizationSettingsTool, UpdateOrganizationSettingsInput{}, map[string][]string{"default_bundle_access": {"NONE", "ALL_PROJECTS"}})
	withEnums(ListPolicyAttributeValuesTool, ListPolicyAttributeValuesInput{}, map[string][]string{"scope": scopes})
	withEnums(CreatePolicyTool, CreatePolicyInput{}, map[string][]string{"effect": effects})
	withEnums(UpdatePolicyTool, UpdatePolicyInput{}, map[string][]string{"effect": effects})
	withEnums(ExplainPolicyTool, ExplainPolicyInput{}, map[string][]string{"effect": effects})
	withEnums(GetURLTool, GetURLInput{}, map[string][]string{
		"type": {"organization", "projects", "project", "environment", "instance", "bundle", "repo_instances"},
	})
}

// applyMaxLengths attaches JSON Schema maxLength constraints mirroring API-side
// limits that otherwise surface only as runtime mutation errors. Descriptions
// are stored in 255-character columns; creation-time identifier slugs are
// capped at 20 characters. Only slugs chosen at creation are constrained — the
// `id` on get/update tools is a lookup reference (e.g. 'myproj-staging') that
// can exceed the slug limit.
func applyMaxLengths() {
	const (
		descriptionMax = 255
		identifierMax  = 20
	)
	withMaxLengths(CreateProjectTool, CreateProjectInput{}, map[string]int{"description": descriptionMax, "id": identifierMax})
	withMaxLengths(CloneProjectTool, CloneProjectInput{}, map[string]int{"description": descriptionMax, "id": identifierMax})
	withMaxLengths(UpdateProjectTool, UpdateProjectInput{}, map[string]int{"description": descriptionMax})
	withMaxLengths(CreateEnvironmentTool, CreateEnvironmentInput{}, map[string]int{"description": descriptionMax, "id": identifierMax})
	withMaxLengths(ForkEnvironmentTool, ForkEnvironmentInput{}, map[string]int{"description": descriptionMax, "id": identifierMax})
	withMaxLengths(UpdateEnvironmentTool, UpdateEnvironmentInput{}, map[string]int{"description": descriptionMax})
	withMaxLengths(AddComponentTool, AddComponentInput{}, map[string]int{"description": descriptionMax, "id": identifierMax})
	withMaxLengths(UpdateComponentTool, UpdateComponentInput{}, map[string]int{"description": descriptionMax})
}

// toolSchema returns the tool's input schema for constraint editing: the one
// already assigned by an earlier helper, or a freshly inferred schema for the
// given zero-value input (the same way mcp-go's AddTool would) assigned to the
// tool so AddTool uses it verbatim. Sharing one schema instance lets multiple
// constraint helpers compose on the same tool.
func toolSchema(tool *mcpsdk.Tool, in any) *jsonschema.Schema {
	if tool.InputSchema != nil {
		if s, ok := tool.InputSchema.(*jsonschema.Schema); ok {
			return s
		}
		panic(fmt.Sprintf("toolSchema: %s has a non-*jsonschema.Schema InputSchema (%T)", tool.Name, tool.InputSchema))
	}
	schema, err := jsonschema.ForType(reflect.TypeOf(in), &jsonschema.ForOptions{})
	if err != nil {
		panic(fmt.Sprintf("toolSchema: inferring schema for %T: %v", in, err))
	}
	tool.InputSchema = schema
	return schema
}

// mustProp returns the named property from the schema, panicking on a
// misconfigured field name since that is a programming error caught at startup.
func mustProp(schema *jsonschema.Schema, in any, field string) *jsonschema.Schema {
	prop, ok := schema.Properties[field]
	if !ok {
		panic(fmt.Sprintf("%T has no property %q", in, field))
	}
	return prop
}

// withEnums applies enum constraints to the named string properties of the
// tool's input schema.
func withEnums(tool *mcpsdk.Tool, in any, enums map[string][]string) {
	schema := toolSchema(tool, in)
	for field, values := range enums {
		anyVals := make([]any, len(values))
		for i, v := range values {
			anyVals[i] = v
		}
		mustProp(schema, in, field).Enum = anyVals
	}
}

// withMaxLengths applies maxLength constraints to the named string properties
// of the tool's input schema.
func withMaxLengths(tool *mcpsdk.Tool, in any, limits map[string]int) {
	schema := toolSchema(tool, in)
	for field, limit := range limits {
		l := limit
		mustProp(schema, in, field).MaxLength = &l
	}
}
