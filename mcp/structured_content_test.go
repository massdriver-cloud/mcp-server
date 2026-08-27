package mcp

import (
	"context"
	"encoding/json"
	"io"
	"testing"
	"time"

	"github.com/massdriver-cloud/massdriver-sdk-go/massdriver/platform/auditlogs"
	"github.com/massdriver-cloud/massdriver-sdk-go/massdriver/platform/bundles"
	"github.com/massdriver-cloud/massdriver-sdk-go/massdriver/platform/components"
	"github.com/massdriver-cloud/massdriver-sdk-go/massdriver/platform/deployments"
	"github.com/massdriver-cloud/massdriver-sdk-go/massdriver/platform/environments"
	"github.com/massdriver-cloud/massdriver-sdk-go/massdriver/platform/groups"
	"github.com/massdriver-cloud/massdriver-sdk-go/massdriver/platform/instances"
	"github.com/massdriver-cloud/massdriver-sdk-go/massdriver/platform/ocirepos"
	"github.com/massdriver-cloud/massdriver-sdk-go/massdriver/platform/organizations"
	"github.com/massdriver-cloud/massdriver-sdk-go/massdriver/platform/policies"
	"github.com/massdriver-cloud/massdriver-sdk-go/massdriver/platform/projects"
	"github.com/massdriver-cloud/massdriver-sdk-go/massdriver/platform/resources"
	"github.com/massdriver-cloud/massdriver-sdk-go/massdriver/platform/resourcetypes"
	"github.com/massdriver-cloud/massdriver-sdk-go/massdriver/platform/server"
	"github.com/massdriver-cloud/massdriver-sdk-go/massdriver/platform/serviceaccounts"
	"github.com/massdriver-cloud/massdriver-sdk-go/massdriver/platform/types"
	"github.com/massdriver-cloud/massdriver-sdk-go/massdriver/platform/urls"
	"github.com/massdriver-cloud/massdriver-sdk-go/massdriver/platform/viewer"
	"github.com/massdriver-cloud/mcp-server/mcp/tools"
	mcpsdk "github.com/modelcontextprotocol/go-sdk/mcp"
)

// TestStructuredContentIsObject is a class-level guardrail against the bug where
// a tool returns a bare JSON array or scalar as its structured content. The MCP
// spec requires a tool result's structuredContent to be a JSON object; a naked
// top-level array fails client-side schema validation (this is exactly how
// list_components et al. broke).
//
// It drives EVERY registered tool end-to-end against stub services that always
// succeed, synthesizing arguments from each tool's advertised input schema, and
// asserts that whenever a tool emits structured content it is a JSON object.
// Because it enumerates tools from the real registration, a newly added tool is
// covered automatically — no per-tool entry to maintain.
func TestStructuredContentIsObject(t *testing.T) {
	srv := newServerWithClient(stubToolsClient())

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	clientTransport, serverTransport := mcpsdk.NewInMemoryTransports()
	serverSession, err := srv.Connect(ctx, serverTransport)
	if err != nil {
		t.Fatalf("server.Connect: %v", err)
	}
	defer func() { _ = serverSession.Close() }()

	client := mcpsdk.NewClient(&mcpsdk.Implementation{Name: "guardrail", Version: "1.0.0"}, nil)
	cs, err := client.Connect(ctx, clientTransport, nil)
	if err != nil {
		t.Fatalf("client.Connect: %v", err)
	}
	defer func() { _ = cs.Close() }()

	listed, err := cs.ListTools(ctx, nil)
	if err != nil {
		t.Fatalf("ListTools: %v", err)
	}
	if len(listed.Tools) == 0 {
		t.Fatal("no tools registered")
	}

	objectCount := 0
	for _, tool := range listed.Tools {
		schema, _ := tool.InputSchema.(map[string]any)
		args := synthObject(schema)

		res, err := cs.CallTool(ctx, &mcpsdk.CallToolParams{Name: tool.Name, Arguments: args})
		if err != nil {
			// A protocol error here means our synthesized args didn't satisfy
			// the input schema — not the invariant under test. Log so coverage
			// gaps are visible, but don't fail.
			t.Logf("skip %q: call error: %v", tool.Name, err)
			continue
		}
		if res.IsError {
			t.Logf("skip %q: tool returned an error result", tool.Name)
			continue
		}
		if res.StructuredContent == nil {
			continue // text-only tools (deletes, get_url, schema dumps) — allowed
		}
		if _, ok := res.StructuredContent.(map[string]any); !ok {
			t.Errorf("tool %q: structuredContent must be a JSON object, got %T (%v) — wrap it in an object like {\"items\": [...]}",
				tool.Name, res.StructuredContent, res.StructuredContent)
			continue
		}
		objectCount++
	}

	// Sanity floor: if arg synthesis or the stubs regress, most tools would fall
	// into the skip paths and the guardrail would pass vacuously. Require that a
	// healthy majority actually exercised the structured-content path.
	if min := len(listed.Tools) / 2; objectCount < min {
		t.Fatalf("guardrail exercised too few tools (%d object results of %d registered); arg synthesis or stubs likely broken",
			objectCount, len(listed.Tools))
	}
	t.Logf("verified structured content is an object for %d/%d tools", objectCount, len(listed.Tools))
}

