package tools

import (
	"context"
	"strings"
	"testing"

	"github.com/massdriver-cloud/massdriver-sdk-go/massdriver/platform/bundles"
)

type stubBundles struct {
	getFn func(context.Context, string) (*bundles.Bundle, error)
}

func (s *stubBundles) Get(ctx context.Context, id string) (*bundles.Bundle, error) {
	return s.getFn(ctx, id)
}

func TestHandleGetBundle(t *testing.T) {
	tests := []struct {
		name     string
		input    GetBundleInput
		stub     *stubBundles
		wantErr  string
		wantText string
	}{
		{
			name:    "missing id",
			input:   GetBundleInput{},
			stub:    &stubBundles{},
			wantErr: "id is required",
		},
		{
			name:  "returns bundle JSON",
			input: GetBundleInput{ID: "aws-aurora-postgres@latest"},
			stub: &stubBundles{
				getFn: func(_ context.Context, id string) (*bundles.Bundle, error) {
					return &bundles.Bundle{ID: "aws-aurora-postgres", Name: "Aurora PostgreSQL"}, nil
				},
			},
			wantText: "Aurora PostgreSQL",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &Client{Bundles: tt.stub}
			handler := HandleGetBundle(c)
			result, _, err := handler(context.Background(), nil, tt.input)
			if tt.wantErr != "" {
				if err == nil || !strings.Contains(err.Error(), tt.wantErr) {
					t.Fatalf("expected error containing %q, got: %v", tt.wantErr, err)
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
