package tools

import (
	"context"
	"strings"
	"testing"

	"github.com/massdriver-cloud/massdriver-sdk-go/massdriver/platform/deployments"
	"github.com/massdriver-cloud/massdriver-sdk-go/massdriver/platform/types"
)

type stubDeployments struct {
	listPageFn func(context.Context, deployments.ListInput) (types.Page[deployments.Deployment], error)
	getFn      func(context.Context, string) (*deployments.Deployment, error)
	getLogsFn  func(context.Context, string) (string, error)
	createFn   func(context.Context, string, deployments.CreateInput) (*deployments.Deployment, error)
	proposeFn  func(context.Context, string, deployments.ProposeInput) (*deployments.Deployment, error)
	approveFn  func(context.Context, string) (*deployments.Deployment, error)
	rejectFn   func(context.Context, string) (*deployments.Deployment, error)
	abortFn    func(context.Context, string) (*deployments.Deployment, error)
}

func (s *stubDeployments) ListPage(ctx context.Context, input deployments.ListInput) (types.Page[deployments.Deployment], error) {
	return s.listPageFn(ctx, input)
}
func (s *stubDeployments) Get(ctx context.Context, id string) (*deployments.Deployment, error) {
	return s.getFn(ctx, id)
}
func (s *stubDeployments) GetLogs(ctx context.Context, id string) (string, error) {
	return s.getLogsFn(ctx, id)
}
func (s *stubDeployments) Create(ctx context.Context, instanceID string, input deployments.CreateInput) (*deployments.Deployment, error) {
	return s.createFn(ctx, instanceID, input)
}
func (s *stubDeployments) Propose(ctx context.Context, instanceID string, input deployments.ProposeInput) (*deployments.Deployment, error) {
	return s.proposeFn(ctx, instanceID, input)
}
func (s *stubDeployments) Approve(ctx context.Context, id string) (*deployments.Deployment, error) {
	return s.approveFn(ctx, id)
}
func (s *stubDeployments) Reject(ctx context.Context, id string) (*deployments.Deployment, error) {
	return s.rejectFn(ctx, id)
}
func (s *stubDeployments) Abort(ctx context.Context, id string) (*deployments.Deployment, error) {
	return s.abortFn(ctx, id)
}

