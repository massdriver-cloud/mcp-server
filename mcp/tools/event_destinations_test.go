package tools

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
	"testing"

	"github.com/massdriver-cloud/massdriver-sdk-go/massdriver/platform/eventdestinations"
	"github.com/massdriver-cloud/massdriver-sdk-go/massdriver/platform/types"
	mcpsdk "github.com/modelcontextprotocol/go-sdk/mcp"
)

type stubEventDestinations struct {
	getFn                func(context.Context, string) (*eventdestinations.EventDestination, error)
	listPageFn           func(context.Context, eventdestinations.ListInput) (types.Page[eventdestinations.EventDestination], error)
	listDeliveriesPageFn func(context.Context, string, eventdestinations.ListDeliveriesInput) (types.Page[eventdestinations.Delivery], error)
	previewPageFn        func(context.Context, eventdestinations.PreviewInput) (types.Page[eventdestinations.AuditLog], error)
	listAdaptersFn       func(context.Context) ([]eventdestinations.Adapter, error)
	listEventTypesFn     func(context.Context, eventdestinations.ListEventTypesInput) ([]eventdestinations.EventType, error)
	createFn             func(context.Context, eventdestinations.CreateInput) (*eventdestinations.EventDestination, error)
	updateFn             func(context.Context, string, eventdestinations.UpdateInput) (*eventdestinations.EventDestination, error)
	pauseFn              func(context.Context, string) (*eventdestinations.EventDestination, error)
	resumeFn             func(context.Context, string) (*eventdestinations.EventDestination, error)
	deleteFn             func(context.Context, string) (*eventdestinations.EventDestination, error)
	redeliverFn          func(context.Context, string, []string) (*eventdestinations.EventDestination, error)
}

func (s *stubEventDestinations) Get(ctx context.Context, id string) (*eventdestinations.EventDestination, error) {
	return s.getFn(ctx, id)
}
func (s *stubEventDestinations) ListPage(ctx context.Context, input eventdestinations.ListInput) (types.Page[eventdestinations.EventDestination], error) {
	return s.listPageFn(ctx, input)
}
func (s *stubEventDestinations) ListDeliveriesPage(ctx context.Context, id string, input eventdestinations.ListDeliveriesInput) (types.Page[eventdestinations.Delivery], error) {
	return s.listDeliveriesPageFn(ctx, id, input)
}
func (s *stubEventDestinations) PreviewPage(ctx context.Context, input eventdestinations.PreviewInput) (types.Page[eventdestinations.AuditLog], error) {
	return s.previewPageFn(ctx, input)
}
func (s *stubEventDestinations) ListAdapters(ctx context.Context) ([]eventdestinations.Adapter, error) {
	return s.listAdaptersFn(ctx)
}
func (s *stubEventDestinations) ListEventTypes(ctx context.Context, input eventdestinations.ListEventTypesInput) ([]eventdestinations.EventType, error) {
	return s.listEventTypesFn(ctx, input)
}
func (s *stubEventDestinations) Create(ctx context.Context, input eventdestinations.CreateInput) (*eventdestinations.EventDestination, error) {
	return s.createFn(ctx, input)
}
func (s *stubEventDestinations) Update(ctx context.Context, id string, input eventdestinations.UpdateInput) (*eventdestinations.EventDestination, error) {
	return s.updateFn(ctx, id, input)
}
func (s *stubEventDestinations) Pause(ctx context.Context, id string) (*eventdestinations.EventDestination, error) {
	return s.pauseFn(ctx, id)
}
func (s *stubEventDestinations) Resume(ctx context.Context, id string) (*eventdestinations.EventDestination, error) {
	return s.resumeFn(ctx, id)
}
func (s *stubEventDestinations) Delete(ctx context.Context, id string) (*eventdestinations.EventDestination, error) {
	return s.deleteFn(ctx, id)
}
func (s *stubEventDestinations) Redeliver(ctx context.Context, id string, deliveryIDs []string) (*eventdestinations.EventDestination, error) {
	return s.redeliverFn(ctx, id, deliveryIDs)
}

const testSigningSecret = "whsec_c2VjcmV0LXNpZ25pbmcta2V5"

