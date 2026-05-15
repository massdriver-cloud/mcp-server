package tools

import (
	"context"
	"fmt"

	mcpsdk "github.com/modelcontextprotocol/go-sdk/mcp"
)

var GetViewerTool = &mcpsdk.Tool{
	Name:        "get_viewer",
	Description: "Gets the currently authenticated identity (user account or service account).",
}

type GetViewerInput struct{}

func HandleGetViewer(c *Client) func(context.Context, *mcpsdk.CallToolRequest, GetViewerInput) (*mcpsdk.CallToolResult, any, error) {
	return func(ctx context.Context, _ *mcpsdk.CallToolRequest, _ GetViewerInput) (*mcpsdk.CallToolResult, any, error) {
		viewer, err := c.Viewer.Get(ctx)
		if err != nil {
			return nil, nil, fmt.Errorf("get_viewer: %w", err)
		}

		result, err := jsonResult(viewer)
		if err != nil {
			return nil, nil, err
		}
		return result, viewer, nil
	}
}
