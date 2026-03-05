package tools

import (
	"context"
	"strings"
	"testing"

	"github.com/massdriver-cloud/mcp-server/internal/gqlmock"
)

func toolProjectResp(id, name string) any {
	return gqlmock.MockQueryResponse("projects", map[string]any{
		"cursor": map[string]any{"next": ""},
		"items": []map[string]any{{
			"id": id, "name": name, "description": "",
			"createdAt": "2024-01-01T00:00:00Z", "updatedAt": "2024-01-01T00:00:00Z",
		}},
	})
}

func toolProjectMutationResp(op, id, name string) any {
	return gqlmock.MockMutationResponse(op, map[string]any{
		"id": id, "name": name, "description": "",
		"createdAt": "2024-01-01T00:00:00Z", "updatedAt": "2024-01-01T00:00:00Z",
	})
}

func toolMutationFailureResp(op, field, msg string) any {
	return map[string]any{
		"data": map[string]any{
			op: map[string]any{
				"successful": false,
				"result":     map[string]any{"id": "", "name": ""},
				"messages":   []map[string]any{{"field": field, "message": msg, "code": "invalid"}},
			},
		},
	}
}

func TestHandleListProjects(t *testing.T) {
	tests := []struct {
		name      string
		responses []any
		wantErr   bool
		wantText  string
	}{
		{
			name:      "returns all projects as JSON",
			responses: []any{toolProjectResp("myproj", "My Project")},
			wantText:  "myproj",
		},
		{
			name:      "returns null for empty project list",
			responses: []any{gqlmock.MockQueryResponse("projects", map[string]any{"cursor": map[string]any{"next": ""}, "items": []any{}})},
			wantText:  "null",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler := HandleListProjects(newToolClient(tt.responses))
			result, _, err := handler(context.Background(), nil, ListProjectsInput{})
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

func TestHandleGetProject(t *testing.T) {
	tests := []struct {
		name      string
		input     GetProjectInput
		responses []any
		wantErr   string
		wantText  string
	}{
		{
			name:    "missing id",
			input:   GetProjectInput{},
			wantErr: "id is required",
		},
		{
			name:  "returns project JSON",
			input: GetProjectInput{ID: "myproj"},
			responses: []any{gqlmock.MockQueryResponse("project", map[string]any{
				"id": "myproj", "name": "My Project", "description": "",
				"createdAt": "2024-01-01T00:00:00Z", "updatedAt": "2024-01-01T00:00:00Z",
				"environments": map[string]any{"cursor": map[string]any{"next": ""}, "items": []any{}},
			})},
			wantText: "myproj",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler := HandleGetProject(newToolClient(tt.responses))
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

func TestHandleCreateProject(t *testing.T) {
	tests := []struct {
		name      string
		input     CreateProjectInput
		responses []any
		wantErr   string
		wantText  string
	}{
		{
			name:    "missing id",
			input:   CreateProjectInput{Name: "My Project"},
			wantErr: "id is required",
		},
		{
			name:    "missing name",
			input:   CreateProjectInput{ID: "myproj"},
			wantErr: "name is required",
		},
		{
			name:      "success returns project JSON",
			input:     CreateProjectInput{ID: "myproj", Name: "My Project"},
			responses: []any{toolProjectMutationResp("createProject", "myproj", "My Project")},
			wantText:  "myproj",
		},
		{
			name:      "payload failure returns error message",
			input:     CreateProjectInput{ID: "myproj", Name: "My Project"},
			responses: []any{toolMutationFailureResp("createProject", "id", "already taken")},
			wantText:  "create_project failed",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler := HandleCreateProject(newToolClient(tt.responses))
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

func TestHandleUpdateProject(t *testing.T) {
	tests := []struct {
		name      string
		input     UpdateProjectInput
		responses []any
		wantErr   string
		wantText  string
	}{
		{
			name:    "missing id",
			input:   UpdateProjectInput{Name: "New Name"},
			wantErr: "id is required",
		},
		{
			name:      "success returns updated project JSON",
			input:     UpdateProjectInput{ID: "myproj", Name: "Updated Name"},
			responses: []any{toolProjectMutationResp("updateProject", "myproj", "Updated Name")},
			wantText:  "Updated Name",
		},
		{
			name:      "payload failure returns error message",
			input:     UpdateProjectInput{ID: "myproj"},
			responses: []any{toolMutationFailureResp("updateProject", "name", "too long")},
			wantText:  "update_project failed",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler := HandleUpdateProject(newToolClient(tt.responses))
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

func TestHandleDeleteProject(t *testing.T) {
	tests := []struct {
		name      string
		input     DeleteProjectInput
		responses []any
		wantErr   string
		wantText  string
	}{
		{
			name:    "missing id",
			input:   DeleteProjectInput{},
			wantErr: "id is required",
		},
		{
			name:      "success returns confirmation message",
			input:     DeleteProjectInput{ID: "myproj"},
			responses: []any{gqlmock.MockMutationResponse("deleteProject", map[string]any{"id": "myproj", "name": "My Project"})},
			wantText:  "deleted successfully",
		},
		{
			name:      "payload failure returns error message",
			input:     DeleteProjectInput{ID: "myproj"},
			responses: []any{toolMutationFailureResp("deleteProject", "", "not empty")},
			wantText:  "delete_project failed",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler := HandleDeleteProject(newToolClient(tt.responses))
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
