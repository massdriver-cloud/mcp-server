package tools

import (
	"context"
	"strings"
	"testing"

	"github.com/massdriver-cloud/massdriver-sdk-go/massdriver/platform/environments"
)

type stubEnvironments struct {
	listFn          func(context.Context, environments.ListInput) ([]environments.Environment, error)
	getFn           func(context.Context, string) (*environments.Environment, error)
	createFn        func(context.Context, string, environments.CreateInput) (*environments.Environment, error)
	updateFn        func(context.Context, string, environments.UpdateInput) (*environments.Environment, error)
	deleteFn        func(context.Context, string) (*environments.Environment, error)
	setDefaultFn    func(context.Context, string, string) (*environments.EnvironmentDefault, error)
	removeDefaultFn func(context.Context, string) (*environments.EnvironmentDefault, error)
}

func (s *stubEnvironments) List(ctx context.Context, input environments.ListInput) ([]environments.Environment, error) {
	return s.listFn(ctx, input)
}
func (s *stubEnvironments) Get(ctx context.Context, id string) (*environments.Environment, error) {
	return s.getFn(ctx, id)
}
func (s *stubEnvironments) Create(ctx context.Context, projectID string, input environments.CreateInput) (*environments.Environment, error) {
	return s.createFn(ctx, projectID, input)
}
func (s *stubEnvironments) Update(ctx context.Context, id string, input environments.UpdateInput) (*environments.Environment, error) {
	return s.updateFn(ctx, id, input)
}
func (s *stubEnvironments) Delete(ctx context.Context, id string) (*environments.Environment, error) {
	return s.deleteFn(ctx, id)
}
func (s *stubEnvironments) SetDefault(ctx context.Context, environmentID, resourceID string) (*environments.EnvironmentDefault, error) {
	return s.setDefaultFn(ctx, environmentID, resourceID)
}
func (s *stubEnvironments) RemoveDefault(ctx context.Context, id string) (*environments.EnvironmentDefault, error) {
	return s.removeDefaultFn(ctx, id)
}

