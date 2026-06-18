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

// errorResult builds a CallToolResult flagged as a tool error (IsError=true)
// with a single text content item. Use this for handled failures (such as
// validation errors surfaced by the API) so MCP clients can distinguish them
// from successful calls, rather than seeing a normal result whose text happens
// to describe a failure.
func errorResult(text string) *mcpsdk.CallToolResult {
	return &mcpsdk.CallToolResult{
		IsError: true,
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

// jsonResultStripping serializes v to indented JSON with the named keys removed
// at every level of the structure. It exists to drop large, model-useless
// fields — chiefly inline SVG `icon` blobs on bundles and OCI repos — that would
// otherwise bloat the context with no value to an AI caller. The stripped value
// is returned as both the text content and the structured content.
func jsonResultStripping(v any, keys ...string) (*mcpsdk.CallToolResult, any, error) {
	data, err := json.Marshal(v)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to marshal result: %w", err)
	}
	var generic any
	if err := json.Unmarshal(data, &generic); err != nil {
		return nil, nil, fmt.Errorf("failed to normalize result: %w", err)
	}
	stripKeys(generic, keys)

	result, err := jsonResult(generic)
	if err != nil {
		return nil, nil, err
	}
	return result, generic, nil
}

// stripKeys recursively deletes the given keys from any maps within v.
func stripKeys(v any, keys []string) {
	switch t := v.(type) {
	case map[string]any:
		for _, k := range keys {
			delete(t, k)
		}
		for _, child := range t {
			stripKeys(child, keys)
		}
	case []any:
		for _, child := range t {
			stripKeys(child, keys)
		}
	}
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
