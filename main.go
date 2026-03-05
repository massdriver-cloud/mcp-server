package main

import (
	"context"
	"log"

	"github.com/massdriver-cloud/mcp-server/internal/api"
	"github.com/massdriver-cloud/mcp-server/mcp"
	mcpsdk "github.com/modelcontextprotocol/go-sdk/mcp"
)

func main() {
	ctx := context.Background()

	// Initialize the Massdriver client. Configuration is read from environment variables:
	//   MASSDRIVER_API_KEY         – required
	//   MASSDRIVER_ORGANIZATION_ID – required
	//   MASSDRIVER_URL             – optional, defaults to https://api.massdriver.cloud
	client, err := api.NewClient()
	if err != nil {
		log.Fatalf("failed to initialize Massdriver client: %v", err)
	}

	server := mcp.NewServer(client)

	// Stdio is the standard MCP transport for local tool integrations.
	transport := &mcpsdk.StdioTransport{}

	if err := server.Run(ctx, transport); err != nil {
		log.Fatalf("server exited with error: %v", err)
	}
}