func TestHandleListDeployments(t *testing.T) {
	tests := []struct {
		name     string
		input    ListDeploymentsInput
		stub     *stubDeployments
		wantErr  bool
		wantText string
	}{
		{
			name:  "returns page of deployments",
			input: ListDeploymentsInput{},
			stub: &stubDeployments{
				listPageFn: func(_ context.Context, _ deployments.ListInput) (types.Page[deployments.Deployment], error) {
					return types.Page[deployments.Deployment]{
						Items: []deployments.Deployment{{ID: "dep1", Status: "COMPLETED"}},
					}, nil
				},
			},
			wantText: "dep1",
		},
		{
			name:  "empty page surfaces has_more false",
			input: ListDeploymentsInput{},
			stub: &stubDeployments{
				listPageFn: func(context.Context, deployments.ListInput) (types.Page[deployments.Deployment], error) {
					return types.Page[deployments.Deployment]{}, nil
				},
			},
			wantText: "\"has_more\": false",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &Client{Deployments: tt.stub}
			handler := HandleListDeployments(c)
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

func TestHandleGetDeployment(t *testing.T) {
	tests := []struct {
		name     string
		input    GetDeploymentInput
		stub     *stubDeployments
		wantErr  string
		wantText string
	}{
		{
			name:    "missing id",
			input:   GetDeploymentInput{},
			stub:    &stubDeployments{},
			wantErr: "id is required",
		},
		{
			name:  "returns deployment JSON",
			input: GetDeploymentInput{ID: "dep1"},
			stub: &stubDeployments{
				getFn: func(_ context.Context, id string) (*deployments.Deployment, error) {
					return &deployments.Deployment{ID: id, Status: "COMPLETED", Action: "PROVISION"}, nil
				},
			},
			wantText: "dep1",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &Client{Deployments: tt.stub}
			handler := HandleGetDeployment(c)
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

func TestHandleGetDeploymentLogs(t *testing.T) {
	tests := []struct {
		name     string
		input    GetDeploymentLogsInput
		stub     *stubDeployments
		wantErr  string
		wantText string
	}{
		{
			name:    "missing id",
			input:   GetDeploymentLogsInput{},
			stub:    &stubDeployments{},
			wantErr: "id is required",
		},
		{
			name:  "returns logs",
			input: GetDeploymentLogsInput{ID: "dep1"},
			stub: &stubDeployments{
				getLogsFn: func(context.Context, string) (string, error) {
					return "Apply complete! Resources: 3 added.", nil
				},
			},
			wantText: "Apply complete",
		},
		{
			name:  "empty logs",
			input: GetDeploymentLogsInput{ID: "dep1"},
			stub: &stubDeployments{
				getLogsFn: func(context.Context, string) (string, error) {
					return "", nil
				},
			},
			wantText: "no logs available",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &Client{Deployments: tt.stub}
			handler := HandleGetDeploymentLogs(c)
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

func TestHandleCreateDeployment(t *testing.T) {
	tests := []struct {
		name     string
		input    CreateDeploymentInput
		stub     *stubDeployments
		wantErr  string
		wantText string
	}{
		{
			name:    "missing instance_id",
			input:   CreateDeploymentInput{Action: "PROVISION"},
			stub:    &stubDeployments{},
			wantErr: "instance_id is required",
		},
		{
			name:    "missing action",
			input:   CreateDeploymentInput{InstanceID: "inst1"},
			stub:    &stubDeployments{},
			wantErr: "action is required",
		},
		{
			name:  "success returns deployment JSON",
			input: CreateDeploymentInput{InstanceID: "inst1", Action: "PROVISION"},
			stub: &stubDeployments{
				createFn: func(_ context.Context, _ string, input deployments.CreateInput) (*deployments.Deployment, error) {
					return &deployments.Deployment{ID: "dep1", Status: "PENDING", Action: string(input.Action)}, nil
				},
			},
			wantText: "dep1",
		},
		{
			name:  "mutation failure returns error message",
			input: CreateDeploymentInput{InstanceID: "inst1", Action: "PROVISION"},
			stub: &stubDeployments{
				createFn: func(context.Context, string, deployments.CreateInput) (*deployments.Deployment, error) {
					return nil, mutationFailedErr("create deployment", "instance", "already deploying")
				},
			},
			wantText: "create_deployment failed",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &Client{Deployments: tt.stub}
			handler := HandleCreateDeployment(c)
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

func TestHandleAbortDeployment(t *testing.T) {
	tests := []struct {
		name     string
		input    AbortDeploymentInput
		stub     *stubDeployments
		wantErr  string
		wantText string
	}{
		{
			name:    "missing id",
			input:   AbortDeploymentInput{},
			stub:    &stubDeployments{},
			wantErr: "id is required",
		},
		{
			name:  "success returns deployment JSON",
			input: AbortDeploymentInput{ID: "dep1"},
			stub: &stubDeployments{
				abortFn: func(_ context.Context, id string) (*deployments.Deployment, error) {
					return &deployments.Deployment{ID: id, Status: "ABORTED"}, nil
				},
			},
			wantText: "ABORTED",
		},
		{
			name:  "mutation failure returns error message",
			input: AbortDeploymentInput{ID: "dep1"},
			stub: &stubDeployments{
				abortFn: func(context.Context, string) (*deployments.Deployment, error) {
					return nil, mutationFailedErr("abort deployment", "", "already completed")
				},
			},
			wantText: "abort_deployment failed",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &Client{Deployments: tt.stub}
			handler := HandleAbortDeployment(c)
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

func TestHandleProposeDeployment(t *testing.T) {
	tests := []struct {
		name     string
		input    ProposeDeploymentInput
		stub     *stubDeployments
		wantErr  string
		wantText string
	}{
		{
			name:    "missing instance_id",
			input:   ProposeDeploymentInput{Action: "PROVISION"},
			stub:    &stubDeployments{},
			wantErr: "instance_id is required",
		},
		{
			name:    "missing action",
			input:   ProposeDeploymentInput{InstanceID: "inst1"},
			stub:    &stubDeployments{},
			wantErr: "action is required",
		},
		{
			name:  "success returns deployment JSON",
			input: ProposeDeploymentInput{InstanceID: "inst1", Action: "PROVISION"},
			stub: &stubDeployments{
				proposeFn: func(_ context.Context, _ string, input deployments.ProposeInput) (*deployments.Deployment, error) {
					return &deployments.Deployment{ID: "dep1", Status: "PROPOSED", Action: string(input.Action)}, nil
				},
			},
			wantText: "dep1",
		},
		{
			name:  "mutation failure returns error message",
			input: ProposeDeploymentInput{InstanceID: "inst1", Action: "PROVISION"},
			stub: &stubDeployments{
				proposeFn: func(context.Context, string, deployments.ProposeInput) (*deployments.Deployment, error) {
					return nil, mutationFailedErr("propose deployment", "instance", "already deploying")
				},
			},
			wantText: "propose_deployment failed",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &Client{Deployments: tt.stub}
			handler := HandleProposeDeployment(c)
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

func TestHandleApproveDeployment(t *testing.T) {
	tests := []struct {
		name     string
		input    ApproveDeploymentInput
		stub     *stubDeployments
		wantErr  string
		wantText string
	}{
		{
			name:    "missing id",
			input:   ApproveDeploymentInput{},
			stub:    &stubDeployments{},
			wantErr: "id is required",
		},
		{
			name:  "success returns deployment JSON",
			input: ApproveDeploymentInput{ID: "dep1"},
			stub: &stubDeployments{
				approveFn: func(_ context.Context, id string) (*deployments.Deployment, error) {
					return &deployments.Deployment{ID: id, Status: "APPROVED"}, nil
				},
			},
			wantText: "APPROVED",
		},
		{
			name:  "mutation failure returns error message",
			input: ApproveDeploymentInput{ID: "dep1"},
			stub: &stubDeployments{
				approveFn: func(context.Context, string) (*deployments.Deployment, error) {
					return nil, mutationFailedErr("approve deployment", "", "not in proposed state")
				},
			},
			wantText: "approve_deployment failed",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &Client{Deployments: tt.stub}
			handler := HandleApproveDeployment(c)
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

func TestHandleRejectDeployment(t *testing.T) {
	tests := []struct {
		name     string
		input    RejectDeploymentInput
		stub     *stubDeployments
		wantErr  string
		wantText string
	}{
		{
			name:    "missing id",
			input:   RejectDeploymentInput{},
			stub:    &stubDeployments{},
			wantErr: "id is required",
		},
		{
			name:  "success returns deployment JSON",
			input: RejectDeploymentInput{ID: "dep1"},
			stub: &stubDeployments{
				rejectFn: func(_ context.Context, id string) (*deployments.Deployment, error) {
					return &deployments.Deployment{ID: id, Status: "REJECTED"}, nil
				},
			},
			wantText: "REJECTED",
		},
		{
			name:  "mutation failure returns error message",
			input: RejectDeploymentInput{ID: "dep1"},
			stub: &stubDeployments{
				rejectFn: func(context.Context, string) (*deployments.Deployment, error) {
					return nil, mutationFailedErr("reject deployment", "", "not in proposed state")
				},
			},
			wantText: "reject_deployment failed",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &Client{Deployments: tt.stub}
			handler := HandleRejectDeployment(c)
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
