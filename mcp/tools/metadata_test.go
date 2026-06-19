package tools

import (
	"context"
	"strings"
	"testing"

	"github.com/google/jsonschema-go/jsonschema"
	"github.com/massdriver-cloud/massdriver-sdk-go/massdriver/platform/projects"
	mcpsdk "github.com/modelcontextprotocol/go-sdk/mcp"
)

// TestMutationFailureSetsIsError verifies that a handled mutation validation
// failure is surfaced as a tool error (IsError=true), not a normal result whose
// text merely describes a failure.
func TestMutationFailureSetsIsError(t *testing.T) {
	c := &Client{Projects: &stubProjects{
		createFn: func(context.Context, projects.CreateInput) (*projects.Project, error) {
			return nil, mutationFailedErr("createProject", "team", "Required property team was not present")
		},
	}}

	result, _, err := HandleCreateProject(c)(context.Background(), nil, CreateProjectInput{ID: "p", Name: "P"})
	if err != nil {
		t.Fatalf("expected handled failure (nil error), got: %v", err)
	}
	if !result.IsError {
		t.Errorf("expected IsError=true on mutation failure, got false")
	}
	if got := resultText(t, result); !strings.Contains(got, "team") {
		t.Errorf("expected failure text to mention the field, got: %s", got)
	}
}

// TestEveryToolHasAnnotations asserts that applyAnnotations covered every
// registered tool. The list mirrors mcp/server.go's registration.
func TestEveryToolHasAnnotations(t *testing.T) {
	tools := registeredTools()
	if len(tools) != 84 {
		t.Fatalf("registeredTools lists %d tools, want 84 (keep in sync with registerTools)", len(tools))
	}
	seen := make(map[string]bool, len(tools))
	for _, tool := range tools {
		if seen[tool.Name] {
			t.Errorf("duplicate tool in registeredTools: %q", tool.Name)
		}
		seen[tool.Name] = true
		if tool.Annotations == nil {
			t.Errorf("tool %q is missing behavioral annotations", tool.Name)
		}
	}
}

// TestAnnotationClassification spot-checks representative tools across the
// read-only / additive / destructive categories.
func TestAnnotationClassification(t *testing.T) {
	cases := []struct {
		tool        *mcpsdk.Tool
		readOnly    bool
		destructive bool // only meaningful when readOnly is false
	}{
		{ListProjectsTool, true, false},
		{GetURLTool, true, false},
		{CreateProjectTool, false, false},
		{UpdateProjectTool, false, false},
		{DeleteProjectTool, false, true},
		{AbortDeploymentTool, false, true},
		{ApproveDeploymentTool, false, true},
	}
	for _, tc := range cases {
		a := tc.tool.Annotations
		if a == nil {
			t.Errorf("%s: nil annotations", tc.tool.Name)
			continue
		}
		if a.ReadOnlyHint != tc.readOnly {
			t.Errorf("%s: ReadOnlyHint = %v, want %v", tc.tool.Name, a.ReadOnlyHint, tc.readOnly)
		}
		if !tc.readOnly {
			if a.DestructiveHint == nil || *a.DestructiveHint != tc.destructive {
				t.Errorf("%s: DestructiveHint = %v, want %v", tc.tool.Name, a.DestructiveHint, tc.destructive)
			}
		}
	}
}

// TestEnumConstraintsApplied verifies that enum tools carry the expected closed
// value sets in their input schema.
func TestEnumConstraintsApplied(t *testing.T) {
	cases := []struct {
		tool  *mcpsdk.Tool
		field string
		want  []string
	}{
		{GetURLTool, "type", []string{"organization", "projects", "project", "environment", "instance", "bundle", "repo_instances"}},
		{CreateDeploymentTool, "action", []string{"PROVISION", "DECOMMISSION", "PLAN"}},
		{CreatePolicyTool, "effect", []string{"ALLOW", "DENY"}},
		{CreateCustomAttributeTool, "scope", []string{"PROJECT", "ENVIRONMENT", "COMPONENT", "REPO"}},
	}
	for _, tc := range cases {
		schema, ok := tc.tool.InputSchema.(*jsonschema.Schema)
		if !ok {
			t.Errorf("%s: InputSchema is %T, want *jsonschema.Schema", tc.tool.Name, tc.tool.InputSchema)
			continue
		}
		prop := schema.Properties[tc.field]
		if prop == nil {
			t.Errorf("%s: no property %q", tc.tool.Name, tc.field)
			continue
		}
		if len(prop.Enum) != len(tc.want) {
			t.Fatalf("%s.%s: enum len %d, want %d", tc.tool.Name, tc.field, len(prop.Enum), len(tc.want))
		}
		for i, v := range tc.want {
			if prop.Enum[i] != any(v) {
				t.Errorf("%s.%s[%d] = %v, want %q", tc.tool.Name, tc.field, i, prop.Enum[i], v)
			}
		}
	}
}

