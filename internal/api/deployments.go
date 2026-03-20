package api

import (
	"context"
	"fmt"
	"time"

	"github.com/massdriver-cloud/massdriver-sdk-go/massdriver/client"
	"github.com/massdriver-cloud/mcp-server/internal/api/scalars"
)

// Deployment is the domain representation of a Massdriver deployment.
type Deployment struct {
	ID                 string     `json:"id"`
	Status             string     `json:"status"`
	Action             string     `json:"action"`
	Version            string     `json:"version"`
	Message            string     `json:"message"`
	CreatedAt          time.Time  `json:"createdAt"`
	UpdatedAt          time.Time  `json:"updatedAt"`
	LastTransitionedAt *time.Time `json:"lastTransitionedAt,omitempty"`
	ElapsedTime        int        `json:"elapsedTime"`
	DeployedBy         string     `json:"deployedBy"`
	Instance           *DeploymentInstance `json:"instance,omitempty"`
}

// DeploymentInstance is a lightweight instance reference embedded in a Deployment.
type DeploymentInstance struct {
	ID          string              `json:"id"`
	Name        string              `json:"name"`
	Environment *DeploymentEnv      `json:"environment,omitempty"`
}

// DeploymentEnv is a lightweight environment reference embedded in a DeploymentInstance.
type DeploymentEnv struct {
	ID      string             `json:"id"`
	Name    string             `json:"name"`
	Project *DeploymentProject `json:"project,omitempty"`
}

// DeploymentProject is a lightweight project reference embedded in a DeploymentEnv.
type DeploymentProject struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

// ListDeployments returns every deployment in the organization, following pagination
// cursors until all pages have been collected.
// Pass a non-zero filter to scope results (e.g., by instance ID or status).
func ListDeployments(ctx context.Context, c *client.Client, filter DeploymentsFilter) ([]Deployment, error) {
	var all []Deployment
	cursor := scalars.Cursor{}

	for {
		resp, err := listDeployments(ctx, c.GQL, c.Config.OrganizationID, filter, cursor)
		if err != nil {
			return nil, fmt.Errorf("list deployments: %w", err)
		}

		for _, item := range resp.Deployments.Items {
			d := Deployment{
				ID:          item.Id,
				Status:      string(item.Status),
				Action:      string(item.Action),
				Version:     item.Version,
				Message:     item.Message,
				CreatedAt:   item.CreatedAt,
				UpdatedAt:   item.UpdatedAt,
				ElapsedTime: item.ElapsedTime,
				DeployedBy:  item.DeployedBy,
			}
			if !item.LastTransitionedAt.IsZero() {
				t := item.LastTransitionedAt
				d.LastTransitionedAt = &t
			}
			if item.Instance.Id != "" {
				d.Instance = deploymentInstanceFromLeaf(
					item.Instance.Id,
					item.Instance.Name,
					item.Instance.Environment.Id,
					item.Instance.Environment.Name,
					item.Instance.Environment.Project.Id,
					item.Instance.Environment.Project.Name,
				)
			}
			all = append(all, d)
		}

		if resp.Deployments.Cursor.Next == "" {
			break
		}
		cursor = scalars.Cursor{Next: resp.Deployments.Cursor.Next}
	}

	return all, nil
}

// GetDeployment returns a single deployment by ID.
// Returns an error if the deployment does not exist.
func GetDeployment(ctx context.Context, c *client.Client, id string) (*Deployment, error) {
	resp, err := getDeployment(ctx, c.GQL, c.Config.OrganizationID, id)
	if err != nil {
		return nil, fmt.Errorf("get deployment %q: %w", id, err)
	}

	dep := resp.Deployment
	if dep.Id == "" {
		return nil, fmt.Errorf("deployment %q not found", id)
	}

	d := &Deployment{
		ID:          dep.Id,
		Status:      string(dep.Status),
		Action:      string(dep.Action),
		Version:     dep.Version,
		Message:     dep.Message,
		CreatedAt:   dep.CreatedAt,
		UpdatedAt:   dep.UpdatedAt,
		ElapsedTime: dep.ElapsedTime,
		DeployedBy:  dep.DeployedBy,
	}
	if !dep.LastTransitionedAt.IsZero() {
		t := dep.LastTransitionedAt
		d.LastTransitionedAt = &t
	}
	if dep.Instance.Id != "" {
		d.Instance = deploymentInstanceFromLeaf(
			dep.Instance.Id,
			dep.Instance.Name,
			dep.Instance.Environment.Id,
			dep.Instance.Environment.Name,
			dep.Instance.Environment.Project.Id,
			dep.Instance.Environment.Project.Name,
		)
	}

	return d, nil
}

func deploymentInstanceFromLeaf(instanceID, instanceName, envID, envName, projID, projName string) *DeploymentInstance {
	inst := &DeploymentInstance{ID: instanceID, Name: instanceName}
	if envID != "" {
		env := &DeploymentEnv{ID: envID, Name: envName}
		if projID != "" {
			env.Project = &DeploymentProject{ID: projID, Name: projName}
		}
		inst.Environment = env
	}
	return inst
}
