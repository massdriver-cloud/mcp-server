package api

import (
	"context"
	"testing"

	"github.com/massdriver-cloud/mcp-server/internal/gqlmock"
)

func componentResult(id, name string) map[string]any {
	return map[string]any{
		"id": id, "name": name, "description": "",
		"createdAt": "2024-01-01T00:00:00Z", "updatedAt": "2024-01-01T00:00:00Z",
	}
}

func TestAddComponent(t *testing.T) {
	tests := []struct {
		name        string
		projectID   string
		input       AddComponentInput
		response    any
		wantSuccess bool
		wantMsgLen  int
		wantID      string
	}{
		{
			name:      "success",
			projectID: "proj1",
			input:     AddComponentInput{BundleName: "aws-rds-postgres", Id: "db", Name: "Database"},
			response: gqlmock.MockMutationResponse("addComponent", componentResult("db", "Database")),
			wantSuccess: true,
			wantID:      "db",
		},
		{
			name:      "validation failure",
			projectID: "proj1",
			input:     AddComponentInput{BundleName: "aws-rds-postgres", Id: "db"},
			response: map[string]any{
				"data": map[string]any{
					"addComponent": map[string]any{
						"successful": false,
						"result":     map[string]any{"id": "", "name": ""},
						"messages":   []map[string]any{{"field": "id", "message": "already exists", "code": "taken"}},
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
			payload, err := AddComponent(context.Background(), c, tt.projectID, tt.input)
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

func TestRemoveComponent(t *testing.T) {
	tests := []struct {
		name        string
		projectID   string
		id          string
		response    any
		wantSuccess bool
	}{
		{
			name:      "success",
			projectID: "proj1",
			id:        "db",
			response: gqlmock.MockMutationResponse("removeComponent", map[string]any{
				"id": "db", "name": "Database",
			}),
			wantSuccess: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := newTestClient(t, gqlmock.NewClientWithJSONResponseArray([]any{tt.response}))
			payload, err := RemoveComponent(context.Background(), c, tt.projectID, tt.id)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if payload.Successful != tt.wantSuccess {
				t.Errorf("Successful: want %v, got %v", tt.wantSuccess, payload.Successful)
			}
		})
	}
}

func TestLinkComponents(t *testing.T) {
	tests := []struct {
		name        string
		projectID   string
		input       LinkComponentsInput
		response    any
		wantSuccess bool
		wantMsgLen  int
		wantID      string
	}{
		{
			name:      "success",
			projectID: "proj1",
			input: LinkComponentsInput{
				From: "db", FromField: "database", FromVersion: "~1.0",
				To: "app", ToField: "database", ToVersion: "~1.0",
			},
			response: gqlmock.MockMutationResponse("linkComponents", map[string]any{
				"id": "link1", "fromField": "database", "toField": "database",
				"fromComponent": map[string]any{"id": "db", "name": "Database"},
				"toComponent":   map[string]any{"id": "app", "name": "App"},
			}),
			wantSuccess: true,
			wantID:      "link1",
		},
		{
			name:      "validation failure",
			projectID: "proj1",
			input:     LinkComponentsInput{From: "db", FromField: "database", FromVersion: "~1.0", To: "app", ToField: "database", ToVersion: "~1.0"},
			response: map[string]any{
				"data": map[string]any{
					"linkComponents": map[string]any{
						"successful": false,
						"result":     map[string]any{"id": "", "fromField": "", "toField": ""},
						"messages":   []map[string]any{{"field": "from", "message": "already linked", "code": "conflict"}},
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
			payload, err := LinkComponents(context.Background(), c, tt.projectID, tt.input)
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

func TestUnlinkComponents(t *testing.T) {
	tests := []struct {
		name        string
		id          string
		response    any
		wantSuccess bool
	}{
		{
			name: "success",
			id:   "link1",
			response: gqlmock.MockMutationResponse("unlinkComponents", map[string]any{
				"id": "link1",
			}),
			wantSuccess: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := newTestClient(t, gqlmock.NewClientWithJSONResponseArray([]any{tt.response}))
			payload, err := UnlinkComponents(context.Background(), c, tt.id)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if payload.Successful != tt.wantSuccess {
				t.Errorf("Successful: want %v, got %v", tt.wantSuccess, payload.Successful)
			}
		})
	}
}
