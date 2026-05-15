package tools

import (
	"context"
	"strings"
	"testing"

	"github.com/massdriver-cloud/massdriver-sdk-go/massdriver/platform/organizations"
)

type stubOrganizations struct {
	getFn                   func(context.Context) (*organizations.Organization, error)
	createCustomAttributeFn func(context.Context, organizations.CreateCustomAttributeInput) (*organizations.CustomAttribute, error)
	updateCustomAttributeFn func(context.Context, string, organizations.UpdateCustomAttributeInput) (*organizations.CustomAttribute, error)
	deleteCustomAttributeFn func(context.Context, string) (*organizations.CustomAttribute, error)
}

func (s *stubOrganizations) Get(ctx context.Context) (*organizations.Organization, error) {
	return s.getFn(ctx)
}
func (s *stubOrganizations) CreateCustomAttribute(ctx context.Context, input organizations.CreateCustomAttributeInput) (*organizations.CustomAttribute, error) {
	return s.createCustomAttributeFn(ctx, input)
}
func (s *stubOrganizations) UpdateCustomAttribute(ctx context.Context, id string, input organizations.UpdateCustomAttributeInput) (*organizations.CustomAttribute, error) {
	return s.updateCustomAttributeFn(ctx, id, input)
}
func (s *stubOrganizations) DeleteCustomAttribute(ctx context.Context, id string) (*organizations.CustomAttribute, error) {
	return s.deleteCustomAttributeFn(ctx, id)
}

func TestHandleGetOrganization(t *testing.T) {
	tests := []struct {
		name     string
		stub     *stubOrganizations
		wantErr  bool
		wantText string
	}{
		{
			name: "returns organization JSON",
			stub: &stubOrganizations{
				getFn: func(context.Context) (*organizations.Organization, error) {
					return &organizations.Organization{ID: "org1", Name: "Acme Corp"}, nil
				},
			},
			wantText: "Acme Corp",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &Client{Organizations: tt.stub}
			handler := HandleGetOrganization(c)
			result, _, err := handler(context.Background(), nil, GetOrganizationInput{})
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

func TestHandleCreateCustomAttribute(t *testing.T) {
	tests := []struct {
		name     string
		input    CreateCustomAttributeInput
		stub     *stubOrganizations
		wantErr  string
		wantText string
	}{
		{
			name:    "missing key",
			input:   CreateCustomAttributeInput{Scope: "PROJECT"},
			stub:    &stubOrganizations{},
			wantErr: "key is required",
		},
		{
			name:    "missing scope",
			input:   CreateCustomAttributeInput{Key: "team"},
			stub:    &stubOrganizations{},
			wantErr: "scope is required",
		},
		{
			name:  "success returns custom attribute JSON",
			input: CreateCustomAttributeInput{Key: "team", Scope: "PROJECT"},
			stub: &stubOrganizations{
				createCustomAttributeFn: func(_ context.Context, input organizations.CreateCustomAttributeInput) (*organizations.CustomAttribute, error) {
					return &organizations.CustomAttribute{ID: "attr1", Key: input.Key, Scope: string(input.Scope)}, nil
				},
			},
			wantText: "attr1",
		},
		{
			name:  "mutation failure returns error message",
			input: CreateCustomAttributeInput{Key: "team", Scope: "PROJECT"},
			stub: &stubOrganizations{
				createCustomAttributeFn: func(context.Context, organizations.CreateCustomAttributeInput) (*organizations.CustomAttribute, error) {
					return nil, mutationFailedErr("create custom attribute", "key", "already exists")
				},
			},
			wantText: "create_custom_attribute failed",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &Client{Organizations: tt.stub}
			handler := HandleCreateCustomAttribute(c)
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

func TestHandleUpdateCustomAttribute(t *testing.T) {
	tests := []struct {
		name     string
		input    UpdateCustomAttributeInput
		stub     *stubOrganizations
		wantErr  string
		wantText string
	}{
		{
			name:    "missing id",
			input:   UpdateCustomAttributeInput{},
			stub:    &stubOrganizations{},
			wantErr: "id is required",
		},
		{
			name:  "success returns custom attribute JSON",
			input: UpdateCustomAttributeInput{ID: "attr1"},
			stub: &stubOrganizations{
				updateCustomAttributeFn: func(_ context.Context, id string, _ organizations.UpdateCustomAttributeInput) (*organizations.CustomAttribute, error) {
					return &organizations.CustomAttribute{ID: id, Key: "team"}, nil
				},
			},
			wantText: "attr1",
		},
		{
			name:  "mutation failure returns error message",
			input: UpdateCustomAttributeInput{ID: "attr1"},
			stub: &stubOrganizations{
				updateCustomAttributeFn: func(context.Context, string, organizations.UpdateCustomAttributeInput) (*organizations.CustomAttribute, error) {
					return nil, mutationFailedErr("update custom attribute", "values", "invalid")
				},
			},
			wantText: "update_custom_attribute failed",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &Client{Organizations: tt.stub}
			handler := HandleUpdateCustomAttribute(c)
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

func TestHandleDeleteCustomAttribute(t *testing.T) {
	tests := []struct {
		name     string
		input    DeleteCustomAttributeInput
		stub     *stubOrganizations
		wantErr  string
		wantText string
	}{
		{
			name:    "missing id",
			input:   DeleteCustomAttributeInput{},
			stub:    &stubOrganizations{},
			wantErr: "id is required",
		},
		{
			name:  "success returns confirmation message",
			input: DeleteCustomAttributeInput{ID: "attr1"},
			stub: &stubOrganizations{
				deleteCustomAttributeFn: func(_ context.Context, id string) (*organizations.CustomAttribute, error) {
					return &organizations.CustomAttribute{ID: id}, nil
				},
			},
			wantText: "deleted successfully",
		},
		{
			name:  "mutation failure returns error message",
			input: DeleteCustomAttributeInput{ID: "attr1"},
			stub: &stubOrganizations{
				deleteCustomAttributeFn: func(context.Context, string) (*organizations.CustomAttribute, error) {
					return nil, mutationFailedErr("delete custom attribute", "", "in use")
				},
			},
			wantText: "delete_custom_attribute failed",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &Client{Organizations: tt.stub}
			handler := HandleDeleteCustomAttribute(c)
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
