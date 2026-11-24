package tools

import (
	"context"
	"fmt"

	"github.com/massdriver-cloud/massdriver-sdk-go/massdriver/platform/eventdestinations"
	mcpsdk "github.com/modelcontextprotocol/go-sdk/mcp"
)

const eventDestinationSecretNote = "The signing secret and auth credentials are not returned. " +
	"The user can see the signing secret in the Massdriver UI or with `mass event destination get <id> -o json`."

// withoutSigningSecret clears the destination's signing secret, which is a
// credential and must never reach the model.
func withoutSigningSecret(d *eventdestinations.EventDestination) *eventdestinations.EventDestination {
	d.SigningSecret = ""
	return d
}

func eventDestinationResult(d *eventdestinations.EventDestination) (*mcpsdk.CallToolResult, any, error) {
	d = withoutSigningSecret(d)
	result, err := jsonResult(d)
	if err != nil {
		return nil, nil, err
	}
	return result, d, nil
}

var GetEventDestinationTool = &mcpsdk.Tool{
	Name: "get_event_destination",
	Description: "Gets an event destination by ID: its name, adapter, config, event types, attributes, condition, " +
		"status, and the time and error of its last delivery. " + eventDestinationSecretNote,
}

type GetEventDestinationInput struct {
	ID string `json:"id" jsonschema:"The event destination ID."`
}

func HandleGetEventDestination(c *Client) func(context.Context, *mcpsdk.CallToolRequest, GetEventDestinationInput) (*mcpsdk.CallToolResult, any, error) {
	return func(ctx context.Context, _ *mcpsdk.CallToolRequest, args GetEventDestinationInput) (*mcpsdk.CallToolResult, any, error) {
		if args.ID == "" {
			return nil, nil, fmt.Errorf("get_event_destination: id is required")
		}

		destination, err := c.EventDestinations.Get(ctx, args.ID)
		if err != nil {
			return nil, nil, fmt.Errorf("get_event_destination: %w", err)
		}

		return eventDestinationResult(destination)
	}
}

var ListEventDestinationsTool = &mcpsdk.Tool{
	Name: "list_event_destinations",
	Description: "Lists the organization's event destinations, places outside Massdriver that receive its audit log events. " +
		"Optionally filter by `status`. " +
		"Returns up to `page_size` destinations (default 25, max 100) plus a `next_cursor` for the following page. " +
		"To continue, call again with `cursor` set to the previous `next_cursor`. " +
		"Stop once you have what you need — do NOT paginate to exhaustion unless the user explicitly asked for every destination. " +
		eventDestinationSecretNote,
}

type ListEventDestinationsInput struct {
	Status   string `json:"status,omitempty"    jsonschema:"Optional. Filter by status: ACTIVE or PAUSED."`
	SortBy   string `json:"sort_by,omitempty"   jsonschema:"Optional. Sort field: NAME or CREATED_AT. Default NAME."`
	Order    string `json:"order,omitempty"     jsonschema:"Optional. Sort order: ASC or DESC. Default ASC."`
	Cursor   string `json:"cursor,omitempty"    jsonschema:"Optional. Opaque cursor from a prior call's next_cursor. Omit for the first page."`
	PageSize int    `json:"page_size,omitempty" jsonschema:"Optional. Page size (1-100, default 25)."`
}

func HandleListEventDestinations(c *Client) func(context.Context, *mcpsdk.CallToolRequest, ListEventDestinationsInput) (*mcpsdk.CallToolResult, any, error) {
	return func(ctx context.Context, _ *mcpsdk.CallToolRequest, args ListEventDestinationsInput) (*mcpsdk.CallToolResult, any, error) {
		input := eventdestinations.ListInput{
			SortBy:    eventdestinations.SortField(args.SortBy),
			SortOrder: eventdestinations.SortOrder(args.Order),
			PageSize:  clampPageSize(args.PageSize),
			After:     args.Cursor,
		}
		if args.Status != "" {
			input.Statuses = []eventdestinations.Status{eventdestinations.Status(args.Status)}
		}

		page, err := c.EventDestinations.ListPage(ctx, input)
		if err != nil {
			return nil, nil, fmt.Errorf("list_event_destinations: %w", err)
		}

		for i := range page.Items {
			withoutSigningSecret(&page.Items[i])
		}
		out := pageResult(page)
		result, err := jsonResult(out)
		if err != nil {
			return nil, nil, err
		}
		return result, out, nil
	}
}

