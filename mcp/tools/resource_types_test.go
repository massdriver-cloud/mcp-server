package tools

import (
	"context"
	"errors"
	"strings"
	"testing"

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
	}{
		{
			name:    "missing id",
			input:   GetResourceTypeInput{},
			stub:    &stubResourceTypes{},
			wantErr: "id is required",
		},
		{
			name:  "returns resource type JSON with icon stripped",
			input: GetResourceTypeInput{ID: "aws-iam-role@~1"},
			stub: &stubResourceTypes{
				getFn: func(_ context.Context, id string) (*resourcetypes.ResourceType, error) {
					if id != "aws-iam-role@~1" {
						t.Errorf("expected id %q, got %q", "aws-iam-role@~1", id)
					}
					return &resourcetypes.ResourceType{
						ID:      "aws-iam-role@1.2.3",
						Name:    "AWS IAM Role",
						Version: "1.2.3",
						Icon:    "<svg>blob</svg>",
					}, nil
				},
			},
			wantText: "aws-iam-role@1.2.3",
		},
		{
			name:  "error is surfaced",
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
			handler := HandleGetResourceType(c)
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
			text := resultText(t, result)
			if !strings.Contains(text, tt.wantText) {
				t.Errorf("expected %q in result, got: %s", tt.wantText, text)
			}
			if strings.Contains(text, "<svg>") {
				t.Errorf("expected icon to be stripped, got: %s", text)
			}
		})
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
			name:  "returns dependents list",
			input: ListResourceTypeDependentsInput{EnvironmentID: "myproj-staging", ResourceTypeID: "aws-vpc"},
			stub: &stubResourceTypes{
				dependentsFn: func(_ context.Context, environmentID, resourceTypeID string) ([]resourcetypes.Dependent, error) {
					if environmentID != "myproj-staging" || resourceTypeID != "aws-vpc" {
						t.Errorf("unexpected args: %q %q", environmentID, resourceTypeID)
					}
					return []resourcetypes.Dependent{{
						Instance: types.Instance{ID: "myproj-staging-api"},
						Field:    "network",
					}}, nil
				},
			},
			wantText: "myproj-staging-api",
		},
		{
			name:  "empty list returns items array",
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
			handler := HandleListResourceTypeDependents(c)
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
