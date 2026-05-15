package tools

import (
	"context"
	"strings"
	"testing"

	"github.com/massdriver-cloud/massdriver-sdk-go/massdriver/platform/groups"
)

type stubGroups struct {
	listFn                 func(context.Context, groups.ListInput) ([]groups.Group, error)
	getFn                  func(context.Context, string) (*groups.Group, error)
	createFn               func(context.Context, groups.CreateInput) (*groups.Group, error)
	updateFn               func(context.Context, string, groups.UpdateInput) (*groups.Group, error)
	deleteFn               func(context.Context, string) (*groups.Group, error)
	addUserFn              func(context.Context, string, string) (*groups.AddUserResult, error)
	removeUserFn           func(context.Context, string, string) error
	revokeInvitationFn     func(context.Context, string, string) error
	addServiceAccountFn    func(context.Context, string, string) error
	removeServiceAccountFn func(context.Context, string, string) error
}

func (s *stubGroups) List(ctx context.Context, input groups.ListInput) ([]groups.Group, error) {
	return s.listFn(ctx, input)
}
func (s *stubGroups) Get(ctx context.Context, id string) (*groups.Group, error) {
	return s.getFn(ctx, id)
}
func (s *stubGroups) Create(ctx context.Context, input groups.CreateInput) (*groups.Group, error) {
	return s.createFn(ctx, input)
}
func (s *stubGroups) Update(ctx context.Context, id string, input groups.UpdateInput) (*groups.Group, error) {
	return s.updateFn(ctx, id, input)
}
func (s *stubGroups) Delete(ctx context.Context, id string) (*groups.Group, error) {
	return s.deleteFn(ctx, id)
}
func (s *stubGroups) AddUser(ctx context.Context, groupID, email string) (*groups.AddUserResult, error) {
	return s.addUserFn(ctx, groupID, email)
}
func (s *stubGroups) RemoveUser(ctx context.Context, groupID, email string) error {
	return s.removeUserFn(ctx, groupID, email)
}
func (s *stubGroups) RevokeInvitation(ctx context.Context, groupID, email string) error {
	return s.revokeInvitationFn(ctx, groupID, email)
}
func (s *stubGroups) AddServiceAccount(ctx context.Context, groupID, serviceAccountID string) error {
	return s.addServiceAccountFn(ctx, groupID, serviceAccountID)
}
func (s *stubGroups) RemoveServiceAccount(ctx context.Context, groupID, serviceAccountID string) error {
	return s.removeServiceAccountFn(ctx, groupID, serviceAccountID)
}