var ListEventDestinationDeliveriesTool = &mcpsdk.Tool{
	Name: "list_event_destination_deliveries",
	Description: "Lists an event destination's deliveries, newest first. Each delivery has its status, attempts, last error, and the audit log event. " +
		"Use it to find why a destination fails. Optionally filter by `statuses` or `created_after`/`created_before`. " +
		"Returns up to `page_size` deliveries (default 25, max 100) plus a `next_cursor` for the following page. " +
		"To continue, call again with `cursor` set to the previous `next_cursor`. " +
		"Do NOT paginate to exhaustion unless the user explicitly asked for the full history.",
}

type ListEventDestinationDeliveriesInput struct {
	ID            string   `json:"id"                       jsonschema:"The event destination ID."`
	Statuses      []string `json:"statuses,omitempty"       jsonschema:"Optional. Only deliveries with any of these statuses: DELIVERED, SKIPPED (the condition did not pass the event), FAILED (no more attempts will run), or PENDING (waits to be sent or retried)."`
	CreatedAfter  string   `json:"created_after,omitempty"  jsonschema:"Optional. Only deliveries created at or after this RFC 3339 timestamp (e.g., '2026-01-15T00:00:00Z'). Inclusive."`
	CreatedBefore string   `json:"created_before,omitempty" jsonschema:"Optional. Only deliveries created at or before this RFC 3339 timestamp (e.g., '2026-02-01T00:00:00Z'). Inclusive."`
	Cursor        string   `json:"cursor,omitempty"         jsonschema:"Optional. Opaque cursor from a prior call's next_cursor. Omit for the first page."`
	PageSize      int      `json:"page_size,omitempty"      jsonschema:"Optional. Page size (1-100, default 25)."`
}

func HandleListEventDestinationDeliveries(c *Client) func(context.Context, *mcpsdk.CallToolRequest, ListEventDestinationDeliveriesInput) (*mcpsdk.CallToolResult, any, error) {
	return func(ctx context.Context, _ *mcpsdk.CallToolRequest, args ListEventDestinationDeliveriesInput) (*mcpsdk.CallToolResult, any, error) {
		if args.ID == "" {
			return nil, nil, fmt.Errorf("list_event_destination_deliveries: id is required")
		}
		createdAfter, err := parseTimeFilter("list_event_destination_deliveries", "created_after", args.CreatedAfter)
		if err != nil {
			return nil, nil, err
		}
		createdBefore, err := parseTimeFilter("list_event_destination_deliveries", "created_before", args.CreatedBefore)
		if err != nil {
			return nil, nil, err
		}

		var statuses []eventdestinations.DeliveryStatus
		for _, s := range args.Statuses {
			statuses = append(statuses, eventdestinations.DeliveryStatus(s))
		}

		page, err := c.EventDestinations.ListDeliveriesPage(ctx, args.ID, eventdestinations.ListDeliveriesInput{
			Statuses:      statuses,
			CreatedAfter:  createdAfter,
			CreatedBefore: createdBefore,
			PageSize:      clampPageSize(args.PageSize),
			After:         args.Cursor,
		})
		if err != nil {
			return nil, nil, fmt.Errorf("list_event_destination_deliveries: %w", err)
		}

		out := pageResult(page)
		result, err := jsonResult(out)
		if err != nil {
			return nil, nil, err
		}
		return result, out, nil
	}
}

