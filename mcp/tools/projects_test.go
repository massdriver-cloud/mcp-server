package tools

import (
	"context"
	"strings"
	"testing"

	"github.com/massdriver-cloud/massdriver-sdk-go/massdriver/platform/projects"
	"github.com/massdriver-cloud/massdriver-sdk-go/massdriver/platform/types"
)

type stubProjects struct {
	listPageFn func(context.Context, projects.ListInput) (types.Page[projects.Project], error)
	getFn      func(context.Context, string) (*projects.Project, error)
	createFn   func(context.Context, projects.CreateInput) (*projects.Project, error)
	updateFn   func(context.Context, string, projects.UpdateInput) (*projects.Project, error)
	deleteFn   func(context.Context, string) (*projects.Project, error)
}

func (s *stubProjects) ListPage(ctx context.Context, input projects.ListInput) (types.Page[projects.Project], error) {
	return s.listPageFn(ctx, input)
}
func (s *stubProjects) Get(ctx context.Context, id string) (*projects.Project, error) {
	return s.getFn(ctx, id)
}
func (s *stubProjects) Create(ctx context.Context, input projects.CreateInput) (*projects.Project, error) {
	return s.createFn(ctx, input)
}
func (s *stubProjects) Update(ctx context.Context, id string, input projects.UpdateInput) (*projects.Project, error) {
	return s.updateFn(ctx, id, input)
}
func (s *stubProjects) Delete(ctx context.Context, id string) (*projects.Project, error) {
	return s.deleteFn(ctx, id)
}

func TestHandleListProjects(t *testing.T) {
	tests := []struct {
		name     string
		stub     *stubProjects
		wantErr  bool
		wantText string
	}{
		{
			name: "returns page of projects as JSON",
			stub: &stubProjects{
				listPageFn: func(context.Context, projects.ListInput) (types.Page[projects.Project], error) {
					return types.Page[projects.Project]{
						Items: []projects.Project{{ID: "myproj", Name: "My Project"}},
						Next:  "cursor-2",
					}, nil
				},
			},
			wantText: "myproj",
		},
		{
			name: "empty page surfaces empty items and has_more false",
			stub: &stubProjects{
				listPageFn: func(context.Context, projects.ListInput) (types.Page[projects.Project], error) {
					return types.Page[projects.Project]{}, nil
				},
			},
			wantText: "\"has_more\": false",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &Client{Projects: tt.stub}
			handler := HandleListProjects(c)
			result, _, err := handler(context.Background(), nil, ListProjectsInput{})
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

func TestHandleGetProject(t *testing.T) {
	tests := []struct {
		name     string
		input    GetProjectInput
		stub     *stubProjects
		wantErr  string
		wantText string
	}{
		{
			name:    "missing id",
			input:   GetProjectInput{},
			stub:    &stubProjects{},
			wantErr: "id is required",
		},
		{
			name:  "returns project JSON",
			input: GetProjectInput{ID: "myproj"},
			stub: &stubProjects{
				getFn: func(_ context.Context, id string) (*projects.Project, error) {
					return &projects.Project{ID: id, Name: "My Project"}, nil
				},
			},
			wantText: "myproj",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &Client{Projects: tt.stub}
			handler := HandleGetProject(c)
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

func TestHandleCreateProject(t *testing.T) {
	tests := []struct {
		name     string
		input    CreateProjectInput
		stub     *stubProjects
		wantErr  string
		wantText string
	}{
		{
			name:    "missing id",
			input:   CreateProjectInput{Name: "My Project"},
			stub:    &stubProjects{},
			wantErr: "id is required",
		},
		{
			name:    "missing name",
			input:   CreateProjectInput{ID: "myproj"},
			stub:    &stubProjects{},
			wantErr: "name is required",
		},
		{
			name:  "success returns project JSON",
			input: CreateProjectInput{ID: "myproj", Name: "My Project"},
			stub: &stubProjects{
				createFn: func(_ context.Context, input projects.CreateInput) (*projects.Project, error) {
					return &projects.Project{ID: input.ID, Name: input.Name}, nil
				},
			},
			wantText: "myproj",
		},
		{
			name:  "mutation failure returns error message",
			input: CreateProjectInput{ID: "myproj", Name: "My Project"},
			stub: &stubProjects{
				createFn: func(context.Context, projects.CreateInput) (*projects.Project, error) {
					return nil, mutationFailedErr("create project", "id", "already taken")
				},
			},
			wantText: "create_project failed",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &Client{Projects: tt.stub}
			handler := HandleCreateProject(c)
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

func TestHandleUpdateProject(t *testing.T) {
	tests := []struct {
		name     string
		input    UpdateProjectInput
		stub     *stubProjects
		wantErr  string
		wantText string
	}{
		{
			name:    "missing id",
			input:   UpdateProjectInput{Name: "New Name"},
			stub:    &stubProjects{},
			wantErr: "id is required",
		},
		{
			name:  "success returns updated project JSON",
			input: UpdateProjectInput{ID: "myproj", Name: "Updated Name"},
			stub: &stubProjects{
				updateFn: func(_ context.Context, id string, input projects.UpdateInput) (*projects.Project, error) {
					return &projects.Project{ID: id, Name: input.Name}, nil
				},
			},
			wantText: "Updated Name",
		},
		{
			name:  "mutation failure returns error message",
			input: UpdateProjectInput{ID: "myproj"},
			stub: &stubProjects{
				updateFn: func(context.Context, string, projects.UpdateInput) (*projects.Project, error) {
					return nil, mutationFailedErr("update project", "name", "too long")
				},
			},
			wantText: "update_project failed",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &Client{Projects: tt.stub}
			handler := HandleUpdateProject(c)
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

func TestHandleDeleteProject(t *testing.T) {
	tests := []struct {
		name     string
		input    DeleteProjectInput
		stub     *stubProjects
		wantErr  string
		wantText string
	}{
		{
			name:    "missing id",
			input:   DeleteProjectInput{},
			stub:    &stubProjects{},
			wantErr: "id is required",
		},
		{
			name:  "success returns confirmation message",
			input: DeleteProjectInput{ID: "myproj"},
			stub: &stubProjects{
				deleteFn: func(_ context.Context, id string) (*projects.Project, error) {
					return &projects.Project{ID: id}, nil
				},
			},
			wantText: "deleted successfully",
		},
		{
			name:  "mutation failure returns error message",
			input: DeleteProjectInput{ID: "myproj"},
			stub: &stubProjects{
				deleteFn: func(context.Context, string) (*projects.Project, error) {
					return nil, mutationFailedErr("delete project", "", "not empty")
				},
			},
			wantText: "delete_project failed",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &Client{Projects: tt.stub}
			handler := HandleDeleteProject(c)
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
