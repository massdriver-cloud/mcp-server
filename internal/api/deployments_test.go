package api

import (
	"context"
	"testing"

	"github.com/massdriver-cloud/mcp-server/internal/gqlmock"
)

func deploymentItem(id, status, action string) map[string]any {
	return map[string]any{
		"id": id, "status": status, "action": action,
		"version": "1.0.0", "message": "Deployed via CI",
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
	}
}

func TestListDeployments(t *testing.T) {
	tests := []struct {
		name      string
		filter    DeploymentsFilter
		responses []any
		wantIDs   []string
	}{
		{
			name:   "single page no filter",
			filter: DeploymentsFilter{},
			responses: []any{gqlmock.MockQueryResponse("deployments", map[string]any{
				"cursor": map[string]any{"next": ""},
				"items":  []map[string]any{deploymentItem("dep1", "COMPLETED", "PROVISION")},
			})},
			wantIDs: []string{"dep1"},
		},
		{
			name:   "multi-page pagination",
			filter: DeploymentsFilter{},
			responses: []any{
				gqlmock.MockQueryResponse("deployments", map[string]any{
					"cursor": map[string]any{"next": "cursor-p2"},
					"items":  []map[string]any{deploymentItem("dep1", "COMPLETED", "PROVISION")},
				}),
				gqlmock.MockQueryResponse("deployments", map[string]any{
					"cursor": map[string]any{"next": ""},
					"items":  []map[string]any{deploymentItem("dep2", "FAILED", "PROVISION")},
				}),
			},
			wantIDs: []string{"dep1", "dep2"},
		},
		{
			name:   "filter by instance ID",
			filter: DeploymentsFilter{InstanceId: IdFilter{Eq: "proj1-staging-db"}},
			responses: []any{gqlmock.MockQueryResponse("deployments", map[string]any{
				"cursor": map[string]any{"next": ""},
				"items":  []map[string]any{deploymentItem("dep1", "COMPLETED", "PROVISION")},
			})},
			wantIDs: []string{"dep1"},
		},
		{
			name:   "filter by status",
			filter: DeploymentsFilter{Status: DeploymentStatusFilter{Eq: DeploymentStatusRunning}},
			responses: []any{gqlmock.MockQueryResponse("deployments", map[string]any{
				"cursor": map[string]any{"next": ""},
				"items":  []map[string]any{deploymentItem("dep3", "RUNNING", "PROVISION")},
			})},
			wantIDs: []string{"dep3"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := newTestClient(t, gqlmock.NewClientWithJSONResponseArray(tt.responses))
			deployments, err := ListDeployments(context.Background(), c, tt.filter)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if len(deployments) != len(tt.wantIDs) {
				t.Fatalf("expected %d deployments, got %d", len(tt.wantIDs), len(deployments))
			}
			for i, id := range tt.wantIDs {
				if deployments[i].ID != id {
					t.Errorf("deployments[%d].ID: want %q, got %q", i, id, deployments[i].ID)
				}
			}
		})
	}
}

func TestGetDeployment(t *testing.T) {
	tests := []struct {
		name         string
		id           string
		response     any
		wantErr      bool
		wantStatus   string
		wantInstance string
	}{
		{
			name:     "returns deployment with instance",
			id:       "dep1",
			response: gqlmock.MockQueryResponse("deployment", deploymentItem("dep1", "COMPLETED", "PROVISION")),
			wantStatus:   "COMPLETED",
			wantInstance: "proj1-staging-db",
		},
		{
			name: "returns error when not found",
			id:   "missing",
			response: gqlmock.MockQueryResponse("deployment", map[string]any{
				"id": "", "status": "", "action": "", "version": "", "message": "",
				"createdAt": "0001-01-01T00:00:00Z", "updatedAt": "0001-01-01T00:00:00Z",
				"lastTransitionedAt": "0001-01-01T00:00:00Z",
				"elapsedTime": 0, "deployedBy": "",
				"instance": map[string]any{
					"id": "", "name": "",
					"environment": map[string]any{
						"id": "", "name": "",
						"project": map[string]any{"id": "", "name": ""},
					},
				},
			}),
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := newTestClient(t, gqlmock.NewClientWithJSONResponseArray([]any{tt.response}))
			deployment, err := GetDeployment(context.Background(), c, tt.id)
			if tt.wantErr {
				if err == nil {
					t.Fatal("expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if deployment.ID != tt.id {
				t.Errorf("ID: want %q, got %q", tt.id, deployment.ID)
			}
			if deployment.Status != tt.wantStatus {
				t.Errorf("Status: want %q, got %q", tt.wantStatus, deployment.Status)
			}
			if deployment.Instance == nil || deployment.Instance.ID != tt.wantInstance {
				t.Errorf("Instance.ID: want %q, got %v", tt.wantInstance, deployment.Instance)
			}
		})
	}
}
