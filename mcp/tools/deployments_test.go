package tools

import (
	"context"
	"strings"
	"testing"

	"github.com/massdriver-cloud/mcp-server/internal/gqlmock"
)

func toolDeploymentResp(id, status string) any {
	return gqlmock.MockQueryResponse("deployments", map[string]any{
		"cursor": map[string]any{"next": ""},
		"items": []map[string]any{{
			"id": id, "status": status, "action": "PROVISION",
			"version": "1.0.0", "message": "Deployed",
			"createdAt": "2024-01-01T00:00:00Z", "updatedAt": "2024-01-01T00:00:00Z",
			"lastTransitionedAt": "2024-01-01T00:01:00Z",
			"elapsedTime": 60, "deployedBy": "user@example.com",
			"instance": map[string]any{
				"id": "proj1-staging-db", "name": "Database",
				"environment": map[string]any{
					"id": "proj1-staging", "name": "Staging",
					"project": map[string]any{"id": "proj1", "name": "Project One"},
				},
			},
		}},
	})
}

func TestHandleListDeployments(t *testing.T) {
	tests := []struct {
		name      string
		input     ListDeploymentsInput
		responses []any
		wantErr   bool
		wantText  string
	}{
		{
			name:      "returns all deployments",
			input:     ListDeploymentsInput{},
			responses: []any{toolDeploymentResp("dep1", "COMPLETED")},
			wantText:  "dep1",
		},
		{
			name:      "filters by instance ID",
			input:     ListDeploymentsInput{InstanceID: "proj1-staging-db"},
			responses: []any{toolDeploymentResp("dep1", "COMPLETED")},
			wantText:  "COMPLETED",
		},
		{
			name:      "filters by status",
			input:     ListDeploymentsInput{Status: "RUNNING"},
			responses: []any{toolDeploymentResp("dep2", "RUNNING")},
			wantText:  "RUNNING",
		},
		{
			name:      "filters by action",
			input:     ListDeploymentsInput{Action: "PROVISION"},
			responses: []any{toolDeploymentResp("dep1", "COMPLETED")},
			wantText:  "PROVISION",
		},
		{
			name: "returns null for empty list",
			input: ListDeploymentsInput{},
			responses: []any{gqlmock.MockQueryResponse("deployments", map[string]any{
				"cursor": map[string]any{"next": ""},
				"items":  []any{},
			})},
			wantText: "null",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler := HandleListDeployments(newToolClient(tt.responses))
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
		name      string
		input     GetDeploymentInput
		responses []any
		wantErr   string
		wantText  string
	}{
		{
			name:    "missing id",
			input:   GetDeploymentInput{},
			wantErr: "id is required",
		},
		{
			name:  "returns deployment JSON",
			input: GetDeploymentInput{ID: "dep1"},
			responses: []any{gqlmock.MockQueryResponse("deployment", map[string]any{
				"id": "dep1", "status": "COMPLETED", "action": "PROVISION",
				"version": "1.0.0", "message": "Deployed",
				"createdAt": "2024-01-01T00:00:00Z", "updatedAt": "2024-01-01T00:00:00Z",
				"lastTransitionedAt": "2024-01-01T00:01:00Z",
				"elapsedTime": 60, "deployedBy": "user@example.com",
				"instance": map[string]any{
					"id": "proj1-staging-db", "name": "Database",
					"environment": map[string]any{
						"id": "proj1-staging", "name": "Staging",
						"project": map[string]any{"id": "proj1", "name": "Project One"},
					},
				},
			})},
			wantText: "dep1",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler := HandleGetDeployment(newToolClient(tt.responses))
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
