package tools

import (
	"context"
	"strings"
	"testing"

	"github.com/massdriver-cloud/massdriver-sdk-go/massdriver/platform/serviceaccounts"
	"github.com/massdriver-cloud/massdriver-sdk-go/massdriver/platform/types"
)

type stubServiceAccounts struct {
	listPageFn func(context.Context, serviceaccounts.ListInput) (types.Page[serviceaccounts.ServiceAccount], error)
	getFn      func(context.Context, string) (*serviceaccounts.ServiceAccount, error)
	createFn   func(context.Context, serviceaccounts.CreateInput) (*serviceaccounts.Created, error)
	updateFn   func(context.Context, string, serviceaccounts.UpdateInput) (*serviceaccounts.ServiceAccount, error)
	deleteFn   func(context.Context, string) (*serviceaccounts.ServiceAccount, error)
}

func (s *stubServiceAccounts) ListPage(ctx context.Context, input serviceaccounts.ListInput) (types.Page[serviceaccounts.ServiceAccount], error) {
	return s.listPageFn(ctx, input)
}
func (s *stubServiceAccounts) Get(ctx context.Context, id string) (*serviceaccounts.ServiceAccount, error) {
	return s.getFn(ctx, id)
}
func (s *stubServiceAccounts) Create(ctx context.Context, input serviceaccounts.CreateInput) (*serviceaccounts.Created, error) {
	return s.createFn(ctx, input)
}
func (s *stubServiceAccounts) Update(ctx context.Context, id string, input serviceaccounts.UpdateInput) (*serviceaccounts.ServiceAccount, error) {
	return s.updateFn(ctx, id, input)
}
func (s *stubServiceAccounts) Delete(ctx context.Context, id string) (*serviceaccounts.ServiceAccount, error) {
	return s.deleteFn(ctx, id)
}

