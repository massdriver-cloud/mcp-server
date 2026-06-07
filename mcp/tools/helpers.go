package tools

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/massdriver-cloud/massdriver-sdk-go/massdriver/gql"
	"github.com/massdriver-cloud/massdriver-sdk-go/massdriver/platform/types"
	mcpsdk "github.com/modelcontextprotocol/go-sdk/mcp"
)

const (
	// defaultPageSize is the page size used when a List tool caller omits page_size.
	defaultPageSize = 25
	// maxPageSize matches the SDK's per-request upper bound.
	maxPageSize = 100
)

// clampPageSize maps a user-supplied page size onto the SDK's accepted range,
// substituting the default when zero/negative and capping the upper bound.
func clampPageSize(n int) int {
	if n <= 0 {
		return defaultPageSize
	}
	if n > maxPageSize {
		return maxPageSize
	}
	return n
}

// PageResult is the JSON shape every paginated List tool returns. Keeping the
// fields flat and consistent across tools means the model only has to learn the
// pattern once.
type PageResult[T any] struct {
	Items      []T    `json:"items"`
	NextCursor string `json:"next_cursor,omitempty"`
	HasMore    bool   `json:"has_more"`
}

// pageResult wraps an SDK page into the tool-facing PageResult shape.
func pageResult[T any](p types.Page[T]) PageResult[T] {
	return PageResult[T]{
		Items:      p.Items,
		NextCursor: p.Next,
		HasMore:    p.Next != "",
	}
}

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
