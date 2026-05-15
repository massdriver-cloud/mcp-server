package tools

import (
	"testing"

	"github.com/massdriver-cloud/massdriver-sdk-go/massdriver/gql"
	mcpsdk "github.com/modelcontextprotocol/go-sdk/mcp"
)

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

// mutationFailedErr creates a *gql.MutationFailedError for testing.
func mutationFailedErr(op, field, msg string) error {
	return gql.NewMutationFailedError(op, []gql.MutationMessage{
		{Code: "invalid", Field: field, Message: msg},
	})
}
