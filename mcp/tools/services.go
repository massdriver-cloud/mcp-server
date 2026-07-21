package tools

import (
	"context"
	"encoding/json"
	"io"

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
	"github.com/massdriver-cloud/massdriver-sdk-go/massdriver/platform/server"
	"github.com/massdriver-cloud/massdriver-sdk-go/massdriver/platform/serviceaccounts"
	"github.com/massdriver-cloud/massdriver-sdk-go/massdriver/platform/types"
	"github.com/massdriver-cloud/massdriver-sdk-go/massdriver/platform/urls"
	"github.com/massdriver-cloud/massdriver-sdk-go/massdriver/platform/viewer"
)

// Client holds service interfaces used by tool handlers. In production the
// fields are backed by the real SDK service types; in tests they can be
// replaced with simple stubs.
type Client struct {
	Projects        ProjectsService
	Environments    EnvironmentsService
	Instances       InstancesService
	Deployments     DeploymentsService
	Components      ComponentsService
	Bundles         BundlesService
	Resources       ResourcesService
	Organizations   OrganizationsService
	Viewer          ViewerService
	AuditLogs       AuditLogsService
	Groups          GroupsService
	ServiceAccounts ServiceAccountsService
	OciRepos        OciReposService
	Policies        PoliciesService
	Server          ServerService
	URLs            URLsService
}

// ProjectsService defines the project operations used by tool handlers.
type ProjectsService interface {
	ListPage(ctx context.Context, input projects.ListInput) (types.Page[projects.Project], error)
	Get(ctx context.Context, id string) (*projects.Project, error)
	Create(ctx context.Context, input projects.CreateInput) (*projects.Project, error)
	Clone(ctx context.Context, sourceProjectID string, input projects.CloneInput) (*projects.Project, error)
	Update(ctx context.Context, id string, input projects.UpdateInput) (*projects.Project, error)
	Delete(ctx context.Context, id string) (*projects.Project, error)
}

// EnvironmentsService defines the environment operations used by tool handlers.
type EnvironmentsService interface {
	ListPage(ctx context.Context, input environments.ListInput) (types.Page[environments.Environment], error)
	Get(ctx context.Context, id string) (*environments.Environment, error)
	Create(ctx context.Context, projectID string, input environments.CreateInput) (*environments.Environment, error)
	Update(ctx context.Context, id string, input environments.UpdateInput) (*environments.Environment, error)
	Delete(ctx context.Context, id string) (*environments.Environment, error)
	SetDefault(ctx context.Context, environmentID, resourceID string) (*environments.EnvironmentDefault, error)
	RemoveDefault(ctx context.Context, id string) (*environments.EnvironmentDefault, error)
	Compare(ctx context.Context, sourceID, targetID string) (*environments.Comparison, error)
}

// InstancesService defines the instance operations used by tool handlers.
type InstancesService interface {
	ListPage(ctx context.Context, input instances.ListInput) (types.Page[instances.Instance], error)
	Get(ctx context.Context, id string) (*instances.Instance, error)
	Update(ctx context.Context, id string, input instances.UpdateInput) (*instances.Instance, error)
	SetSecret(ctx context.Context, instanceID, name, value string) (*instances.Secret, error)
	RemoveSecret(ctx context.Context, instanceID, name string) (*instances.Secret, error)
	SetRemoteReference(ctx context.Context, instanceID, resourceID, field string) (*instances.RemoteReference, error)
	RemoveRemoteReference(ctx context.Context, instanceID, field string) (*instances.RemoteReference, error)
	ListAlarmsPage(ctx context.Context, input instances.ListAlarmsInput) (types.Page[instances.Alarm], error)
}

// DeploymentsService defines the deployment operations used by tool handlers.
type DeploymentsService interface {
	ListPage(ctx context.Context, input deployments.ListInput) (types.Page[deployments.Deployment], error)
	Get(ctx context.Context, id string) (*deployments.Deployment, error)
	GetLogs(ctx context.Context, id string) (string, error)
	TailLogs(ctx context.Context, id string, w io.Writer) error
	Create(ctx context.Context, instanceID string, input deployments.CreateInput) (*deployments.Deployment, error)
	Propose(ctx context.Context, instanceID string, input deployments.ProposeInput) (*deployments.Deployment, error)
	Approve(ctx context.Context, id string) (*deployments.Deployment, error)
	Reject(ctx context.Context, id string) (*deployments.Deployment, error)
	Abort(ctx context.Context, id string) (*deployments.Deployment, error)
	Plan(ctx context.Context, id string) (*deployments.Deployment, error)
	Rollback(ctx context.Context, id string) (*deployments.Deployment, error)
	Compare(ctx context.Context, sourceID, targetID string) (*deployments.Comparison, error)
}

// ComponentsService defines the component operations used by tool handlers.
type ComponentsService interface {
	List(ctx context.Context, input components.ListInput) ([]components.Component, error)
	Get(ctx context.Context, id string) (*components.Component, error)
	Add(ctx context.Context, projectID string, input components.AddInput) (*components.Component, error)
	Update(ctx context.Context, id string, input components.UpdateInput) (*components.Component, error)
	Remove(ctx context.Context, id string) (*components.Component, error)
	AddLink(ctx context.Context, input components.AddLinkInput) (*components.Link, error)
	RemoveLink(ctx context.Context, linkID string) (*components.Link, error)
}

// BundlesService defines the bundle operations used by tool handlers.
type BundlesService interface {
	Get(ctx context.Context, id string) (*bundles.Bundle, error)
}

