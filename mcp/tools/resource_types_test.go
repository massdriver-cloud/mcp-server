package tools

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/massdriver-cloud/massdriver-sdk-go/massdriver/platform/instances"
	"github.com/massdriver-cloud/massdriver-sdk-go/massdriver/platform/resourcetypes"
	"github.com/massdriver-cloud/massdriver-sdk-go/massdriver/platform/types"
)

type stubResourceTypes struct {
	getFn        func(context.Context, string) (*resourcetypes.ResourceType, error)
	dependentsFn func(context.Context, string, string) ([]resourcetypes.Dependent, error)
}

func (s *stubResourceTypes) Get(ctx context.Context, id string) (*resourcetypes.ResourceType, error) {
	return s.getFn(ctx, id)
}
func (s *stubResourceTypes) Dependents(ctx context.Context, environmentID, resourceTypeID string) ([]resourcetypes.Dependent, error) {
	return s.dependentsFn(ctx, environmentID, resourceTypeID)
}

func TestHandleGetResourceType(t *testing.T) {
	tests := []struct {
		name     string
		input    GetResourceTypeInput
		stub     *stubResourceTypes
		wantErr  string
		wantText string
		absent   string
	}{
		{
			name:    "missing id",
			input:   GetResourceTypeInput{},
			stub:    &stubResourceTypes{},
			wantErr: "id is required",
		},
		{
			name:  "returns the schema the payload must satisfy",
			input: GetResourceTypeInput{ID: "aws-iam-role"},
			stub: &stubResourceTypes{
				getFn: func(_ context.Context, id string) (*resourcetypes.ResourceType, error) {
					return &resourcetypes.ResourceType{
						ID:      id + "@1.2.3",
						Name:    "AWS IAM Role",
						Version: "1.2.3",
						Schema:  map[string]any{"required": []any{"arn"}},
					}, nil
				},
			},
			wantText: "aws-iam-role@1.2.3",
		},
		{
			name:  "strips the icon blob",
			input: GetResourceTypeInput{ID: "aws-iam-role"},
			stub: &stubResourceTypes{
				getFn: func(context.Context, string) (*resourcetypes.ResourceType, error) {
					return &resourcetypes.ResourceType{ID: "aws-iam-role@1.0.0", Icon: "<svg/>"}, nil
				},
			},
			wantText: "aws-iam-role@1.0.0",
			absent:   "svg",
		},
		{
			name:  "propagates lookup failure",
			input: GetResourceTypeInput{ID: "nope"},
			stub: &stubResourceTypes{
				getFn: func(context.Context, string) (*resourcetypes.ResourceType, error) {
					return nil, errors.New("not found")
				},
			},
			wantErr: "get_resource_type",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &Client{ResourceTypes: tt.stub}
			result, _, err := HandleGetResourceType(c)(context.Background(), nil, tt.input)
			if tt.wantErr != "" {
				if err == nil || !strings.Contains(err.Error(), tt.wantErr) {
					t.Fatalf("expected error containing %q, got: %v", tt.wantErr, err)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			got := resultText(t, result)
			if !strings.Contains(got, tt.wantText) {
				t.Errorf("expected %q in result, got: %s", tt.wantText, got)
			}
			if tt.absent != "" && strings.Contains(got, tt.absent) {
				t.Errorf("expected %q to be stripped from result, got: %s", tt.absent, got)
			}
		})
	}
}

func TestHandleGetResourceTypePassesVersionedIDThrough(t *testing.T) {
	var gotID string
	c := &Client{ResourceTypes: &stubResourceTypes{
		getFn: func(_ context.Context, id string) (*resourcetypes.ResourceType, error) {
			gotID = id
			return &resourcetypes.ResourceType{ID: id}, nil
		},
	}}

	if _, _, err := HandleGetResourceType(c)(context.Background(), nil, GetResourceTypeInput{ID: "aws-vpc@~1.2"}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if gotID != "aws-vpc@~1.2" {
		t.Errorf("resource type id = %q, want the version suffix preserved", gotID)
	}
}

func TestHandleListResourceTypeDependents(t *testing.T) {
	tests := []struct {
		name     string
		input    ListResourceTypeDependentsInput
		stub     *stubResourceTypes
		wantErr  string
		wantText string
	}{
		{
			name:    "missing environment_id",
			input:   ListResourceTypeDependentsInput{ResourceTypeID: "aws-vpc"},
			stub:    &stubResourceTypes{},
			wantErr: "environment_id is required",
		},
		{
			name:    "missing resource_type_id",
			input:   ListResourceTypeDependentsInput{EnvironmentID: "myproj-staging"},
			stub:    &stubResourceTypes{},
			wantErr: "resource_type_id is required",
		},
		{
			name:  "returns one entry per instance and field",
			input: ListResourceTypeDependentsInput{EnvironmentID: "myproj-staging", ResourceTypeID: "aws-vpc"},
			stub: &stubResourceTypes{
				dependentsFn: func(context.Context, string, string) ([]resourcetypes.Dependent, error) {
					return []resourcetypes.Dependent{
						{Instance: instances.Instance{ID: "inst1"}, Field: "network"},
						{Instance: instances.Instance{ID: "inst1"}, Field: "peer_network"},
					}, nil
				},
			},
			wantText: "peer_network",
		},
		{
			name:  "no dependents yields an empty array, not null",
			input: ListResourceTypeDependentsInput{EnvironmentID: "myproj-staging", ResourceTypeID: "aws-vpc"},
			stub: &stubResourceTypes{
				dependentsFn: func(context.Context, string, string) ([]resourcetypes.Dependent, error) {
					return nil, nil
				},
			},
			wantText: "\"items\": []",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &Client{ResourceTypes: tt.stub}
			result, _, err := HandleListResourceTypeDependents(c)(context.Background(), nil, tt.input)
			if tt.wantErr != "" {
				if err == nil || !strings.Contains(err.Error(), tt.wantErr) {
					t.Fatalf("expected error containing %q, got: %v", tt.wantErr, err)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got := resultText(t, result); !strings.Contains(got, tt.wantText) {
				t.Errorf("expected %q in result, got: %s", tt.wantText, got)
			}
		})
	}
}

// TestListResourceTypeDependentsScopesToEnvironment guards the argument order —
// swapping the environment and resource type would silently return the wrong set.
func TestListResourceTypeDependentsScopesToEnvironment(t *testing.T) {
	var gotEnv, gotType string
	c := &Client{ResourceTypes: &stubResourceTypes{
		dependentsFn: func(_ context.Context, environmentID, resourceTypeID string) ([]resourcetypes.Dependent, error) {
			gotEnv, gotType = environmentID, resourceTypeID
			return []resourcetypes.Dependent{{ResourceType: types.ResourceType{ID: resourceTypeID}}}, nil
		},
	}}

	_, _, err := HandleListResourceTypeDependents(c)(context.Background(), nil, ListResourceTypeDependentsInput{
		EnvironmentID:  "myproj-staging",
		ResourceTypeID: "aws-vpc",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if gotEnv != "myproj-staging" || gotType != "aws-vpc" {
		t.Errorf("Dependents(env=%q, type=%q), want (myproj-staging, aws-vpc)", gotEnv, gotType)
	}
}
