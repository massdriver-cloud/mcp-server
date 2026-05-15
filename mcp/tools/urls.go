package tools

import (
	"context"
	"fmt"

	mcpsdk "github.com/modelcontextprotocol/go-sdk/mcp"
)

var GetURLTool = &mcpsdk.Tool{
	Name:        "get_url",
	Description: "Generates a deep link URL into the Massdriver web UI. Supported types: organization, projects, project, environment, instance, bundle, repo_instances.",
}

type GetURLInput struct {
	Type    string `json:"type"    jsonschema:"URL type: organization, projects, project, environment, instance, bundle, or repo_instances."`
	ID      string `json:"id"     jsonschema:"Optional. Resource ID (required for project, environment, instance, bundle, repo_instances)."`
	Version string `json:"version" jsonschema:"Optional. Version string (required for bundle, repo_instances)."`
}

func HandleGetURL(c *Client) func(context.Context, *mcpsdk.CallToolRequest, GetURLInput) (*mcpsdk.CallToolResult, any, error) {
	return func(ctx context.Context, _ *mcpsdk.CallToolRequest, args GetURLInput) (*mcpsdk.CallToolResult, any, error) {
		if args.Type == "" {
			return nil, nil, fmt.Errorf("get_url: type is required")
		}

		h := c.URLs.Helper(ctx)

		var url string
		switch args.Type {
		case "organization":
			url = h.OrganizationURL()
		case "projects":
			url = h.ProjectsURL()
		case "project":
			if args.ID == "" {
				return nil, nil, fmt.Errorf("get_url: id is required for type %q", args.Type)
			}
			url = h.ProjectURL(args.ID)
		case "environment":
			if args.ID == "" {
				return nil, nil, fmt.Errorf("get_url: id is required for type %q", args.Type)
			}
			url = h.EnvironmentURL(args.ID)
		case "instance":
			if args.ID == "" {
				return nil, nil, fmt.Errorf("get_url: id is required for type %q", args.Type)
			}
			url = h.InstanceURL(args.ID)
		case "bundle":
			if args.ID == "" {
				return nil, nil, fmt.Errorf("get_url: id is required for type %q", args.Type)
			}
			if args.Version == "" {
				return nil, nil, fmt.Errorf("get_url: version is required for type %q", args.Type)
			}
			url = h.BundleURL(args.ID, args.Version)
		case "repo_instances":
			if args.ID == "" {
				return nil, nil, fmt.Errorf("get_url: id is required for type %q", args.Type)
			}
			if args.Version == "" {
				return nil, nil, fmt.Errorf("get_url: version is required for type %q", args.Type)
			}
			url = h.RepoInstancesURL(args.ID, args.Version)
		default:
			return nil, nil, fmt.Errorf("get_url: unknown type %q", args.Type)
		}

		return textResult(url), nil, nil
	}
}