var PreviewEventDestinationTool = &mcpsdk.Tool{
	Name: "preview_event_destination",
	Description: "Lists the recent audit log events that an event destination selection would receive, newest first. " +
		"Takes the same `event_types`, `attributes`, and `condition` as create_event_destination, so a selection can be checked before it is used. " +
		"A page can hold fewer items than `page_size` because the condition runs after the selection. " +
		"Returns up to `page_size` events (default 25, max 100) plus a `next_cursor` for the following page. " +
		"To continue, call again with `cursor` set to the previous `next_cursor`. " +
		"Stop once you have what you need — do NOT paginate to exhaustion.",
}

type PreviewEventDestinationInput struct {
	EventTypes []string          `json:"event_types,omitempty" jsonschema:"Optional. Event type IDs (e.g., 'deployment.completed'), or '<entity>.*' for every event of an entity (e.g., 'deployment.*'). Omit or pass an empty list to select every event type. Use list_event_destination_event_types to find IDs."`
	Attributes map[string]string `json:"attributes,omitempty"  jsonschema:"Optional. Attribute values the event's entity must have had, e.g. {\"md-environment\":\"prod\"}. All must match. Omit to select events of any attributes."`
	Condition  string            `json:"condition,omitempty"   jsonschema:"Optional. A jq expression run on each selected CloudEvent. The event is selected only when it outputs true."`
	Cursor     string            `json:"cursor,omitempty"      jsonschema:"Optional. Opaque cursor from a prior call's next_cursor. Omit for the first page."`
	PageSize   int               `json:"page_size,omitempty"   jsonschema:"Optional. Page size (1-100, default 25)."`
}

func HandlePreviewEventDestination(c *Client) func(context.Context, *mcpsdk.CallToolRequest, PreviewEventDestinationInput) (*mcpsdk.CallToolResult, any, error) {
	return func(ctx context.Context, _ *mcpsdk.CallToolRequest, args PreviewEventDestinationInput) (*mcpsdk.CallToolResult, any, error) {
		page, err := c.EventDestinations.PreviewPage(ctx, eventdestinations.PreviewInput{
			EventTypes: args.EventTypes,
			Attributes: args.Attributes,
			Condition:  args.Condition,
			PageSize:   clampPageSize(args.PageSize),
			After:      args.Cursor,
		})
		if err != nil {
			return nil, nil, fmt.Errorf("preview_event_destination: %w", err)
		}

		out := pageResult(page)
		result, err := jsonResult(out)
		if err != nil {
			return nil, nil, err
		}
		return result, out, nil
	}
}

var ListEventDestinationAdaptersTool = &mcpsdk.Tool{
	Name: "list_event_destination_adapters",
	Description: "Lists the adapters an event destination can use to send events, with the JSON schemas of each adapter's `config` and `auth`. " +
		"Use it before create_event_destination to learn the config and auth shape.",
}

type ListEventDestinationAdaptersInput struct{}

func HandleListEventDestinationAdapters(c *Client) func(context.Context, *mcpsdk.CallToolRequest, ListEventDestinationAdaptersInput) (*mcpsdk.CallToolResult, any, error) {
	return func(ctx context.Context, _ *mcpsdk.CallToolRequest, _ ListEventDestinationAdaptersInput) (*mcpsdk.CallToolResult, any, error) {
		adapters, err := c.EventDestinations.ListAdapters(ctx)
		if err != nil {
			return nil, nil, fmt.Errorf("list_event_destination_adapters: %w", err)
		}

		out := listResult(adapters)
		result, err := jsonResult(out)
		if err != nil {
			return nil, nil, err
		}
		return result, out, nil
	}
}

var ListEventDestinationEventTypesTool = &mcpsdk.Tool{
	Name: "list_event_destination_event_types",
	Description: "Lists the event types an event destination can receive, sorted by ID. " +
		"Each has an ID such as 'deployment.completed', its entity, action, description, and the metadata recorded with it. " +
		"Optionally filter by `entity`.",
}

type ListEventDestinationEventTypesInput struct {
	Entity string `json:"entity,omitempty" jsonschema:"Optional. Only the event types of this entity (e.g., 'deployment')."`
}

