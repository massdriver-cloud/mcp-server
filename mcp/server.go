package mcp

import (
	"context"
	"net/http"

	"github.com/massdriver-cloud/massdriver-sdk-go/massdriver"
	"github.com/massdriver-cloud/mcp-server/mcp/tools"
	mcpsdk "github.com/modelcontextprotocol/go-sdk/mcp"
)

// Version is the server version reported to MCP clients. It is overridden at
// build time via -ldflags "-X github.com/massdriver-cloud/mcp-server/mcp.Version=v1.2.3"
// (see Makefile, .goreleaser.yaml, and the Dockerfile).
var Version = "dev"

// Server wraps the MCP server with Massdriver-specific functionality.
type Server struct {
	mcpServer *mcpsdk.Server
	client    *tools.Client
}

// NewServer creates a new Massdriver MCP server and registers all tools.
func NewServer(client *massdriver.Client) *Server {
	impl := &mcpsdk.Implementation{
		Name:    "massdriver-mcp-server",
		Version: Version,
	}

	mcpServer := mcpsdk.NewServer(impl, nil)

	tc := clientFromSDK(client)

	s := &Server{
		mcpServer: mcpServer,
		client:    tc,
	}
	s.registerTools()
	return s
}

// clientFromSDK adapts a *massdriver.Client to the tools.Client interface bag.
// When client is nil (e.g. protocol-level tests) the fields are left nil;
// handlers will panic but that is acceptable for tests that never invoke them.
func clientFromSDK(client *massdriver.Client) *tools.Client {
	if client == nil {
		return &tools.Client{}
	}
	return &tools.Client{
		Projects:        client.Projects,
		Environments:    client.Environments,
		Instances:       client.Instances,
		Deployments:     client.Deployments,
		Components:      client.Components,
		Bundles:         client.Bundles,
		Resources:       client.Resources,
		Organizations:   client.Organizations,
		Viewer:          client.Viewer,
		AuditLogs:       client.AuditLogs,
		Groups:          client.Groups,
		ServiceAccounts: client.ServiceAccounts,
		OciRepos:        client.OciRepos,
		Policies:        client.Policies,
		Server:          client.Server,
		URLs:            client.URLs,
	}
}