// TestStripKeysRemovesNestedKeys verifies icon-style stripping at every level.
func TestStripKeysRemovesNestedKeys(t *testing.T) {
	v := map[string]any{
		"icon": "<svg/>",
		"items": []any{
			map[string]any{"name": "a", "icon": "<svg/>"},
			map[string]any{"name": "b", "icon": "<svg/>"},
		},
	}
	stripKeys(v, []string{"icon"})
	if _, ok := v["icon"]; ok {
		t.Error("top-level icon not stripped")
	}
	for _, item := range v["items"].([]any) {
		if _, ok := item.(map[string]any)["icon"]; ok {
			t.Error("nested icon not stripped")
		}
		if _, ok := item.(map[string]any)["name"]; !ok {
			t.Error("stripping removed a non-targeted key")
		}
	}
}

// registeredTools returns every tool that mcp/server.go registers, used to
// assert metadata coverage. Keep in sync with registerTools.
func registeredTools() []*mcpsdk.Tool {
	return []*mcpsdk.Tool{
		ListProjectsTool, GetProjectTool, CreateProjectTool, UpdateProjectTool, DeleteProjectTool,
		ListEnvironmentsTool, GetEnvironmentTool, CreateEnvironmentTool, UpdateEnvironmentTool, DeleteEnvironmentTool, SetEnvironmentDefaultTool, RemoveEnvironmentDefaultTool,
		ListInstancesTool, GetInstanceTool, UpdateInstanceTool, SetInstanceSecretTool, RemoveInstanceSecretTool, ListAlarmsTool,
		ListDeploymentsTool, GetDeploymentTool, GetDeploymentLogsTool, CreateDeploymentTool, ProposeDeploymentTool, ApproveDeploymentTool, RejectDeploymentTool, AbortDeploymentTool,
		ListComponentsTool, GetComponentTool, AddComponentTool, UpdateComponentTool, RemoveComponentTool, LinkComponentsTool, UnlinkComponentsTool,
		ListBundlesTool, GetBundleTool,
		ListResourcesTool, GetResourceTool, CreateResourceTool, UpdateResourceTool, DeleteResourceTool, ExportResourceTool, CreateResourceGrantTool, DeleteResourceGrantTool,
		GetOrganizationTool, CreateCustomAttributeTool, UpdateCustomAttributeTool, DeleteCustomAttributeTool,
		GetViewerTool,
		GetAuditLogTool, ListAuditLogsTool, ListAuditLogEventTypesTool,
		ListGroupsTool, GetGroupTool, CreateGroupTool, UpdateGroupTool, DeleteGroupTool, AddGroupUserTool, RemoveGroupUserTool, RevokeGroupInvitationTool, AddGroupServiceAccountTool, RemoveGroupServiceAccountTool,
		ListServiceAccountsTool, GetServiceAccountTool, CreateServiceAccountTool, UpdateServiceAccountTool, DeleteServiceAccountTool,
		ListOciReposTool, GetOciRepoTool, CreateOciRepoTool, UpdateOciRepoTool, DeleteOciRepoTool,
		GetPolicyTool, CreatePolicyTool, UpdatePolicyTool, DeletePolicyTool, ListPolicyActionsTool, ListPolicyEntitiesTool, EvaluatePolicyTool, EvaluatePoliciesBatchTool, ExplainPolicyTool, GetPolicyAttributeSchemaTool, ListPolicyAttributeValuesTool,
		GetServerTool,
		GetURLTool,
	}
}