// synthObject builds an arguments object satisfying the required properties of a
// tool input schema, filling each with a schema-appropriate placeholder.
func synthObject(schema map[string]any) map[string]any {
	out := map[string]any{}
	props, _ := schema["properties"].(map[string]any)
	req, _ := schema["required"].([]any)
	for _, r := range req {
		name, ok := r.(string)
		if !ok {
			continue
		}
		if ps, ok := props[name].(map[string]any); ok {
			out[name] = synthValue(ps)
		} else {
			out[name] = "x"
		}
	}
	return out
}

// synthValue produces a placeholder value for a single property schema.
func synthValue(schema map[string]any) any {
	if enum, ok := schema["enum"].([]any); ok && len(enum) > 0 {
		return enum[0]
	}
	switch schema["type"] {
	case "integer", "number":
		return 1
	case "boolean":
		return true
	case "array":
		if items, ok := schema["items"].(map[string]any); ok {
			return []any{synthValue(items)}
		}
		return []any{}
	case "object":
		return synthObject(schema)
	default: // "string" and anything unspecified
		return "x"
	}
}

// stubToolsClient assembles a tools.Client whose every service returns
// non-nil, success-shaped data so each handler runs its full body and populates
// structured content.
func stubToolsClient() *tools.Client {
	return &tools.Client{
		Projects:        scProjects{},
		Environments:    scEnvironments{},
		Instances:       scInstances{},
		Deployments:     scDeployments{},
		Components:      scComponents{},
		Bundles:         scBundles{},
		Resources:       scResources{},
		ResourceTypes:   scResourceTypes{},
		Organizations:   scOrganizations{},
		Viewer:          scViewer{},
		AuditLogs:       scAuditLogs{},
		Groups:          scGroups{},
		ServiceAccounts: scServiceAccounts{},
		OciRepos:        scOciRepos{},
		Policies:        scPolicies{},
		Server:          scServer{},
		URLs:            scURLs{},
	}
}

type scProjects struct{}

func (scProjects) ListPage(context.Context, projects.ListInput) (types.Page[projects.Project], error) {
	return types.Page[projects.Project]{Items: []projects.Project{{}}}, nil
}
func (scProjects) Get(context.Context, string) (*projects.Project, error) {
	return &projects.Project{}, nil
}
func (scProjects) Create(context.Context, projects.CreateInput) (*projects.Project, error) {
	return &projects.Project{}, nil
}
func (scProjects) Clone(context.Context, string, projects.CloneInput) (*projects.Project, error) {
	return &projects.Project{}, nil
}
func (scProjects) Update(context.Context, string, projects.UpdateInput) (*projects.Project, error) {
	return &projects.Project{}, nil
}
func (scProjects) Delete(context.Context, string) (*projects.Project, error) {
	return &projects.Project{}, nil
}

type scEnvironments struct{}