func (s *Server) registerTools() {
	c := s.client

	// Projects
	mcpsdk.AddTool(s.mcpServer, tools.ListProjectsTool, tools.HandleListProjects(c))
	mcpsdk.AddTool(s.mcpServer, tools.GetProjectTool, tools.HandleGetProject(c))
	mcpsdk.AddTool(s.mcpServer, tools.CreateProjectTool, tools.HandleCreateProject(c))
	mcpsdk.AddTool(s.mcpServer, tools.UpdateProjectTool, tools.HandleUpdateProject(c))
	mcpsdk.AddTool(s.mcpServer, tools.DeleteProjectTool, tools.HandleDeleteProject(c))

	// Environments
	mcpsdk.AddTool(s.mcpServer, tools.ListEnvironmentsTool, tools.HandleListEnvironments(c))
	mcpsdk.AddTool(s.mcpServer, tools.GetEnvironmentTool, tools.HandleGetEnvironment(c))
	mcpsdk.AddTool(s.mcpServer, tools.CreateEnvironmentTool, tools.HandleCreateEnvironment(c))
	mcpsdk.AddTool(s.mcpServer, tools.UpdateEnvironmentTool, tools.HandleUpdateEnvironment(c))
	mcpsdk.AddTool(s.mcpServer, tools.DeleteEnvironmentTool, tools.HandleDeleteEnvironment(c))
	mcpsdk.AddTool(s.mcpServer, tools.SetEnvironmentDefaultTool, tools.HandleSetEnvironmentDefault(c))
	mcpsdk.AddTool(s.mcpServer, tools.RemoveEnvironmentDefaultTool, tools.HandleRemoveEnvironmentDefault(c))

	// Instances
	mcpsdk.AddTool(s.mcpServer, tools.ListInstancesTool, tools.HandleListInstances(c))
	mcpsdk.AddTool(s.mcpServer, tools.GetInstanceTool, tools.HandleGetInstance(c))
	mcpsdk.AddTool(s.mcpServer, tools.UpdateInstanceTool, tools.HandleUpdateInstance(c))
	mcpsdk.AddTool(s.mcpServer, tools.SetInstanceSecretTool, tools.HandleSetInstanceSecret(c))
	mcpsdk.AddTool(s.mcpServer, tools.RemoveInstanceSecretTool, tools.HandleRemoveInstanceSecret(c))
	mcpsdk.AddTool(s.mcpServer, tools.ListAlarmsTool, tools.HandleListAlarms(c))

	// Deployments
	mcpsdk.AddTool(s.mcpServer, tools.ListDeploymentsTool, tools.HandleListDeployments(c))
	mcpsdk.AddTool(s.mcpServer, tools.GetDeploymentTool, tools.HandleGetDeployment(c))
	mcpsdk.AddTool(s.mcpServer, tools.GetDeploymentLogsTool, tools.HandleGetDeploymentLogs(c))
	mcpsdk.AddTool(s.mcpServer, tools.CreateDeploymentTool, tools.HandleCreateDeployment(c))
	mcpsdk.AddTool(s.mcpServer, tools.ProposeDeploymentTool, tools.HandleProposeDeployment(c))
	mcpsdk.AddTool(s.mcpServer, tools.ApproveDeploymentTool, tools.HandleApproveDeployment(c))
	mcpsdk.AddTool(s.mcpServer, tools.RejectDeploymentTool, tools.HandleRejectDeployment(c))
	mcpsdk.AddTool(s.mcpServer, tools.AbortDeploymentTool, tools.HandleAbortDeployment(c))

	// Components
	mcpsdk.AddTool(s.mcpServer, tools.ListComponentsTool, tools.HandleListComponents(c))
	mcpsdk.AddTool(s.mcpServer, tools.GetComponentTool, tools.HandleGetComponent(c))
	mcpsdk.AddTool(s.mcpServer, tools.AddComponentTool, tools.HandleAddComponent(c))
	mcpsdk.AddTool(s.mcpServer, tools.UpdateComponentTool, tools.HandleUpdateComponent(c))
	mcpsdk.AddTool(s.mcpServer, tools.RemoveComponentTool, tools.HandleRemoveComponent(c))
	mcpsdk.AddTool(s.mcpServer, tools.LinkComponentsTool, tools.HandleLinkComponents(c))
	mcpsdk.AddTool(s.mcpServer, tools.UnlinkComponentsTool, tools.HandleUnlinkComponents(c))

	// Bundles
	mcpsdk.AddTool(s.mcpServer, tools.GetBundleTool, tools.HandleGetBundle(c))

	// Resources
	mcpsdk.AddTool(s.mcpServer, tools.ListResourcesTool, tools.HandleListResources(c))
	mcpsdk.AddTool(s.mcpServer, tools.GetResourceTool, tools.HandleGetResource(c))
	mcpsdk.AddTool(s.mcpServer, tools.CreateResourceTool, tools.HandleCreateResource(c))
	mcpsdk.AddTool(s.mcpServer, tools.UpdateResourceTool, tools.HandleUpdateResource(c))
	mcpsdk.AddTool(s.mcpServer, tools.DeleteResourceTool, tools.HandleDeleteResource(c))
	mcpsdk.AddTool(s.mcpServer, tools.ExportResourceTool, tools.HandleExportResource(c))
	mcpsdk.AddTool(s.mcpServer, tools.CreateResourceGrantTool, tools.HandleCreateResourceGrant(c))
	mcpsdk.AddTool(s.mcpServer, tools.DeleteResourceGrantTool, tools.HandleDeleteResourceGrant(c))
	mcpsdk.AddTool(s.mcpServer, tools.ListResourceGrantsTool, tools.HandleListResourceGrants(c))

	// Organization
	mcpsdk.AddTool(s.mcpServer, tools.GetOrganizationTool, tools.HandleGetOrganization(c))
	mcpsdk.AddTool(s.mcpServer, tools.CreateCustomAttributeTool, tools.HandleCreateCustomAttribute(c))
	mcpsdk.AddTool(s.mcpServer, tools.UpdateCustomAttributeTool, tools.HandleUpdateCustomAttribute(c))
	mcpsdk.AddTool(s.mcpServer, tools.DeleteCustomAttributeTool, tools.HandleDeleteCustomAttribute(c))

	// Viewer
	mcpsdk.AddTool(s.mcpServer, tools.GetViewerTool, tools.HandleGetViewer(c))

	// Audit Logs
	mcpsdk.AddTool(s.mcpServer, tools.GetAuditLogTool, tools.HandleGetAuditLog(c))
	mcpsdk.AddTool(s.mcpServer, tools.ListAuditLogsTool, tools.HandleListAuditLogs(c))
	mcpsdk.AddTool(s.mcpServer, tools.ListAuditLogEventTypesTool, tools.HandleListAuditLogEventTypes(c))

	// Groups
	mcpsdk.AddTool(s.mcpServer, tools.ListGroupsTool, tools.HandleListGroups(c))
	mcpsdk.AddTool(s.mcpServer, tools.GetGroupTool, tools.HandleGetGroup(c))
	mcpsdk.AddTool(s.mcpServer, tools.CreateGroupTool, tools.HandleCreateGroup(c))
	mcpsdk.AddTool(s.mcpServer, tools.UpdateGroupTool, tools.HandleUpdateGroup(c))
	mcpsdk.AddTool(s.mcpServer, tools.DeleteGroupTool, tools.HandleDeleteGroup(c))
	mcpsdk.AddTool(s.mcpServer, tools.AddGroupUserTool, tools.HandleAddGroupUser(c))
	mcpsdk.AddTool(s.mcpServer, tools.RemoveGroupUserTool, tools.HandleRemoveGroupUser(c))
	mcpsdk.AddTool(s.mcpServer, tools.RevokeGroupInvitationTool, tools.HandleRevokeGroupInvitation(c))
	mcpsdk.AddTool(s.mcpServer, tools.AddGroupServiceAccountTool, tools.HandleAddGroupServiceAccount(c))
	mcpsdk.AddTool(s.mcpServer, tools.RemoveGroupServiceAccountTool, tools.HandleRemoveGroupServiceAccount(c))

	// Service Accounts
	mcpsdk.AddTool(s.mcpServer, tools.ListServiceAccountsTool, tools.HandleListServiceAccounts(c))
	mcpsdk.AddTool(s.mcpServer, tools.GetServiceAccountTool, tools.HandleGetServiceAccount(c))
	mcpsdk.AddTool(s.mcpServer, tools.CreateServiceAccountTool, tools.HandleCreateServiceAccount(c))
	mcpsdk.AddTool(s.mcpServer, tools.UpdateServiceAccountTool, tools.HandleUpdateServiceAccount(c))
	mcpsdk.AddTool(s.mcpServer, tools.DeleteServiceAccountTool, tools.HandleDeleteServiceAccount(c))

	// OCI Repos
	mcpsdk.AddTool(s.mcpServer, tools.ListOciReposTool, tools.HandleListOciRepos(c))
	mcpsdk.AddTool(s.mcpServer, tools.GetOciRepoTool, tools.HandleGetOciRepo(c))
	mcpsdk.AddTool(s.mcpServer, tools.CreateOciRepoTool, tools.HandleCreateOciRepo(c))
	mcpsdk.AddTool(s.mcpServer, tools.UpdateOciRepoTool, tools.HandleUpdateOciRepo(c))
	mcpsdk.AddTool(s.mcpServer, tools.DeleteOciRepoTool, tools.HandleDeleteOciRepo(c))
	mcpsdk.AddTool(s.mcpServer, tools.CreateOciRepoGrantTool, tools.HandleCreateOciRepoGrant(c))
	mcpsdk.AddTool(s.mcpServer, tools.DeleteOciRepoGrantTool, tools.HandleDeleteOciRepoGrant(c))
	mcpsdk.AddTool(s.mcpServer, tools.ListOciRepoGrantsTool, tools.HandleListOciRepoGrants(c))

	// Policies
	mcpsdk.AddTool(s.mcpServer, tools.GetPolicyTool, tools.HandleGetPolicy(c))
	mcpsdk.AddTool(s.mcpServer, tools.CreatePolicyTool, tools.HandleCreatePolicy(c))
	mcpsdk.AddTool(s.mcpServer, tools.UpdatePolicyTool, tools.HandleUpdatePolicy(c))
	mcpsdk.AddTool(s.mcpServer, tools.DeletePolicyTool, tools.HandleDeletePolicy(c))
	mcpsdk.AddTool(s.mcpServer, tools.ListPolicyActionsTool, tools.HandleListPolicyActions(c))
	mcpsdk.AddTool(s.mcpServer, tools.ListPolicyEntitiesTool, tools.HandleListPolicyEntities(c))
	mcpsdk.AddTool(s.mcpServer, tools.EvaluatePolicyTool, tools.HandleEvaluatePolicy(c))
	mcpsdk.AddTool(s.mcpServer, tools.EvaluatePoliciesBatchTool, tools.HandleEvaluatePoliciesBatch(c))
	mcpsdk.AddTool(s.mcpServer, tools.ExplainPolicyTool, tools.HandleExplainPolicy(c))
	mcpsdk.AddTool(s.mcpServer, tools.GetPolicyAttributeSchemaTool, tools.HandleGetPolicyAttributeSchema(c))
	mcpsdk.AddTool(s.mcpServer, tools.ListPolicyAttributeValuesTool, tools.HandleListPolicyAttributeValues(c))

	// Server
	mcpsdk.AddTool(s.mcpServer, tools.GetServerTool, tools.HandleGetServer(c))

	// URLs
	mcpsdk.AddTool(s.mcpServer, tools.GetURLTool, tools.HandleGetURL(c))
}

// Run starts the MCP server with the specified transport.
func (s *Server) Run(ctx context.Context, transport mcpsdk.Transport) error {
	return s.mcpServer.Run(ctx, transport)
}

// Connect attaches the server to a transport and returns the session.
// Primarily used for testing.
func (s *Server) Connect(ctx context.Context, transport mcpsdk.Transport) (*mcpsdk.ServerSession, error) {
	return s.mcpServer.Connect(ctx, transport, nil)
}

// HTTPHandler returns an http.Handler that serves this server over the MCP
// Streamable HTTP transport. The handler is stateless (no Mcp-Session-Id
// affinity, which suits horizontal scaling since all state lives in the
// Massdriver API) and replies with application/json rather than SSE, since
// every tool is a simple request/response. The same underlying server instance
// is reused for all requests.
func (s *Server) HTTPHandler() http.Handler {
	return mcpsdk.NewStreamableHTTPHandler(
		func(*http.Request) *mcpsdk.Server { return s.mcpServer },
		&mcpsdk.StreamableHTTPOptions{Stateless: true, JSONResponse: true},
	)
}
