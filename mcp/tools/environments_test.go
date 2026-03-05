package tools

import (
	"context"
	"strings"
	"testing"

	"github.com/massdriver-cloud/mcp-server/internal/gqlmock"
)

func toolEnvResp(id string) any {
	return gqlmock.MockQueryResponse("environments", map[string]any{
		"cursor": map[string]any{"next": ""},
		"items": []map[string]any{{
			"id": id, "name": id, "description": "",
			"createdAt": "2024-01-01T00:00:00Z", "updatedAt": "2024-01-01T00:00:00Z",
			"project": map[string]any{
				"id": "myproj", "name": "My Project", "description": "",
				"createdAt": "2024-01-01T00:00:00Z", "updatedAt": "2024-01-01T00:00:00Z",
			},
		}},
	})
}

func toolEnvMutationResp(op, id, name string) any {
	return gqlmock.MockMutationResponse(op, map[string]any{
		"id": id, "name": name, "description": "",
		"createdAt": "2024-01-01T00:00:00Z", "updatedAt": "2024-01-01T00:00:00Z",
	})
}

func TestHandleListEnvironments(t *testing.T) {
	tests := []struct {
		name      string
		input     ListEnvironmentsInput
		responses []any
		wantErr   bool
		wantText  string
	}{
		{
			name:      "returns all environments",
			input:     ListEnvironmentsInput{},
			responses: []any{toolEnvResp("myproj-staging")},
			wantText:  "myproj-staging",
		},
		{
			name:      "filters by project ID",
			input:     ListEnvironmentsInput{ProjectID: "myproj"},
			responses: []any{toolEnvResp("myproj-prod")},
			wantText:  "myproj-prod",
		},
		{
			name:      "returns null for empty environment list",
			input:     ListEnvironmentsInput{},
			responses: []any{gqlmock.MockQueryResponse("environments", map[string]any{"cursor": map[string]any{"next": ""}, "items": []any{}})},
			wantText:  "null",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler := HandleListEnvironments(newToolClient(tt.responses))
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

func TestHandleGetEnvironment(t *testing.T) {
	tests := []struct {
		name      string
		input     GetEnvironmentInput
		responses []any
		wantErr   string
		wantText  string
	}{
		{
			name:    "missing id",
			input:   GetEnvironmentInput{},
			wantErr: "id is required",
		},
		{
			name:  "returns environment JSON",
			input: GetEnvironmentInput{ID: "myproj-staging"},
			responses: []any{gqlmock.MockQueryResponse("environment", map[string]any{
				"id": "myproj-staging", "name": "Staging", "description": "",
				"createdAt": "2024-01-01T00:00:00Z", "updatedAt": "2024-01-01T00:00:00Z",
				"project": map[string]any{
					"id": "myproj", "name": "My Project", "description": "",
					"createdAt": "2024-01-01T00:00:00Z", "updatedAt": "2024-01-01T00:00:00Z",
				},
			})},
			wantText: "myproj-staging",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler := HandleGetEnvironment(newToolClient(tt.responses))
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

func TestHandleCreateEnvironment(t *testing.T) {
	tests := []struct {
		name      string
		input     CreateEnvironmentInput
		responses []any
		wantErr   string
		wantText  string
	}{
		{
			name:    "missing project_id",
			input:   CreateEnvironmentInput{ID: "staging", Name: "Staging"},
			wantErr: "project_id is required",
		},
		{
			name:    "missing id",
			input:   CreateEnvironmentInput{ProjectID: "myproj", Name: "Staging"},
			wantErr: "id is required",
		},
		{
			name:    "missing name",
			input:   CreateEnvironmentInput{ProjectID: "myproj", ID: "staging"},
			wantErr: "name is required",
		},
		{
			name:      "success returns environment JSON",
			input:     CreateEnvironmentInput{ProjectID: "myproj", ID: "staging", Name: "Staging"},
			responses: []any{toolEnvMutationResp("createEnvironment", "myproj-staging", "Staging")},
			wantText:  "myproj-staging",
		},
		{
			name:      "payload failure returns error message",
			input:     CreateEnvironmentInput{ProjectID: "myproj", ID: "staging", Name: "Staging"},
			responses: []any{toolMutationFailureResp("createEnvironment", "id", "already exists")},
			wantText:  "create_environment failed",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler := HandleCreateEnvironment(newToolClient(tt.responses))
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

func TestHandleUpdateEnvironment(t *testing.T) {
	tests := []struct {
		name      string
		input     UpdateEnvironmentInput
		responses []any
		wantErr   string
		wantText  string
	}{
		{
			name:    "missing id",
			input:   UpdateEnvironmentInput{Name: "New Name"},
			wantErr: "id is required",
		},
		{
			name:      "success returns updated environment JSON",
			input:     UpdateEnvironmentInput{ID: "myproj-staging", Name: "Production"},
			responses: []any{toolEnvMutationResp("updateEnvironment", "myproj-staging", "Production")},
			wantText:  "Production",
		},
		{
			name:      "payload failure returns error message",
			input:     UpdateEnvironmentInput{ID: "myproj-staging"},
			responses: []any{toolMutationFailureResp("updateEnvironment", "name", "too long")},
			wantText:  "update_environment failed",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler := HandleUpdateEnvironment(newToolClient(tt.responses))
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

func TestHandleDeleteEnvironment(t *testing.T) {
	tests := []struct {
		name      string
		input     DeleteEnvironmentInput
		responses []any
		wantErr   string
		wantText  string
	}{
		{
			name:    "missing id",
			input:   DeleteEnvironmentInput{},
			wantErr: "id is required",
		},
		{
			name:      "success returns confirmation message",
			input:     DeleteEnvironmentInput{ID: "myproj-staging"},
			responses: []any{gqlmock.MockMutationResponse("deleteEnvironment", map[string]any{"id": "myproj-staging", "name": "Staging"})},
			wantText:  "deleted successfully",
		},
		{
			name:      "payload failure returns error message",
			input:     DeleteEnvironmentInput{ID: "myproj-staging"},
			responses: []any{toolMutationFailureResp("deleteEnvironment", "", "packages still active")},
			wantText:  "delete_environment failed",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler := HandleDeleteEnvironment(newToolClient(tt.responses))
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
