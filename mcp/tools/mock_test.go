package tools

import (
	"testing"

	mdclient "github.com/massdriver-cloud/massdriver-sdk-go/massdriver/client"
	"github.com/massdriver-cloud/mcp-server/internal/gqlmock"
	mcpsdk "github.com/modelcontextprotocol/go-sdk/mcp"
)

// newToolClient creates a *client.Client backed by an in-order mock GQL response sequence.
func newToolClient(responses []any) *mdclient.Client {
	return &mdclient.Client{GQL: gqlmock.NewClientWithJSONResponseArray(responses)}
}

// resultText extracts the text from the first TextContent item in a CallToolResult.
func resultText(t *testing.T, result *mcpsdk.CallToolResult) string {
	t.Helper()
	if result == nil {
		t.Fatal("result is nil")
	}
	if len(result.Content) == 0 {
		t.Fatal("result.Content is empty")
	}
	tc, ok := result.Content[0].(*mcpsdk.TextContent)
	if !ok {
		t.Fatalf("content[0] is %T, not *mcpsdk.TextContent", result.Content[0])
	}
	return tc.Text
}
