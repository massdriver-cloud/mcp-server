package api

import (
	"context"
	"testing"

	"github.com/massdriver-cloud/mcp-server/internal/gqlmock"
)

func projectItem(id, name string) map[string]any {
	return map[string]any{
		"id": id, "name": name, "description": "",
		"createdAt": "2024-01-01T00:00:00Z", "updatedAt": "2024-01-01T00:00:00Z",
	}
}

func TestListProjects(t *testing.T) {
	tests := []struct {
		name      string
		responses []any
		wantIDs   []string
	}{
		{
			name: "single page",
			responses: []any{gqlmock.MockQueryResponse("projects", map[string]any{
				"cursor": map[string]any{"next": ""},
				"items":  []map[string]any{projectItem("proj1", "Project One")},
			})},
			wantIDs: []string{"proj1"},
		},
		{
			name: "multi-page pagination",
			responses: []any{
				gqlmock.MockQueryResponse("projects", map[string]any{
					"cursor": map[string]any{"next": "cursor-p2"},
					"items":  []map[string]any{projectItem("proj1", "P1")},
				}),
				gqlmock.MockQueryResponse("projects", map[string]any{
					"cursor": map[string]any{"next": ""},
					"items":  []map[string]any{projectItem("proj2", "P2")},
				}),
			},
			wantIDs: []string{"proj1", "proj2"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := newTestClient(t, gqlmock.NewClientWithJSONResponseArray(tt.responses))
			projects, err := ListProjects(context.Background(), c)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if len(projects) != len(tt.wantIDs) {
				t.Fatalf("expected %d projects, got %d", len(tt.wantIDs), len(projects))
			}
			for i, id := range tt.wantIDs {
				if projects[i].ID != id {
					t.Errorf("projects[%d].ID: want %q, got %q", i, id, projects[i].ID)
				}
			}
		})
	}
}

func TestGetProject(t *testing.T) {
	tests := []struct {
		name     string
		id       string
		response any
		wantErr  bool
		wantEnvs int
	}{
		{
			name: "returns project with environments",
			id:   "proj1",
			response: gqlmock.MockQueryResponse("project", map[string]any{
				"id": "proj1", "name": "Project One", "description": "First project",
				"createdAt": "2024-01-01T00:00:00Z", "updatedAt": "2024-01-01T00:00:00Z",
				"environments": map[string]any{
					"cursor": map[string]any{"next": ""},
					"items":  []map[string]any{{"id": "proj1-staging", "name": "Staging", "description": "", "createdAt": "2024-01-01T00:00:00Z", "updatedAt": "2024-01-01T00:00:00Z"}},
				},
			}),
			wantEnvs: 1,
		},
		{
			name: "returns error when not found",
			id:   "missing",
			response: gqlmock.MockQueryResponse("project", map[string]any{
				"id": "", "name": "",
				"environments": map[string]any{"cursor": map[string]any{"next": ""}, "items": []any{}},
			}),
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := newTestClient(t, gqlmock.NewClientWithJSONResponseArray([]any{tt.response}))
			project, err := GetProject(context.Background(), c, tt.id)
			if tt.wantErr {
				if err == nil {
					t.Fatal("expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if project.ID != tt.id {
				t.Errorf("ID: want %q, got %q", tt.id, project.ID)
			}
			if len(project.Environments) != tt.wantEnvs {
				t.Errorf("environments: want %d, got %d", tt.wantEnvs, len(project.Environments))
			}
		})
	}
}

func TestCreateProject(t *testing.T) {
	tests := []struct {
		name        string
		input       CreateProjectInput
		response    any
		wantSuccess bool
		wantMsgLen  int
	}{
		{
			name:  "success",
			input: CreateProjectInput{Id: "newproj", Name: "New Project"},
			response: gqlmock.MockMutationResponse("createProject", map[string]any{
				"id": "newproj", "name": "New Project", "description": "",
				"createdAt": "2024-01-01T00:00:00Z", "updatedAt": "2024-01-01T00:00:00Z",
			}),
			wantSuccess: true,
		},
		{
			name:  "validation failure",
			input: CreateProjectInput{Id: "existing"},
			response: map[string]any{
				"data": map[string]any{
					"createProject": map[string]any{
						"successful": false,
						"result":     map[string]any{"id": "", "name": ""},
						"messages":   []map[string]any{{"field": "id", "message": "ID is already taken", "code": "taken"}},
					},
				},
			},
			wantSuccess: false,
			wantMsgLen:  1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := newTestClient(t, gqlmock.NewClientWithJSONResponseArray([]any{tt.response}))
			payload, err := CreateProject(context.Background(), c, tt.input)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if payload.Successful != tt.wantSuccess {
				t.Errorf("Successful: want %v, got %v", tt.wantSuccess, payload.Successful)
			}
			if len(payload.Messages) != tt.wantMsgLen {
				t.Errorf("Messages: want %d, got %d", tt.wantMsgLen, len(payload.Messages))
			}
		})
	}
}

func TestUpdateProject(t *testing.T) {
	tests := []struct {
		name     string
		id       string
		input    UpdateProjectInput
		response any
		wantName string
	}{
		{
			name:  "updates name and description",
			id:    "proj1",
			input: UpdateProjectInput{Name: "Updated Name", Description: "Updated desc"},
			response: gqlmock.MockMutationResponse("updateProject", map[string]any{
				"id": "proj1", "name": "Updated Name", "description": "Updated desc",
				"createdAt": "2024-01-01T00:00:00Z", "updatedAt": "2024-06-01T00:00:00Z",
			}),
			wantName: "Updated Name",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := newTestClient(t, gqlmock.NewClientWithJSONResponseArray([]any{tt.response}))
			payload, err := UpdateProject(context.Background(), c, tt.id, tt.input)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if !payload.Successful {
				t.Error("expected successful payload")
			}
			if payload.Result == nil || payload.Result.Name != tt.wantName {
				t.Errorf("Result.Name: want %q, got %v", tt.wantName, payload.Result)
			}
		})
	}
}

func TestDeleteProject(t *testing.T) {
	tests := []struct {
		name     string
		id       string
		response any
	}{
		{
			name: "returns successful payload",
			id:   "proj1",
			response: gqlmock.MockMutationResponse("deleteProject", map[string]any{
				"id": "proj1", "name": "Project One",
			}),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := newTestClient(t, gqlmock.NewClientWithJSONResponseArray([]any{tt.response}))
			payload, err := DeleteProject(context.Background(), c, tt.id)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if !payload.Successful {
				t.Error("expected successful payload")
			}
		})
	}
}
