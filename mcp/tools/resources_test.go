package tools

import (
	"context"
	"strings"
	"testing"

	"github.com/massdriver-cloud/massdriver-sdk-go/massdriver/platform/resources"
)

type stubResources struct {
	listFn        func(context.Context, resources.ListInput) ([]resources.Resource, error)
	getFn         func(context.Context, string) (*resources.Resource, error)
	createFn      func(context.Context, string, resources.CreateInput) (*resources.Resource, error)
	updateFn      func(context.Context, string, resources.UpdateInput) (*resources.Resource, error)
	deleteFn      func(context.Context, string) (*resources.Resource, error)
	exportFn      func(context.Context, string, string) (*resources.Exported, error)
	createGrantFn func(context.Context, string, resources.CreateGrantInput) (*resources.Grant, error)
	deleteGrantFn func(context.Context, string) error
}

func (s *stubResources) List(ctx context.Context, input resources.ListInput) ([]resources.Resource, error) {
	return s.listFn(ctx, input)
}
func (s *stubResources) Get(ctx context.Context, id string) (*resources.Resource, error) {
	return s.getFn(ctx, id)
}
func (s *stubResources) Create(ctx context.Context, resourceTypeID string, input resources.CreateInput) (*resources.Resource, error) {
	return s.createFn(ctx, resourceTypeID, input)
}
func (s *stubResources) Update(ctx context.Context, id string, input resources.UpdateInput) (*resources.Resource, error) {
	return s.updateFn(ctx, id, input)
}
func (s *stubResources) Delete(ctx context.Context, id string) (*resources.Resource, error) {
	return s.deleteFn(ctx, id)
}
func (s *stubResources) Export(ctx context.Context, id, format string) (*resources.Exported, error) {
	return s.exportFn(ctx, id, format)
}
func (s *stubResources) CreateGrant(ctx context.Context, resourceID string, input resources.CreateGrantInput) (*resources.Grant, error) {
	return s.createGrantFn(ctx, resourceID, input)
}
func (s *stubResources) DeleteGrant(ctx context.Context, grantID string) error {
	return s.deleteGrantFn(ctx, grantID)
}

