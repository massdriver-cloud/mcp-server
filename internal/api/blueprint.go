package api

import (
	"context"
	"fmt"
	"time"

	"github.com/massdriver-cloud/massdriver-sdk-go/massdriver/client"
)

// Component is the domain representation of a blueprint component.
type Component struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	CreatedAt   time.Time `json:"createdAt"`
	UpdatedAt   time.Time `json:"updatedAt"`
}

// ComponentPayload is the result of an addComponent or removeComponent mutation.
type ComponentPayload struct {
	Successful bool                `json:"successful"`
	Messages   []ValidationMessage `json:"messages"`
	Result     *Component          `json:"result,omitempty"`
}

// Link is the domain representation of a blueprint link between components.
type Link struct {
	ID            string     `json:"id"`
	FromField     string     `json:"fromField"`
	ToField       string     `json:"toField"`
	FromComponent *LinkEndpoint `json:"fromComponent,omitempty"`
	ToComponent   *LinkEndpoint `json:"toComponent,omitempty"`
}

// LinkEndpoint is a lightweight component reference embedded in a Link.
type LinkEndpoint struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

// LinkPayload is the result of a linkComponents or unlinkComponents mutation.
type LinkPayload struct {
	Successful bool                `json:"successful"`
	Messages   []ValidationMessage `json:"messages"`
	Result     *Link               `json:"result,omitempty"`
}

// AddComponent adds a component to a project's blueprint and returns the mutation payload.
func AddComponent(ctx context.Context, c *client.Client, projectID string, input AddComponentInput) (*ComponentPayload, error) {
	resp, err := addComponent(ctx, c.GQL, c.Config.OrganizationID, projectID, input)
	if err != nil {
		return nil, fmt.Errorf("add component: %w", err)
	}

	payload := resp.AddComponent
	return &ComponentPayload{
		Successful: payload.Successful,
		Messages:   toValidationMessages(payload.Messages),
		Result:     componentFromAddResult(payload.Result),
	}, nil
}

// RemoveComponent removes a component from a project's blueprint and returns the mutation payload.
func RemoveComponent(ctx context.Context, c *client.Client, projectID, id string) (*ComponentPayload, error) {
	resp, err := removeComponent(ctx, c.GQL, c.Config.OrganizationID, projectID, id)
	if err != nil {
		return nil, fmt.Errorf("remove component %q: %w", id, err)
	}

	payload := resp.RemoveComponent
	return &ComponentPayload{
		Successful: payload.Successful,
		Messages:   toValidationMessages(payload.Messages),
	}, nil
}

// LinkComponents creates a link between two components and returns the mutation payload.
func LinkComponents(ctx context.Context, c *client.Client, projectID string, input LinkComponentsInput) (*LinkPayload, error) {
	resp, err := linkComponents(ctx, c.GQL, c.Config.OrganizationID, projectID, input)
	if err != nil {
		return nil, fmt.Errorf("link components: %w", err)
	}

	payload := resp.LinkComponents
	return &LinkPayload{
		Successful: payload.Successful,
		Messages:   toValidationMessages(payload.Messages),
		Result:     linkFromResult(payload.Result),
	}, nil
}

// UnlinkComponents removes a link between two components and returns the mutation payload.
func UnlinkComponents(ctx context.Context, c *client.Client, id string) (*LinkPayload, error) {
	resp, err := unlinkComponents(ctx, c.GQL, c.Config.OrganizationID, id)
	if err != nil {
		return nil, fmt.Errorf("unlink components %q: %w", id, err)
	}

	payload := resp.UnlinkComponents
	return &LinkPayload{
		Successful: payload.Successful,
		Messages:   toValidationMessages(payload.Messages),
	}, nil
}

func componentFromAddResult(r addComponentAddComponentComponentPayloadResultComponent) *Component {
	if r.Id == "" {
		return nil
	}
	return &Component{
		ID:          r.Id,
		Name:        r.Name,
		Description: r.Description,
		CreatedAt:   r.CreatedAt,
		UpdatedAt:   r.UpdatedAt,
	}
}

func linkFromResult(r linkComponentsLinkComponentsLinkPayloadResultLink) *Link {
	if r.Id == "" {
		return nil
	}
	link := &Link{
		ID:        r.Id,
		FromField: r.FromField,
		ToField:   r.ToField,
	}
	if r.FromComponent.Id != "" {
		link.FromComponent = &LinkEndpoint{ID: r.FromComponent.Id, Name: r.FromComponent.Name}
	}
	if r.ToComponent.Id != "" {
		link.ToComponent = &LinkEndpoint{ID: r.ToComponent.Id, Name: r.ToComponent.Name}
	}
	return link
}
