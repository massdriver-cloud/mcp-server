package tools

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/massdriver-cloud/massdriver-sdk-go/massdriver/platform/environments"
	"github.com/massdriver-cloud/massdriver-sdk-go/massdriver/platform/types"
)

type stubEnvironments struct {
	listPageFn      func(context.Context, environments.ListInput) (types.Page[environments.Environment], error)
	getFn           func(context.Context, string) (*environments.Environment, error)
	createFn        func(context.Context, string, environments.CreateInput) (*environments.Environment, error)
	updateFn        func(context.Context, string, environments.UpdateInput) (*environments.Environment, error)
	deleteFn        func(context.Context, string) (*environments.Environment, error)
	setDefaultFn    func(context.Context, string, string) (*environments.EnvironmentDefault, error)
	removeDefaultFn func(context.Context, string) (*environments.EnvironmentDefault, error)
	compareFn       func(context.Context, string, string) (*environments.Comparison, error)
	forkFn          func(context.Context, string, environments.ForkInput) (*environments.Environment, error)
	deployFn        func(context.Context, string) (*environments.Environment, error)
	decommissionFn  func(context.Context, string) (*environments.Environment, error)
	linksFn         func(context.Context, string) ([]types.Link, error)
	unfulfilledFn   func(context.Context, string) ([]environments.UnfulfilledDependency, error)
}