func (scEnvironments) ListPage(context.Context, environments.ListInput) (types.Page[environments.Environment], error) {
	return types.Page[environments.Environment]{Items: []environments.Environment{{}}}, nil
}
func (scEnvironments) Get(context.Context, string) (*environments.Environment, error) {
	return &environments.Environment{}, nil
}
func (scEnvironments) Links(context.Context, string) ([]types.Link, error) {
	return []types.Link{{}}, nil
}
func (scEnvironments) UnfulfilledDependencies(context.Context, string) ([]environments.UnfulfilledDependency, error) {
	return []environments.UnfulfilledDependency{{}}, nil
}
func (scEnvironments) Create(context.Context, string, environments.CreateInput) (*environments.Environment, error) {
	return &environments.Environment{}, nil
}
func (scEnvironments) Update(context.Context, string, environments.UpdateInput) (*environments.Environment, error) {
	return &environments.Environment{}, nil
}
func (scEnvironments) Delete(context.Context, string) (*environments.Environment, error) {
	return &environments.Environment{}, nil
}
func (scEnvironments) SetDefault(context.Context, string, string) (*environments.EnvironmentDefault, error) {
	return &environments.EnvironmentDefault{}, nil
}
func (scEnvironments) RemoveDefault(context.Context, string) (*environments.EnvironmentDefault, error) {
	return &environments.EnvironmentDefault{}, nil
}
func (scEnvironments) Compare(context.Context, string, string) (*environments.Comparison, error) {
	return &environments.Comparison{}, nil
}
func (scEnvironments) Fork(context.Context, string, environments.ForkInput) (*environments.Environment, error) {
	return &environments.Environment{}, nil
}
func (scEnvironments) Deploy(context.Context, string) (*environments.Environment, error) {
	return &environments.Environment{}, nil
}
func (scEnvironments) Decommission(context.Context, string) (*environments.Environment, error) {
	return &environments.Environment{}, nil
}

type scInstances struct{}

func (scInstances) ListPage(context.Context, instances.ListInput) (types.Page[instances.Instance], error) {
	return types.Page[instances.Instance]{Items: []instances.Instance{{}}}, nil
}
func (scInstances) Get(context.Context, string) (*instances.Instance, error) {
	return &instances.Instance{}, nil
}
func (scInstances) Update(context.Context, string, instances.UpdateInput) (*instances.Instance, error) {
	return &instances.Instance{}, nil
}
func (scInstances) SetSecret(context.Context, string, string, string) (*instances.Secret, error) {
	return &instances.Secret{}, nil
}
func (scInstances) RemoveSecret(context.Context, string, string) (*instances.Secret, error) {
	return &instances.Secret{}, nil
}
func (scInstances) SetRemoteReference(context.Context, string, string, string) (*instances.RemoteReference, error) {
	return &instances.RemoteReference{}, nil
}
func (scInstances) RemoveRemoteReference(context.Context, string, string) (*instances.RemoteReference, error) {
	return &instances.RemoteReference{}, nil
}
func (scInstances) Copy(context.Context, string, string, instances.CopyInput) (*instances.Instance, error) {
	return &instances.Instance{}, nil
}
func (scInstances) Orphan(context.Context, string, instances.OrphanInput) (*instances.Instance, error) {
	return &instances.Instance{}, nil
}
func (scInstances) ListAlarmsPage(context.Context, instances.ListAlarmsInput) (types.Page[instances.Alarm], error) {
	return types.Page[instances.Alarm]{Items: []instances.Alarm{{}}}, nil
}

type scDeployments struct{}

