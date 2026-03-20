package api

import (
	"context"
	"testing"

	"github.com/massdriver-cloud/mcp-server/internal/gqlmock"
)

func envItem(id string) map[string]any {
	return map[string]any{
		"id": id, "name": id, "description": "",
		"createdAt": "2024-01-01T00:00:00Z", "updatedAt": "2024-01-01T00:00:00Z",
		"project": map[string]any{
			"id": "proj1", "name": "Project One", "description": "",
			"createdAt": "2024-01-01T00:00:00Z", "updatedAt": "2024-01-01T00:00:00Z",
		},
	}
}

func TestListEnvironments(t *testing.T) {
	tests := []struct {
		name      string
		filter    EnvironmentsFilter
		responses []any
		wantIDs   []string
	}{
		{
			name:   "single page with project",
			filter: EnvironmentsFilter{},
			responses: []any{gqlmock.MockQueryResponse("environments", map[string]any{
				"cursor": map[string]any{"next": ""},
				"items":  []map[string]any{envItem("proj1-staging")},
			})},
			wantIDs: []string{"proj1-staging"},
		},
		{
			name:   "multi-page pagination",
			filter: EnvironmentsFilter{},
			responses: []any{
				gqlmock.MockQueryResponse("environments", map[string]any{
					"cursor": map[string]any{"next": "cursor-p2"},
					"items":  []map[string]any{envItem("env1")},
				}),
				gqlmock.MockQueryResponse("environments", map[string]any{
					"cursor": map[string]any{"next": ""},
					"items":  []map[string]any{envItem("env2")},
				}),
			},
			wantIDs: []string{"env1", "env2"},
		},
		{
			name:   "filter by project ID",
			filter: EnvironmentsFilter{ProjectId: IdFilter{Eq: "proj1"}},
			responses: []any{gqlmock.MockQueryResponse("environments", map[string]any{
				"cursor": map[string]any{"next": ""},
				"items":  []map[string]any{envItem("proj1-prod")},
			})},
			wantIDs: []string{"proj1-prod"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := newTestClient(t, gqlmock.NewClientWithJSONResponseArray(tt.responses))
			envs, err := ListEnvironments(context.Background(), c, tt.filter)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if len(envs) != len(tt.wantIDs) {
				t.Fatalf("expected %d environments, got %d", len(tt.wantIDs), len(envs))
			}
			for i, id := range tt.wantIDs {
				if envs[i].ID != id {
					t.Errorf("envs[%d].ID: want %q, got %q", i, id, envs[i].ID)
				}
			}
		})
	}
}

func TestGetEnvironment(t *testing.T) {
	tests := []struct {
		name        string
		id          string
		response    any
		wantErr     bool
		wantProject string
	}{
		{
			name: "returns environment with project",
			id:   "proj1-staging",
			response: gqlmock.MockQueryResponse("environment", map[string]any{
				"id": "proj1-staging", "name": "Staging", "description": "Staging env",
				"createdAt": "2024-01-01T00:00:00Z", "updatedAt": "2024-01-01T00:00:00Z",
				"project": map[string]any{
					"id": "proj1", "name": "Project One", "description": "",
					"createdAt": "2024-01-01T00:00:00Z", "updatedAt": "2024-01-01T00:00:00Z",
				},
			}),
			wantProject: "proj1",
		},
		{
			name: "returns error when not found",
			id:   "missing",
			response: gqlmock.MockQueryResponse("environment", map[string]any{
				"id": "", "name": "",
				"project": map[string]any{
					"id": "", "name": "", "description": "",
					"createdAt": "0001-01-01T00:00:00Z", "updatedAt": "0001-01-01T00:00:00Z",
				},
			}),
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := newTestClient(t, gqlmock.NewClientWithJSONResponseArray([]any{tt.response}))
			env, err := GetEnvironment(context.Background(), c, tt.id)
			if tt.wantErr {
				if err == nil {
					t.Fatal("expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if env.ID != tt.id {
				t.Errorf("ID: want %q, got %q", tt.id, env.ID)
			}
			if env.Project == nil || env.Project.ID != tt.wantProject {
				t.Errorf("Project.ID: want %q, got %v", tt.wantProject, env.Project)
			}
		})
	}
}

func TestCreateEnvironment(t *testing.T) {
	tests := []struct {
		name        string
		projectID   string
		input       CreateEnvironmentInput
		response    any
		wantSuccess bool
		wantMsgLen  int
		wantID      string
	}{
		{
			name:      "success",
			projectID: "proj1",
			input:     CreateEnvironmentInput{Id: "staging", Name: "Staging"},
			response: gqlmock.MockMutationResponse("createEnvironment", map[string]any{
				"id": "proj1-staging", "name": "Staging", "description": "",
				"createdAt": "2024-01-01T00:00:00Z", "updatedAt": "2024-01-01T00:00:00Z",
			}),
			wantSuccess: true,
			wantID:      "proj1-staging",
		},
		{
			name:      "validation failure",
			projectID: "proj1",
			input:     CreateEnvironmentInput{Id: "existing"},
			response: map[string]any{
				"data": map[string]any{
					"createEnvironment": map[string]any{
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
			payload, err := CreateEnvironment(context.Background(), c, tt.projectID, tt.input)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if payload.Successful != tt.wantSuccess {
				t.Errorf("Successful: want %v, got %v", tt.wantSuccess, payload.Successful)
			}
			if len(payload.Messages) != tt.wantMsgLen {
				t.Errorf("Messages: want %d, got %d", tt.wantMsgLen, len(payload.Messages))
			}
			if tt.wantID != "" && (payload.Result == nil || payload.Result.ID != tt.wantID) {
				t.Errorf("Result.ID: want %q, got %v", tt.wantID, payload.Result)
			}
		})
	}
}

func TestUpdateEnvironment(t *testing.T) {
	tests := []struct {
		name     string
		id       string
		input    UpdateEnvironmentInput
		response any
		wantName string
	}{
		{
			name:  "updates name and description",
			id:    "proj1-staging",
			input: UpdateEnvironmentInput{Name: "Staging Updated", Description: "Updated desc"},
			response: gqlmock.MockMutationResponse("updateEnvironment", map[string]any{
				"id": "proj1-staging", "name": "Staging Updated", "description": "Updated desc",
				"createdAt": "2024-01-01T00:00:00Z", "updatedAt": "2024-06-01T00:00:00Z",
			}),
			wantName: "Staging Updated",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := newTestClient(t, gqlmock.NewClientWithJSONResponseArray([]any{tt.response}))
			payload, err := UpdateEnvironment(context.Background(), c, tt.id, tt.input)
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

func TestDeleteEnvironment(t *testing.T) {
	tests := []struct {
		name     string
		id       string
		response any
	}{
		{
			name: "returns successful payload",
			id:   "proj1-staging",
			response: gqlmock.MockMutationResponse("deleteEnvironment", map[string]any{
				"id": "proj1-staging", "name": "Staging",
			}),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := newTestClient(t, gqlmock.NewClientWithJSONResponseArray([]any{tt.response}))
			payload, err := DeleteEnvironment(context.Background(), c, tt.id)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if !payload.Successful {
				t.Error("expected successful payload")
			}
		})
	}
}
