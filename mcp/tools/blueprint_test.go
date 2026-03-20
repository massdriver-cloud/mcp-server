package tools

import (
	"context"
	"strings"
	"testing"

	"github.com/massdriver-cloud/mcp-server/internal/gqlmock"
)

func toolComponentMutationResp(op, id, name string) any {
	return gqlmock.MockMutationResponse(op, map[string]any{
		"id": id, "name": name, "description": "",
		"createdAt": "2024-01-01T00:00:00Z", "updatedAt": "2024-01-01T00:00:00Z",
	})
}

func TestHandleAddComponent(t *testing.T) {
	tests := []struct {
		name      string
		input     AddComponentInput
		responses []any
		wantErr   string
		wantText  string
	}{
		{
			name:    "missing project_id",
			input:   AddComponentInput{BundleName: "aws-rds", ID: "db", Name: "DB"},
			wantErr: "project_id is required",
		},
		{
			name:    "missing bundle_name",
			input:   AddComponentInput{ProjectID: "proj1", ID: "db", Name: "DB"},
			wantErr: "bundle_name is required",
		},
		{
			name:    "missing id",
			input:   AddComponentInput{ProjectID: "proj1", BundleName: "aws-rds", Name: "DB"},
			wantErr: "id is required",
		},
		{
			name:    "missing name",
			input:   AddComponentInput{ProjectID: "proj1", BundleName: "aws-rds", ID: "db"},
			wantErr: "name is required",
		},
		{
			name:      "success returns component JSON",
			input:     AddComponentInput{ProjectID: "proj1", BundleName: "aws-rds-postgres", ID: "db", Name: "Database"},
			responses: []any{toolComponentMutationResp("addComponent", "db", "Database")},
			wantText:  "Database",
		},
		{
			name:      "payload failure returns error message",
			input:     AddComponentInput{ProjectID: "proj1", BundleName: "aws-rds-postgres", ID: "db", Name: "Database"},
			responses: []any{toolMutationFailureResp("addComponent", "id", "already exists")},
			wantText:  "add_component failed",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler := HandleAddComponent(newToolClient(tt.responses))
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
		name      string
		input     RemoveComponentInput
		responses []any
		wantErr   string
		wantText  string
	}{
		{
			name:    "missing project_id",
			input:   RemoveComponentInput{ID: "db"},
			wantErr: "project_id is required",
		},
		{
			name:    "missing id",
			input:   RemoveComponentInput{ProjectID: "proj1"},
			wantErr: "id is required",
		},
		{
			name:      "success returns confirmation",
			input:     RemoveComponentInput{ProjectID: "proj1", ID: "db"},
			responses: []any{gqlmock.MockMutationResponse("removeComponent", map[string]any{"id": "db", "name": "Database"})},
			wantText:  "removed successfully",
		},
		{
			name:      "payload failure returns error message",
			input:     RemoveComponentInput{ProjectID: "proj1", ID: "db"},
			responses: []any{toolMutationFailureResp("removeComponent", "", "has active instances")},
			wantText:  "remove_component failed",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler := HandleRemoveComponent(newToolClient(tt.responses))
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
		ProjectID: "proj1", From: "db", FromField: "database", FromVersion: "~1.0",
		To: "app", ToField: "database", ToVersion: "~1.0",
	}
	tests := []struct {
		name      string
		input     LinkComponentsInput
		responses []any
		wantErr   string
		wantText  string
	}{
		{
			name:    "missing project_id",
			input:   LinkComponentsInput{From: "db", FromField: "f", FromVersion: "~1", To: "app", ToField: "f", ToVersion: "~1"},
			wantErr: "project_id is required",
		},
		{
			name:    "missing from",
			input:   LinkComponentsInput{ProjectID: "proj1", FromField: "f", FromVersion: "~1", To: "app", ToField: "f", ToVersion: "~1"},
			wantErr: "from is required",
		},
		{
			name:      "success returns link JSON",
			input:     validInput,
			responses: []any{gqlmock.MockMutationResponse("linkComponents", map[string]any{
				"id": "link1", "fromField": "database", "toField": "database",
				"fromComponent": map[string]any{"id": "db", "name": "Database"},
				"toComponent":   map[string]any{"id": "app", "name": "App"},
			})},
			wantText: "link1",
		},
		{
			name:      "payload failure returns error message",
			input:     validInput,
			responses: []any{toolMutationFailureResp("linkComponents", "from", "already linked")},
			wantText:  "link_components failed",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler := HandleLinkComponents(newToolClient(tt.responses))
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
		name      string
		input     UnlinkComponentsInput
		responses []any
		wantErr   string
		wantText  string
	}{
		{
			name:    "missing id",
			input:   UnlinkComponentsInput{},
			wantErr: "id is required",
		},
		{
			name:      "success returns confirmation",
			input:     UnlinkComponentsInput{ID: "link1"},
			responses: []any{gqlmock.MockMutationResponse("unlinkComponents", map[string]any{"id": "link1"})},
			wantText:  "removed successfully",
		},
		{
			name:      "payload failure returns error message",
			input:     UnlinkComponentsInput{ID: "link1"},
			responses: []any{toolMutationFailureResp("unlinkComponents", "", "not found")},
			wantText:  "unlink_components failed",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler := HandleUnlinkComponents(newToolClient(tt.responses))
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
