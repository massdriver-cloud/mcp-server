package tools

import (
	"context"
	"strings"
	"testing"

	"github.com/massdriver-cloud/massdriver-sdk-go/massdriver/platform/viewer"
)

type stubViewer struct {
	getFn func(context.Context) (*viewer.Viewer, error)
}

func (s *stubViewer) Get(ctx context.Context) (*viewer.Viewer, error) {
	return s.getFn(ctx)
}

func TestHandleGetViewer(t *testing.T) {
	tests := []struct {
		name     string
		stub     *stubViewer
		wantErr  bool
		wantText string
	}{
		{
			name: "returns viewer JSON",
			stub: &stubViewer{
				getFn: func(context.Context) (*viewer.Viewer, error) {
					return &viewer.Viewer{ID: "user1", Email: "user@example.com"}, nil
				},
			},
			wantText: "user@example.com",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &Client{Viewer: tt.stub}
			handler := HandleGetViewer(c)
			result, _, err := handler(context.Background(), nil, GetViewerInput{})
			if tt.wantErr {
				if err == nil {
					t.Fatal("expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if !strings.Contains(resultText(t, result), tt.wantText) {
				t.Errorf("expected %q in result, got: %s", tt.wantText, resultText(t, result))
			}
		})
	}
}