func TestHandleListServiceAccounts(t *testing.T) {
	tests := []struct {
		name     string
		input    ListServiceAccountsInput
		stub     *stubServiceAccounts
		wantErr  bool
		wantText string
	}{
		{
			name:  "returns service accounts page",
			input: ListServiceAccountsInput{},
			stub: &stubServiceAccounts{
				listPageFn: func(_ context.Context, _ serviceaccounts.ListInput) (types.Page[serviceaccounts.ServiceAccount], error) {
					return types.Page[serviceaccounts.ServiceAccount]{
						Items: []serviceaccounts.ServiceAccount{{ID: "sa1", Name: "CI Bot"}},
					}, nil
				},
			},
			wantText: "CI Bot",
		},
		{
			name:  "empty page surfaces has_more false",
			input: ListServiceAccountsInput{},
			stub: &stubServiceAccounts{
				listPageFn: func(context.Context, serviceaccounts.ListInput) (types.Page[serviceaccounts.ServiceAccount], error) {
					return types.Page[serviceaccounts.ServiceAccount]{}, nil
				},
			},
			wantText: "\"has_more\": false",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &Client{ServiceAccounts: tt.stub}
			handler := HandleListServiceAccounts(c)
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

func TestHandleGetServiceAccount(t *testing.T) {
	tests := []struct {
		name     string
		input    GetServiceAccountInput
		stub     *stubServiceAccounts
		wantErr  string
		wantText string
	}{
		{
			name:    "missing id",
			input:   GetServiceAccountInput{},
			stub:    &stubServiceAccounts{},
			wantErr: "id is required",
		},
		{
			name:  "returns service account JSON",
			input: GetServiceAccountInput{ID: "sa1"},
			stub: &stubServiceAccounts{
				getFn: func(_ context.Context, id string) (*serviceaccounts.ServiceAccount, error) {
					return &serviceaccounts.ServiceAccount{ID: id, Name: "CI Bot"}, nil
				},
			},
			wantText: "CI Bot",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &Client{ServiceAccounts: tt.stub}
			handler := HandleGetServiceAccount(c)
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

func TestHandleCreateServiceAccount(t *testing.T) {
	tests := []struct {
		name     string
		input    CreateServiceAccountInput
		stub     *stubServiceAccounts
		wantErr  string
		wantText string
	}{
		{
			name:    "missing name",
			input:   CreateServiceAccountInput{},
			stub:    &stubServiceAccounts{},
			wantErr: "name is required",
		},
		{
			name:  "success returns created JSON",
			input: CreateServiceAccountInput{Name: "CI Bot"},
			stub: &stubServiceAccounts{
				createFn: func(_ context.Context, input serviceaccounts.CreateInput) (*serviceaccounts.Created, error) {
					return &serviceaccounts.Created{
						ServiceAccount: serviceaccounts.ServiceAccount{ID: "sa1", Name: input.Name},
						DefaultToken:   "tok_secret",
					}, nil
				},
			},
			wantText: "sa1",
		},
		{
			name:  "mutation failure returns error message",
			input: CreateServiceAccountInput{Name: "CI Bot"},
			stub: &stubServiceAccounts{
				createFn: func(context.Context, serviceaccounts.CreateInput) (*serviceaccounts.Created, error) {
					return nil, mutationFailedErr("create service account", "name", "already exists")
				},
			},
			wantText: "create_service_account failed",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &Client{ServiceAccounts: tt.stub}
			handler := HandleCreateServiceAccount(c)
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

func TestHandleUpdateServiceAccount(t *testing.T) {
	tests := []struct {
		name     string
		input    UpdateServiceAccountInput
		stub     *stubServiceAccounts
		wantErr  string
		wantText string
	}{
		{
			name:    "missing id",
			input:   UpdateServiceAccountInput{Name: "New Name"},
			stub:    &stubServiceAccounts{},
			wantErr: "id is required",
		},
		{
			name:  "success returns service account JSON",
			input: UpdateServiceAccountInput{ID: "sa1", Name: "Updated Bot"},
			stub: &stubServiceAccounts{
				updateFn: func(_ context.Context, id string, input serviceaccounts.UpdateInput) (*serviceaccounts.ServiceAccount, error) {
					return &serviceaccounts.ServiceAccount{ID: id, Name: input.Name}, nil
				},
			},
			wantText: "Updated Bot",
		},
		{
			name:  "mutation failure returns error message",
			input: UpdateServiceAccountInput{ID: "sa1"},
			stub: &stubServiceAccounts{
				updateFn: func(context.Context, string, serviceaccounts.UpdateInput) (*serviceaccounts.ServiceAccount, error) {
					return nil, mutationFailedErr("update service account", "name", "too long")
				},
			},
			wantText: "update_service_account failed",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &Client{ServiceAccounts: tt.stub}
			handler := HandleUpdateServiceAccount(c)
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

func TestHandleDeleteServiceAccount(t *testing.T) {
	tests := []struct {
		name     string
		input    DeleteServiceAccountInput
		stub     *stubServiceAccounts
		wantErr  string
		wantText string
	}{
		{
			name:    "missing id",
			input:   DeleteServiceAccountInput{},
			stub:    &stubServiceAccounts{},
			wantErr: "id is required",
		},
		{
			name:  "success returns confirmation message",
			input: DeleteServiceAccountInput{ID: "sa1"},
			stub: &stubServiceAccounts{
				deleteFn: func(_ context.Context, id string) (*serviceaccounts.ServiceAccount, error) {
					return &serviceaccounts.ServiceAccount{ID: id}, nil
				},
			},
			wantText: "deleted successfully",
		},
		{
			name:  "mutation failure returns error message",
			input: DeleteServiceAccountInput{ID: "sa1"},
			stub: &stubServiceAccounts{
				deleteFn: func(context.Context, string) (*serviceaccounts.ServiceAccount, error) {
					return nil, mutationFailedErr("delete service account", "", "still in use")
				},
			},
			wantText: "delete_service_account failed",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &Client{ServiceAccounts: tt.stub}
			handler := HandleDeleteServiceAccount(c)
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
