package tools

import (
	"context"
	"strings"
	"testing"

	"github.com/massdriver-cloud/mcp-server/internal/gqlmock"
)

func toolInstanceResp(id string) any {
	return gqlmock.MockQueryResponse("instances", map[string]any{
		"cursor": map[string]any{"next": ""},
		"items": []map[string]any{{
			"id": id, "name": id, "status": "PROVISIONED",
			"version": "~1.0", "releaseStrategy": "STABLE",
			"createdAt": "2024-01-01T00:00:00Z", "updatedAt": "2024-01-01T00:00:00Z",
			"environment": map[string]any{
				"id": "proj1-staging", "name": "Staging",
				"project": map[string]any{"id": "proj1", "name": "Project One"},
			},
			"release": map[string]any{"id": "rel1", "name": "aws-rds", "version": "1.2.3"},
		}},
	})
}

func TestHandleListInstances(t *testing.T) {
	tests := []struct {
		name      string
		input     ListInstancesInput
		responses []any
		wantErr   bool
		wantText  string
	}{
		{
			name:      "returns all instances",
			input:     ListInstancesInput{},
			responses: []any{toolInstanceResp("proj1-staging-db")},
			wantText:  "proj1-staging-db",
		},
		{
			name:      "filters by environment ID",
			input:     ListInstancesInput{EnvironmentID: "proj1-staging"},
			responses: []any{toolInstanceResp("proj1-staging-db")},
			wantText:  "proj1-staging-db",
		},
		{
			name:      "filters by status",
			input:     ListInstancesInput{Status: "PROVISIONED"},
			responses: []any{toolInstanceResp("proj1-staging-db")},
			wantText:  "PROVISIONED",
		},
		{
			name: "returns null for empty instance list",
			input: ListInstancesInput{},
			responses: []any{gqlmock.MockQueryResponse("instances", map[string]any{
				"cursor": map[string]any{"next": ""},
				"items":  []any{},
			})},
			wantText: "null",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler := HandleListInstances(newToolClient(tt.responses))
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

func TestHandleGetInstance(t *testing.T) {
	tests := []struct {
		name      string
		input     GetInstanceInput
		responses []any
		wantErr   string
		wantText  string
	}{
		{
			name:    "missing id",
			input:   GetInstanceInput{},
			wantErr: "id is required",
		},
		{
			name:  "returns instance JSON",
			input: GetInstanceInput{ID: "proj1-staging-db"},
			responses: []any{gqlmock.MockQueryResponse("instance", map[string]any{
				"id": "proj1-staging-db", "name": "Database", "status": "PROVISIONED",
				"version": "~1.0", "releaseStrategy": "STABLE",
				"createdAt": "2024-01-01T00:00:00Z", "updatedAt": "2024-01-01T00:00:00Z",
				"environment": map[string]any{
					"id": "proj1-staging", "name": "Staging",
					"project": map[string]any{"id": "proj1", "name": "Project One"},
				},
				"release": map[string]any{"id": "rel1", "name": "aws-rds", "version": "1.2.3"},
			})},
			wantText: "proj1-staging-db",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler := HandleGetInstance(newToolClient(tt.responses))
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
