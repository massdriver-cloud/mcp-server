package tools

import (
	"context"
	"strings"
	"testing"

	"github.com/massdriver-cloud/massdriver-sdk-go/massdriver/platform/instances"
	"github.com/massdriver-cloud/massdriver-sdk-go/massdriver/platform/types"
)

type stubInstances struct {
	listPageFn        func(context.Context, instances.ListInput) (types.Page[instances.Instance], error)
	getFn             func(context.Context, string) (*instances.Instance, error)
	updateFn          func(context.Context, string, instances.UpdateInput) (*instances.Instance, error)
	setSecretFn       func(context.Context, string, string, string) (*instances.Secret, error)
	removeSecretFn    func(context.Context, string, string) (*instances.Secret, error)
	setRemoteRefFn    func(context.Context, string, string, string) (*instances.RemoteReference, error)
	removeRemoteRefFn func(context.Context, string, string) (*instances.RemoteReference, error)
	listAlarmsPageFn  func(context.Context, instances.ListAlarmsInput) (types.Page[instances.Alarm], error)
}

func (s *stubInstances) ListPage(ctx context.Context, input instances.ListInput) (types.Page[instances.Instance], error) {
	return s.listPageFn(ctx, input)
}
func (s *stubInstances) Get(ctx context.Context, id string) (*instances.Instance, error) {
	return s.getFn(ctx, id)
}
func (s *stubInstances) Update(ctx context.Context, id string, input instances.UpdateInput) (*instances.Instance, error) {
	return s.updateFn(ctx, id, input)
}
func (s *stubInstances) SetSecret(ctx context.Context, instanceID, name, value string) (*instances.Secret, error) {
	return s.setSecretFn(ctx, instanceID, name, value)
}
func (s *stubInstances) RemoveSecret(ctx context.Context, instanceID, name string) (*instances.Secret, error) {
	return s.removeSecretFn(ctx, instanceID, name)
}
func (s *stubInstances) SetRemoteReference(ctx context.Context, instanceID, resourceID, field string) (*instances.RemoteReference, error) {
	return s.setRemoteRefFn(ctx, instanceID, resourceID, field)
}
func (s *stubInstances) RemoveRemoteReference(ctx context.Context, instanceID, field string) (*instances.RemoteReference, error) {
	return s.removeRemoteRefFn(ctx, instanceID, field)
}
func (s *stubInstances) ListAlarmsPage(ctx context.Context, input instances.ListAlarmsInput) (types.Page[instances.Alarm], error) {
	return s.listAlarmsPageFn(ctx, input)
}

