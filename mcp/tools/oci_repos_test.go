package tools

import (
	"context"
	"strings"
	"testing"

	"github.com/massdriver-cloud/massdriver-sdk-go/massdriver/platform/ocirepos"
	"github.com/massdriver-cloud/massdriver-sdk-go/massdriver/platform/types"
)

type stubOciRepos struct {
	listPageFn func(context.Context, ocirepos.ListInput) (types.Page[ocirepos.OciRepo], error)
	getFn      func(context.Context, string) (*ocirepos.OciRepo, error)
	createFn   func(context.Context, ocirepos.CreateInput) (*ocirepos.OciRepo, error)
	updateFn   func(context.Context, string, ocirepos.UpdateInput) (*ocirepos.OciRepo, error)
}

func (s *stubOciRepos) ListPage(ctx context.Context, input ocirepos.ListInput) (types.Page[ocirepos.OciRepo], error) {
	return s.listPageFn(ctx, input)
}
func (s *stubOciRepos) Get(ctx context.Context, id string) (*ocirepos.OciRepo, error) {
	return s.getFn(ctx, id)
}
func (s *stubOciRepos) Create(ctx context.Context, input ocirepos.CreateInput) (*ocirepos.OciRepo, error) {
	return s.createFn(ctx, input)
}
func (s *stubOciRepos) Update(ctx context.Context, id string, input ocirepos.UpdateInput) (*ocirepos.OciRepo, error) {
	return s.updateFn(ctx, id, input)
}

func TestHandleListOciRepos(t *testing.T) {
	tests := []struct {
		name     string
		stub     *stubOciRepos
		wantText string
	}{
		{
			name: "success",
			stub: &stubOciRepos{
				listPageFn: func(context.Context, ocirepos.ListInput) (types.Page[ocirepos.OciRepo], error) {
					return types.Page[ocirepos.OciRepo]{
						Items: []ocirepos.OciRepo{{ID: "repo1", Name: "my-repo"}},
					}, nil
				},
			},
			wantText: "repo1",
		},
		{
			name: "empty page surfaces has_more false",
			stub: &stubOciRepos{
				listPageFn: func(context.Context, ocirepos.ListInput) (types.Page[ocirepos.OciRepo], error) {
					return types.Page[ocirepos.OciRepo]{}, nil
				},
			},
			wantText: "\"has_more\": false",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &Client{OciRepos: tt.stub}
			handler := HandleListOciRepos(c)
			result, _, err := handler(context.Background(), nil, ListOciReposInput{})
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if !strings.Contains(resultText(t, result), tt.wantText) {
				t.Errorf("expected %q in result, got: %s", tt.wantText, resultText(t, result))
			}
		})
	}
}

func TestHandleGetOciRepo(t *testing.T) {
	tests := []struct {
		name     string
		input    GetOciRepoInput
		stub     *stubOciRepos
		wantErr  string
		wantText string
	}{
		{
			name:    "missing id",
			input:   GetOciRepoInput{},
			stub:    &stubOciRepos{},
			wantErr: "id is required",
		},
		{
			name:  "success",
			input: GetOciRepoInput{ID: "repo1"},
			stub: &stubOciRepos{
				getFn: func(_ context.Context, id string) (*ocirepos.OciRepo, error) {
					return &ocirepos.OciRepo{ID: id, Name: "my-repo"}, nil
				},
			},
			wantText: "repo1",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &Client{OciRepos: tt.stub}
			handler := HandleGetOciRepo(c)
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

func TestHandleCreateOciRepo(t *testing.T) {
	tests := []struct {
		name     string
		input    CreateOciRepoInput
		stub     *stubOciRepos
		wantErr  string
		wantText string
	}{
		{
			name:    "missing id",
			input:   CreateOciRepoInput{ArtifactType: "BUNDLE"},
			stub:    &stubOciRepos{},
			wantErr: "id is required",
		},
		{
			name:    "missing artifact_type",
			input:   CreateOciRepoInput{ID: "repo1"},
			stub:    &stubOciRepos{},
			wantErr: "artifact_type is required",
		},
		{
			name:  "success",
			input: CreateOciRepoInput{ID: "repo1", ArtifactType: "BUNDLE"},
			stub: &stubOciRepos{
				createFn: func(_ context.Context, input ocirepos.CreateInput) (*ocirepos.OciRepo, error) {
					return &ocirepos.OciRepo{ID: input.ID, Name: input.ID}, nil
				},
			},
			wantText: "repo1",
		},
		{
			name:  "mutation failure",
			input: CreateOciRepoInput{ID: "repo1", ArtifactType: "BUNDLE"},
			stub: &stubOciRepos{
				createFn: func(context.Context, ocirepos.CreateInput) (*ocirepos.OciRepo, error) {
					return nil, mutationFailedErr("create oci repo", "id", "already taken")
				},
			},
			wantText: "create_oci_repo failed",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &Client{OciRepos: tt.stub}
			handler := HandleCreateOciRepo(c)
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

func TestHandleUpdateOciRepo(t *testing.T) {
	tests := []struct {
		name     string
		input    UpdateOciRepoInput
		stub     *stubOciRepos
		wantErr  string
		wantText string
	}{
		{
			name:    "missing id",
			input:   UpdateOciRepoInput{},
			stub:    &stubOciRepos{},
			wantErr: "id is required",
		},
		{
			name:  "success",
			input: UpdateOciRepoInput{ID: "repo1", Attributes: map[string]any{"key": "val"}},
			stub: &stubOciRepos{
				updateFn: func(_ context.Context, id string, _ ocirepos.UpdateInput) (*ocirepos.OciRepo, error) {
					return &ocirepos.OciRepo{ID: id, Name: "my-repo"}, nil
				},
			},
			wantText: "repo1",
		},
		{
			name:  "mutation failure",
			input: UpdateOciRepoInput{ID: "repo1"},
			stub: &stubOciRepos{
				updateFn: func(context.Context, string, ocirepos.UpdateInput) (*ocirepos.OciRepo, error) {
					return nil, mutationFailedErr("update oci repo", "attributes", "invalid")
				},
			},
			wantText: "update_oci_repo failed",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &Client{OciRepos: tt.stub}
			handler := HandleUpdateOciRepo(c)
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
