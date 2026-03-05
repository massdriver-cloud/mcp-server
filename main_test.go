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
	t.Cleanup(func() { serverSession.Close() })

	mcpClient := mcpsdk.NewClient(&mcpsdk.Implementation{Name: "test-client", Version: "1.0.0"}, nil)
	clientSession, err := mcpClient.Connect(ctx, clientTransport, nil)
	if err != nil {
		cancel()
		t.Fatalf("client.Connect: %v", err)
	}
	t.Cleanup(func() { clientSession.Close() })

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
		"list_projects",
		"get_project",
		"create_project",
		"update_project",
		"delete_project",
		"list_environments",
		"get_environment",
		"create_environment",
		"update_environment",
		"delete_environment",
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
