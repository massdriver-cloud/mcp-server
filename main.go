package main

import (
	"context"
	"log"

	"github.com/massdriver-cloud/massdriver-sdk-go/massdriver"
	"github.com/massdriver-cloud/mcp-server/mcp"
	mcpsdk "github.com/modelcontextprotocol/go-sdk/mcp"
)

func main() {
	ctx := context.Background()

	// Initialize the Massdriver client. Configuration is resolved from:
	//   1. Environment variables: MASSDRIVER_API_KEY, MASSDRIVER_ORGANIZATION_ID, MASSDRIVER_URL
	//   2. Profile in ~/.config/massdriver/config.yaml
	client, err := massdriver.NewClient()
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