func signedDestination(id string) *eventdestinations.EventDestination {
	return &eventdestinations.EventDestination{
		ID:            id,
		Name:          "ops-webhook",
		Adapter:       "HTTP",
		Config:        map[string]any{"url": "https://hooks.example.com/massdriver"},
		Status:        "ACTIVE",
		SigningSecret: testSigningSecret,
	}
}

func TestEventDestinationResultsOmitSigningSecret(t *testing.T) {
	c := &Client{EventDestinations: &stubEventDestinations{
		getFn: func(_ context.Context, id string) (*eventdestinations.EventDestination, error) {
			return signedDestination(id), nil
		},
		listPageFn: func(context.Context, eventdestinations.ListInput) (types.Page[eventdestinations.EventDestination], error) {
			return types.Page[eventdestinations.EventDestination]{
				Items: []eventdestinations.EventDestination{*signedDestination("ed1"), *signedDestination("ed2")},
			}, nil
		},
		createFn: func(context.Context, eventdestinations.CreateInput) (*eventdestinations.EventDestination, error) {
			return signedDestination("ed1"), nil
		},
		updateFn: func(_ context.Context, id string, _ eventdestinations.UpdateInput) (*eventdestinations.EventDestination, error) {
			return signedDestination(id), nil
		},
		pauseFn: func(_ context.Context, id string) (*eventdestinations.EventDestination, error) {
			return signedDestination(id), nil
		},
		resumeFn: func(_ context.Context, id string) (*eventdestinations.EventDestination, error) {
			return signedDestination(id), nil
		},
		deleteFn: func(_ context.Context, id string) (*eventdestinations.EventDestination, error) {
			return signedDestination(id), nil
		},
		redeliverFn: func(_ context.Context, id string, _ []string) (*eventdestinations.EventDestination, error) {
			return signedDestination(id), nil
		},
	}}

	calls := map[string]func() (*mcpsdk.CallToolResult, any, error){
		"get_event_destination": func() (*mcpsdk.CallToolResult, any, error) {
			return HandleGetEventDestination(c)(context.Background(), nil, GetEventDestinationInput{ID: "ed1"})
		},
		"list_event_destinations": func() (*mcpsdk.CallToolResult, any, error) {
			return HandleListEventDestinations(c)(context.Background(), nil, ListEventDestinationsInput{})
		},
		"create_event_destination": func() (*mcpsdk.CallToolResult, any, error) {
			return HandleCreateEventDestination(c)(context.Background(), nil, CreateEventDestinationInput{
				Name: "ops-webhook", Config: map[string]any{"url": "https://hooks.example.com/massdriver"},
			})
		},
		"update_event_destination": func() (*mcpsdk.CallToolResult, any, error) {
			return HandleUpdateEventDestination(c)(context.Background(), nil, UpdateEventDestinationInput{ID: "ed1", Name: ptr("renamed")})
		},
		"pause_event_destination": func() (*mcpsdk.CallToolResult, any, error) {
			return HandlePauseEventDestination(c)(context.Background(), nil, PauseEventDestinationInput{ID: "ed1"})
		},
		"resume_event_destination": func() (*mcpsdk.CallToolResult, any, error) {
			return HandleResumeEventDestination(c)(context.Background(), nil, ResumeEventDestinationInput{ID: "ed1"})
		},
		"delete_event_destination": func() (*mcpsdk.CallToolResult, any, error) {
			return HandleDeleteEventDestination(c)(context.Background(), nil, DeleteEventDestinationInput{ID: "ed1"})
		},
		"redeliver_event_destination_deliveries": func() (*mcpsdk.CallToolResult, any, error) {
			return HandleRedeliverEventDestinationDeliveries(c)(context.Background(), nil, RedeliverEventDestinationDeliveriesInput{ID: "ed1"})
		},
	}

	for name, call := range calls {
		t.Run(name, func(t *testing.T) {
			result, structured, err := call()
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			text := resultText(t, result)
			if strings.Contains(text, testSigningSecret) || strings.Contains(text, "signingSecret") {
				t.Errorf("signing secret in text result: %s", text)
			}
			data, err := json.Marshal(structured)
			if err != nil {
				t.Fatalf("marshal structured content: %v", err)
			}
			if strings.Contains(string(data), testSigningSecret) {
				t.Errorf("signing secret in structured content: %s", data)
			}
		})
	}
}

