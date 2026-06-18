package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/massdriver-cloud/massdriver-sdk-go/massdriver"
	"github.com/massdriver-cloud/mcp-server/mcp"
	mcpsdk "github.com/modelcontextprotocol/go-sdk/mcp"
)

func main() {
	if err := run(); err != nil {
		log.Fatal(err)
	}
}

func run() error {
	// Cancel the server context on interrupt/termination so the stdio loop
	// shuts down cleanly when the host signals or the client disconnects.
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	// Initialize the Massdriver client. Configuration is resolved from:
	//   1. Environment variables: MASSDRIVER_API_KEY, MASSDRIVER_ORGANIZATION_ID, MASSDRIVER_URL
	//   2. Profile in ~/.config/massdriver/config.yaml
	client, err := massdriver.NewClient()
	if err != nil {
		return fmt.Errorf("failed to initialize Massdriver client: %w", err)
	}

	server := mcp.NewServer(client)

	// Stdio is the standard MCP transport for local tool integrations.
	transport := &mcpsdk.StdioTransport{}

	if err := server.Run(ctx, transport); err != nil {
		return fmt.Errorf("server exited with error: %w", err)
	}
	return nil
}