func TestHandleListInstances(t *testing.T) {
	tests := []struct {
		name     string
		input    ListInstancesInput
		stub     *stubInstances
		wantErr  bool
		wantText string
	}{
		{
			name:  "returns page of instances",
			input: ListInstancesInput{},
			stub: &stubInstances{
				listPageFn: func(_ context.Context, _ instances.ListInput) (types.Page[instances.Instance], error) {
					return types.Page[instances.Instance]{
						Items: []instances.Instance{{ID: "proj1-staging-db", Name: "Database"}},
					}, nil
				},
			},
			wantText: "proj1-staging-db",
		},
		{
			name:  "empty page surfaces has_more false",
			input: ListInstancesInput{},
			stub: &stubInstances{
				listPageFn: func(context.Context, instances.ListInput) (types.Page[instances.Instance], error) {
					return types.Page[instances.Instance]{}, nil
				},
			},
			wantText: "\"has_more\": false",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &Client{Instances: tt.stub}
			handler := HandleListInstances(c)
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

func TestHandleGetInstance(t *testing.T) {
	tests := []struct {
		name     string
		input    GetInstanceInput
		stub     *stubInstances
		wantErr  string
		wantText string
	}{
		{
			name:    "missing id",
			input:   GetInstanceInput{},
			stub:    &stubInstances{},
			wantErr: "id is required",
		},
		{
			name:  "returns instance JSON",
			input: GetInstanceInput{ID: "proj1-staging-db"},
			stub: &stubInstances{
				getFn: func(_ context.Context, id string) (*instances.Instance, error) {
					return &instances.Instance{ID: id, Name: "Database"}, nil
				},
			},
			wantText: "proj1-staging-db",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &Client{Instances: tt.stub}
			handler := HandleGetInstance(c)
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

func TestHandleSetInstanceSecret(t *testing.T) {
	tests := []struct {
		name     string
		input    SetInstanceSecretInput
		stub     *stubInstances
		wantErr  string
		wantText string
	}{
		{
			name:    "missing instance_id",
			input:   SetInstanceSecretInput{Name: "DB_PASS", Value: "secret"},
			stub:    &stubInstances{},
			wantErr: "instance_id is required",
		},
		{
			name:    "missing name",
			input:   SetInstanceSecretInput{InstanceID: "inst1", Value: "secret"},
			stub:    &stubInstances{},
			wantErr: "name is required",
		},
		{
			name:    "missing value",
			input:   SetInstanceSecretInput{InstanceID: "inst1", Name: "DB_PASS"},
			stub:    &stubInstances{},
			wantErr: "value is required",
		},
		{
			name:  "success returns secret JSON",
			input: SetInstanceSecretInput{InstanceID: "inst1", Name: "DB_PASS", Value: "secret"},
			stub: &stubInstances{
				setSecretFn: func(_ context.Context, _, name, _ string) (*instances.Secret, error) {
					return &instances.Secret{Name: name}, nil
				},
			},
			wantText: "DB_PASS",
		},
		{
			name:  "mutation failure returns error message",
			input: SetInstanceSecretInput{InstanceID: "inst1", Name: "DB_PASS", Value: "secret"},
			stub: &stubInstances{
				setSecretFn: func(context.Context, string, string, string) (*instances.Secret, error) {
					return nil, mutationFailedErr("set secret", "name", "invalid characters")
				},
			},
			wantText: "set_instance_secret failed",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &Client{Instances: tt.stub}
			handler := HandleSetInstanceSecret(c)
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

func TestHandleRemoveInstanceSecret(t *testing.T) {
	tests := []struct {
		name     string
		input    RemoveInstanceSecretInput
		stub     *stubInstances
		wantErr  string
		wantText string
	}{
		{
			name:    "missing instance_id",
			input:   RemoveInstanceSecretInput{Name: "DB_PASS"},
			stub:    &stubInstances{},
			wantErr: "instance_id is required",
		},
		{
			name:    "missing name",
			input:   RemoveInstanceSecretInput{InstanceID: "inst1"},
			stub:    &stubInstances{},
			wantErr: "name is required",
		},
		{
			name:  "success returns secret JSON",
			input: RemoveInstanceSecretInput{InstanceID: "inst1", Name: "DB_PASS"},
			stub: &stubInstances{
				removeSecretFn: func(_ context.Context, _, name string) (*instances.Secret, error) {
					return &instances.Secret{Name: name}, nil
				},
			},
			wantText: "DB_PASS",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &Client{Instances: tt.stub}
			handler := HandleRemoveInstanceSecret(c)
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

func TestHandleListAlarms(t *testing.T) {
	tests := []struct {
		name     string
		input    ListAlarmsInput
		stub     *stubInstances
		wantErr  bool
		wantText string
	}{
		{
			name:  "returns alarms page",
			input: ListAlarmsInput{InstanceID: "inst1"},
			stub: &stubInstances{
				listAlarmsPageFn: func(_ context.Context, _ instances.ListAlarmsInput) (types.Page[instances.Alarm], error) {
					return types.Page[instances.Alarm]{
						Items: []instances.Alarm{{ID: "alarm1", DisplayName: "High CPU"}},
					}, nil
				},
			},
			wantText: "High CPU",
		},
		{
			name:  "empty page surfaces has_more false",
			input: ListAlarmsInput{},
			stub: &stubInstances{
				listAlarmsPageFn: func(context.Context, instances.ListAlarmsInput) (types.Page[instances.Alarm], error) {
					return types.Page[instances.Alarm]{}, nil
				},
			},
			wantText: "\"has_more\": false",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &Client{Instances: tt.stub}
			handler := HandleListAlarms(c)
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

func TestHandleUpdateInstance(t *testing.T) {
	tests := []struct {
		name     string
		input    UpdateInstanceInput
		stub     *stubInstances
		wantErr  string
		wantText string
	}{
		{
			name:    "missing id",
			input:   UpdateInstanceInput{Version: "~1.2"},
			stub:    &stubInstances{},
			wantErr: "id is required",
		},
		{
			name:    "missing version",
			input:   UpdateInstanceInput{ID: "inst1"},
			stub:    &stubInstances{},
			wantErr: "version is required",
		},
		{
			name:  "success returns instance JSON",
			input: UpdateInstanceInput{ID: "inst1", Version: "~1.2"},
			stub: &stubInstances{
				updateFn: func(_ context.Context, id string, _ instances.UpdateInput) (*instances.Instance, error) {
					return &instances.Instance{ID: id, Name: "Database"}, nil
				},
			},
			wantText: "inst1",
		},
		{
			name:  "mutation failure returns error message",
			input: UpdateInstanceInput{ID: "inst1", Version: "~1.2"},
			stub: &stubInstances{
				updateFn: func(context.Context, string, instances.UpdateInput) (*instances.Instance, error) {
					return nil, mutationFailedErr("update instance", "version", "invalid constraint")
				},
			},
			wantText: "update_instance failed",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &Client{Instances: tt.stub}
			handler := HandleUpdateInstance(c)
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

func TestHandleSetRemoteReference(t *testing.T) {
	tests := []struct {
		name     string
		input    SetRemoteReferenceInput
		stub     *stubInstances
		wantErr  string
		wantText string
	}{
		{
			name:    "missing instance_id",
			input:   SetRemoteReferenceInput{ResourceID: "res1", Field: "database"},
			stub:    &stubInstances{},
			wantErr: "instance_id is required",
		},
		{
			name:    "missing resource_id",
			input:   SetRemoteReferenceInput{InstanceID: "inst1", Field: "database"},
			stub:    &stubInstances{},
			wantErr: "resource_id is required",
		},
		{
			name:    "missing field",
			input:   SetRemoteReferenceInput{InstanceID: "inst1", ResourceID: "res1"},
			stub:    &stubInstances{},
			wantErr: "field is required",
		},
		{
			name:  "success returns remote reference JSON",
			input: SetRemoteReferenceInput{InstanceID: "inst1", ResourceID: "res1", Field: "database"},
			stub: &stubInstances{
				setRemoteRefFn: func(_ context.Context, _, _, field string) (*instances.RemoteReference, error) {
					return &instances.RemoteReference{Field: field}, nil
				},
			},
			wantText: "database",
		},
		{
			name:  "mutation failure returns error message",
			input: SetRemoteReferenceInput{InstanceID: "inst1", ResourceID: "res1", Field: "database"},
			stub: &stubInstances{
				setRemoteRefFn: func(context.Context, string, string, string) (*instances.RemoteReference, error) {
					return nil, mutationFailedErr("set remote reference", "field", "unknown slot")
				},
			},
			wantText: "set_remote_reference failed",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &Client{Instances: tt.stub}
			handler := HandleSetRemoteReference(c)
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

func TestHandleRemoveRemoteReference(t *testing.T) {
	tests := []struct {
		name     string
		input    RemoveRemoteReferenceInput
		stub     *stubInstances
		wantErr  string
		wantText string
	}{
		{
			name:    "missing instance_id",
			input:   RemoveRemoteReferenceInput{Field: "database"},
			stub:    &stubInstances{},
			wantErr: "instance_id is required",
		},
		{
			name:    "missing field",
			input:   RemoveRemoteReferenceInput{InstanceID: "inst1"},
			stub:    &stubInstances{},
			wantErr: "field is required",
		},
		{
			name:  "success returns remote reference JSON",
			input: RemoveRemoteReferenceInput{InstanceID: "inst1", Field: "database"},
			stub: &stubInstances{
				removeRemoteRefFn: func(_ context.Context, _, field string) (*instances.RemoteReference, error) {
					return &instances.RemoteReference{Field: field}, nil
				},
			},
			wantText: "database",
		},
		{
			name:  "mutation failure returns error message",
			input: RemoveRemoteReferenceInput{InstanceID: "inst1", Field: "database"},
			stub: &stubInstances{
				removeRemoteRefFn: func(context.Context, string, string) (*instances.RemoteReference, error) {
					return nil, mutationFailedErr("remove remote reference", "field", "not set")
				},
			},
			wantText: "remove_remote_reference failed",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &Client{Instances: tt.stub}
			handler := HandleRemoveRemoteReference(c)
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