func TestHandleGetEventDestination(t *testing.T) {
	tests := []struct {
		name     string
		input    GetEventDestinationInput
		stub     *stubEventDestinations
		wantErr  string
		wantText string
	}{
		{
			name:    "missing id",
			input:   GetEventDestinationInput{},
			stub:    &stubEventDestinations{},
			wantErr: "id is required",
		},
		{
			name:  "returns destination JSON",
			input: GetEventDestinationInput{ID: "ed1"},
			stub: &stubEventDestinations{
				getFn: func(_ context.Context, id string) (*eventdestinations.EventDestination, error) {
					return signedDestination(id), nil
				},
			},
			wantText: "https://hooks.example.com/massdriver",
		},
		{
			name:  "api error",
			input: GetEventDestinationInput{ID: "ed1"},
			stub: &stubEventDestinations{
				getFn: func(context.Context, string) (*eventdestinations.EventDestination, error) {
					return nil, errors.New("not found")
				},
			},
			wantErr: "get_event_destination: not found",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &Client{EventDestinations: tt.stub}
			result, _, err := HandleGetEventDestination(c)(context.Background(), nil, tt.input)
			if tt.wantErr != "" {
				if err == nil || !strings.Contains(err.Error(), tt.wantErr) {
					t.Fatalf("expected error containing %q, got: %v", tt.wantErr, err)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if !strings.Contains(resultText(t, result), tt.wantText) {
				t.Errorf("expected %q in result, got: %s", tt.wantText, resultText(t, result))
			}
		})
	}
}

func TestHandleListEventDestinations(t *testing.T) {
	var got eventdestinations.ListInput
	c := &Client{EventDestinations: &stubEventDestinations{
		listPageFn: func(_ context.Context, input eventdestinations.ListInput) (types.Page[eventdestinations.EventDestination], error) {
			got = input
			return types.Page[eventdestinations.EventDestination]{
				Items: []eventdestinations.EventDestination{{ID: "ed1", Name: "ops-webhook", Status: "PAUSED"}},
				Next:  "cursor-2",
			}, nil
		},
	}}

	result, _, err := HandleListEventDestinations(c)(context.Background(), nil, ListEventDestinationsInput{
		Status: "PAUSED", SortBy: "CREATED_AT", Order: "DESC", Cursor: "cursor-1", PageSize: 500,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(got.Statuses) != 1 || got.Statuses[0] != eventdestinations.StatusPaused {
		t.Errorf("Statuses = %v, want [PAUSED]", got.Statuses)
	}
	if got.SortBy != eventdestinations.SortByCreatedAt {
		t.Errorf("SortBy = %q, want CREATED_AT", got.SortBy)
	}
	if got.SortOrder != eventdestinations.SortDesc {
		t.Errorf("SortOrder = %q, want DESC", got.SortOrder)
	}
	if got.After != "cursor-1" {
		t.Errorf("After = %q, want cursor-1", got.After)
	}
	if got.PageSize != 100 {
		t.Errorf("PageSize = %d, want 100", got.PageSize)
	}

	text := resultText(t, result)
	for _, want := range []string{"ops-webhook", "\"next_cursor\": \"cursor-2\"", "\"has_more\": true"} {
		if !strings.Contains(text, want) {
			t.Errorf("expected %q in result, got: %s", want, text)
		}
	}
}

func TestHandleListEventDestinationsWithoutStatus(t *testing.T) {
	var got eventdestinations.ListInput
	c := &Client{EventDestinations: &stubEventDestinations{
		listPageFn: func(_ context.Context, input eventdestinations.ListInput) (types.Page[eventdestinations.EventDestination], error) {
			got = input
			return types.Page[eventdestinations.EventDestination]{}, nil
		},
	}}

	result, _, err := HandleListEventDestinations(c)(context.Background(), nil, ListEventDestinationsInput{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got.Statuses != nil {
		t.Errorf("Statuses = %v, want nil", got.Statuses)
	}
	if got.PageSize != 25 {
		t.Errorf("PageSize = %d, want 25", got.PageSize)
	}
	if text := resultText(t, result); !strings.Contains(text, "\"has_more\": false") {
		t.Errorf("expected has_more false, got: %s", text)
	}
}

func TestHandleListEventDestinationDeliveries(t *testing.T) {
	tests := []struct {
		name     string
		input    ListEventDestinationDeliveriesInput
		wantErr  string
		wantText string
		check    func(*testing.T, string, eventdestinations.ListDeliveriesInput)
	}{
		{
			name:    "missing id",
			input:   ListEventDestinationDeliveriesInput{},
			wantErr: "id is required",
		},
		{
			name:    "invalid created_after",
			input:   ListEventDestinationDeliveriesInput{ID: "ed1", CreatedAfter: "yesterday"},
			wantErr: "created_after must be an RFC 3339 timestamp",
		},
		{
			name: "passes filters to the SDK",
			input: ListEventDestinationDeliveriesInput{
				ID:            "ed1",
				Statuses:      []string{"FAILED", "PENDING"},
				CreatedAfter:  "2026-01-15T00:00:00Z",
				CreatedBefore: "2026-02-01T00:00:00Z",
				Cursor:        "cursor-1",
			},
			wantText: "connection refused",
			check: func(t *testing.T, id string, input eventdestinations.ListDeliveriesInput) {
				t.Helper()
				if id != "ed1" {
					t.Errorf("id = %q, want ed1", id)
				}
				if len(input.Statuses) != 2 || input.Statuses[0] != eventdestinations.DeliveryFailed || input.Statuses[1] != eventdestinations.DeliveryPending {
					t.Errorf("Statuses = %v, want [FAILED PENDING]", input.Statuses)
				}
				if got := input.CreatedAfter.Format("2006-01-02T15:04:05Z07:00"); got != "2026-01-15T00:00:00Z" {
					t.Errorf("CreatedAfter = %s, want 2026-01-15T00:00:00Z", got)
				}
				if got := input.CreatedBefore.Format("2006-01-02T15:04:05Z07:00"); got != "2026-02-01T00:00:00Z" {
					t.Errorf("CreatedBefore = %s, want 2026-02-01T00:00:00Z", got)
				}
				if input.After != "cursor-1" {
					t.Errorf("After = %q, want cursor-1", input.After)
				}
				if input.PageSize != 25 {
					t.Errorf("PageSize = %d, want 25", input.PageSize)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var gotID string
			var gotInput eventdestinations.ListDeliveriesInput
			c := &Client{EventDestinations: &stubEventDestinations{
				listDeliveriesPageFn: func(_ context.Context, id string, input eventdestinations.ListDeliveriesInput) (types.Page[eventdestinations.Delivery], error) {
					gotID, gotInput = id, input
					return types.Page[eventdestinations.Delivery]{
						Items: []eventdestinations.Delivery{{ID: "dl1", Status: "FAILED", Attempts: 8, LastError: "connection refused"}},
					}, nil
				},
			}}
			result, _, err := HandleListEventDestinationDeliveries(c)(context.Background(), nil, tt.input)
			if tt.wantErr != "" {
				if err == nil || !strings.Contains(err.Error(), tt.wantErr) {
					t.Fatalf("expected error containing %q, got: %v", tt.wantErr, err)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if !strings.Contains(resultText(t, result), tt.wantText) {
				t.Errorf("expected %q in result, got: %s", tt.wantText, resultText(t, result))
			}
			tt.check(t, gotID, gotInput)
		})
	}
}

func TestHandlePreviewEventDestination(t *testing.T) {
	var got eventdestinations.PreviewInput
	c := &Client{EventDestinations: &stubEventDestinations{
		previewPageFn: func(_ context.Context, input eventdestinations.PreviewInput) (types.Page[eventdestinations.AuditLog], error) {
			got = input
			return types.Page[eventdestinations.AuditLog]{
				Items: []eventdestinations.AuditLog{{ID: "log1", Type: "deployment.completed"}},
			}, nil
		},
	}}

	result, _, err := HandlePreviewEventDestination(c)(context.Background(), nil, PreviewEventDestinationInput{
		EventTypes: []string{"deployment.*"},
		Attributes: map[string]string{"md-environment": "prod"},
		Condition:  ".data.status == \"FAILED\"",
		PageSize:   10,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(got.EventTypes) != 1 || got.EventTypes[0] != "deployment.*" {
		t.Errorf("EventTypes = %v, want [deployment.*]", got.EventTypes)
	}
	if got.Attributes["md-environment"] != "prod" {
		t.Errorf("Attributes = %v, want md-environment=prod", got.Attributes)
	}
	if got.Condition != ".data.status == \"FAILED\"" {
		t.Errorf("Condition = %q", got.Condition)
	}
	if got.PageSize != 10 {
		t.Errorf("PageSize = %d, want 10", got.PageSize)
	}
	if text := resultText(t, result); !strings.Contains(text, "deployment.completed") {
		t.Errorf("expected event in result, got: %s", text)
	}
}

func TestHandlePreviewEventDestinationError(t *testing.T) {
	c := &Client{EventDestinations: &stubEventDestinations{
		previewPageFn: func(context.Context, eventdestinations.PreviewInput) (types.Page[eventdestinations.AuditLog], error) {
			return types.Page[eventdestinations.AuditLog]{}, errors.New("condition failed on event log1")
		},
	}}

	_, _, err := HandlePreviewEventDestination(c)(context.Background(), nil, PreviewEventDestinationInput{Condition: ".bad"})
	if err == nil || !strings.Contains(err.Error(), "preview_event_destination: condition failed on event log1") {
		t.Fatalf("expected preview error, got: %v", err)
	}
}

func TestHandleListEventDestinationAdapters(t *testing.T) {
	c := &Client{EventDestinations: &stubEventDestinations{
		listAdaptersFn: func(context.Context) ([]eventdestinations.Adapter, error) {
			return []eventdestinations.Adapter{{
				ID:           "HTTP",
				Name:         "HTTP",
				ConfigSchema: map[string]any{"required": []any{"url"}},
				Signs:        true,
			}}, nil
		},
	}}

	result, _, err := HandleListEventDestinationAdapters(c)(context.Background(), nil, ListEventDestinationAdaptersInput{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	text := resultText(t, result)
	for _, want := range []string{"\"items\"", "\"configSchema\"", "\"signs\": true"} {
		if !strings.Contains(text, want) {
			t.Errorf("expected %q in result, got: %s", want, text)
		}
	}
}

func TestHandleListEventDestinationEventTypes(t *testing.T) {
	var got eventdestinations.ListEventTypesInput
	c := &Client{EventDestinations: &stubEventDestinations{
		listEventTypesFn: func(_ context.Context, input eventdestinations.ListEventTypesInput) ([]eventdestinations.EventType, error) {
			got = input
			return []eventdestinations.EventType{{ID: "deployment.completed", Entity: "deployment", Action: "completed"}}, nil
		},
	}}

	result, _, err := HandleListEventDestinationEventTypes(c)(context.Background(), nil, ListEventDestinationEventTypesInput{Entity: "deployment"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got.Entity != "deployment" {
		t.Errorf("Entity = %q, want deployment", got.Entity)
	}
	if text := resultText(t, result); !strings.Contains(text, "deployment.completed") {
		t.Errorf("expected event type in result, got: %s", text)
	}
}

func TestHandleListEventDestinationEventTypesEmpty(t *testing.T) {
	c := &Client{EventDestinations: &stubEventDestinations{
		listEventTypesFn: func(context.Context, eventdestinations.ListEventTypesInput) ([]eventdestinations.EventType, error) {
			return nil, nil
		},
	}}

	result, _, err := HandleListEventDestinationEventTypes(c)(context.Background(), nil, ListEventDestinationEventTypesInput{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if text := resultText(t, result); !strings.Contains(text, "\"items\": []") {
		t.Errorf("expected empty items array, got: %s", text)
	}
}

func TestHandleCreateEventDestination(t *testing.T) {
	config := map[string]any{"url": "https://hooks.example.com/massdriver"}
	tests := []struct {
		name        string
		input       CreateEventDestinationInput
		createErr   error
		wantErr     string
		wantIsError bool
		wantText    string
		wantAdapter string
	}{
		{
			name:    "missing name",
			input:   CreateEventDestinationInput{Config: config},
			wantErr: "name is required",
		},
		{
			name:    "missing config",
			input:   CreateEventDestinationInput{Name: "ops-webhook"},
			wantErr: "config is required",
		},
		{
			name:        "defaults adapter to HTTP",
			input:       CreateEventDestinationInput{Name: "ops-webhook", Config: config},
			wantText:    "ops-webhook",
			wantAdapter: "HTTP",
		},
		{
			name:        "mutation failure is a tool error",
			input:       CreateEventDestinationInput{Name: "ops-webhook", Config: config},
			createErr:   mutationFailedErr("create event destination", "name", "has already been taken"),
			wantIsError: true,
			wantText:    "name: has already been taken",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var got eventdestinations.CreateInput
			c := &Client{EventDestinations: &stubEventDestinations{
				createFn: func(_ context.Context, input eventdestinations.CreateInput) (*eventdestinations.EventDestination, error) {
					got = input
					if tt.createErr != nil {
						return nil, tt.createErr
					}
					return signedDestination("ed1"), nil
				},
			}}
			result, _, err := HandleCreateEventDestination(c)(context.Background(), nil, tt.input)
			if tt.wantErr != "" {
				if err == nil || !strings.Contains(err.Error(), tt.wantErr) {
					t.Fatalf("expected error containing %q, got: %v", tt.wantErr, err)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if result.IsError != tt.wantIsError {
				t.Errorf("IsError = %v, want %v", result.IsError, tt.wantIsError)
			}
			if !strings.Contains(resultText(t, result), tt.wantText) {
				t.Errorf("expected %q in result, got: %s", tt.wantText, resultText(t, result))
			}
			if tt.wantAdapter != "" && got.Adapter != tt.wantAdapter {
				t.Errorf("Adapter = %q, want %q", got.Adapter, tt.wantAdapter)
			}
		})
	}
}

func TestHandleCreateEventDestinationPassesSelection(t *testing.T) {
	var got eventdestinations.CreateInput
	c := &Client{EventDestinations: &stubEventDestinations{
		createFn: func(_ context.Context, input eventdestinations.CreateInput) (*eventdestinations.EventDestination, error) {
			got = input
			return signedDestination("ed1"), nil
		},
	}}

	_, _, err := HandleCreateEventDestination(c)(context.Background(), nil, CreateEventDestinationInput{
		Name:       "ops-webhook",
		Adapter:    "HTTP",
		Config:     map[string]any{"url": "https://hooks.example.com/massdriver"},
		Auth:       map[string]any{"headers": map[string]any{"Authorization": "Bearer abc"}},
		EventTypes: []string{"deployment.*", "instance.updated"},
		Attributes: map[string]string{"md-environment": "prod"},
		Condition:  ".data.status == \"FAILED\"",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got.Name != "ops-webhook" || got.Adapter != "HTTP" {
		t.Errorf("Name/Adapter = %q/%q", got.Name, got.Adapter)
	}
	if got.Config["url"] != "https://hooks.example.com/massdriver" {
		t.Errorf("Config = %v", got.Config)
	}
	if got.Auth == nil {
		t.Error("Auth not passed")
	}
	if len(got.EventTypes) != 2 || got.EventTypes[0] != "deployment.*" {
		t.Errorf("EventTypes = %v", got.EventTypes)
	}
	if got.Attributes["md-environment"] != "prod" {
		t.Errorf("Attributes = %v", got.Attributes)
	}
	if got.Condition != ".data.status == \"FAILED\"" {
		t.Errorf("Condition = %q", got.Condition)
	}
}

func TestHandleUpdateEventDestination(t *testing.T) {
	tests := []struct {
		name  string
		args  string
		check func(*testing.T, eventdestinations.UpdateInput)
	}{
		{
			name: "omitted fields stay nil",
			args: `{"id":"ed1","name":"renamed"}`,
			check: func(t *testing.T, in eventdestinations.UpdateInput) {
				t.Helper()
				if derefStr(in.Name) != "renamed" {
					t.Errorf("Name = %v, want renamed", in.Name)
				}
				if in.Config != nil || in.Auth != nil || in.EventTypes != nil || in.Attributes != nil || in.Condition != nil {
					t.Errorf("expected omitted fields to be nil, got %+v", in)
				}
			},
		},
		{
			name: "empty values clear",
			args: `{"id":"ed1","auth":{},"event_types":[],"attributes":{},"condition":""}`,
			check: func(t *testing.T, in eventdestinations.UpdateInput) {
				t.Helper()
				if in.Auth == nil || len(in.Auth) != 0 {
					t.Errorf("Auth = %v, want empty non-nil map", in.Auth)
				}
				if in.EventTypes == nil || len(*in.EventTypes) != 0 {
					t.Errorf("EventTypes = %v, want pointer to empty slice", in.EventTypes)
				}
				if in.Attributes == nil || len(in.Attributes) != 0 {
					t.Errorf("Attributes = %v, want empty non-nil map", in.Attributes)
				}
				if in.Condition == nil || *in.Condition != "" {
					t.Errorf("Condition = %v, want pointer to empty string", in.Condition)
				}
			},
		},
		{
			name: "values replace",
			args: `{"id":"ed1","config":{"url":"https://new.example.com"},"event_types":["deployment.*"],"condition":".data.ok"}`,
			check: func(t *testing.T, in eventdestinations.UpdateInput) {
				t.Helper()
				if in.Config["url"] != "https://new.example.com" {
					t.Errorf("Config = %v", in.Config)
				}
				if in.EventTypes == nil || len(*in.EventTypes) != 1 || (*in.EventTypes)[0] != "deployment.*" {
					t.Errorf("EventTypes = %v", in.EventTypes)
				}
				if derefStr(in.Condition) != ".data.ok" {
					t.Errorf("Condition = %v", in.Condition)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var args UpdateEventDestinationInput
			if err := json.Unmarshal([]byte(tt.args), &args); err != nil {
				t.Fatalf("unmarshal args: %v", err)
			}
			var got eventdestinations.UpdateInput
			c := &Client{EventDestinations: &stubEventDestinations{
				updateFn: func(_ context.Context, id string, input eventdestinations.UpdateInput) (*eventdestinations.EventDestination, error) {
					got = input
					return signedDestination(id), nil
				},
			}}
			if _, _, err := HandleUpdateEventDestination(c)(context.Background(), nil, args); err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			tt.check(t, got)
		})
	}
}

func TestHandleUpdateEventDestinationErrors(t *testing.T) {
	c := &Client{EventDestinations: &stubEventDestinations{
		updateFn: func(context.Context, string, eventdestinations.UpdateInput) (*eventdestinations.EventDestination, error) {
			return nil, mutationFailedErr("update event destination", "condition", "is not a valid jq expression")
		},
	}}

	if _, _, err := HandleUpdateEventDestination(c)(context.Background(), nil, UpdateEventDestinationInput{}); err == nil || !strings.Contains(err.Error(), "id is required") {
		t.Fatalf("expected id is required, got: %v", err)
	}

	result, _, err := HandleUpdateEventDestination(c)(context.Background(), nil, UpdateEventDestinationInput{ID: "ed1", Condition: ptr(".bad")})
	if err != nil {
		t.Fatalf("expected handled failure, got: %v", err)
	}
	if !result.IsError {
		t.Error("expected IsError=true")
	}
	if text := resultText(t, result); !strings.Contains(text, "condition: is not a valid jq expression") {
		t.Errorf("unexpected text: %s", text)
	}
}

func TestHandlePauseAndResumeEventDestination(t *testing.T) {
	c := &Client{EventDestinations: &stubEventDestinations{
		pauseFn: func(_ context.Context, id string) (*eventdestinations.EventDestination, error) {
			d := signedDestination(id)
			d.Status = "PAUSED"
			return d, nil
		},
		resumeFn: func(_ context.Context, id string) (*eventdestinations.EventDestination, error) {
			return signedDestination(id), nil
		},
	}}

	if _, _, err := HandlePauseEventDestination(c)(context.Background(), nil, PauseEventDestinationInput{}); err == nil || !strings.Contains(err.Error(), "id is required") {
		t.Fatalf("pause: expected id is required, got: %v", err)
	}
	if _, _, err := HandleResumeEventDestination(c)(context.Background(), nil, ResumeEventDestinationInput{}); err == nil || !strings.Contains(err.Error(), "id is required") {
		t.Fatalf("resume: expected id is required, got: %v", err)
	}

	result, _, err := HandlePauseEventDestination(c)(context.Background(), nil, PauseEventDestinationInput{ID: "ed1"})
	if err != nil {
		t.Fatalf("pause: unexpected error: %v", err)
	}
	if text := resultText(t, result); !strings.Contains(text, "\"status\": \"PAUSED\"") {
		t.Errorf("pause: expected PAUSED, got: %s", text)
	}

	result, _, err = HandleResumeEventDestination(c)(context.Background(), nil, ResumeEventDestinationInput{ID: "ed1"})
	if err != nil {
		t.Fatalf("resume: unexpected error: %v", err)
	}
	if text := resultText(t, result); !strings.Contains(text, "\"status\": \"ACTIVE\"") {
		t.Errorf("resume: expected ACTIVE, got: %s", text)
	}
}

func TestHandleDeleteEventDestination(t *testing.T) {
	tests := []struct {
		name        string
		input       DeleteEventDestinationInput
		deleteErr   error
		wantErr     string
		wantIsError bool
		wantText    string
	}{
		{
			name:    "missing id",
			input:   DeleteEventDestinationInput{},
			wantErr: "id is required",
		},
		{
			name:     "deletes destination",
			input:    DeleteEventDestinationInput{ID: "ed1"},
			wantText: "event destination \"ed1\" deleted successfully",
		},
		{
			name:        "mutation failure is a tool error",
			input:       DeleteEventDestinationInput{ID: "ed1"},
			deleteErr:   mutationFailedErr("delete event destination", "", "not found"),
			wantIsError: true,
			wantText:    "delete_event_destination failed: not found",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &Client{EventDestinations: &stubEventDestinations{
				deleteFn: func(_ context.Context, id string) (*eventdestinations.EventDestination, error) {
					if tt.deleteErr != nil {
						return nil, tt.deleteErr
					}
					return &eventdestinations.EventDestination{ID: id, Name: "ops-webhook"}, nil
				},
			}}
			result, _, err := HandleDeleteEventDestination(c)(context.Background(), nil, tt.input)
			if tt.wantErr != "" {
				if err == nil || !strings.Contains(err.Error(), tt.wantErr) {
					t.Fatalf("expected error containing %q, got: %v", tt.wantErr, err)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if result.IsError != tt.wantIsError {
				t.Errorf("IsError = %v, want %v", result.IsError, tt.wantIsError)
			}
			if !strings.Contains(resultText(t, result), tt.wantText) {
				t.Errorf("expected %q in result, got: %s", tt.wantText, resultText(t, result))
			}
		})
	}
}

func TestHandleRedeliverEventDestinationDeliveries(t *testing.T) {
	tests := []struct {
		name    string
		input   RedeliverEventDestinationDeliveriesInput
		wantErr string
		wantIDs []string
	}{
		{
			name:    "missing id",
			input:   RedeliverEventDestinationDeliveriesInput{},
			wantErr: "id is required",
		},
		{
			name:    "omitted delivery_ids sends every failed delivery",
			input:   RedeliverEventDestinationDeliveriesInput{ID: "ed1"},
			wantIDs: nil,
		},
		{
			name:    "passes delivery_ids",
			input:   RedeliverEventDestinationDeliveriesInput{ID: "ed1", DeliveryIDs: []string{"dl1", "dl2"}},
			wantIDs: []string{"dl1", "dl2"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var gotIDs []string
			c := &Client{EventDestinations: &stubEventDestinations{
				redeliverFn: func(_ context.Context, id string, deliveryIDs []string) (*eventdestinations.EventDestination, error) {
					gotIDs = deliveryIDs
					return &eventdestinations.EventDestination{ID: id, Name: "ops-webhook"}, nil
				},
			}}
			result, _, err := HandleRedeliverEventDestinationDeliveries(c)(context.Background(), nil, tt.input)
			if tt.wantErr != "" {
				if err == nil || !strings.Contains(err.Error(), tt.wantErr) {
					t.Fatalf("expected error containing %q, got: %v", tt.wantErr, err)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if strings.Join(gotIDs, ",") != strings.Join(tt.wantIDs, ",") || (gotIDs == nil) != (tt.wantIDs == nil) {
				t.Errorf("deliveryIDs = %v, want %v", gotIDs, tt.wantIDs)
			}
			if text := resultText(t, result); !strings.Contains(text, "ops-webhook") {
				t.Errorf("expected destination in result, got: %s", text)
			}
		})
	}
}