func (s *stubEnvironments) ListPage(ctx context.Context, input environments.ListInput) (types.Page[environments.Environment], error) {
	return s.listPageFn(ctx, input)
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
func (s *stubEnvironments) Compare(ctx context.Context, sourceID, targetID string) (*environments.Comparison, error) {
	return s.compareFn(ctx, sourceID, targetID)
}
func (s *stubEnvironments) Fork(ctx context.Context, parentID string, input environments.ForkInput) (*environments.Environment, error) {
	return s.forkFn(ctx, parentID, input)
}
func (s *stubEnvironments) Deploy(ctx context.Context, id string) (*environments.Environment, error) {
	return s.deployFn(ctx, id)
}
func (s *stubEnvironments) Decommission(ctx context.Context, id string) (*environments.Environment, error) {
	return s.decommissionFn(ctx, id)
}
func (s *stubEnvironments) Links(ctx context.Context, id string) ([]types.Link, error) {
	return s.linksFn(ctx, id)
}
func (s *stubEnvironments) UnfulfilledDependencies(ctx context.Context, id string) ([]environments.UnfulfilledDependency, error) {
	return s.unfulfilledFn(ctx, id)
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
			name:  "returns page of environments",
			input: ListEnvironmentsInput{},
			stub: &stubEnvironments{
				listPageFn: func(_ context.Context, _ environments.ListInput) (types.Page[environments.Environment], error) {
					return types.Page[environments.Environment]{
						Items: []environments.Environment{{ID: "myproj-staging", Name: "Staging"}},
					}, nil
				},
			},
			wantText: "myproj-staging",
		},
		{
			name:  "empty page surfaces has_more false",
			input: ListEnvironmentsInput{},
			stub: &stubEnvironments{
				listPageFn: func(context.Context, environments.ListInput) (types.Page[environments.Environment], error) {
					return types.Page[environments.Environment]{}, nil
				},
			},
			wantText: "\"has_more\": false",
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
			name: "protection toggles are passed through",
			input: CreateEnvironmentInput{
				ProjectID: "myproj", ID: "prod", Name: "Production",
				DecommissionProtection: true, SeparationOfDuty: true,
			},
			stub: &stubEnvironments{
				createFn: func(_ context.Context, projectID string, input environments.CreateInput) (*environments.Environment, error) {
					if !input.DecommissionProtection || !input.SeparationOfDuty {
						t.Errorf("protection toggles not passed through: %+v", input)
					}
					return &environments.Environment{ID: projectID + "-" + input.ID, Name: input.Name}, nil
				},
			},
			wantText: "myproj-prod",
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
			input:   UpdateEnvironmentInput{Name: ptr("New Name")},
			stub:    &stubEnvironments{},
			wantErr: "id is required",
		},
		{
			name:  "success returns updated environment JSON",
			input: UpdateEnvironmentInput{ID: "myproj-staging", Name: ptr("Production")},
			stub: &stubEnvironments{
				updateFn: func(_ context.Context, id string, input environments.UpdateInput) (*environments.Environment, error) {
					return &environments.Environment{ID: id, Name: derefStr(input.Name)}, nil
				},
			},
			wantText: "Production",
		},
		{
			name:  "separation_of_duty toggle is passed through",
			input: UpdateEnvironmentInput{ID: "myproj-prod", SeparationOfDuty: ptr(true)},
			stub: &stubEnvironments{
				updateFn: func(_ context.Context, id string, input environments.UpdateInput) (*environments.Environment, error) {
					if input.SeparationOfDuty == nil || !*input.SeparationOfDuty {
						t.Errorf("separation_of_duty not passed through: %+v", input)
					}
					return &environments.Environment{ID: id, SeparationOfDuty: true}, nil
				},
			},
			wantText: "myproj-prod",
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

func TestHandleCompareEnvironments(t *testing.T) {
	tests := []struct {
		name     string
		input    CompareEnvironmentsInput
		stub     *stubEnvironments
		wantErr  string
		wantText string
	}{
		{
			name:    "missing source_id",
			input:   CompareEnvironmentsInput{TargetID: "env2"},
			stub:    &stubEnvironments{},
			wantErr: "source_id is required",
		},
		{
			name:    "missing target_id",
			input:   CompareEnvironmentsInput{SourceID: "env1"},
			stub:    &stubEnvironments{},
			wantErr: "target_id is required",
		},
		{
			name:  "success returns comparison JSON",
			input: CompareEnvironmentsInput{SourceID: "env1", TargetID: "env2"},
			stub: &stubEnvironments{
				compareFn: func(_ context.Context, sourceID, targetID string) (*environments.Comparison, error) {
					return &environments.Comparison{
						Source: environments.Environment{ID: sourceID},
						Target: environments.Environment{ID: targetID},
					}, nil
				},
			},
			wantText: "env2",
		},
		{
			name:  "error is surfaced",
			input: CompareEnvironmentsInput{SourceID: "env1", TargetID: "env2"},
			stub: &stubEnvironments{
				compareFn: func(context.Context, string, string) (*environments.Comparison, error) {
					return nil, errors.New("cross-project comparison forbidden")
				},
			},
			wantErr: "compare_environments",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &Client{Environments: tt.stub}
			handler := HandleCompareEnvironments(c)
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

func TestHandleForkEnvironment(t *testing.T) {
	tests := []struct {
		name     string
		input    ForkEnvironmentInput
		stub     *stubEnvironments
		wantErr  string
		wantText string
	}{
		{
			name:    "missing parent_id",
			input:   ForkEnvironmentInput{ID: "staging", Name: "Staging"},
			stub:    &stubEnvironments{},
			wantErr: "parent_id is required",
		},
		{
			name:    "missing id",
			input:   ForkEnvironmentInput{ParentID: "myproj-prod", Name: "Staging"},
			stub:    &stubEnvironments{},
			wantErr: "id is required",
		},
		{
			name:    "missing name",
			input:   ForkEnvironmentInput{ParentID: "myproj-prod", ID: "staging"},
			stub:    &stubEnvironments{},
			wantErr: "name is required",
		},
		{
			name: "success passes copy toggles and returns environment JSON",
			input: ForkEnvironmentInput{
				ParentID: "myproj-prod", ID: "staging", Name: "Staging",
				CopySecrets: true, CopyEnvironmentDefaults: true,
				DecommissionProtection: true, SeparationOfDuty: true,
			},
			stub: &stubEnvironments{
				forkFn: func(_ context.Context, parentID string, input environments.ForkInput) (*environments.Environment, error) {
					if parentID != "myproj-prod" {
						t.Errorf("expected parent %q, got %q", "myproj-prod", parentID)
					}
					if !input.CopySecrets || !input.CopyEnvironmentDefaults || input.CopyRemoteReferences {
						t.Errorf("copy toggles not passed through: %+v", input)
					}
					if !input.DecommissionProtection || !input.SeparationOfDuty {
						t.Errorf("protection toggles not passed through: %+v", input)
					}
					return &environments.Environment{ID: "myproj-" + input.ID, Name: input.Name}, nil
				},
			},
			wantText: "myproj-staging",
		},
		{
			name:  "mutation failure returns error message",
			input: ForkEnvironmentInput{ParentID: "myproj-prod", ID: "staging", Name: "Staging"},
			stub: &stubEnvironments{
				forkFn: func(context.Context, string, environments.ForkInput) (*environments.Environment, error) {
					return nil, mutationFailedErr("fork environment", "id", "already exists")
				},
			},
			wantText: "fork_environment failed",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &Client{Environments: tt.stub}
			handler := HandleForkEnvironment(c)
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

func TestHandleDeployEnvironment(t *testing.T) {
	tests := []struct {
		name     string
		input    DeployEnvironmentInput
		stub     *stubEnvironments
		wantErr  string
		wantText string
	}{
		{
			name:    "missing id",
			input:   DeployEnvironmentInput{},
			stub:    &stubEnvironments{},
			wantErr: "id is required",
		},
		{
			name:  "success returns environment JSON",
			input: DeployEnvironmentInput{ID: "myproj-staging"},
			stub: &stubEnvironments{
				deployFn: func(_ context.Context, id string) (*environments.Environment, error) {
					return &environments.Environment{ID: id, Name: "Staging"}, nil
				},
			},
			wantText: "myproj-staging",
		},
		{
			name:  "mutation failure returns error message",
			input: DeployEnvironmentInput{ID: "myproj-staging"},
			stub: &stubEnvironments{
				deployFn: func(context.Context, string) (*environments.Environment, error) {
					return nil, mutationFailedErr("deploy environment", "", "no instances to deploy")
				},
			},
			wantText: "deploy_environment failed",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &Client{Environments: tt.stub}
			handler := HandleDeployEnvironment(c)
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

func TestHandleDecommissionEnvironment(t *testing.T) {
	tests := []struct {
		name     string
		input    DecommissionEnvironmentInput
		stub     *stubEnvironments
		wantErr  string
		wantText string
	}{
		{
			name:    "missing id",
			input:   DecommissionEnvironmentInput{},
			stub:    &stubEnvironments{},
			wantErr: "id is required",
		},
		{
			name:  "success returns environment JSON",
			input: DecommissionEnvironmentInput{ID: "myproj-staging"},
			stub: &stubEnvironments{
				decommissionFn: func(_ context.Context, id string) (*environments.Environment, error) {
					return &environments.Environment{ID: id, Name: "Staging"}, nil
				},
			},
			wantText: "myproj-staging",
		},
		{
			name:  "mutation failure returns error message",
			input: DecommissionEnvironmentInput{ID: "myproj-staging"},
			stub: &stubEnvironments{
				decommissionFn: func(context.Context, string) (*environments.Environment, error) {
					return nil, mutationFailedErr("decommission environment", "", "decommission protection enabled")
				},
			},
			wantText: "decommission_environment failed",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &Client{Environments: tt.stub}
			handler := HandleDecommissionEnvironment(c)
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

func TestHandleListEnvironmentLinks(t *testing.T) {
	tests := []struct {
		name     string
		input    ListEnvironmentLinksInput
		stub     *stubEnvironments
		wantErr  string
		wantText string
	}{
		{
			name:    "missing id",
			input:   ListEnvironmentLinksInput{},
			stub:    &stubEnvironments{},
			wantErr: "id is required",
		},
		{
			name:  "returns links list",
			input: ListEnvironmentLinksInput{ID: "myproj-staging"},
			stub: &stubEnvironments{
				linksFn: func(_ context.Context, id string) ([]types.Link, error) {
					if id != "myproj-staging" {
						t.Errorf("expected id %q, got %q", "myproj-staging", id)
					}
					return []types.Link{{ID: "link1", FromField: "network", ToField: "vpc", FromVersionConstraint: "~1"}}, nil
				},
			},
			wantText: "link1",
		},
		{
			name:  "empty list returns items array",
			input: ListEnvironmentLinksInput{ID: "myproj-staging"},
			stub: &stubEnvironments{
				linksFn: func(context.Context, string) ([]types.Link, error) {
					return nil, nil
				},
			},
			wantText: "\"items\": []",
		},
		{
			name:  "error is surfaced",
			input: ListEnvironmentLinksInput{ID: "myproj-staging"},
			stub: &stubEnvironments{
				linksFn: func(context.Context, string) ([]types.Link, error) {
					return nil, errors.New("not found")
				},
			},
			wantErr: "list_environment_links",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &Client{Environments: tt.stub}
			handler := HandleListEnvironmentLinks(c)
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

func TestHandleListEnvironmentUnfulfilledDependencies(t *testing.T) {
	tests := []struct {
		name     string
		input    ListEnvironmentUnfulfilledDependenciesInput
		stub     *stubEnvironments
		wantErr  string
		wantText string
	}{
		{
			name:    "missing id",
			input:   ListEnvironmentUnfulfilledDependenciesInput{},
			stub:    &stubEnvironments{},
			wantErr: "id is required",
		},
		{
			name:  "returns unfulfilled dependencies list",
			input: ListEnvironmentUnfulfilledDependenciesInput{ID: "myproj-staging"},
			stub: &stubEnvironments{
				unfulfilledFn: func(_ context.Context, id string) ([]environments.UnfulfilledDependency, error) {
					if id != "myproj-staging" {
						t.Errorf("expected id %q, got %q", "myproj-staging", id)
					}
					return []environments.UnfulfilledDependency{{
						Instance:     types.Instance{ID: "myproj-staging-api"},
						Field:        "network",
						ResourceType: types.ResourceType{ID: "aws-vpc@1.0.0", Name: "AWS VPC"},
					}}, nil
				},
			},
			wantText: "myproj-staging-api",
		},
		{
			name:  "empty list returns items array",
			input: ListEnvironmentUnfulfilledDependenciesInput{ID: "myproj-staging"},
			stub: &stubEnvironments{
				unfulfilledFn: func(context.Context, string) ([]environments.UnfulfilledDependency, error) {
					return nil, nil
				},
			},
			wantText: "\"items\": []",
		},
		{
			name:  "error is surfaced",
			input: ListEnvironmentUnfulfilledDependenciesInput{ID: "myproj-staging"},
			stub: &stubEnvironments{
				unfulfilledFn: func(context.Context, string) ([]environments.UnfulfilledDependency, error) {
					return nil, errors.New("not found")
				},
			},
			wantErr: "list_environment_unfulfilled_dependencies",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &Client{Environments: tt.stub}
			handler := HandleListEnvironmentUnfulfilledDependencies(c)
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