func TestHandleListEnvironments(t *testing.T) {
	tests := []struct {
		name     string
		input    ListEnvironmentsInput
		stub     *stubEnvironments
		wantErr  bool
		wantText string
	}{
		{
			name:  "returns all environments",
			input: ListEnvironmentsInput{},
			stub: &stubEnvironments{
				listFn: func(_ context.Context, _ environments.ListInput) ([]environments.Environment, error) {
					return []environments.Environment{{ID: "myproj-staging", Name: "Staging"}}, nil
				},
			},
			wantText: "myproj-staging",
		},
		{
			name:  "returns null for empty list",
			input: ListEnvironmentsInput{},
			stub: &stubEnvironments{
				listFn: func(context.Context, environments.ListInput) ([]environments.Environment, error) {
					return nil, nil
				},
			},
			wantText: "null",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &Client{Environments: tt.stub}
			handler := HandleListEnvironments(c)
			result, _, err := handler(context.Background(), nil, tt.input)
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

func TestHandleGetEnvironment(t *testing.T) {
	tests := []struct {
		name     string
		input    GetEnvironmentInput
		stub     *stubEnvironments
		wantErr  string
		wantText string
	}{
		{
			name:    "missing id",
			input:   GetEnvironmentInput{},
			stub:    &stubEnvironments{},
			wantErr: "id is required",
		},
		{
			name:  "returns environment JSON",
			input: GetEnvironmentInput{ID: "myproj-staging"},
			stub: &stubEnvironments{
				getFn: func(_ context.Context, id string) (*environments.Environment, error) {
					return &environments.Environment{ID: id, Name: "Staging"}, nil
				},
			},
			wantText: "myproj-staging",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &Client{Environments: tt.stub}
			handler := HandleGetEnvironment(c)
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

func TestHandleCreateEnvironment(t *testing.T) {
	tests := []struct {
		name     string
		input    CreateEnvironmentInput
		stub     *stubEnvironments
		wantErr  string
		wantText string
	}{
		{
			name:    "missing project_id",
			input:   CreateEnvironmentInput{ID: "staging", Name: "Staging"},
			stub:    &stubEnvironments{},
			wantErr: "project_id is required",
		},
		{
			name:    "missing id",
			input:   CreateEnvironmentInput{ProjectID: "myproj", Name: "Staging"},
			stub:    &stubEnvironments{},
			wantErr: "id is required",
		},
		{
			name:    "missing name",
			input:   CreateEnvironmentInput{ProjectID: "myproj", ID: "staging"},
			stub:    &stubEnvironments{},
			wantErr: "name is required",
		},
		{
			name:  "success returns environment JSON",
			input: CreateEnvironmentInput{ProjectID: "myproj", ID: "staging", Name: "Staging"},
			stub: &stubEnvironments{
				createFn: func(_ context.Context, projectID string, input environments.CreateInput) (*environments.Environment, error) {
					return &environments.Environment{ID: projectID + "-" + input.ID, Name: input.Name}, nil
				},
			},
			wantText: "myproj-staging",
		},
		{
			name:  "mutation failure returns error message",
			input: CreateEnvironmentInput{ProjectID: "myproj", ID: "staging", Name: "Staging"},
			stub: &stubEnvironments{
				createFn: func(context.Context, string, environments.CreateInput) (*environments.Environment, error) {
					return nil, mutationFailedErr("create environment", "id", "already exists")
				},
			},
			wantText: "create_environment failed",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &Client{Environments: tt.stub}
			handler := HandleCreateEnvironment(c)
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

func TestHandleUpdateEnvironment(t *testing.T) {
	tests := []struct {
		name     string
		input    UpdateEnvironmentInput
		stub     *stubEnvironments
		wantErr  string
		wantText string
	}{
		{
			name:    "missing id",
			input:   UpdateEnvironmentInput{Name: "New Name"},
			stub:    &stubEnvironments{},
			wantErr: "id is required",
		},
		{
			name:  "success returns updated environment JSON",
			input: UpdateEnvironmentInput{ID: "myproj-staging", Name: "Production"},
			stub: &stubEnvironments{
				updateFn: func(_ context.Context, id string, input environments.UpdateInput) (*environments.Environment, error) {
					return &environments.Environment{ID: id, Name: input.Name}, nil
				},
			},
			wantText: "Production",
		},
		{
			name:  "mutation failure returns error message",
			input: UpdateEnvironmentInput{ID: "myproj-staging"},
			stub: &stubEnvironments{
				updateFn: func(context.Context, string, environments.UpdateInput) (*environments.Environment, error) {
					return nil, mutationFailedErr("update environment", "name", "too long")
				},
			},
			wantText: "update_environment failed",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &Client{Environments: tt.stub}
			handler := HandleUpdateEnvironment(c)
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

func TestHandleDeleteEnvironment(t *testing.T) {
	tests := []struct {
		name     string
		input    DeleteEnvironmentInput
		stub     *stubEnvironments
		wantErr  string
		wantText string
	}{
		{
			name:    "missing id",
			input:   DeleteEnvironmentInput{},
			stub:    &stubEnvironments{},
			wantErr: "id is required",
		},
		{
			name:  "success returns confirmation message",
			input: DeleteEnvironmentInput{ID: "myproj-staging"},
			stub: &stubEnvironments{
				deleteFn: func(_ context.Context, id string) (*environments.Environment, error) {
					return &environments.Environment{ID: id}, nil
				},
			},
			wantText: "deleted successfully",
		},
		{
			name:  "mutation failure returns error message",
			input: DeleteEnvironmentInput{ID: "myproj-staging"},
			stub: &stubEnvironments{
				deleteFn: func(context.Context, string) (*environments.Environment, error) {
					return nil, mutationFailedErr("delete environment", "", "packages still active")
				},
			},
			wantText: "delete_environment failed",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &Client{Environments: tt.stub}
			handler := HandleDeleteEnvironment(c)
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

func TestHandleSetEnvironmentDefault(t *testing.T) {
	tests := []struct {
		name     string
		input    SetEnvironmentDefaultInput
		stub     *stubEnvironments
		wantErr  string
		wantText string
	}{
		{
			name:    "missing environment_id",
			input:   SetEnvironmentDefaultInput{ResourceID: "res1"},
			stub:    &stubEnvironments{},
			wantErr: "environment_id is required",
		},
		{
			name:    "missing resource_id",
			input:   SetEnvironmentDefaultInput{EnvironmentID: "env1"},
			stub:    &stubEnvironments{},
			wantErr: "resource_id is required",
		},
		{
			name:  "success returns environment default JSON",
			input: SetEnvironmentDefaultInput{EnvironmentID: "env1", ResourceID: "res1"},
			stub: &stubEnvironments{
				setDefaultFn: func(_ context.Context, _, _ string) (*environments.EnvironmentDefault, error) {
					return &environments.EnvironmentDefault{ID: "def1"}, nil
				},
			},
			wantText: "def1",
		},
		{
			name:  "mutation failure returns error message",
			input: SetEnvironmentDefaultInput{EnvironmentID: "env1", ResourceID: "res1"},
			stub: &stubEnvironments{
				setDefaultFn: func(context.Context, string, string) (*environments.EnvironmentDefault, error) {
					return nil, mutationFailedErr("set environment default", "resource", "incompatible type")
				},
			},
			wantText: "set_environment_default failed",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &Client{Environments: tt.stub}
			handler := HandleSetEnvironmentDefault(c)
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

func TestHandleRemoveEnvironmentDefault(t *testing.T) {
	tests := []struct {
		name     string
		input    RemoveEnvironmentDefaultInput
		stub     *stubEnvironments
		wantErr  string
		wantText string
	}{
		{
			name:    "missing id",
			input:   RemoveEnvironmentDefaultInput{},
			stub:    &stubEnvironments{},
			wantErr: "id is required",
		},
		{
			name:  "success returns environment default JSON",
			input: RemoveEnvironmentDefaultInput{ID: "def1"},
			stub: &stubEnvironments{
				removeDefaultFn: func(_ context.Context, id string) (*environments.EnvironmentDefault, error) {
					return &environments.EnvironmentDefault{ID: id}, nil
				},
			},
			wantText: "def1",
		},
		{
			name:  "mutation failure returns error message",
			input: RemoveEnvironmentDefaultInput{ID: "def1"},
			stub: &stubEnvironments{
				removeDefaultFn: func(context.Context, string) (*environments.EnvironmentDefault, error) {
					return nil, mutationFailedErr("remove environment default", "", "not found")
				},
			},
			wantText: "remove_environment_default failed",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &Client{Environments: tt.stub}
			handler := HandleRemoveEnvironmentDefault(c)
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
