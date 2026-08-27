package tools

import (
	"strings"
	"testing"
	"time"

	"github.com/massdriver-cloud/massdriver-sdk-go/massdriver/gql"
	mcpsdk "github.com/modelcontextprotocol/go-sdk/mcp"
)

// ptr returns a pointer to v — a test helper for building the *string/*bool
// fields on the partial-update tool inputs.
func ptr[T any](v T) *T { return &v }

// derefStr safely dereferences a *string, returning "" for nil.
func derefStr(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}

// resultText extracts the text from the first TextContent item in a CallToolResult.
func resultText(t *testing.T, result *mcpsdk.CallToolResult) string {
	t.Helper()
	if result == nil {
		t.Fatal("result is nil")
		return ""
	}
	if len(result.Content) == 0 {
		t.Fatal("result.Content is empty")
		return ""
	}
	tc, ok := result.Content[0].(*mcpsdk.TextContent)
	if !ok {
		t.Fatalf("content[0] is %T, not *mcpsdk.TextContent", result.Content[0])
		return ""
	}
	return tc.Text
}

// mutationFailedErr creates a *gql.MutationFailedError for testing.
func mutationFailedErr(op, field, msg string) error {
	return gql.NewMutationFailedError(op, []gql.MutationMessage{
		{Code: "invalid", Field: field, Message: msg},
	})
}

func TestToAttributeFilters(t *testing.T) {
	if got := toAttributeFilters(nil); got != nil {
		t.Errorf("empty input should map to nil so the filter stays off the wire, got %v", got)
	}

	got := toAttributeFilters([]AttributeFilterInput{
		{Key: "team", Eq: "platform"},
		{Key: "cost_center", In: []string{"a", "b"}},
	})
	if len(got) != 2 {
		t.Fatalf("len = %d, want 2", len(got))
	}
	if got[0].Key != "team" || got[0].Eq != "platform" {
		t.Errorf("got[0] = %+v, want key=team eq=platform", got[0])
	}
	if got[1].Key != "cost_center" || len(got[1].In) != 2 {
		t.Errorf("got[1] = %+v, want key=cost_center with 2 values", got[1])
	}
}

func TestParseTimestamp(t *testing.T) {
	t.Run("empty stays zero so the bound is open", func(t *testing.T) {
		got, err := parseTimestamp("list_resources", "created_after", "")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !got.IsZero() {
			t.Errorf("got %v, want the zero time", got)
		}
	})

	t.Run("parses RFC 3339", func(t *testing.T) {
		got, err := parseTimestamp("list_resources", "created_after", "2026-01-15T00:00:00Z")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if got.Year() != 2026 || got.Month() != time.January || got.Day() != 15 {
			t.Errorf("got %v, want 2026-01-15", got)
		}
	})

	t.Run("rejects a malformed value rather than dropping the filter", func(t *testing.T) {
		_, err := parseTimestamp("list_resources", "created_after", "yesterday")
		if err == nil {
			t.Fatal("expected an error, got nil")
		}
		for _, want := range []string{"list_resources", "created_after", "RFC 3339"} {
			if !strings.Contains(err.Error(), want) {
				t.Errorf("error %q should mention %q", err, want)
			}
		}
	})
}
