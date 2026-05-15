package tools

import (
	"context"
	"encoding/json"
	"strings"
	"testing"

	"github.com/massdriver-cloud/massdriver-sdk-go/massdriver/platform/organizations"
	"github.com/massdriver-cloud/massdriver-sdk-go/massdriver/platform/policies"
)

type stubPolicies struct {
	getFn                   func(context.Context, string) (*policies.Policy, error)
	createFn                func(context.Context, string, policies.CreatePolicyInput) (*policies.Policy, error)
	updateFn                func(context.Context, string, policies.UpdatePolicyInput) (*policies.Policy, error)
	deleteFn                func(context.Context, string) (*policies.Policy, error)
	listActionsFn           func(context.Context) ([]policies.Action, error)
	listEntitiesFn          func(context.Context) ([]policies.Entity, error)
	evaluateFn              func(context.Context, string, string) (*policies.Decision, error)
	evaluateBatchFn         func(context.Context, []policies.Check) ([]policies.Decision, error)
	explainFn               func(context.Context, policies.ExplainInput) ([]string, error)
	customAttributeSchemaFn func(context.Context, string) (json.RawMessage, error)
	customAttributeValuesFn func(context.Context, organizations.AttributeScope, string) ([]string, error)
}

func (s *stubPolicies) Get(ctx context.Context, id string) (*policies.Policy, error) {
	return s.getFn(ctx, id)
}
func (s *stubPolicies) Create(ctx context.Context, groupID string, input policies.CreatePolicyInput) (*policies.Policy, error) {
	return s.createFn(ctx, groupID, input)
}
func (s *stubPolicies) Update(ctx context.Context, id string, input policies.UpdatePolicyInput) (*policies.Policy, error) {
	return s.updateFn(ctx, id, input)
}
func (s *stubPolicies) Delete(ctx context.Context, id string) (*policies.Policy, error) {
	return s.deleteFn(ctx, id)
}
func (s *stubPolicies) ListActions(ctx context.Context) ([]policies.Action, error) {
	return s.listActionsFn(ctx)
}
func (s *stubPolicies) ListEntities(ctx context.Context) ([]policies.Entity, error) {
	return s.listEntitiesFn(ctx)
}
func (s *stubPolicies) Evaluate(ctx context.Context, action, entityID string) (*policies.Decision, error) {
	return s.evaluateFn(ctx, action, entityID)
}
func (s *stubPolicies) EvaluateBatch(ctx context.Context, checks []policies.Check) ([]policies.Decision, error) {
	return s.evaluateBatchFn(ctx, checks)
}
func (s *stubPolicies) Explain(ctx context.Context, input policies.ExplainInput) ([]string, error) {
	return s.explainFn(ctx, input)
}
func (s *stubPolicies) CustomAttributeSchema(ctx context.Context, action string) (json.RawMessage, error) {
	return s.customAttributeSchemaFn(ctx, action)
}
func (s *stubPolicies) CustomAttributeValues(ctx context.Context, scope organizations.AttributeScope, key string) ([]string, error) {
	return s.customAttributeValuesFn(ctx, scope, key)
}