func TestHandleListResources(t *testing.T) {
	tests := []struct {
		name     string
		input    ListResourcesInput
		stub     *stubResources
		wantErr  bool
		wantText string
	}{
		{
			name:  "returns resources",
			input: ListResourcesInput{},
			stub: &stubResources{
				listFn: func(_ context.Context, _ resources.ListInput) ([]resources.Resource, error) {
					return []resources.Resource{{ID: "res1", Name: "My Database"}}, nil
				},
			},
			wantText: "My Database",
		},
		{
			name:  "returns null for empty list",
			input: ListResourcesInput{},
			stub: &stubResources{
				listFn: func(context.Context, resources.ListInput) ([]resources.Resource, error) {
					return nil, nil
				},
			},
			wantText: "null",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &Client{Resources: tt.stub}
			handler := HandleListResources(c)
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

func TestHandleGetResource(t *testing.T) {
	tests := []struct {
		name     string
		input    GetResourceInput
		stub     *stubResources
		wantErr  string
		wantText string
	}{
		{
			name:    "missing id",
			input:   GetResourceInput{},
			stub:    &stubResources{},
			wantErr: "id is required",
		},
		{
			name:  "returns resource JSON",
			input: GetResourceInput{ID: "res1"},
			stub: &stubResources{
				getFn: func(_ context.Context, id string) (*resources.Resource, error) {
					return &resources.Resource{ID: id, Name: "My Database"}, nil
				},
			},
			wantText: "res1",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &Client{Resources: tt.stub}
			handler := HandleGetResource(c)
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

func TestHandleExportResource(t *testing.T) {
	tests := []struct {
		name     string
		input    ExportResourceInput
		stub     *stubResources
		wantErr  string
		wantText string
	}{
		{
			name:    "missing id",
			input:   ExportResourceInput{},
			stub:    &stubResources{},
			wantErr: "id is required",
		},
		{
			name:  "returns exported resource JSON",
			input: ExportResourceInput{ID: "res1"},
			stub: &stubResources{
				exportFn: func(_ context.Context, id, _ string) (*resources.Exported, error) {
					return &resources.Exported{
						Resource: resources.Resource{ID: id, Name: "My Database"},
						Rendered: `{"host": "db.example.com"}`,
					}, nil
				},
			},
			wantText: "db.example.com",
		},
		{
			name:  "defaults format to json",
			input: ExportResourceInput{ID: "res1"},
			stub: &stubResources{
				exportFn: func(_ context.Context, _, format string) (*resources.Exported, error) {
					if format != resources.FormatJSON {
						t.Errorf("expected format %q, got %q", resources.FormatJSON, format)
					}
					return &resources.Exported{Resource: resources.Resource{ID: "res1"}}, nil
				},
			},
			wantText: "res1",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &Client{Resources: tt.stub}
			handler := HandleExportResource(c)
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

func TestHandleCreateResource(t *testing.T) {
	tests := []struct {
		name     string
		input    CreateResourceInput
		stub     *stubResources
		wantErr  string
		wantText string
	}{
		{
			name:    "missing resource_type_id",
			input:   CreateResourceInput{Name: "My DB"},
			stub:    &stubResources{},
			wantErr: "resource_type_id is required",
		},
		{
			name:    "missing name",
			input:   CreateResourceInput{ResourceTypeID: "rt1"},
			stub:    &stubResources{},
			wantErr: "name is required",
		},
		{
			name:  "success returns resource JSON",
			input: CreateResourceInput{ResourceTypeID: "rt1", Name: "My DB"},
			stub: &stubResources{
				createFn: func(_ context.Context, _ string, input resources.CreateInput) (*resources.Resource, error) {
					return &resources.Resource{ID: "res1", Name: input.Name}, nil
				},
			},
			wantText: "res1",
		},
		{
			name:  "mutation failure returns error message",
			input: CreateResourceInput{ResourceTypeID: "rt1", Name: "My DB"},
			stub: &stubResources{
				createFn: func(context.Context, string, resources.CreateInput) (*resources.Resource, error) {
					return nil, mutationFailedErr("create resource", "payload", "invalid schema")
				},
			},
			wantText: "create_resource failed",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &Client{Resources: tt.stub}
			handler := HandleCreateResource(c)
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

func TestHandleUpdateResource(t *testing.T) {
	tests := []struct {
		name     string
		input    UpdateResourceInput
		stub     *stubResources
		wantErr  string
		wantText string
	}{
		{
			name:    "missing id",
			input:   UpdateResourceInput{Name: "New Name"},
			stub:    &stubResources{},
			wantErr: "id is required",
		},
		{
			name:    "missing name",
			input:   UpdateResourceInput{ID: "res1"},
			stub:    &stubResources{},
			wantErr: "name is required",
		},
		{
			name:  "success returns resource JSON",
			input: UpdateResourceInput{ID: "res1", Name: "Updated DB"},
			stub: &stubResources{
				updateFn: func(_ context.Context, id string, input resources.UpdateInput) (*resources.Resource, error) {
					return &resources.Resource{ID: id, Name: input.Name}, nil
				},
			},
			wantText: "Updated DB",
		},
		{
			name:  "mutation failure returns error message",
			input: UpdateResourceInput{ID: "res1", Name: "Updated DB"},
			stub: &stubResources{
				updateFn: func(context.Context, string, resources.UpdateInput) (*resources.Resource, error) {
					return nil, mutationFailedErr("update resource", "payload", "invalid schema")
				},
			},
			wantText: "update_resource failed",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &Client{Resources: tt.stub}
			handler := HandleUpdateResource(c)
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

func TestHandleDeleteResource(t *testing.T) {
	tests := []struct {
		name     string
		input    DeleteResourceInput
		stub     *stubResources
		wantErr  string
		wantText string
	}{
		{
			name:    "missing id",
			input:   DeleteResourceInput{},
			stub:    &stubResources{},
			wantErr: "id is required",
		},
		{
			name:  "success returns confirmation message",
			input: DeleteResourceInput{ID: "res1"},
			stub: &stubResources{
				deleteFn: func(_ context.Context, id string) (*resources.Resource, error) {
					return &resources.Resource{ID: id}, nil
				},
			},
			wantText: "deleted successfully",
		},
		{
			name:  "mutation failure returns error message",
			input: DeleteResourceInput{ID: "res1"},
			stub: &stubResources{
				deleteFn: func(context.Context, string) (*resources.Resource, error) {
					return nil, mutationFailedErr("delete resource", "", "still in use")
				},
			},
			wantText: "delete_resource failed",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &Client{Resources: tt.stub}
			handler := HandleDeleteResource(c)
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

func TestHandleCreateResourceGrant(t *testing.T) {
	tests := []struct {
		name     string
		input    CreateResourceGrantInput
		stub     *stubResources
		wantErr  string
		wantText string
	}{
		{
			name:    "missing resource_id",
			input:   CreateResourceGrantInput{Action: "resource:export"},
			stub:    &stubResources{},
			wantErr: "resource_id is required",
		},
		{
			name:    "missing action",
			input:   CreateResourceGrantInput{ResourceID: "res1"},
			stub:    &stubResources{},
			wantErr: "action is required",
		},
		{
			name:  "success returns grant JSON",
			input: CreateResourceGrantInput{ResourceID: "res1", Action: "resource:export"},
			stub: &stubResources{
				createGrantFn: func(_ context.Context, _ string, input resources.CreateGrantInput) (*resources.Grant, error) {
					return &resources.Grant{ID: "grant1", Action: input.Action}, nil
				},
			},
			wantText: "grant1",
		},
		{
			name:  "mutation failure returns error message",
			input: CreateResourceGrantInput{ResourceID: "res1", Action: "resource:export"},
			stub: &stubResources{
				createGrantFn: func(context.Context, string, resources.CreateGrantInput) (*resources.Grant, error) {
					return nil, mutationFailedErr("create resource grant", "action", "invalid action")
				},
			},
			wantText: "create_resource_grant failed",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &Client{Resources: tt.stub}
			handler := HandleCreateResourceGrant(c)
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

func TestHandleDeleteResourceGrant(t *testing.T) {
	tests := []struct {
		name     string
		input    DeleteResourceGrantInput
		stub     *stubResources
		wantErr  string
		wantText string
	}{
		{
			name:    "missing grant_id",
			input:   DeleteResourceGrantInput{},
			stub:    &stubResources{},
			wantErr: "grant_id is required",
		},
		{
			name:  "success returns confirmation message",
			input: DeleteResourceGrantInput{GrantID: "grant1"},
			stub: &stubResources{
				deleteGrantFn: func(_ context.Context, _ string) error {
					return nil
				},
			},
			wantText: "deleted successfully",
		},
		{
			name:  "mutation failure returns error message",
			input: DeleteResourceGrantInput{GrantID: "grant1"},
			stub: &stubResources{
				deleteGrantFn: func(context.Context, string) error {
					return mutationFailedErr("delete resource grant", "", "not found")
				},
			},
			wantText: "delete_resource_grant failed",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &Client{Resources: tt.stub}
			handler := HandleDeleteResourceGrant(c)
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
