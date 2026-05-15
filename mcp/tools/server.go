package tools

import (
	"context"
	"fmt"

	mcpsdk "github.com/modelcontextprotocol/go-sdk/mcp"
)

var GetServerTool = &mcpsdk.Tool{
	Name:        "get_server",
	Description: "Gets server metadata including version information and available authentication methods.",
}

type GetServerInput struct{}

func HandleGetServer(c *Client) func(context.Context, *mcpsdk.CallToolRequest, GetServerInput) (*mcpsdk.CallToolResult, any, error) {
	return func(ctx context.Context, _ *mcpsdk.CallToolRequest, _ GetServerInput) (*mcpsdk.CallToolResult, any, error) {
		srv, err := c.Server.Get(ctx)
		if err != nil {
			return nil, nil, fmt.Errorf("get_server: %w", err)
		}

		result, err := jsonResult(srv)
		if err != nil {
			return nil, nil, err
		}
		return result, srv, nil
	}
}
