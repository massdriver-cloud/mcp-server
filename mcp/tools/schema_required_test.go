package tools

import (
	"reflect"
	"strings"
	"testing"

	"github.com/google/jsonschema-go/jsonschema"
)

// allToolInputs is every tool input struct registered on the server. Generating
// the JSON Schema for each lets us assert invariants about the wire contract the
// MCP client sees, independent of the handler logic.
//
// Keep this list in sync with the input types in this package. The
// TestEveryInputTypeIsCovered test below guards against drift.
var allToolInputs = []any{
	AbortDeploymentInput{}, AddComponentInput{}, AddGroupServiceAccountInput{}, AddGroupUserInput{},
	ApproveDeploymentInput{}, CreateCustomAttributeInput{}, CreateDeploymentInput{}, CreateEnvironmentInput{},
	CreateGroupInput{}, CreateOciRepoInput{}, CreatePolicyInput{}, CreateProjectInput{},
	CreateResourceGrantInput{}, CreateResourceInput{}, CreateServiceAccountInput{}, DeleteCustomAttributeInput{},
	DeleteEnvironmentInput{}, DeleteGroupInput{}, DeletePolicyInput{}, DeleteProjectInput{},
	DeleteResourceGrantInput{}, DeleteResourceInput{}, DeleteServiceAccountInput{}, EvaluatePoliciesBatchInput{},
	EvaluatePolicyInput{}, ExplainPolicyInput{}, ExportResourceInput{}, GetAuditLogInput{},
	GetBundleInput{}, GetComponentInput{}, GetDeploymentInput{}, GetDeploymentLogsInput{},
	GetEnvironmentInput{}, GetGroupInput{}, GetInstanceInput{}, GetOciRepoInput{},
	GetOrganizationInput{}, GetPolicyAttributeSchemaInput{}, GetPolicyInput{}, GetProjectInput{},
	GetResourceInput{}, GetServerInput{}, GetServiceAccountInput{}, GetURLInput{},
	GetViewerInput{}, LinkComponentsInput{}, ListAlarmsInput{}, ListAuditLogEventTypesInput{},
	ListAuditLogsInput{}, ListBundlesInput{}, ListComponentsInput{}, ListDeploymentsInput{},
	ListEnvironmentsInput{}, ListGroupsInput{}, ListInstancesInput{}, ListOciReposInput{},
	ListPolicyActionsInput{}, ListPolicyAttributeValuesInput{}, ListPolicyEntitiesInput{}, ListProjectsInput{},
	ListResourcesInput{}, ListServiceAccountsInput{}, ProposeDeploymentInput{}, RejectDeploymentInput{},
	RemoveComponentInput{}, RemoveEnvironmentDefaultInput{}, RemoveGroupServiceAccountInput{}, RemoveGroupUserInput{},
	RemoveInstanceSecretInput{}, RevokeGroupInvitationInput{}, SetEnvironmentDefaultInput{}, SetInstanceSecretInput{},
	UnlinkComponentsInput{}, UpdateComponentInput{}, UpdateCustomAttributeInput{}, UpdateEnvironmentInput{},
	UpdateGroupInput{}, UpdateInstanceInput{}, UpdateOciRepoInput{}, UpdatePolicyInput{},
	UpdateProjectInput{}, UpdateResourceInput{}, UpdateServiceAccountInput{},
	DeleteOciRepoInput{},
}

// TestOptionalFieldsAreNotRequired guards against a subtle schema bug: the
// go-sdk infers a property as "required" whenever its json tag lacks
// `,omitempty`. A field documented as "Optional." but missing omitempty would
// therefore be advertised as required, causing MCP clients to reject valid
// calls (e.g. a name-only update_project). This test fails if any field whose
// description begins with "Optional." appears in the schema's required set.
func TestOptionalFieldsAreNotRequired(t *testing.T) {
	for _, in := range allToolInputs {
		typ := reflect.TypeOf(in)
		schema, err := jsonschema.ForType(typ, &jsonschema.ForOptions{})
		if err != nil {
			t.Fatalf("%s: generating schema: %v", typ.Name(), err)
		}

		required := make(map[string]bool, len(schema.Required))
		for _, r := range schema.Required {
			required[r] = true
		}

		for name, prop := range schema.Properties {
			if prop == nil {
				continue
			}
			if strings.HasPrefix(strings.TrimSpace(prop.Description), "Optional") && required[name] {
				t.Errorf("%s: field %q is documented Optional but is marked required in the JSON schema; add `,omitempty` to its json tag",
					typ.Name(), name)
			}
		}
	}
}

// TestInputCoverageMatchesRegisteredTools keeps allToolInputs honest: there is
// exactly one input struct per registered tool, so this list must have one
// entry per tool registered in mcp/server.go (registerTools). If you add or
// remove a tool, update wantTools to match. The dedup check catches copy/paste
// mistakes in the list above.
func TestInputCoverageMatchesRegisteredTools(t *testing.T) {
	const wantTools = 84 // must equal the number of AddTool calls in mcp/server.go

	covered := make(map[string]bool, len(allToolInputs))
	for _, in := range allToolInputs {
		name := reflect.TypeOf(in).Name()
		if covered[name] {
			t.Errorf("duplicate entry in allToolInputs: %s", name)
		}
		covered[name] = true
	}
	if len(covered) != wantTools {
		t.Errorf("allToolInputs covers %d input types, want %d (one per registered tool); update the list or wantTools", len(covered), wantTools)
	}
}
