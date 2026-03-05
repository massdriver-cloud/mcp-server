package api

import (
	"context"
	"fmt"
	"time"

	"github.com/massdriver-cloud/massdriver-sdk-go/massdriver/client"
	"github.com/massdriver-cloud/mcp-server/internal/api/scalars"
)

// Project is the domain representation of a Massdriver project.
type Project struct {
	ID           string        `json:"id"`
	Name         string        `json:"name"`
	Description  string        `json:"description"`
	CreatedAt    time.Time     `json:"createdAt"`
	UpdatedAt    time.Time     `json:"updatedAt"`
	Environments []Environment `json:"environments,omitempty"`
}

// ProjectPayload is the result of a project create, update, or delete mutation.
type ProjectPayload struct {
	Successful bool                `json:"successful"`
	Messages   []ValidationMessage `json:"messages"`
	Result     *Project            `json:"result,omitempty"`
}

// ValidationMessage is a validation error returned by a mutation.
type ValidationMessage struct {
	Field   string `json:"field"`
	Message string `json:"message"`
	Code    string `json:"code"`
}

// ListProjects returns every project in the organization, following pagination
// cursors until all pages have been collected.
func ListProjects(ctx context.Context, c *client.Client) ([]Project, error) {
	var all []Project
	cursor := scalars.Cursor{}

	for {
		resp, err := listProjects(ctx, c.GQL, c.Config.OrganizationID, cursor)
		if err != nil {
			return nil, fmt.Errorf("list projects: %w", err)
		}

		for _, item := range resp.Projects.Items {
			all = append(all, Project{
				ID:          item.Id,
				Name:        item.Name,
				Description: item.Description,
				CreatedAt:   item.CreatedAt,
				UpdatedAt:   item.UpdatedAt,
			})
		}

		if resp.Projects.Cursor.Next == "" {
			break
		}
		cursor = scalars.Cursor{Next: resp.Projects.Cursor.Next}
	}

	return all, nil
}

// GetProject returns a single project by ID, including its environments.
// Returns an error if the project does not exist.
func GetProject(ctx context.Context, c *client.Client, id string) (*Project, error) {
	resp, err := getProject(ctx, c.GQL, c.Config.OrganizationID, id)
	if err != nil {
		return nil, fmt.Errorf("get project %q: %w", id, err)
	}

	p := resp.Project
	if p.Id == "" {
		return nil, fmt.Errorf("project %q not found", id)
	}

	var envs []Environment
	for _, item := range p.Environments.Items {
		envs = append(envs, Environment{
			ID:          item.Id,
			Name:        item.Name,
			Description: item.Description,
			CreatedAt:   item.CreatedAt,
			UpdatedAt:   item.UpdatedAt,
		})
	}

	return &Project{
		ID:           p.Id,
		Name:         p.Name,
		Description:  p.Description,
		CreatedAt:    p.CreatedAt,
		UpdatedAt:    p.UpdatedAt,
		Environments: envs,
	}, nil
}

// CreateProject creates a new project and returns the mutation payload.
func CreateProject(ctx context.Context, c *client.Client, input CreateProjectInput) (*ProjectPayload, error) {
	resp, err := createProject(ctx, c.GQL, c.Config.OrganizationID, input)
	if err != nil {
		return nil, fmt.Errorf("create project: %w", err)
	}

	payload := resp.CreateProject
	return &ProjectPayload{
		Successful: payload.Successful,
		Messages:   toValidationMessages(payload.Messages),
		Result:     projectFromCreateResult(payload.Result),
	}, nil
}

// UpdateProject updates a project's name or description and returns the mutation payload.
func UpdateProject(ctx context.Context, c *client.Client, id string, input UpdateProjectInput) (*ProjectPayload, error) {
	resp, err := updateProject(ctx, c.GQL, c.Config.OrganizationID, id, input)
	if err != nil {
		return nil, fmt.Errorf("update project %q: %w", id, err)
	}

	payload := resp.UpdateProject
	return &ProjectPayload{
		Successful: payload.Successful,
		Messages:   toValidationMessages(payload.Messages),
		Result:     projectFromUpdateResult(payload.Result),
	}, nil
}

// DeleteProject deletes a project and returns the mutation payload.
// All environments must be empty before a project can be deleted.
func DeleteProject(ctx context.Context, c *client.Client, id string) (*ProjectPayload, error) {
	resp, err := deleteProject(ctx, c.GQL, c.Config.OrganizationID, id)
	if err != nil {
		return nil, fmt.Errorf("delete project %q: %w", id, err)
	}

	payload := resp.DeleteProject
	return &ProjectPayload{
		Successful: payload.Successful,
		Messages:   toValidationMessages(payload.Messages),
	}, nil
}

// projectFromCreateResult converts the generated create-result type to Project.
func projectFromCreateResult(r createProjectCreateProjectProjectPayloadResultProject) *Project {
	if r.Id == "" {
		return nil
	}
	return &Project{
		ID:          r.Id,
		Name:        r.Name,
		Description: r.Description,
		CreatedAt:   r.CreatedAt,
		UpdatedAt:   r.UpdatedAt,
	}
}

// projectFromUpdateResult converts the generated update-result type to Project.
func projectFromUpdateResult(r updateProjectUpdateProjectProjectPayloadResultProject) *Project {
	if r.Id == "" {
		return nil
	}
	return &Project{
		ID:          r.Id,
		Name:        r.Name,
		Description: r.Description,
		CreatedAt:   r.CreatedAt,
		UpdatedAt:   r.UpdatedAt,
	}
}

// toValidationMessages converts the per-operation validation message slices to the
// shared ValidationMessage type. The two-parameter generic handles all the
// differently-named generated message types whose methods have pointer receivers.
func toValidationMessages[T any, PT interface {
	*T
	GetField() string
	GetMessage() string
	GetCode() string
}](msgs []T) []ValidationMessage {
	result := make([]ValidationMessage, len(msgs))
	for i := range msgs {
		m := PT(&msgs[i])
		result[i] = ValidationMessage{
			Field:   m.GetField(),
			Message: m.GetMessage(),
			Code:    m.GetCode(),
		}
	}
	return result
}