func (scDeployments) ListPage(context.Context, deployments.ListInput) (types.Page[deployments.Deployment], error) {
	return types.Page[deployments.Deployment]{Items: []deployments.Deployment{{}}}, nil
}
func (scDeployments) Get(context.Context, string) (*deployments.Deployment, error) {
	return &deployments.Deployment{}, nil
}
func (scDeployments) GetLogs(context.Context, string) (string, error)   { return "logs", nil }
func (scDeployments) TailLogs(context.Context, string, io.Writer) error { return nil }
func (scDeployments) Create(context.Context, string, deployments.CreateInput) (*deployments.Deployment, error) {
	return &deployments.Deployment{}, nil
}
func (scDeployments) Propose(context.Context, string, deployments.ProposeInput) (*deployments.Deployment, error) {
	return &deployments.Deployment{}, nil
}
func (scDeployments) Approve(context.Context, string) (*deployments.Deployment, error) {
	return &deployments.Deployment{}, nil
}
func (scDeployments) Reject(context.Context, string) (*deployments.Deployment, error) {
	return &deployments.Deployment{}, nil
}
func (scDeployments) Abort(context.Context, string) (*deployments.Deployment, error) {
	return &deployments.Deployment{}, nil
}
func (scDeployments) Plan(context.Context, string) (*deployments.Deployment, error) {
	return &deployments.Deployment{}, nil
}
func (scDeployments) Rollback(context.Context, string) (*deployments.Deployment, error) {
	return &deployments.Deployment{}, nil
}
func (scDeployments) Compare(context.Context, string, string) (*deployments.Comparison, error) {
	return &deployments.Comparison{}, nil
}

type scComponents struct{}

func (scComponents) List(context.Context, components.ListInput) ([]components.Component, error) {
	return []components.Component{{}}, nil
}
func (scComponents) Get(context.Context, string) (*components.Component, error) {
	return &components.Component{}, nil
}
func (scComponents) Add(context.Context, string, components.AddInput) (*components.Component, error) {
	return &components.Component{}, nil
}
func (scComponents) Update(context.Context, string, components.UpdateInput) (*components.Component, error) {
	return &components.Component{}, nil
}
func (scComponents) SetPosition(context.Context, string, components.Position) (*components.Component, error) {
	return &components.Component{}, nil
}
func (scComponents) Remove(context.Context, string) (*components.Component, error) {
	return &components.Component{}, nil
}
func (scComponents) AddLink(context.Context, components.AddLinkInput) (*components.Link, error) {
	return &components.Link{}, nil
}
func (scComponents) RemoveLink(context.Context, string) (*components.Link, error) {
	return &components.Link{}, nil
}

type scBundles struct{}

func (scBundles) Get(context.Context, string) (*bundles.Bundle, error) { return &bundles.Bundle{}, nil }

type scResources struct{}

func (scResources) ListPage(context.Context, resources.ListInput) (types.Page[resources.Resource], error) {
	return types.Page[resources.Resource]{Items: []resources.Resource{{}}}, nil
}
func (scResources) Get(context.Context, string) (*resources.Resource, error) {
	return &resources.Resource{}, nil
}
func (scResources) Create(context.Context, string, resources.CreateInput) (*resources.Resource, error) {
	return &resources.Resource{}, nil
}
func (scResources) Update(context.Context, string, resources.UpdateInput) (*resources.Resource, error) {
	return &resources.Resource{}, nil
}
func (scResources) Delete(context.Context, string) (*resources.Resource, error) {
	return &resources.Resource{}, nil
}
func (scResources) Export(context.Context, string, string) (*resources.Exported, error) {
	return &resources.Exported{}, nil
}
func (scResources) CreateGrant(context.Context, string, resources.CreateGrantInput) (*resources.Grant, error) {
	return &resources.Grant{}, nil
}
func (scResources) DeleteGrant(context.Context, string) error { return nil }
func (scResources) ListGrantsPage(context.Context, string, resources.ListGrantsInput) (types.Page[resources.Grant], error) {
	return types.Page[resources.Grant]{Items: []resources.Grant{{}}}, nil
}

type scResourceTypes struct{}

func (scResourceTypes) Get(context.Context, string) (*resourcetypes.ResourceType, error) {
	return &resourcetypes.ResourceType{}, nil
}
func (scResourceTypes) Dependents(context.Context, string, string) ([]resourcetypes.Dependent, error) {
	return []resourcetypes.Dependent{{}}, nil
}

type scOrganizations struct{}

