package tools

import (
	"fmt"
	"reflect"

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

func applyAnnotations() {
	readers := []*mcpsdk.Tool{
		GetProjectTool, ListProjectsTool,
		GetEnvironmentTool, ListEnvironmentsTool,
		GetInstanceTool, ListInstancesTool, ListAlarmsTool,
		GetDeploymentTool, ListDeploymentsTool, GetDeploymentLogsTool,
		GetComponentTool, ListComponentsTool,
		GetBundleTool, ListBundlesTool,
		GetResourceTool, ListResourcesTool, ExportResourceTool,
		GetOrganizationTool,
		GetViewerTool,
		GetAuditLogTool, ListAuditLogsTool, ListAuditLogEventTypesTool,
		GetGroupTool, ListGroupsTool,
		GetServiceAccountTool, ListServiceAccountsTool,
		GetOciRepoTool, ListOciReposTool,
		GetPolicyTool, ListPolicyActionsTool, ListPolicyEntitiesTool,
		EvaluatePolicyTool, EvaluatePoliciesBatchTool, ExplainPolicyTool,
		GetPolicyAttributeSchemaTool, ListPolicyAttributeValuesTool,
		GetServerTool, GetURLTool,
	}
	for _, t := range readers {
		t.Annotations = readOnly()
	}

	// Additive creators: not destructive, not idempotent (a second identical
	// call creates a duplicate or fails).
	additive := []*mcpsdk.Tool{
		CreateProjectTool, CreateEnvironmentTool, AddComponentTool, LinkComponentsTool,
		CreateResourceTool, CreateResourceGrantTool, CreateCustomAttributeTool,
		CreateGroupTool, CreateServiceAccountTool, CreateOciRepoTool, CreatePolicyTool,
		ProposeDeploymentTool, RejectDeploymentTool,
	}
	for _, t := range additive {
		t.Annotations = writeHints(false, false)
	}

	// In-place mutations: not destructive in the data-loss sense, and idempotent
	// (re-applying the same values is a no-op).
	updates := []*mcpsdk.Tool{
		UpdateProjectTool, UpdateEnvironmentTool, SetEnvironmentDefaultTool,
		UpdateInstanceTool, SetInstanceSecretTool,
		UpdateComponentTool, UpdateResourceTool, UpdateCustomAttributeTool,
		UpdateGroupTool, AddGroupUserTool, AddGroupServiceAccountTool,
		UpdateServiceAccountTool, UpdateOciRepoTool, UpdatePolicyTool,
	}
	for _, t := range updates {
		t.Annotations = writeHints(false, true)
	}

	// Destructive removals: idempotent (the target ends up absent either way).
	destructiveIdempotent := []*mcpsdk.Tool{
		DeleteProjectTool, DeleteEnvironmentTool, RemoveEnvironmentDefaultTool,
		RemoveInstanceSecretTool, RemoveComponentTool, UnlinkComponentsTool,
		DeleteResourceTool, DeleteResourceGrantTool, DeleteCustomAttributeTool,
		DeleteGroupTool, RemoveGroupUserTool, RevokeGroupInvitationTool,
		RemoveGroupServiceAccountTool, DeleteServiceAccountTool, DeletePolicyTool,
	}
	for _, t := range destructiveIdempotent {
		t.Annotations = writeHints(true, true)
	}

	// Deployment lifecycle actions that execute or interrupt infrastructure
	// changes: potentially destructive and not idempotent.
	destructiveNonIdempotent := []*mcpsdk.Tool{
		CreateDeploymentTool, ApproveDeploymentTool, AbortDeploymentTool,
	}
	for _, t := range destructiveNonIdempotent {
		t.Annotations = writeHints(true, false)
	}
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
	withEnums(CreateCustomAttributeTool, CreateCustomAttributeInput{}, map[string][]string{"scope": scopes})
	withEnums(ListPolicyAttributeValuesTool, ListPolicyAttributeValuesInput{}, map[string][]string{"scope": scopes})
	withEnums(CreatePolicyTool, CreatePolicyInput{}, map[string][]string{"effect": effects})
	withEnums(UpdatePolicyTool, UpdatePolicyInput{}, map[string][]string{"effect": effects})
	withEnums(ExplainPolicyTool, ExplainPolicyInput{}, map[string][]string{"effect": effects})
	withEnums(GetURLTool, GetURLInput{}, map[string][]string{
		"type": {"organization", "projects", "project", "environment", "instance", "bundle", "repo_instances"},
	})
}

// withEnums infers the input schema for the given zero-value input (the same way
// mcp-go's AddTool would), applies enum constraints to the named string
// properties, and assigns the result to the tool's InputSchema so AddTool uses
// it verbatim. It panics on a misconfigured field name since that is a
// programming error caught at startup.
func withEnums(tool *mcpsdk.Tool, in any, enums map[string][]string) {
	schema, err := jsonschema.ForType(reflect.TypeOf(in), &jsonschema.ForOptions{})
	if err != nil {
		panic(fmt.Sprintf("withEnums: inferring schema for %T: %v", in, err))
	}
	for field, values := range enums {
		prop, ok := schema.Properties[field]
		if !ok {
			panic(fmt.Sprintf("withEnums: %T has no property %q", in, field))
		}
		anyVals := make([]any, len(values))
		for i, v := range values {
			anyVals[i] = v
		}
		prop.Enum = anyVals
	}
	tool.InputSchema = schema
}
