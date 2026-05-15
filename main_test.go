package main

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/massdriver-cloud/mcp-server/mcp"
	mcpsdk "github.com/modelcontextprotocol/go-sdk/mcp"
)

// newTestServer creates an MCP server with a nil client.
// No actual API calls are made; this only validates protocol-level behavior.
func newTestServer(t *testing.T) *mcp.Server {
	t.Helper()
	return mcp.NewServer(nil)
}

func connectTestClient(t *testing.T, server *mcp.Server) (*mcpsdk.ClientSession, context.CancelFunc) {
	t.Helper()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)

	clientTransport, serverTransport := mcpsdk.NewInMemoryTransports()

	serverSession, err := server.Connect(ctx, serverTransport)
	if err != nil {
		cancel()
		t.Fatalf("server.Connect: %v", err)
	}
	t.Cleanup(func() { _ = serverSession.Close() })

	mcpClient := mcpsdk.NewClient(&mcpsdk.Implementation{Name: "test-client", Version: "1.0.0"}, nil)
	clientSession, err := mcpClient.Connect(ctx, clientTransport, nil)
	if err != nil {
		cancel()
		t.Fatalf("client.Connect: %v", err)
	}
	t.Cleanup(func() { _ = clientSession.Close() })

	return clientSession, cancel
}

// TestMCPServerTools verifies that all expected tools are registered and have descriptions.
func TestMCPServerTools(t *testing.T) {
	server := newTestServer(t)
	clientSession, cancel := connectTestClient(t, server)
	defer cancel()

	result, err := clientSession.ListTools(context.Background(), nil)
	if err != nil {
		t.Fatalf("ListTools: %v", err)
	}

	want := []string{
		// Projects
		"list_projects", "get_project", "create_project", "update_project", "delete_project",
		// Environments
		"list_environments", "get_environment", "create_environment", "update_environment", "delete_environment",
		"set_environment_default", "remove_environment_default",
		// Instances
		"list_instances", "get_instance", "update_instance", "set_instance_secret", "remove_instance_secret", "list_alarms",
		// Deployments
		"list_deployments", "get_deployment", "get_deployment_logs", "create_deployment",
		"propose_deployment", "approve_deployment", "reject_deployment", "abort_deployment",
		// Components
		"list_components", "get_component", "add_component", "update_component", "remove_component", "link_components", "unlink_components",
		// Bundles
		"list_bundles", "get_bundle",
		// Resources
		"list_resources", "get_resource", "create_resource", "update_resource", "delete_resource", "export_resource",
		"create_resource_grant", "delete_resource_grant",
		// Organization
		"get_organization", "create_custom_attribute", "update_custom_attribute", "delete_custom_attribute",
		// Viewer
		"get_viewer",
		// Audit Logs
		"get_audit_log", "list_audit_logs", "list_audit_log_event_types",
		// Groups
		"list_groups", "get_group", "create_group", "update_group", "delete_group",
		"add_group_user", "remove_group_user", "revoke_group_invitation",
		"add_group_service_account", "remove_group_service_account",
		// Service Accounts
		"list_service_accounts", "get_service_account", "create_service_account", "update_service_account", "delete_service_account",
		// OCI Repos
		"list_oci_repos", "get_oci_repo", "create_oci_repo", "update_oci_repo",
		// Policies
		"get_policy", "create_policy", "update_policy", "delete_policy",
		"list_policy_actions", "list_policy_entities",
		"evaluate_policy", "evaluate_policies_batch", "explain_policy",
		"get_policy_attribute_schema", "list_policy_attribute_values",
		// Server
		"get_server",
		// URLs
		"get_url",
	}

	registered := make(map[string]bool, len(result.Tools))
	for _, tool := range result.Tools {
		registered[tool.Name] = true
		if tool.Description == "" {
			t.Errorf("tool %q is missing a description", tool.Name)
		}
	}

	for _, name := range want {
		if !registered[name] {
			t.Errorf("expected tool %q to be registered", name)
		}
	}

	t.Logf("registered %d tools", len(result.Tools))
}

// TestMCPServerToolSchemas logs each tool's name, description, and input schema.
func TestMCPServerToolSchemas(t *testing.T) {
	server := newTestServer(t)
	clientSession, cancel := connectTestClient(t, server)
	defer cancel()

	result, err := clientSession.ListTools(context.Background(), nil)
	if err != nil {
		t.Fatalf("ListTools: %v", err)
	}

	for _, tool := range result.Tools {
		t.Logf("tool: %s", tool.Name)
		t.Logf("  description: %s", tool.Description)
		if tool.InputSchema != nil {
			schema, _ := json.MarshalIndent(tool.InputSchema, "  ", "  ")
			t.Logf("  input_schema: %s", schema)
		}
	}
}