func TestHandleGetPolicy(t *testing.T) {
	tests := []struct {
		name     string
		input    GetPolicyInput
		stub     *stubPolicies
		wantErr  string
		wantText string
	}{
		{
			name:    "missing id",
			input:   GetPolicyInput{},
			stub:    &stubPolicies{},
			wantErr: "id is required",
		},
		{
			name:  "success",
			input: GetPolicyInput{ID: "pol1"},
			stub: &stubPolicies{
				getFn: func(_ context.Context, id string) (*policies.Policy, error) {
					return &policies.Policy{ID: id, Effect: "ALLOW"}, nil
				},
			},
			wantText: "pol1",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &Client{Policies: tt.stub}
			handler := HandleGetPolicy(c)
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

func TestHandleCreatePolicy(t *testing.T) {
	tests := []struct {
		name     string
		input    CreatePolicyInput
		stub     *stubPolicies
		wantErr  string
		wantText string
	}{
		{
			name:    "missing group_id",
			input:   CreatePolicyInput{Effect: "ALLOW"},
			stub:    &stubPolicies{},
			wantErr: "group_id is required",
		},
		{
			name:    "missing effect",
			input:   CreatePolicyInput{GroupID: "grp1"},
			stub:    &stubPolicies{},
			wantErr: "effect is required",
		},
		{
			name:  "success",
			input: CreatePolicyInput{GroupID: "grp1", Effect: "ALLOW", Actions: []string{"project:view"}},
			stub: &stubPolicies{
				createFn: func(_ context.Context, _ string, input policies.CreatePolicyInput) (*policies.Policy, error) {
					return &policies.Policy{ID: "pol1", Effect: string(input.Effect)}, nil
				},
			},
			wantText: "pol1",
		},
		{
			name:  "mutation failure",
			input: CreatePolicyInput{GroupID: "grp1", Effect: "ALLOW"},
			stub: &stubPolicies{
				createFn: func(context.Context, string, policies.CreatePolicyInput) (*policies.Policy, error) {
					return nil, mutationFailedErr("create policy", "effect", "invalid")
				},
			},
			wantText: "create_policy failed",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &Client{Policies: tt.stub}
			handler := HandleCreatePolicy(c)
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

func TestHandleUpdatePolicy(t *testing.T) {
	tests := []struct {
		name     string
		input    UpdatePolicyInput
		stub     *stubPolicies
		wantErr  string
		wantText string
	}{
		{
			name:    "missing id",
			input:   UpdatePolicyInput{Effect: "DENY"},
			stub:    &stubPolicies{},
			wantErr: "id is required",
		},
		{
			name:  "success",
			input: UpdatePolicyInput{ID: "pol1", Effect: "DENY"},
			stub: &stubPolicies{
				updateFn: func(_ context.Context, id string, _ policies.UpdatePolicyInput) (*policies.Policy, error) {
					return &policies.Policy{ID: id, Effect: "DENY"}, nil
				},
			},
			wantText: "pol1",
		},
		{
			name:  "mutation failure",
			input: UpdatePolicyInput{ID: "pol1"},
			stub: &stubPolicies{
				updateFn: func(context.Context, string, policies.UpdatePolicyInput) (*policies.Policy, error) {
					return nil, mutationFailedErr("update policy", "actions", "empty")
				},
			},
			wantText: "update_policy failed",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &Client{Policies: tt.stub}
			handler := HandleUpdatePolicy(c)
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

func TestHandleDeletePolicy(t *testing.T) {
	tests := []struct {
		name     string
		input    DeletePolicyInput
		stub     *stubPolicies
		wantErr  string
		wantText string
	}{
		{
			name:    "missing id",
			input:   DeletePolicyInput{},
			stub:    &stubPolicies{},
			wantErr: "id is required",
		},
		{
			name:  "success",
			input: DeletePolicyInput{ID: "pol1"},
			stub: &stubPolicies{
				deleteFn: func(_ context.Context, id string) (*policies.Policy, error) {
					return &policies.Policy{ID: id}, nil
				},
			},
			wantText: "deleted successfully",
		},
		{
			name:  "mutation failure",
			input: DeletePolicyInput{ID: "pol1"},
			stub: &stubPolicies{
				deleteFn: func(context.Context, string) (*policies.Policy, error) {
					return nil, mutationFailedErr("delete policy", "", "in use")
				},
			},
			wantText: "delete_policy failed",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &Client{Policies: tt.stub}
			handler := HandleDeletePolicy(c)
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

func TestHandleListPolicyActions(t *testing.T) {
	tests := []struct {
		name     string
		stub     *stubPolicies
		wantText string
	}{
		{
			name: "success",
			stub: &stubPolicies{
				listActionsFn: func(context.Context) ([]policies.Action, error) {
					return []policies.Action{{ID: "project:view", Verb: "view"}}, nil
				},
			},
			wantText: "project:view",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &Client{Policies: tt.stub}
			handler := HandleListPolicyActions(c)
			result, _, err := handler(context.Background(), nil, ListPolicyActionsInput{})
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if !strings.Contains(resultText(t, result), tt.wantText) {
				t.Errorf("expected %q in result, got: %s", tt.wantText, resultText(t, result))
			}
		})
	}
}

func TestHandleListPolicyEntities(t *testing.T) {
	tests := []struct {
		name     string
		stub     *stubPolicies
		wantText string
	}{
		{
			name: "success",
			stub: &stubPolicies{
				listEntitiesFn: func(context.Context) ([]policies.Entity, error) {
					return []policies.Entity{{ID: "project", Description: "A project"}}, nil
				},
			},
			wantText: "project",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &Client{Policies: tt.stub}
			handler := HandleListPolicyEntities(c)
			result, _, err := handler(context.Background(), nil, ListPolicyEntitiesInput{})
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if !strings.Contains(resultText(t, result), tt.wantText) {
				t.Errorf("expected %q in result, got: %s", tt.wantText, resultText(t, result))
			}
		})
	}
}

func TestHandleEvaluatePolicy(t *testing.T) {
	tests := []struct {
		name     string
		input    EvaluatePolicyInput
		stub     *stubPolicies
		wantErr  string
		wantText string
	}{
		{
			name:    "missing action",
			input:   EvaluatePolicyInput{EntityID: "proj1"},
			stub:    &stubPolicies{},
			wantErr: "action is required",
		},
		{
			name:    "missing entity_id",
			input:   EvaluatePolicyInput{Action: "project:view"},
			stub:    &stubPolicies{},
			wantErr: "entity_id is required",
		},
		{
			name:  "success",
			input: EvaluatePolicyInput{Action: "project:view", EntityID: "proj1"},
			stub: &stubPolicies{
				evaluateFn: func(_ context.Context, action, entityID string) (*policies.Decision, error) {
					return &policies.Decision{Allowed: true, Action: action, EntityID: entityID}, nil
				},
			},
			wantText: "true",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &Client{Policies: tt.stub}
			handler := HandleEvaluatePolicy(c)
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

func TestHandleEvaluatePoliciesBatch(t *testing.T) {
	tests := []struct {
		name     string
		input    EvaluatePoliciesBatchInput
		stub     *stubPolicies
		wantErr  string
		wantText string
	}{
		{
			name:    "empty checks",
			input:   EvaluatePoliciesBatchInput{},
			stub:    &stubPolicies{},
			wantErr: "checks is required",
		},
		{
			name: "success",
			input: EvaluatePoliciesBatchInput{
				Checks: []PolicyCheck{{Action: "project:view", EntityID: "proj1"}},
			},
			stub: &stubPolicies{
				evaluateBatchFn: func(_ context.Context, checks []policies.Check) ([]policies.Decision, error) {
					return []policies.Decision{{Allowed: true, Action: checks[0].Action, EntityID: checks[0].EntityID}}, nil
				},
			},
			wantText: "proj1",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &Client{Policies: tt.stub}
			handler := HandleEvaluatePoliciesBatch(c)
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

func TestHandleExplainPolicy(t *testing.T) {
	tests := []struct {
		name     string
		input    ExplainPolicyInput
		stub     *stubPolicies
		wantErr  string
		wantText string
	}{
		{
			name:    "missing effect",
			input:   ExplainPolicyInput{},
			stub:    &stubPolicies{},
			wantErr: "effect is required",
		},
		{
			name:  "success",
			input: ExplainPolicyInput{Effect: "ALLOW", Actions: []string{"project:view"}},
			stub: &stubPolicies{
				explainFn: func(_ context.Context, _ policies.ExplainInput) ([]string, error) {
					return []string{"Allows viewing projects"}, nil
				},
			},
			wantText: "Allows viewing projects",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &Client{Policies: tt.stub}
			handler := HandleExplainPolicy(c)
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

func TestHandleGetPolicyAttributeSchema(t *testing.T) {
	tests := []struct {
		name     string
		input    GetPolicyAttributeSchemaInput
		stub     *stubPolicies
		wantErr  string
		wantText string
	}{
		{
			name:    "missing action",
			input:   GetPolicyAttributeSchemaInput{},
			stub:    &stubPolicies{},
			wantErr: "action is required",
		},
		{
			name:  "success",
			input: GetPolicyAttributeSchemaInput{Action: "project:view"},
			stub: &stubPolicies{
				customAttributeSchemaFn: func(_ context.Context, _ string) (json.RawMessage, error) {
					return json.RawMessage(`{"type":"object"}`), nil
				},
			},
			wantText: `{"type":"object"}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &Client{Policies: tt.stub}
			handler := HandleGetPolicyAttributeSchema(c)
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

func TestHandleListPolicyAttributeValues(t *testing.T) {
	tests := []struct {
		name     string
		input    ListPolicyAttributeValuesInput
		stub     *stubPolicies
		wantErr  string
		wantText string
	}{
		{
			name:    "missing scope",
			input:   ListPolicyAttributeValuesInput{Key: "team"},
			stub:    &stubPolicies{},
			wantErr: "scope is required",
		},
		{
			name:    "missing key",
			input:   ListPolicyAttributeValuesInput{Scope: "PROJECT"},
			stub:    &stubPolicies{},
			wantErr: "key is required",
		},
		{
			name:  "success",
			input: ListPolicyAttributeValuesInput{Scope: "PROJECT", Key: "team"},
			stub: &stubPolicies{
				customAttributeValuesFn: func(_ context.Context, _ organizations.AttributeScope, _ string) ([]string, error) {
					return []string{"backend", "frontend"}, nil
				},
			},
			wantText: "backend",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &Client{Policies: tt.stub}
			handler := HandleListPolicyAttributeValues(c)
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
