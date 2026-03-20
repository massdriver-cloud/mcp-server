package api

import (
	"context"
	"testing"

	"github.com/massdriver-cloud/mcp-server/internal/gqlmock"
)

func instanceItem(id string) map[string]any {
	return map[string]any{
		"id": id, "name": id, "status": "PROVISIONED",
		"version": "~1.0", "releaseStrategy": "STABLE",
		"createdAt": "2024-01-01T00:00:00Z", "updatedAt": "2024-01-01T00:00:00Z",
		"environment": map[string]any{
			"id": "proj1-staging", "name": "Staging",
			"project": map[string]any{"id": "proj1", "name": "Project One"},
		},
		"release": map[string]any{"id": "rel1", "name": "aws-rds", "version": "1.2.3"},
	}
}

func TestListInstances(t *testing.T) {
	tests := []struct {
		name      string
		filter    InstancesFilter
		responses []any
		wantIDs   []string
	}{
		{
			name:   "single page no filter",
			filter: InstancesFilter{},
			responses: []any{gqlmock.MockQueryResponse("instances", map[string]any{
				"cursor": map[string]any{"next": ""},
				"items":  []map[string]any{instanceItem("proj1-staging-db")},
			})},
			wantIDs: []string{"proj1-staging-db"},
		},
		{
			name:   "multi-page pagination",
			filter: InstancesFilter{},
			responses: []any{
				gqlmock.MockQueryResponse("instances", map[string]any{
					"cursor": map[string]any{"next": "cursor-p2"},
					"items":  []map[string]any{instanceItem("inst1")},
				}),
				gqlmock.MockQueryResponse("instances", map[string]any{
					"cursor": map[string]any{"next": ""},
					"items":  []map[string]any{instanceItem("inst2")},
				}),
			},
			wantIDs: []string{"inst1", "inst2"},
		},
		{
			name:   "filter by environment ID",
			filter: InstancesFilter{EnvironmentId: IdFilter{Eq: "proj1-staging"}},
			responses: []any{gqlmock.MockQueryResponse("instances", map[string]any{
				"cursor": map[string]any{"next": ""},
				"items":  []map[string]any{instanceItem("proj1-staging-db")},
			})},
			wantIDs: []string{"proj1-staging-db"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := newTestClient(t, gqlmock.NewClientWithJSONResponseArray(tt.responses))
			instances, err := ListInstances(context.Background(), c, tt.filter)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if len(instances) != len(tt.wantIDs) {
				t.Fatalf("expected %d instances, got %d", len(tt.wantIDs), len(instances))
			}
			for i, id := range tt.wantIDs {
				if instances[i].ID != id {
					t.Errorf("instances[%d].ID: want %q, got %q", i, id, instances[i].ID)
				}
			}
		})
	}
}

func TestGetInstance(t *testing.T) {
	tests := []struct {
		name            string
		id              string
		response        any
		wantErr         bool
		wantStatus      string
		wantEnvironment string
	}{
		{
			name: "returns instance with environment and release",
			id:   "proj1-staging-db",
			response: gqlmock.MockQueryResponse("instance", map[string]any{
				"id": "proj1-staging-db", "name": "Database", "status": "PROVISIONED",
				"version": "~1.0", "releaseStrategy": "STABLE",
				"createdAt": "2024-01-01T00:00:00Z", "updatedAt": "2024-01-01T00:00:00Z",
				"environment": map[string]any{
					"id": "proj1-staging", "name": "Staging",
					"project": map[string]any{"id": "proj1", "name": "Project One"},
				},
				"release": map[string]any{"id": "rel1", "name": "aws-rds", "version": "1.2.3"},
			}),
			wantStatus:      "PROVISIONED",
			wantEnvironment: "proj1-staging",
		},
		{
			name: "returns error when not found",
			id:   "missing",
			response: gqlmock.MockQueryResponse("instance", map[string]any{
				"id": "", "name": "", "status": "", "version": "", "releaseStrategy": "",
				"createdAt": "0001-01-01T00:00:00Z", "updatedAt": "0001-01-01T00:00:00Z",
				"environment": map[string]any{
					"id": "", "name": "",
					"project": map[string]any{"id": "", "name": ""},
				},
				"release": map[string]any{"id": "", "name": "", "version": ""},
			}),
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := newTestClient(t, gqlmock.NewClientWithJSONResponseArray([]any{tt.response}))
			instance, err := GetInstance(context.Background(), c, tt.id)
			if tt.wantErr {
				if err == nil {
					t.Fatal("expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if instance.ID != tt.id {
				t.Errorf("ID: want %q, got %q", tt.id, instance.ID)
			}
			if instance.Status != tt.wantStatus {
				t.Errorf("Status: want %q, got %q", tt.wantStatus, instance.Status)
			}
			if instance.Environment == nil || instance.Environment.ID != tt.wantEnvironment {
				t.Errorf("Environment.ID: want %q, got %v", tt.wantEnvironment, instance.Environment)
			}
		})
	}
}
