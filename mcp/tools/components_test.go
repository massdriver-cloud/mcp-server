package tools

import (
	"context"
	"strings"
	"testing"

	"github.com/massdriver-cloud/massdriver-sdk-go/massdriver/platform/components"
)

type stubComponents struct {
	listFn       func(context.Context, components.ListInput) ([]components.Component, error)
	getFn        func(context.Context, string) (*components.Component, error)
	addFn        func(context.Context, string, components.AddInput) (*components.Component, error)
	updateFn     func(context.Context, string, components.UpdateInput) (*components.Component, error)
	removeFn     func(context.Context, string) (*components.Component, error)
	addLinkFn    func(context.Context, components.AddLinkInput) (*components.Link, error)
	removeLinkFn func(context.Context, string) (*components.Link, error)
}

func (s *stubComponents) List(ctx context.Context, input components.ListInput) ([]components.Component, error) {
	return s.listFn(ctx, input)
}
func (s *stubComponents) Get(ctx context.Context, id string) (*components.Component, error) {
	return s.getFn(ctx, id)
}
func (s *stubComponents) Add(ctx context.Context, projectID string, input components.AddInput) (*components.Component, error) {
	return s.addFn(ctx, projectID, input)
}
func (s *stubComponents) Update(ctx context.Context, id string, input components.UpdateInput) (*components.Component, error) {
	return s.updateFn(ctx, id, input)
}
func (s *stubComponents) Remove(ctx context.Context, id string) (*components.Component, error) {
	return s.removeFn(ctx, id)
}
func (s *stubComponents) AddLink(ctx context.Context, input components.AddLinkInput) (*components.Link, error) {
	return s.addLinkFn(ctx, input)
}
func (s *stubComponents) RemoveLink(ctx context.Context, linkID string) (*components.Link, error) {
	return s.removeLinkFn(ctx, linkID)
}