func TestHandleListGroups(t *testing.T) {
	tests := []struct {
		name     string
		stub     *stubGroups
		wantErr  bool
		wantText string
	}{
		{
			name: "returns groups",
			stub: &stubGroups{
				listFn: func(_ context.Context, _ groups.ListInput) ([]groups.Group, error) {
					return []groups.Group{{ID: "grp1", Name: "Admins"}}, nil
				},
			},
			wantText: "Admins",
		},
		{
			name: "returns null for empty list",
			stub: &stubGroups{
				listFn: func(context.Context, groups.ListInput) ([]groups.Group, error) {
					return nil, nil
				},
			},
			wantText: "null",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &Client{Groups: tt.stub}
			handler := HandleListGroups(c)
			result, _, err := handler(context.Background(), nil, ListGroupsInput{})
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

func TestHandleGetGroup(t *testing.T) {
	tests := []struct {
		name     string
		input    GetGroupInput
		stub     *stubGroups
		wantErr  string
		wantText string
	}{
		{
			name:    "missing id",
			input:   GetGroupInput{},
			stub:    &stubGroups{},
			wantErr: "id is required",
		},
		{
			name:  "returns group JSON",
			input: GetGroupInput{ID: "grp1"},
			stub: &stubGroups{
				getFn: func(_ context.Context, id string) (*groups.Group, error) {
					return &groups.Group{ID: id, Name: "Admins"}, nil
				},
			},
			wantText: "Admins",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &Client{Groups: tt.stub}
			handler := HandleGetGroup(c)
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

func TestHandleCreateGroup(t *testing.T) {
	tests := []struct {
		name     string
		input    CreateGroupInput
		stub     *stubGroups
		wantErr  string
		wantText string
	}{
		{
			name:    "missing name",
			input:   CreateGroupInput{},
			stub:    &stubGroups{},
			wantErr: "name is required",
		},
		{
			name:  "success returns group JSON",
			input: CreateGroupInput{Name: "Developers"},
			stub: &stubGroups{
				createFn: func(_ context.Context, input groups.CreateInput) (*groups.Group, error) {
					return &groups.Group{ID: "grp1", Name: input.Name}, nil
				},
			},
			wantText: "Developers",
		},
		{
			name:  "mutation failure returns error message",
			input: CreateGroupInput{Name: "Developers"},
			stub: &stubGroups{
				createFn: func(context.Context, groups.CreateInput) (*groups.Group, error) {
					return nil, mutationFailedErr("create group", "name", "already exists")
				},
			},
			wantText: "create_group failed",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &Client{Groups: tt.stub}
			handler := HandleCreateGroup(c)
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

func TestHandleUpdateGroup(t *testing.T) {
	tests := []struct {
		name     string
		input    UpdateGroupInput
		stub     *stubGroups
		wantErr  string
		wantText string
	}{
		{
			name:    "missing id",
			input:   UpdateGroupInput{Name: "New Name"},
			stub:    &stubGroups{},
			wantErr: "id is required",
		},
		{
			name:  "success returns group JSON",
			input: UpdateGroupInput{ID: "grp1", Name: "Updated"},
			stub: &stubGroups{
				updateFn: func(_ context.Context, id string, input groups.UpdateInput) (*groups.Group, error) {
					return &groups.Group{ID: id, Name: input.Name}, nil
				},
			},
			wantText: "Updated",
		},
		{
			name:  "mutation failure returns error message",
			input: UpdateGroupInput{ID: "grp1"},
			stub: &stubGroups{
				updateFn: func(context.Context, string, groups.UpdateInput) (*groups.Group, error) {
					return nil, mutationFailedErr("update group", "name", "too long")
				},
			},
			wantText: "update_group failed",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &Client{Groups: tt.stub}
			handler := HandleUpdateGroup(c)
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

func TestHandleDeleteGroup(t *testing.T) {
	tests := []struct {
		name     string
		input    DeleteGroupInput
		stub     *stubGroups
		wantErr  string
		wantText string
	}{
		{
			name:    "missing id",
			input:   DeleteGroupInput{},
			stub:    &stubGroups{},
			wantErr: "id is required",
		},
		{
			name:  "success returns confirmation message",
			input: DeleteGroupInput{ID: "grp1"},
			stub: &stubGroups{
				deleteFn: func(_ context.Context, id string) (*groups.Group, error) {
					return &groups.Group{ID: id}, nil
				},
			},
			wantText: "deleted successfully",
		},
		{
			name:  "mutation failure returns error message",
			input: DeleteGroupInput{ID: "grp1"},
			stub: &stubGroups{
				deleteFn: func(context.Context, string) (*groups.Group, error) {
					return nil, mutationFailedErr("delete group", "", "has active members")
				},
			},
			wantText: "delete_group failed",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &Client{Groups: tt.stub}
			handler := HandleDeleteGroup(c)
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

func TestHandleAddGroupUser(t *testing.T) {
	tests := []struct {
		name     string
		input    AddGroupUserInput
		stub     *stubGroups
		wantErr  string
		wantText string
	}{
		{
			name:    "missing group_id",
			input:   AddGroupUserInput{Email: "user@example.com"},
			stub:    &stubGroups{},
			wantErr: "group_id is required",
		},
		{
			name:    "missing email",
			input:   AddGroupUserInput{GroupID: "grp1"},
			stub:    &stubGroups{},
			wantErr: "email is required",
		},
		{
			name:  "success returns result JSON",
			input: AddGroupUserInput{GroupID: "grp1", Email: "user@example.com"},
			stub: &stubGroups{
				addUserFn: func(_ context.Context, _, email string) (*groups.AddUserResult, error) {
					return &groups.AddUserResult{Invitation: &groups.Invitation{Email: email}}, nil
				},
			},
			wantText: "user@example.com",
		},
		{
			name:  "mutation failure returns error message",
			input: AddGroupUserInput{GroupID: "grp1", Email: "user@example.com"},
			stub: &stubGroups{
				addUserFn: func(context.Context, string, string) (*groups.AddUserResult, error) {
					return nil, mutationFailedErr("add group user", "email", "already a member")
				},
			},
			wantText: "add_group_user failed",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &Client{Groups: tt.stub}
			handler := HandleAddGroupUser(c)
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

func TestHandleRemoveGroupUser(t *testing.T) {
	tests := []struct {
		name     string
		input    RemoveGroupUserInput
		stub     *stubGroups
		wantErr  string
		wantText string
	}{
		{
			name:    "missing group_id",
			input:   RemoveGroupUserInput{Email: "user@example.com"},
			stub:    &stubGroups{},
			wantErr: "group_id is required",
		},
		{
			name:    "missing email",
			input:   RemoveGroupUserInput{GroupID: "grp1"},
			stub:    &stubGroups{},
			wantErr: "email is required",
		},
		{
			name:  "success returns confirmation message",
			input: RemoveGroupUserInput{GroupID: "grp1", Email: "user@example.com"},
			stub: &stubGroups{
				removeUserFn: func(_ context.Context, _, _ string) error {
					return nil
				},
			},
			wantText: "removed from group",
		},
		{
			name:  "mutation failure returns error message",
			input: RemoveGroupUserInput{GroupID: "grp1", Email: "user@example.com"},
			stub: &stubGroups{
				removeUserFn: func(context.Context, string, string) error {
					return mutationFailedErr("remove group user", "email", "not a member")
				},
			},
			wantText: "remove_group_user failed",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &Client{Groups: tt.stub}
			handler := HandleRemoveGroupUser(c)
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

func TestHandleRevokeGroupInvitation(t *testing.T) {
	tests := []struct {
		name     string
		input    RevokeGroupInvitationInput
		stub     *stubGroups
		wantErr  string
		wantText string
	}{
		{
			name:    "missing group_id",
			input:   RevokeGroupInvitationInput{Email: "user@example.com"},
			stub:    &stubGroups{},
			wantErr: "group_id is required",
		},
		{
			name:    "missing email",
			input:   RevokeGroupInvitationInput{GroupID: "grp1"},
			stub:    &stubGroups{},
			wantErr: "email is required",
		},
		{
			name:  "success returns confirmation message",
			input: RevokeGroupInvitationInput{GroupID: "grp1", Email: "user@example.com"},
			stub: &stubGroups{
				revokeInvitationFn: func(_ context.Context, _, _ string) error {
					return nil
				},
			},
			wantText: "revoked from group",
		},
		{
			name:  "mutation failure returns error message",
			input: RevokeGroupInvitationInput{GroupID: "grp1", Email: "user@example.com"},
			stub: &stubGroups{
				revokeInvitationFn: func(context.Context, string, string) error {
					return mutationFailedErr("revoke group invitation", "email", "no pending invitation")
				},
			},
			wantText: "revoke_group_invitation failed",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &Client{Groups: tt.stub}
			handler := HandleRevokeGroupInvitation(c)
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

func TestHandleAddGroupServiceAccount(t *testing.T) {
	tests := []struct {
		name     string
		input    AddGroupServiceAccountInput
		stub     *stubGroups
		wantErr  string
		wantText string
	}{
		{
			name:    "missing group_id",
			input:   AddGroupServiceAccountInput{ServiceAccountID: "sa1"},
			stub:    &stubGroups{},
			wantErr: "group_id is required",
		},
		{
			name:    "missing service_account_id",
			input:   AddGroupServiceAccountInput{GroupID: "grp1"},
			stub:    &stubGroups{},
			wantErr: "service_account_id is required",
		},
		{
			name:  "success returns confirmation message",
			input: AddGroupServiceAccountInput{GroupID: "grp1", ServiceAccountID: "sa1"},
			stub: &stubGroups{
				addServiceAccountFn: func(_ context.Context, _, _ string) error {
					return nil
				},
			},
			wantText: "added to group",
		},
		{
			name:  "mutation failure returns error message",
			input: AddGroupServiceAccountInput{GroupID: "grp1", ServiceAccountID: "sa1"},
			stub: &stubGroups{
				addServiceAccountFn: func(context.Context, string, string) error {
					return mutationFailedErr("add group service account", "service_account", "already a member")
				},
			},
			wantText: "add_group_service_account failed",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &Client{Groups: tt.stub}
			handler := HandleAddGroupServiceAccount(c)
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

func TestHandleRemoveGroupServiceAccount(t *testing.T) {
	tests := []struct {
		name     string
		input    RemoveGroupServiceAccountInput
		stub     *stubGroups
		wantErr  string
		wantText string
	}{
		{
			name:    "missing group_id",
			input:   RemoveGroupServiceAccountInput{ServiceAccountID: "sa1"},
			stub:    &stubGroups{},
			wantErr: "group_id is required",
		},
		{
			name:    "missing service_account_id",
			input:   RemoveGroupServiceAccountInput{GroupID: "grp1"},
			stub:    &stubGroups{},
			wantErr: "service_account_id is required",
		},
		{
			name:  "success returns confirmation message",
			input: RemoveGroupServiceAccountInput{GroupID: "grp1", ServiceAccountID: "sa1"},
			stub: &stubGroups{
				removeServiceAccountFn: func(_ context.Context, _, _ string) error {
					return nil
				},
			},
			wantText: "removed from group",
		},
		{
			name:  "mutation failure returns error message",
			input: RemoveGroupServiceAccountInput{GroupID: "grp1", ServiceAccountID: "sa1"},
			stub: &stubGroups{
				removeServiceAccountFn: func(context.Context, string, string) error {
					return mutationFailedErr("remove group service account", "service_account", "not a member")
				},
			},
			wantText: "remove_group_service_account failed",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &Client{Groups: tt.stub}
			handler := HandleRemoveGroupServiceAccount(c)
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