// ResourcesService defines the resource operations used by tool handlers.
type ResourcesService interface {
	ListPage(ctx context.Context, input resources.ListInput) (types.Page[resources.Resource], error)
	Get(ctx context.Context, id string) (*resources.Resource, error)
	Create(ctx context.Context, resourceTypeID string, input resources.CreateInput) (*resources.Resource, error)
	Update(ctx context.Context, id string, input resources.UpdateInput) (*resources.Resource, error)
	Delete(ctx context.Context, id string) (*resources.Resource, error)
	Export(ctx context.Context, id, format string) (*resources.Exported, error)
	CreateGrant(ctx context.Context, resourceID string, input resources.CreateGrantInput) (*resources.Grant, error)
	DeleteGrant(ctx context.Context, grantID string) error
	ListGrantsPage(ctx context.Context, resourceID string, input resources.ListGrantsInput) (types.Page[resources.Grant], error)
}

// OrganizationsService defines the organization operations used by tool handlers.
type OrganizationsService interface {
	Get(ctx context.Context) (*organizations.Organization, error)
	CreateCustomAttribute(ctx context.Context, input organizations.CreateCustomAttributeInput) (*organizations.CustomAttribute, error)
	UpdateCustomAttribute(ctx context.Context, id string, input organizations.UpdateCustomAttributeInput) (*organizations.CustomAttribute, error)
	DeleteCustomAttribute(ctx context.Context, id string) (*organizations.CustomAttribute, error)
}

// ViewerService defines the viewer operations used by tool handlers.
type ViewerService interface {
	Get(ctx context.Context) (*viewer.Viewer, error)
}

// AuditLogsService defines the audit log operations used by tool handlers.
type AuditLogsService interface {
	Get(ctx context.Context, id string) (*auditlogs.AuditLog, error)
	ListPage(ctx context.Context, input auditlogs.ListInput) (types.Page[auditlogs.AuditLog], error)
	ListEventTypes(ctx context.Context) ([]string, error)
}

// GroupsService defines the group operations used by tool handlers.
type GroupsService interface {
	ListPage(ctx context.Context, input groups.ListInput) (types.Page[groups.Group], error)
	Get(ctx context.Context, id string) (*groups.Group, error)
	Create(ctx context.Context, input groups.CreateInput) (*groups.Group, error)
	Update(ctx context.Context, id string, input groups.UpdateInput) (*groups.Group, error)
	Delete(ctx context.Context, id string) (*groups.Group, error)
	AddUser(ctx context.Context, groupID, email string) (*groups.AddUserResult, error)
	RemoveUser(ctx context.Context, groupID, email string) error
	RevokeInvitation(ctx context.Context, groupID, email string) error
	AddServiceAccount(ctx context.Context, groupID, serviceAccountID string) error
	RemoveServiceAccount(ctx context.Context, groupID, serviceAccountID string) error
}

// ServiceAccountsService defines the service account operations used by tool handlers.
type ServiceAccountsService interface {
	ListPage(ctx context.Context, input serviceaccounts.ListInput) (types.Page[serviceaccounts.ServiceAccount], error)
	Get(ctx context.Context, id string) (*serviceaccounts.ServiceAccount, error)
	Create(ctx context.Context, input serviceaccounts.CreateInput) (*serviceaccounts.Created, error)
	Update(ctx context.Context, id string, input serviceaccounts.UpdateInput) (*serviceaccounts.ServiceAccount, error)
	Delete(ctx context.Context, id string) (*serviceaccounts.ServiceAccount, error)
}

// OciReposService defines the OCI repository operations used by tool handlers.
type OciReposService interface {
	ListPage(ctx context.Context, input ocirepos.ListInput) (types.Page[ocirepos.OciRepo], error)
	Get(ctx context.Context, id string) (*ocirepos.OciRepo, error)
	Create(ctx context.Context, input ocirepos.CreateInput) (*ocirepos.OciRepo, error)
	Update(ctx context.Context, id string, input ocirepos.UpdateInput) (*ocirepos.OciRepo, error)
	Delete(ctx context.Context, id string) (*ocirepos.OciRepo, error)
	CreateGrant(ctx context.Context, repoID string, input ocirepos.CreateGrantInput) (*ocirepos.Grant, error)
	DeleteGrant(ctx context.Context, grantID string) error
	ListGrantsPage(ctx context.Context, repoID string, input ocirepos.ListGrantsInput) (types.Page[ocirepos.Grant], error)
}

// PoliciesService defines the policy operations used by tool handlers.
type PoliciesService interface {
	Get(ctx context.Context, policyID string) (*policies.Policy, error)
	Create(ctx context.Context, groupID string, input policies.CreatePolicyInput) (*policies.Policy, error)
	Update(ctx context.Context, policyID string, input policies.UpdatePolicyInput) (*policies.Policy, error)
	Delete(ctx context.Context, policyID string) (*policies.Policy, error)
	ListActions(ctx context.Context) ([]policies.Action, error)
	ListEntities(ctx context.Context) ([]policies.Entity, error)
	Evaluate(ctx context.Context, action, entityID string) (*policies.Decision, error)
	EvaluateBatch(ctx context.Context, checks []policies.Check) ([]policies.Decision, error)
	Explain(ctx context.Context, input policies.ExplainInput) ([]string, error)
	CustomAttributeSchema(ctx context.Context, action string) (json.RawMessage, error)
	CustomAttributeValues(ctx context.Context, scope organizations.AttributeScope, key string) ([]string, error)
}

// ServerService defines the server metadata operations used by tool handlers.
type ServerService interface {
	Get(ctx context.Context) (*server.Server, error)
}

// URLsService defines the URL builder operations used by tool handlers.
type URLsService interface {
	Helper(ctx context.Context) *urls.Helper
}