func (scOrganizations) Get(context.Context) (*organizations.Organization, error) {
	return &organizations.Organization{}, nil
}
func (scOrganizations) GetSettings(context.Context) (*organizations.Settings, error) {
	return &organizations.Settings{}, nil
}
func (scOrganizations) UpdateSettings(context.Context, organizations.UpdateSettingsInput) (*organizations.Settings, error) {
	return &organizations.Settings{}, nil
}
func (scOrganizations) CreateCustomAttribute(context.Context, organizations.CreateCustomAttributeInput) (*organizations.CustomAttribute, error) {
	return &organizations.CustomAttribute{}, nil
}
func (scOrganizations) UpdateCustomAttribute(context.Context, string, organizations.UpdateCustomAttributeInput) (*organizations.CustomAttribute, error) {
	return &organizations.CustomAttribute{}, nil
}
func (scOrganizations) DeleteCustomAttribute(context.Context, string) (*organizations.CustomAttribute, error) {
	return &organizations.CustomAttribute{}, nil
}
func (scOrganizations) ListMembersPage(context.Context, organizations.ListMembersInput) (types.Page[organizations.Account], error) {
	return types.Page[organizations.Account]{Items: []organizations.Account{{}}}, nil
}
func (scOrganizations) ListCustomAttributesPage(context.Context, organizations.ListCustomAttributesInput) (types.Page[organizations.CustomAttribute], error) {
	return types.Page[organizations.CustomAttribute]{Items: []organizations.CustomAttribute{{}}}, nil
}

type scViewer struct{}

func (scViewer) Get(context.Context) (*viewer.Viewer, error) { return &viewer.Viewer{}, nil }

type scAuditLogs struct{}

func (scAuditLogs) Get(context.Context, string) (*auditlogs.AuditLog, error) {
	return &auditlogs.AuditLog{}, nil
}
func (scAuditLogs) ListPage(context.Context, auditlogs.ListInput) (types.Page[auditlogs.AuditLog], error) {
	return types.Page[auditlogs.AuditLog]{Items: []auditlogs.AuditLog{{}}}, nil
}
func (scAuditLogs) ListEventTypes(context.Context) ([]string, error) {
	return []string{"deployment.created"}, nil
}

type scGroups struct{}

func (scGroups) ListPage(context.Context, groups.ListInput) (types.Page[groups.Group], error) {
	return types.Page[groups.Group]{Items: []groups.Group{{}}}, nil
}
func (scGroups) Get(context.Context, string) (*groups.Group, error) { return &groups.Group{}, nil }
func (scGroups) Create(context.Context, groups.CreateInput) (*groups.Group, error) {
	return &groups.Group{}, nil
}
func (scGroups) Update(context.Context, string, groups.UpdateInput) (*groups.Group, error) {
	return &groups.Group{}, nil
}
func (scGroups) Delete(context.Context, string) (*groups.Group, error) { return &groups.Group{}, nil }
func (scGroups) AddUser(context.Context, string, string) (*groups.AddUserResult, error) {
	return &groups.AddUserResult{}, nil
}
func (scGroups) RemoveUser(context.Context, string, string) error           { return nil }
func (scGroups) RevokeInvitation(context.Context, string, string) error     { return nil }
func (scGroups) AddServiceAccount(context.Context, string, string) error    { return nil }
func (scGroups) RemoveServiceAccount(context.Context, string, string) error { return nil }
func (scGroups) ListMembersPage(context.Context, string, groups.ListMembersInput) (types.Page[groups.User], error) {
	return types.Page[groups.User]{Items: []groups.User{{}}}, nil
}
func (scGroups) ListServiceAccountsPage(context.Context, string, groups.ListServiceAccountsInput) (types.Page[groups.ServiceAccount], error) {
	return types.Page[groups.ServiceAccount]{Items: []groups.ServiceAccount{{}}}, nil
}
func (scGroups) ListInvitationsPage(context.Context, string, groups.ListInvitationsInput) (types.Page[groups.Invitation], error) {
	return types.Page[groups.Invitation]{Items: []groups.Invitation{{}}}, nil
}
func (scGroups) ListPoliciesPage(context.Context, string, groups.ListPoliciesInput) (types.Page[groups.Policy], error) {
	return types.Page[groups.Policy]{Items: []groups.Policy{{}}}, nil
}

type scServiceAccounts struct{}

