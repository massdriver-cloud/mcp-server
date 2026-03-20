package api

import (
	"context"
	"fmt"
	"time"

	"github.com/massdriver-cloud/massdriver-sdk-go/massdriver/client"
	"github.com/massdriver-cloud/mcp-server/internal/api/scalars"
)

// Environment is the domain representation of a Massdriver environment.
type Environment struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	CreatedAt   time.Time `json:"createdAt"`
	UpdatedAt   time.Time `json:"updatedAt"`
	Project     *Project  `json:"project,omitempty"`
}

// EnvironmentPayload is the result of an environment create, update, or delete mutation.
type EnvironmentPayload struct {
	Successful bool                `json:"successful"`
	Messages   []ValidationMessage `json:"messages"`
	Result     *Environment        `json:"result,omitempty"`
}

// ListEnvironments returns every environment in the organization, following pagination
// cursors until all pages have been collected.
// Pass a non-zero filter to scope results (e.g., by project IDs).
func ListEnvironments(ctx context.Context, c *client.Client, filter EnvironmentsFilter) ([]Environment, error) {
	var all []Environment
	cursor := scalars.Cursor{}

	for {
		resp, err := listEnvironments(ctx, c.GQL, c.Config.OrganizationID, filter, cursor)
		if err != nil {
			return nil, fmt.Errorf("list environments: %w", err)
		}

		for _, item := range resp.Environments.Items {
			proj := Project{
				ID:          item.Project.Id,
				Name:        item.Project.Name,
				Description: item.Project.Description,
				CreatedAt:   item.Project.CreatedAt,
				UpdatedAt:   item.Project.UpdatedAt,
			}
			all = append(all, Environment{
				ID:          item.Id,
				Name:        item.Name,
				Description: item.Description,
				CreatedAt:   item.CreatedAt,
				UpdatedAt:   item.UpdatedAt,
				Project:     &proj,
			})
		}

		if resp.Environments.Cursor.Next == "" {
			break
		}
		cursor = scalars.Cursor{Next: resp.Environments.Cursor.Next}
	}

	return all, nil
}

// GetEnvironment returns a single environment by ID, including its project.
// Returns an error if the environment does not exist.
func GetEnvironment(ctx context.Context, c *client.Client, id string) (*Environment, error) {
	resp, err := getEnvironment(ctx, c.GQL, c.Config.OrganizationID, id)
	if err != nil {
		return nil, fmt.Errorf("get environment %q: %w", id, err)
	}

	e := resp.Environment
	if e.Id == "" {
		return nil, fmt.Errorf("environment %q not found", id)
	}

	proj := Project{
		ID:          e.Project.Id,
		Name:        e.Project.Name,
		Description: e.Project.Description,
		CreatedAt:   e.Project.CreatedAt,
		UpdatedAt:   e.Project.UpdatedAt,
	}

	return &Environment{
		ID:          e.Id,
		Name:        e.Name,
		Description: e.Description,
		CreatedAt:   e.CreatedAt,
		UpdatedAt:   e.UpdatedAt,
		Project:     &proj,
	}, nil
}

// CreateEnvironment creates a new environment within a project and returns the mutation payload.
func CreateEnvironment(ctx context.Context, c *client.Client, projectID string, input CreateEnvironmentInput) (*EnvironmentPayload, error) {
	resp, err := createEnvironment(ctx, c.GQL, c.Config.OrganizationID, projectID, input)
	if err != nil {
		return nil, fmt.Errorf("create environment: %w", err)
	}

	payload := resp.CreateEnvironment
	return &EnvironmentPayload{
		Successful: payload.Successful,
		Messages:   toValidationMessages(payload.Messages),
		Result:     environmentFromCreateResult(payload.Result),
	}, nil
}

// UpdateEnvironment updates an environment's name or description and returns the mutation payload.
func UpdateEnvironment(ctx context.Context, c *client.Client, id string, input UpdateEnvironmentInput) (*EnvironmentPayload, error) {
	resp, err := updateEnvironment(ctx, c.GQL, c.Config.OrganizationID, id, input)
	if err != nil {
		return nil, fmt.Errorf("update environment %q: %w", id, err)
	}

	payload := resp.UpdateEnvironment
	return &EnvironmentPayload{
		Successful: payload.Successful,
		Messages:   toValidationMessages(payload.Messages),
		Result:     environmentFromUpdateResult(payload.Result),
	}, nil
}

// DeleteEnvironment deletes an environment and returns the mutation payload.
// All instances must be decommissioned before an environment can be deleted.
func DeleteEnvironment(ctx context.Context, c *client.Client, id string) (*EnvironmentPayload, error) {
	resp, err := deleteEnvironment(ctx, c.GQL, c.Config.OrganizationID, id)
	if err != nil {
		return nil, fmt.Errorf("delete environment %q: %w", id, err)
	}

	payload := resp.DeleteEnvironment
	return &EnvironmentPayload{
		Successful: payload.Successful,
		Messages:   toValidationMessages(payload.Messages),
	}, nil
}

// environmentFromCreateResult converts the generated create-result type to Environment.
func environmentFromCreateResult(r createEnvironmentCreateEnvironmentEnvironmentPayloadResultEnvironment) *Environment {
	if r.Id == "" {
		return nil
	}
	return &Environment{
		ID:          r.Id,
		Name:        r.Name,
		Description: r.Description,
		CreatedAt:   r.CreatedAt,
		UpdatedAt:   r.UpdatedAt,
	}
}

// environmentFromUpdateResult converts the generated update-result type to Environment.
func environmentFromUpdateResult(r updateEnvironmentUpdateEnvironmentEnvironmentPayloadResultEnvironment) *Environment {
	if r.Id == "" {
		return nil
	}
	return &Environment{
		ID:          r.Id,
		Name:        r.Name,
		Description: r.Description,
		CreatedAt:   r.CreatedAt,
		UpdatedAt:   r.UpdatedAt,
	}
}
