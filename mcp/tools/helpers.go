package tools

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/massdriver-cloud/massdriver-sdk-go/massdriver/gql"
	mcpsdk "github.com/modelcontextprotocol/go-sdk/mcp"
)

// textResult builds a CallToolResult with a single text content item.
func textResult(text string) *mcpsdk.CallToolResult {
	return &mcpsdk.CallToolResult{
		Content: []mcpsdk.Content{
			&mcpsdk.TextContent{Text: text},
		},
	}
}

// jsonResult serializes v to indented JSON and returns it as a text result.
func jsonResult(v any) (*mcpsdk.CallToolResult, error) {
	data, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return nil, fmt.Errorf("failed to marshal result: %w", err)
	}
	return textResult(string(data)), nil
}

// mutationErr formats a MutationFailedError into a human-readable string.
func mutationErr(err error) string {
	mf, ok := gql.AsMutationFailedError(err)
	if !ok {
		return err.Error()
	}
	parts := make([]string, 0, len(mf.Messages))
	for _, m := range mf.Messages {
		text := m.Code
		if m.Message != "" {
			text = m.Message
		}
		if m.Field != "" {
			text = fmt.Sprintf("%s: %s", m.Field, text)
		}
		parts = append(parts, text)
	}
	return strings.Join(parts, "; ")
}

// isMutationFailed returns true if err is a mutation validation failure.
func isMutationFailed(err error) bool {
	_, ok := gql.AsMutationFailedError(err)
	return ok
}
