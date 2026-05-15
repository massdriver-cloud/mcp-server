package tools

import (
	"context"
	"strings"
	"testing"

	"github.com/massdriver-cloud/massdriver-sdk-go/massdriver/platform/auditlogs"
)

type stubAuditLogs struct {
	listFn           func(context.Context, auditlogs.ListInput) ([]auditlogs.AuditLog, error)
	listEventTypesFn func(context.Context) ([]string, error)
	getFn            func(context.Context, string) (*auditlogs.AuditLog, error)
}

func (s *stubAuditLogs) List(ctx context.Context, input auditlogs.ListInput) ([]auditlogs.AuditLog, error) {
	return s.listFn(ctx, input)
}
func (s *stubAuditLogs) ListEventTypes(ctx context.Context) ([]string, error) {
	return s.listEventTypesFn(ctx)
}
func (s *stubAuditLogs) Get(ctx context.Context, id string) (*auditlogs.AuditLog, error) {
	return s.getFn(ctx, id)
}

func TestHandleListAuditLogs(t *testing.T) {
	tests := []struct {
		name     string
		input    ListAuditLogsInput
		stub     *stubAuditLogs
		wantErr  bool
		wantText string
	}{
		{
			name:  "returns audit logs",
			input: ListAuditLogsInput{},
			stub: &stubAuditLogs{
				listFn: func(_ context.Context, _ auditlogs.ListInput) ([]auditlogs.AuditLog, error) {
					return []auditlogs.AuditLog{{ID: "log1", Type: "deployment.created"}}, nil
				},
			},
			wantText: "deployment.created",
		},
		{
			name:  "returns null for empty list",
			input: ListAuditLogsInput{},
			stub: &stubAuditLogs{
				listFn: func(context.Context, auditlogs.ListInput) ([]auditlogs.AuditLog, error) {
					return nil, nil
				},
			},
			wantText: "null",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &Client{AuditLogs: tt.stub}
			handler := HandleListAuditLogs(c)
			result, _, err := handler(context.Background(), nil, tt.input)
			if tt.wantErr {
				if err == nil {
					t.Fatal("expected error, got nil")
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

func TestHandleListAuditLogEventTypes(t *testing.T) {
	tests := []struct {
		name     string
		stub     *stubAuditLogs
		wantErr  bool
		wantText string
	}{
		{
			name: "returns event types",
			stub: &stubAuditLogs{
				listEventTypesFn: func(context.Context) ([]string, error) {
					return []string{"deployment.created", "project.updated"}, nil
				},
			},
			wantText: "deployment.created",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &Client{AuditLogs: tt.stub}
			handler := HandleListAuditLogEventTypes(c)
			result, _, err := handler(context.Background(), nil, ListAuditLogEventTypesInput{})
			if tt.wantErr {
				if err == nil {
					t.Fatal("expected error, got nil")
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

func TestHandleGetAuditLog(t *testing.T) {
	tests := []struct {
		name     string
		input    GetAuditLogInput
		stub     *stubAuditLogs
		wantErr  string
		wantText string
	}{
		{
			name:    "missing id",
			input:   GetAuditLogInput{},
			stub:    &stubAuditLogs{},
			wantErr: "id is required",
		},
		{
			name:  "returns audit log JSON",
			input: GetAuditLogInput{ID: "log1"},
			stub: &stubAuditLogs{
				getFn: func(_ context.Context, id string) (*auditlogs.AuditLog, error) {
					return &auditlogs.AuditLog{ID: id, Type: "deployment.created"}, nil
				},
			},
			wantText: "log1",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &Client{AuditLogs: tt.stub}
			handler := HandleGetAuditLog(c)
			result, _, err := handler(context.Background(), nil, tt.input)
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
