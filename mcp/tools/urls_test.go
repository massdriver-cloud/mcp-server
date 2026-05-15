package tools

import (
	"context"
	"strings"
	"testing"

	"github.com/massdriver-cloud/massdriver-sdk-go/massdriver/platform/urls"
)

type stubURLs struct {
	helperFn func(context.Context) *urls.Helper
}

func (s *stubURLs) Helper(ctx context.Context) *urls.Helper {
	return s.helperFn(ctx)
}

func TestHandleGetURL(t *testing.T) {
	helper := urls.NewWithBaseURL("https://app.example.com", "org1")

	tests := []struct {
		name     string
		input    GetURLInput
		stub     *stubURLs
		wantErr  string
		wantText string
	}{
		{
			name:    "missing type",
			input:   GetURLInput{},
			stub:    &stubURLs{helperFn: func(context.Context) *urls.Helper { return helper }},
			wantErr: "type is required",
		},
		{
			name:    "unknown type",
			input:   GetURLInput{Type: "bogus"},
			stub:    &stubURLs{helperFn: func(context.Context) *urls.Helper { return helper }},
			wantErr: "unknown type",
		},
		{
			name:     "organization",
			input:    GetURLInput{Type: "organization"},
			stub:     &stubURLs{helperFn: func(context.Context) *urls.Helper { return helper }},
			wantText: "https://app.example.com/orgs/org1/",
		},
		{
			name:    "project missing id",
			input:   GetURLInput{Type: "project"},
			stub:    &stubURLs{helperFn: func(context.Context) *urls.Helper { return helper }},
			wantErr: "id is required",
		},
		{
			name:     "project success",
			input:    GetURLInput{Type: "project", ID: "myproj"},
			stub:     &stubURLs{helperFn: func(context.Context) *urls.Helper { return helper }},
			wantText: "https://app.example.com/orgs/org1/projects/myproj/",
		},
		{
			name:    "bundle missing id",
			input:   GetURLInput{Type: "bundle", Version: "1.0.0"},
			stub:    &stubURLs{helperFn: func(context.Context) *urls.Helper { return helper }},
			wantErr: "id is required",
		},
		{
			name:    "bundle missing version",
			input:   GetURLInput{Type: "bundle", ID: "my-bundle"},
			stub:    &stubURLs{helperFn: func(context.Context) *urls.Helper { return helper }},
			wantErr: "version is required",
		},
		{
			name:     "bundle success",
			input:    GetURLInput{Type: "bundle", ID: "my-bundle", Version: "1.0.0"},
			stub:     &stubURLs{helperFn: func(context.Context) *urls.Helper { return helper }},
			wantText: "https://app.example.com/orgs/org1/repos/my-bundle/1.0.0",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &Client{URLs: tt.stub}
			handler := HandleGetURL(c)
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