func HandleListEventDestinationEventTypes(c *Client) func(context.Context, *mcpsdk.CallToolRequest, ListEventDestinationEventTypesInput) (*mcpsdk.CallToolResult, any, error) {
	return func(ctx context.Context, _ *mcpsdk.CallToolRequest, args ListEventDestinationEventTypesInput) (*mcpsdk.CallToolResult, any, error) {
		eventTypes, err := c.EventDestinations.ListEventTypes(ctx, eventdestinations.ListEventTypesInput{Entity: args.Entity})
		if err != nil {
			return nil, nil, fmt.Errorf("list_event_destination_event_types: %w", err)
		}

		out := listResult(eventTypes)
		result, err := jsonResult(out)
		if err != nil {
			return nil, nil, err
		}
		return result, out, nil
	}
}

var CreateEventDestinationTool = &mcpsdk.Tool{
	Name: "create_event_destination",
	Description: "Creates an event destination that receives the organization's audit log events. " +
		"Use list_event_destination_adapters for the config and auth shape, and preview_event_destination to check the selection first. " +
		eventDestinationSecretNote,
}

type CreateEventDestinationInput struct {
	Name       string            `json:"name"                  jsonschema:"The destination name. Unique in the organization."`
	Adapter    string            `json:"adapter,omitempty"     jsonschema:"Optional. How the destination receives events. Default HTTP."`
	Config     map[string]any    `json:"config"                jsonschema:"Where and how events are sent, in the shape of the adapter's config schema (for HTTP: {\"url\": \"https://...\", \"headers\": {...}}). The API returns the config, including config.headers, so put credentials such as an Authorization header in auth.headers."`
	Auth       map[string]any    `json:"auth,omitempty"        jsonschema:"Optional. Credentials in the shape of the adapter's auth schema (for HTTP: {\"headers\": {\"Authorization\": \"...\"}}). Write-only: never returned. Omit to send none."`
	EventTypes []string          `json:"event_types,omitempty" jsonschema:"Optional. Event type IDs (e.g., 'deployment.completed'), or '<entity>.*' for every event of an entity. Omit or pass an empty list to receive every event."`
	Attributes map[string]string `json:"attributes,omitempty"  jsonschema:"Optional. Attribute values the event's entity must have had, e.g. {\"md-environment\":\"prod\"}. All must match. Omit to receive events of any attributes."`
	Condition  string            `json:"condition,omitempty"   jsonschema:"Optional. A jq expression run on each selected CloudEvent. The event is sent only when it outputs true. Omit to send every selected event."`
}

func HandleCreateEventDestination(c *Client) func(context.Context, *mcpsdk.CallToolRequest, CreateEventDestinationInput) (*mcpsdk.CallToolResult, any, error) {
	return func(ctx context.Context, _ *mcpsdk.CallToolRequest, args CreateEventDestinationInput) (*mcpsdk.CallToolResult, any, error) {
		if args.Name == "" {
			return nil, nil, fmt.Errorf("create_event_destination: name is required")
		}
		if args.Config == nil {
			return nil, nil, fmt.Errorf("create_event_destination: config is required")
		}
		adapter := args.Adapter
		if adapter == "" {
			adapter = eventdestinations.AdapterHTTP
		}

		destination, err := c.EventDestinations.Create(ctx, eventdestinations.CreateInput{
			Name:       args.Name,
			Adapter:    adapter,
			Config:     args.Config,
			Auth:       args.Auth,
			EventTypes: args.EventTypes,
			Attributes: args.Attributes,
			Condition:  args.Condition,
		})
		if err != nil {
			if isMutationFailed(err) {
				return errorResult(fmt.Sprintf("create_event_destination failed: %s", mutationErr(err))), nil, nil
			}
			return nil, nil, fmt.Errorf("create_event_destination: %w", err)
		}

		return eventDestinationResult(destination)
	}
}

var UpdateEventDestinationTool = &mcpsdk.Tool{
	Name: "update_event_destination",
	Description: "Updates an event destination's name, config, auth, event types, attributes, or condition. " +
		"Omitted fields keep their current values. The adapter cannot change. " +
		"New event types and attributes apply to events recorded after the change. " +
		eventDestinationSecretNote,
}

