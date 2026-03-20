package mcp

import (
	"context"

	mdclient "github.com/massdriver-cloud/massdriver-sdk-go/massdriver/client"
	"github.com/massdriver-cloud/mcp-server/mcp/tools"
	mcpsdk "github.com/modelcontextprotocol/go-sdk/mcp"
)

// Server wraps the MCP server with Massdriver-specific functionality.
type Server struct {
	mcpServer *mcpsdk.Server
	client    *mdclient.Client
}

// NewServer creates a new Massdriver MCP server and registers all tools.
func NewServer(client *mdclient.Client) *Server {
	impl := &mcpsdk.Implementation{
		Name:    "massdriver-mcp-server",
		Version: "1.0.0",
	}

	mcpServer := mcpsdk.NewServer(impl, nil)

	s := &Server{
		mcpServer: mcpServer,
		client:    client,
	}
	s.registerTools()
	return s
}

func (s *Server) registerTools() {
	// Projects
	mcpsdk.AddTool(s.mcpServer, tools.ListProjectsTool, tools.HandleListProjects(s.client))
	mcpsdk.AddTool(s.mcpServer, tools.GetProjectTool, tools.HandleGetProject(s.client))
	mcpsdk.AddTool(s.mcpServer, tools.CreateProjectTool, tools.HandleCreateProject(s.client))
	mcpsdk.AddTool(s.mcpServer, tools.UpdateProjectTool, tools.HandleUpdateProject(s.client))
	mcpsdk.AddTool(s.mcpServer, tools.DeleteProjectTool, tools.HandleDeleteProject(s.client))

	// Environments
	mcpsdk.AddTool(s.mcpServer, tools.ListEnvironmentsTool, tools.HandleListEnvironments(s.client))
	mcpsdk.AddTool(s.mcpServer, tools.GetEnvironmentTool, tools.HandleGetEnvironment(s.client))
	mcpsdk.AddTool(s.mcpServer, tools.CreateEnvironmentTool, tools.HandleCreateEnvironment(s.client))
	mcpsdk.AddTool(s.mcpServer, tools.UpdateEnvironmentTool, tools.HandleUpdateEnvironment(s.client))
	mcpsdk.AddTool(s.mcpServer, tools.DeleteEnvironmentTool, tools.HandleDeleteEnvironment(s.client))

	// Instances
	mcpsdk.AddTool(s.mcpServer, tools.ListInstancesTool, tools.HandleListInstances(s.client))
	mcpsdk.AddTool(s.mcpServer, tools.GetInstanceTool, tools.HandleGetInstance(s.client))

	// Deployments
	mcpsdk.AddTool(s.mcpServer, tools.ListDeploymentsTool, tools.HandleListDeployments(s.client))
	mcpsdk.AddTool(s.mcpServer, tools.GetDeploymentTool, tools.HandleGetDeployment(s.client))

	// Blueprint
	mcpsdk.AddTool(s.mcpServer, tools.AddComponentTool, tools.HandleAddComponent(s.client))
	mcpsdk.AddTool(s.mcpServer, tools.RemoveComponentTool, tools.HandleRemoveComponent(s.client))
	mcpsdk.AddTool(s.mcpServer, tools.LinkComponentsTool, tools.HandleLinkComponents(s.client))
	mcpsdk.AddTool(s.mcpServer, tools.UnlinkComponentsTool, tools.HandleUnlinkComponents(s.client))
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
