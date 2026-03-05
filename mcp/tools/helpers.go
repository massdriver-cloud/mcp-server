package tools

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/massdriver-cloud/mcp-server/internal/api"
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

// payloadMessages formats mutation validation messages into a human-readable string.
func payloadMessages(messages []api.ValidationMessage) string {
	if len(messages) == 0 {
		return ""
	}
	parts := make([]string, 0, len(messages))
	for _, m := range messages {
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