type UpdateEventDestinationInput struct {
	ID         string            `json:"id"                    jsonschema:"The event destination ID to update."`
	Name       *string           `json:"name,omitempty"        jsonschema:"Optional. New name. Omit to leave unchanged."`
	Config     map[string]any    `json:"config,omitempty"      jsonschema:"Optional. Replacement config in the shape of the adapter's config schema. Omit to leave unchanged; when provided, replaces the whole config. The API returns the config, including config.headers, so put credentials such as an Authorization header in auth.headers."`
	Auth       map[string]any    `json:"auth,omitempty"        jsonschema:"Optional. Replacement credentials in the shape of the adapter's auth schema. Omit to leave unchanged; pass an empty object {} to remove the credentials."`
	EventTypes *[]string         `json:"event_types,omitempty" jsonschema:"Optional. Replacement event type IDs, or '<entity>.*' for every event of an entity. Omit to leave unchanged; pass an empty list [] to receive every event."`
	Attributes map[string]string `json:"attributes,omitempty"  jsonschema:"Optional. Replacement attribute values the event's entity must have had. Omit to leave unchanged; pass an empty object {} to receive events of any attributes."`
	Condition  *string           `json:"condition,omitempty"   jsonschema:"Optional. Replacement jq condition. Omit to leave unchanged; pass an empty string to remove the condition."`
}

func HandleUpdateEventDestination(c *Client) func(context.Context, *mcpsdk.CallToolRequest, UpdateEventDestinationInput) (*mcpsdk.CallToolResult, any, error) {
	return func(ctx context.Context, _ *mcpsdk.CallToolRequest, args UpdateEventDestinationInput) (*mcpsdk.CallToolResult, any, error) {
		if args.ID == "" {
			return nil, nil, fmt.Errorf("update_event_destination: id is required")
		}

		destination, err := c.EventDestinations.Update(ctx, args.ID, eventdestinations.UpdateInput{
			Name:       args.Name,
			Config:     args.Config,
			Auth:       args.Auth,
			EventTypes: args.EventTypes,
			Attributes: args.Attributes,
			Condition:  args.Condition,
		})
		if err != nil {
			if isMutationFailed(err) {
				return errorResult(fmt.Sprintf("update_event_destination failed: %s", mutationErr(err))), nil, nil
			}
			return nil, nil, fmt.Errorf("update_event_destination: %w", err)
		}

		return eventDestinationResult(destination)
	}
}

var PauseEventDestinationTool = &mcpsdk.Tool{
	Name: "pause_event_destination",
	Description: "Pauses an event destination so it sends nothing. Events recorded while it is paused wait and are sent after resume_event_destination. " +
		eventDestinationSecretNote,
}

type PauseEventDestinationInput struct {
	ID string `json:"id" jsonschema:"The event destination ID to pause."`
}

func HandlePauseEventDestination(c *Client) func(context.Context, *mcpsdk.CallToolRequest, PauseEventDestinationInput) (*mcpsdk.CallToolResult, any, error) {
	return func(ctx context.Context, _ *mcpsdk.CallToolRequest, args PauseEventDestinationInput) (*mcpsdk.CallToolResult, any, error) {
		if args.ID == "" {
			return nil, nil, fmt.Errorf("pause_event_destination: id is required")
		}

		destination, err := c.EventDestinations.Pause(ctx, args.ID)
		if err != nil {
			if isMutationFailed(err) {
				return errorResult(fmt.Sprintf("pause_event_destination failed: %s", mutationErr(err))), nil, nil
			}
			return nil, nil, fmt.Errorf("pause_event_destination: %w", err)
		}

		return eventDestinationResult(destination)
	}
}

var ResumeEventDestinationTool = &mcpsdk.Tool{
	Name: "resume_event_destination",
	Description: "Resumes a paused event destination. It sends the events that waited first. " +
		"Massdriver pauses a destination on its own after a period of failed sends; check list_event_destination_deliveries for the cause before resuming. " +
		eventDestinationSecretNote,
}

type ResumeEventDestinationInput struct {
	ID string `json:"id" jsonschema:"The event destination ID to resume."`
}

