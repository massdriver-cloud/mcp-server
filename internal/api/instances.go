package api

import (
	"context"
	"fmt"
	"time"

	"github.com/massdriver-cloud/massdriver-sdk-go/massdriver/client"
	"github.com/massdriver-cloud/mcp-server/internal/api/scalars"
)

// Instance is the domain representation of a Massdriver instance.
type Instance struct {
	ID              string              `json:"id"`
	Name            string              `json:"name"`
	Status          string              `json:"status"`
	Version         string              `json:"version"`
	ReleaseStrategy string              `json:"releaseStrategy"`
	CreatedAt       time.Time           `json:"createdAt"`
	UpdatedAt       time.Time           `json:"updatedAt"`
	Environment     *InstanceEnv        `json:"environment,omitempty"`
	Release         *InstanceRelease    `json:"release,omitempty"`
}

// InstanceEnv is a lightweight environment reference embedded in an Instance.
type InstanceEnv struct {
	ID      string           `json:"id"`
	Name    string           `json:"name"`
	Project *InstanceProject `json:"project,omitempty"`
}

// InstanceProject is a lightweight project reference embedded in an InstanceEnv.
type InstanceProject struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

// InstanceRelease is the bundle release currently deployed for an instance.
type InstanceRelease struct {
	ID      string `json:"id"`
	Name    string `json:"name"`
	Version string `json:"version"`
}

// ListInstances returns every instance in the organization, following pagination
// cursors until all pages have been collected.
// Pass a non-zero filter to scope results (e.g., by project or environment ID).
func ListInstances(ctx context.Context, c *client.Client, filter InstancesFilter) ([]Instance, error) {
	var all []Instance
	cursor := scalars.Cursor{}

	for {
		resp, err := listInstances(ctx, c.GQL, c.Config.OrganizationID, filter, cursor)
		if err != nil {
			return nil, fmt.Errorf("list instances: %w", err)
		}

		for _, item := range resp.Instances.Items {
			inst := Instance{
				ID:              item.Id,
				Name:            item.Name,
				Status:          string(item.Status),
				Version:         item.Version,
				ReleaseStrategy: string(item.ReleaseStrategy),
				CreatedAt:       item.CreatedAt,
				UpdatedAt:       item.UpdatedAt,
			}
			if item.Environment.Id != "" {
				env := &InstanceEnv{
					ID:   item.Environment.Id,
					Name: item.Environment.Name,
				}
				if item.Environment.Project.Id != "" {
					env.Project = &InstanceProject{
						ID:   item.Environment.Project.Id,
						Name: item.Environment.Project.Name,
					}
				}
				inst.Environment = env
			}
			if item.Release.Id != "" {
				inst.Release = &InstanceRelease{
					ID:      item.Release.Id,
					Name:    item.Release.Name,
					Version: item.Release.Version,
				}
			}
			all = append(all, inst)
		}

		if resp.Instances.Cursor.Next == "" {
			break
		}
		cursor = scalars.Cursor{Next: resp.Instances.Cursor.Next}
	}

	return all, nil
}

// GetInstance returns a single instance by ID.
// Returns an error if the instance does not exist.
func GetInstance(ctx context.Context, c *client.Client, id string) (*Instance, error) {
	resp, err := getInstance(ctx, c.GQL, c.Config.OrganizationID, id)
	if err != nil {
		return nil, fmt.Errorf("get instance %q: %w", id, err)
	}

	i := resp.Instance
	if i.Id == "" {
		return nil, fmt.Errorf("instance %q not found", id)
	}

	inst := &Instance{
		ID:              i.Id,
		Name:            i.Name,
		Status:          string(i.Status),
		Version:         i.Version,
		ReleaseStrategy: string(i.ReleaseStrategy),
		CreatedAt:       i.CreatedAt,
		UpdatedAt:       i.UpdatedAt,
	}
	if i.Environment.Id != "" {
		env := &InstanceEnv{
			ID:   i.Environment.Id,
			Name: i.Environment.Name,
		}
		if i.Environment.Project.Id != "" {
			env.Project = &InstanceProject{
				ID:   i.Environment.Project.Id,
				Name: i.Environment.Project.Name,
			}
		}
		inst.Environment = env
	}
	if i.Release.Id != "" {
		inst.Release = &InstanceRelease{
			ID:      i.Release.Id,
			Name:    i.Release.Name,
			Version: i.Release.Version,
		}
	}

	return inst, nil
}
