package tools

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/massdriver-cloud/massdriver-sdk-go/massdriver/platform/organizations"
	"github.com/massdriver-cloud/massdriver-sdk-go/massdriver/platform/types"
)

type stubOrganizations struct {
	getFn                   func(context.Context) (*organizations.Organization, error)
	createCustomAttributeFn func(context.Context, organizations.CreateCustomAttributeInput) (*organizations.CustomAttribute, error)
	updateCustomAttributeFn func(context.Context, string, organizations.UpdateCustomAttributeInput) (*organizations.CustomAttribute, error)
	deleteCustomAttributeFn func(context.Context, string) (*organizations.CustomAttribute, error)
	listMembersPageFn       func(context.Context, organizations.ListMembersInput) (types.Page[organizations.Account], error)
	listCustomAttributesFn  func(context.Context, organizations.ListCustomAttributesInput) (types.Page[organizations.CustomAttribute], error)
	getSettingsFn           func(context.Context) (*organizations.Settings, error)
	updateSettingsFn        func(context.Context, organizations.UpdateSettingsInput) (*organizations.Settings, error)
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
func (s *stubOrganizations) ListMembersPage(ctx context.Context, input organizations.ListMembersInput) (types.Page[organizations.Account], error) {
	return s.listMembersPageFn(ctx, input)
}
func (s *stubOrganizations) ListCustomAttributesPage(ctx context.Context, input organizations.ListCustomAttributesInput) (types.Page[organizations.CustomAttribute], error) {
	return s.listCustomAttributesFn(ctx, input)
}
func (s *stubOrganizations) DeleteCustomAttribute(ctx context.Context, id string) (*organizations.CustomAttribute, error) {
	return s.deleteCustomAttributeFn(ctx, id)
}
func (s *stubOrganizations) GetSettings(ctx context.Context) (*organizations.Settings, error) {
	return s.getSettingsFn(ctx)
}
func (s *stubOrganizations) UpdateSettings(ctx context.Context, input organizations.UpdateSettingsInput) (*organizations.Settings, error) {
	return s.updateSettingsFn(ctx, input)
}

func TestHandleGetOrganizationSettings(t *testing.T) {
	tests := []struct {
		name     string
		stub     *stubOrganizations
		wantErr  string
		wantText string
	}{
		{
			name: "returns settings JSON",
			stub: &stubOrganizations{
				getSettingsFn: func(context.Context) (*organizations.Settings, error) {
					return &organizations.Settings{DefaultBundleAccess: organizations.DefaultBundleAccessAllProjects}, nil
				},
			},
			wantText: "\"default_bundle_access\": \"ALL_PROJECTS\"",
		},
		{
			name: "error is surfaced",
			stub: &stubOrganizations{
				getSettingsFn: func(context.Context) (*organizations.Settings, error) {
					return nil, errors.New("forbidden")
				},
			},
			wantErr: "get_organization_settings",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &Client{Organizations: tt.stub}
			handler := HandleGetOrganizationSettings(c)
			result, _, err := handler(context.Background(), nil, GetOrganizationSettingsInput{})
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

func TestHandleUpdateOrganizationSettings(t *testing.T) {
	tests := []struct {
		name     string
		input    UpdateOrganizationSettingsInput
		stub     *stubOrganizations
		wantErr  string
		wantText string
	}{
		{
			name:    "missing default_bundle_access",
			input:   UpdateOrganizationSettingsInput{},
			stub:    &stubOrganizations{},
			wantErr: "default_bundle_access is required",
		},
		{
			name:  "success returns updated settings JSON",
			input: UpdateOrganizationSettingsInput{DefaultBundleAccess: "ALL_PROJECTS"},
			stub: &stubOrganizations{
				updateSettingsFn: func(_ context.Context, input organizations.UpdateSettingsInput) (*organizations.Settings, error) {
					if input.DefaultBundleAccess != organizations.DefaultBundleAccessAllProjects {
						t.Errorf("expected ALL_PROJECTS, got %q", input.DefaultBundleAccess)
					}
					return &organizations.Settings{DefaultBundleAccess: input.DefaultBundleAccess}, nil
				},
			},
			wantText: "ALL_PROJECTS",
		},
		{
			name:  "mutation failure returns error message",
			input: UpdateOrganizationSettingsInput{DefaultBundleAccess: "NONE"},
			stub: &stubOrganizations{
				updateSettingsFn: func(context.Context, organizations.UpdateSettingsInput) (*organizations.Settings, error) {
					return nil, mutationFailedErr("update organization settings", "", "forbidden")
				},
			},
			wantText: "update_organization_settings failed",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &Client{Organizations: tt.stub}
			handler := HandleUpdateOrganizationSettings(c)
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

// TestHandleCreateCustomAttributeRequiredDefault verifies the handler sends an
// explicit required=false when the caller omits it (the API defaults an omitted
// value to true, which would silently make the attribute mandatory org-wide),
// and passes an explicit value through unchanged.
func TestHandleCreateCustomAttributeRequiredDefault(t *testing.T) {
	tests := []struct {
		name         string
		input        CreateCustomAttributeInput
		wantRequired bool
	}{
		{name: "omitted defaults to false", input: CreateCustomAttributeInput{Key: "team", Scope: "PROJECT"}, wantRequired: false},
		{name: "explicit true preserved", input: CreateCustomAttributeInput{Key: "team", Scope: "PROJECT", Required: ptr(true)}, wantRequired: true},
		{name: "explicit false preserved", input: CreateCustomAttributeInput{Key: "team", Scope: "PROJECT", Required: ptr(false)}, wantRequired: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var got *bool
			c := &Client{Organizations: &stubOrganizations{
				createCustomAttributeFn: func(_ context.Context, input organizations.CreateCustomAttributeInput) (*organizations.CustomAttribute, error) {
					got = input.Required
					return &organizations.CustomAttribute{ID: "attr1"}, nil
				},
			}}
			if _, _, err := HandleCreateCustomAttribute(c)(context.Background(), nil, tt.input); err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got == nil {
				t.Fatal("expected an explicit required value to be sent, got nil")
			}
			if *got != tt.wantRequired {
				t.Errorf("expected required=%v, got %v", tt.wantRequired, *got)
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

func TestHandleListOrganizationMembers(t *testing.T) {
	c := &Client{Organizations: &stubOrganizations{
		listMembersPageFn: func(_ context.Context, _ organizations.ListMembersInput) (types.Page[organizations.Account], error) {
			return types.Page[organizations.Account]{Items: []organizations.Account{{ID: "u1", Email: "member@example.com"}}}, nil
		},
	}}
	result, _, err := HandleListOrganizationMembers(c)(context.Background(), nil, ListOrganizationMembersInput{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(resultText(t, result), "member@example.com") {
		t.Errorf("expected member in result, got: %s", resultText(t, result))
	}
}

func TestHandleListCustomAttributes(t *testing.T) {
	c := &Client{Organizations: &stubOrganizations{
		listCustomAttributesFn: func(_ context.Context, _ organizations.ListCustomAttributesInput) (types.Page[organizations.CustomAttribute], error) {
			return types.Page[organizations.CustomAttribute]{Items: []organizations.CustomAttribute{{ID: "attr1", Key: "team", Scope: "PROJECT"}}}, nil
		},
	}}
	result, _, err := HandleListCustomAttributes(c)(context.Background(), nil, ListCustomAttributesInput{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(resultText(t, result), "team") {
		t.Errorf("expected attribute in result, got: %s", resultText(t, result))
	}
}