func HandleResumeEventDestination(c *Client) func(context.Context, *mcpsdk.CallToolRequest, ResumeEventDestinationInput) (*mcpsdk.CallToolResult, any, error) {
	return func(ctx context.Context, _ *mcpsdk.CallToolRequest, args ResumeEventDestinationInput) (*mcpsdk.CallToolResult, any, error) {
		if args.ID == "" {
			return nil, nil, fmt.Errorf("resume_event_destination: id is required")
		}

		destination, err := c.EventDestinations.Resume(ctx, args.ID)
		if err != nil {
			if isMutationFailed(err) {
				return errorResult(fmt.Sprintf("resume_event_destination failed: %s", mutationErr(err))), nil, nil
			}
			return nil, nil, fmt.Errorf("resume_event_destination: %w", err)
		}

		return eventDestinationResult(destination)
	}
}

var DeleteEventDestinationTool = &mcpsdk.Tool{
	Name:        "delete_event_destination",
	Description: "Deletes an event destination and its deliveries. Events that wait to be sent are not sent.",
}

type DeleteEventDestinationInput struct {
	ID string `json:"id" jsonschema:"The event destination ID to delete."`
}

func HandleDeleteEventDestination(c *Client) func(context.Context, *mcpsdk.CallToolRequest, DeleteEventDestinationInput) (*mcpsdk.CallToolResult, any, error) {
	return func(ctx context.Context, _ *mcpsdk.CallToolRequest, args DeleteEventDestinationInput) (*mcpsdk.CallToolResult, any, error) {
		if args.ID == "" {
			return nil, nil, fmt.Errorf("delete_event_destination: id is required")
		}

		_, err := c.EventDestinations.Delete(ctx, args.ID)
		if err != nil {
			if isMutationFailed(err) {
				return errorResult(fmt.Sprintf("delete_event_destination failed: %s", mutationErr(err))), nil, nil
			}
			return nil, nil, fmt.Errorf("delete_event_destination: %w", err)
		}

		return textResult(fmt.Sprintf("event destination %q deleted successfully", args.ID)), nil, nil
	}
}

var RedeliverEventDestinationDeliveriesTool = &mcpsdk.Tool{
	Name: "redeliver_event_destination_deliveries",
	Description: "Sends an event destination's FAILED deliveries again, with a new set of retries. " +
		"Pass `delivery_ids` to choose deliveries; omit it to send every FAILED delivery again. " +
		"Deliveries that are not FAILED are left as they are. A paused destination sends them after it is resumed. " +
		"Each call sends the events again to the external system. Before you omit `delivery_ids`, list the FAILED " +
		"deliveries with list_event_destination_deliveries and confirm the count with the user.",
}

type RedeliverEventDestinationDeliveriesInput struct {
	ID          string   `json:"id"                     jsonschema:"The event destination ID."`
	DeliveryIDs []string `json:"delivery_ids,omitempty" jsonschema:"Optional. IDs of FAILED deliveries to send again. Omit to send every FAILED delivery again."`
}

func HandleRedeliverEventDestinationDeliveries(c *Client) func(context.Context, *mcpsdk.CallToolRequest, RedeliverEventDestinationDeliveriesInput) (*mcpsdk.CallToolResult, any, error) {
	return func(ctx context.Context, _ *mcpsdk.CallToolRequest, args RedeliverEventDestinationDeliveriesInput) (*mcpsdk.CallToolResult, any, error) {
		if args.ID == "" {
			return nil, nil, fmt.Errorf("redeliver_event_destination_deliveries: id is required")
		}

		destination, err := c.EventDestinations.Redeliver(ctx, args.ID, args.DeliveryIDs)
		if err != nil {
			if isMutationFailed(err) {
				return errorResult(fmt.Sprintf("redeliver_event_destination_deliveries failed: %s", mutationErr(err))), nil, nil
			}
			return nil, nil, fmt.Errorf("redeliver_event_destination_deliveries: %w", err)
		}

		return eventDestinationResult(destination)
	}
}