func (scServiceAccounts) ListPage(context.Context, serviceaccounts.ListInput) (types.Page[serviceaccounts.ServiceAccount], error) {
	return types.Page[serviceaccounts.ServiceAccount]{Items: []serviceaccounts.ServiceAccount{{}}}, nil
}
func (scServiceAccounts) Get(context.Context, string) (*serviceaccounts.ServiceAccount, error) {
	return &serviceaccounts.ServiceAccount{}, nil
}
func (scServiceAccounts) Update(context.Context, string, serviceaccounts.UpdateInput) (*serviceaccounts.ServiceAccount, error) {
	return &serviceaccounts.ServiceAccount{}, nil
}
func (scServiceAccounts) Delete(context.Context, string) (*serviceaccounts.ServiceAccount, error) {
	return &serviceaccounts.ServiceAccount{}, nil
}

type scOciRepos struct{}

func (scOciRepos) ListPage(context.Context, ocirepos.ListInput) (types.Page[ocirepos.OciRepo], error) {
	return types.Page[ocirepos.OciRepo]{Items: []ocirepos.OciRepo{{}}}, nil
}
func (scOciRepos) Get(context.Context, string) (*ocirepos.OciRepo, error) {
	return &ocirepos.OciRepo{}, nil
}
func (scOciRepos) Create(context.Context, ocirepos.CreateInput) (*ocirepos.OciRepo, error) {
	return &ocirepos.OciRepo{}, nil
}
func (scOciRepos) Update(context.Context, string, ocirepos.UpdateInput) (*ocirepos.OciRepo, error) {
	return &ocirepos.OciRepo{}, nil
}
func (scOciRepos) Delete(context.Context, string) (*ocirepos.OciRepo, error) {
	return &ocirepos.OciRepo{}, nil
}
func (scOciRepos) CreateGrant(context.Context, string, ocirepos.CreateGrantInput) (*ocirepos.Grant, error) {
	return &ocirepos.Grant{}, nil
}
func (scOciRepos) DeleteGrant(context.Context, string) error { return nil }
func (scOciRepos) ListGrantsPage(context.Context, string, ocirepos.ListGrantsInput) (types.Page[ocirepos.Grant], error) {
	return types.Page[ocirepos.Grant]{Items: []ocirepos.Grant{{}}}, nil
}

type scPolicies struct{}

func (scPolicies) Get(context.Context, string) (*policies.Policy, error) {
	return &policies.Policy{}, nil
}
func (scPolicies) Create(context.Context, string, policies.CreatePolicyInput) (*policies.Policy, error) {
	return &policies.Policy{}, nil
}
func (scPolicies) Update(context.Context, string, policies.UpdatePolicyInput) (*policies.Policy, error) {
	return &policies.Policy{}, nil
}
func (scPolicies) Delete(context.Context, string) (*policies.Policy, error) {
	return &policies.Policy{}, nil
}
func (scPolicies) ListActions(context.Context) ([]policies.Action, error) {
	return []policies.Action{{}}, nil
}
func (scPolicies) ListEntities(context.Context) ([]policies.Entity, error) {
	return []policies.Entity{{}}, nil
}
func (scPolicies) Evaluate(context.Context, string, string) (*policies.Decision, error) {
	return &policies.Decision{}, nil
}
func (scPolicies) EvaluateBatch(context.Context, []policies.Check) ([]policies.Decision, error) {
	return []policies.Decision{{}}, nil
}
func (scPolicies) Explain(context.Context, policies.ExplainInput) ([]string, error) {
	return []string{"allowed"}, nil
}
func (scPolicies) CustomAttributeSchema(context.Context, string) (json.RawMessage, error) {
	return json.RawMessage(`{}`), nil
}
func (scPolicies) CustomAttributeValues(context.Context, organizations.AttributeScope, string) ([]string, error) {
	return []string{"eng"}, nil
}

type scServer struct{}

func (scServer) Get(context.Context) (*server.Server, error) { return &server.Server{}, nil }

type scURLs struct{}

func (scURLs) Helper(context.Context) *urls.Helper {
	return urls.NewWithBaseURL("https://app.massdriver.cloud", "org")
}