func TestHandleListComponents(t *testing.T) {
	tests := []struct {
		name     string
		input    ListComponentsInput
		stub     *stubComponents
		wantErr  string
		wantText string
	}{
		{
			name:    "missing project_id",
			input:   ListComponentsInput{},
			stub:    &stubComponents{},
			wantErr: "project_id is required",
		},
		{
			name:  "returns components",
			input: ListComponentsInput{ProjectID: "proj1"},
			stub: &stubComponents{
				listFn: func(_ context.Context, _ components.ListInput) ([]components.Component, error) {
					return []components.Component{{ID: "db", Name: "Database"}}, nil
				},
			},
			wantText: "Database",
		},
		{
			name:  "returns null for empty list",
			input: ListComponentsInput{ProjectID: "proj1"},
			stub: &stubComponents{
				listFn: func(context.Context, components.ListInput) ([]components.Component, error) {
					return nil, nil
				},
			},
			wantText: "null",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &Client{Components: tt.stub}
			handler := HandleListComponents(c)
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

func TestHandleGetComponent(t *testing.T) {
	tests := []struct {
		name     string
		input    GetComponentInput
		stub     *stubComponents
		wantErr  string
		wantText string
	}{
		{
			name:    "missing id",
			input:   GetComponentInput{},
			stub:    &stubComponents{},
			wantErr: "id is required",
		},
		{
			name:  "returns component JSON",
			input: GetComponentInput{ID: "db"},
			stub: &stubComponents{
				getFn: func(_ context.Context, id string) (*components.Component, error) {
					return &components.Component{ID: id, Name: "Database"}, nil
				},
			},
			wantText: "Database",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &Client{Components: tt.stub}
			handler := HandleGetComponent(c)
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

func TestHandleAddComponent(t *testing.T) {
	tests := []struct {
		name     string
		input    AddComponentInput
		stub     *stubComponents
		wantErr  string
		wantText string
	}{
		{
			name:    "missing project_id",
			input:   AddComponentInput{BundleName: "aws-rds", ID: "db", Name: "DB"},
			stub:    &stubComponents{},
			wantErr: "project_id is required",
		},
		{
			name:    "missing bundle_name",
			input:   AddComponentInput{ProjectID: "proj1", ID: "db", Name: "DB"},
			stub:    &stubComponents{},
			wantErr: "bundle_name is required",
		},
		{
			name:    "missing id",
			input:   AddComponentInput{ProjectID: "proj1", BundleName: "aws-rds", Name: "DB"},
			stub:    &stubComponents{},
			wantErr: "id is required",
		},
		{
			name:    "missing name",
			input:   AddComponentInput{ProjectID: "proj1", BundleName: "aws-rds", ID: "db"},
			stub:    &stubComponents{},
			wantErr: "name is required",
		},
		{
			name:  "success returns component JSON",
			input: AddComponentInput{ProjectID: "proj1", BundleName: "aws-rds-postgres", ID: "db", Name: "Database"},
			stub: &stubComponents{
				addFn: func(_ context.Context, _ string, input components.AddInput) (*components.Component, error) {
					return &components.Component{ID: input.ID, Name: input.Name}, nil
				},
			},
			wantText: "Database",
		},
		{
			name:  "mutation failure returns error message",
			input: AddComponentInput{ProjectID: "proj1", BundleName: "aws-rds-postgres", ID: "db", Name: "Database"},
			stub: &stubComponents{
				addFn: func(context.Context, string, components.AddInput) (*components.Component, error) {
					return nil, mutationFailedErr("add component", "id", "already exists")
				},
			},
			wantText: "add_component failed",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &Client{Components: tt.stub}
			handler := HandleAddComponent(c)
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

func TestHandleUpdateComponent(t *testing.T) {
	tests := []struct {
		name     string
		input    UpdateComponentInput
		stub     *stubComponents
		wantErr  string
		wantText string
	}{
		{
			name:    "missing id",
			input:   UpdateComponentInput{Name: "New Name"},
			stub:    &stubComponents{},
			wantErr: "id is required",
		},
		{
			name:  "success returns updated component JSON",
			input: UpdateComponentInput{ID: "db", Name: "Updated DB"},
			stub: &stubComponents{
				updateFn: func(_ context.Context, id string, input components.UpdateInput) (*components.Component, error) {
					return &components.Component{ID: id, Name: input.Name}, nil
				},
			},
			wantText: "Updated DB",
		},
		{
			name:  "mutation failure returns error message",
			input: UpdateComponentInput{ID: "db"},
			stub: &stubComponents{
				updateFn: func(context.Context, string, components.UpdateInput) (*components.Component, error) {
					return nil, mutationFailedErr("update component", "name", "too long")
				},
			},
			wantText: "update_component failed",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &Client{Components: tt.stub}
			handler := HandleUpdateComponent(c)
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

func TestHandleRemoveComponent(t *testing.T) {
	tests := []struct {
		name     string
		input    RemoveComponentInput
		stub     *stubComponents
		wantErr  string
		wantText string
	}{
		{
			name:    "missing id",
			input:   RemoveComponentInput{},
			stub:    &stubComponents{},
			wantErr: "id is required",
		},
		{
			name:  "success returns confirmation",
			input: RemoveComponentInput{ID: "db"},
			stub: &stubComponents{
				removeFn: func(_ context.Context, id string) (*components.Component, error) {
					return &components.Component{ID: id}, nil
				},
			},
			wantText: "removed successfully",
		},
		{
			name:  "mutation failure returns error message",
			input: RemoveComponentInput{ID: "db"},
			stub: &stubComponents{
				removeFn: func(context.Context, string) (*components.Component, error) {
					return nil, mutationFailedErr("remove component", "", "has active instances")
				},
			},
			wantText: "remove_component failed",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &Client{Components: tt.stub}
			handler := HandleRemoveComponent(c)
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

func TestHandleLinkComponents(t *testing.T) {
	validInput := LinkComponentsInput{
		From: "db", FromField: "database", FromVersion: "~1.0",
		To: "app", ToField: "database", ToVersion: "~1.0",
	}
	tests := []struct {
		name     string
		input    LinkComponentsInput
		stub     *stubComponents
		wantErr  string
		wantText string
	}{
		{
			name:    "missing from",
			input:   LinkComponentsInput{FromField: "f", FromVersion: "~1", To: "app", ToField: "f", ToVersion: "~1"},
			stub:    &stubComponents{},
			wantErr: "from is required",
		},
		{
			name:  "success returns link JSON",
			input: validInput,
			stub: &stubComponents{
				addLinkFn: func(_ context.Context, _ components.AddLinkInput) (*components.Link, error) {
					return &components.Link{ID: "link1"}, nil
				},
			},
			wantText: "link1",
		},
		{
			name:  "mutation failure returns error message",
			input: validInput,
			stub: &stubComponents{
				addLinkFn: func(context.Context, components.AddLinkInput) (*components.Link, error) {
					return nil, mutationFailedErr("link components", "from", "already linked")
				},
			},
			wantText: "link_components failed",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &Client{Components: tt.stub}
			handler := HandleLinkComponents(c)
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

func TestHandleUnlinkComponents(t *testing.T) {
	tests := []struct {
		name     string
		input    UnlinkComponentsInput
		stub     *stubComponents
		wantErr  string
		wantText string
	}{
		{
			name:    "missing id",
			input:   UnlinkComponentsInput{},
			stub:    &stubComponents{},
			wantErr: "id is required",
		},
		{
			name:  "success returns confirmation",
			input: UnlinkComponentsInput{ID: "link1"},
			stub: &stubComponents{
				removeLinkFn: func(_ context.Context, id string) (*components.Link, error) {
					return &components.Link{ID: id}, nil
				},
			},
			wantText: "removed successfully",
		},
		{
			name:  "mutation failure returns error message",
			input: UnlinkComponentsInput{ID: "link1"},
			stub: &stubComponents{
				removeLinkFn: func(context.Context, string) (*components.Link, error) {
					return nil, mutationFailedErr("unlink components", "", "not found")
				},
			},
			wantText: "unlink_components failed",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &Client{Components: tt.stub}
			handler := HandleUnlinkComponents(c)
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
