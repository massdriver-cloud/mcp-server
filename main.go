package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

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
	// Transport selection. By default the server speaks MCP over stdio (for
	// local tool integrations). Passing -http (or setting MASSDRIVER_MCP_HTTP_ADDR)
	// serves the Streamable HTTP transport on the given address instead.
	httpAddr := flag.String("http", os.Getenv("MASSDRIVER_MCP_HTTP_ADDR"),
		"listen address for the HTTP transport (e.g. 127.0.0.1:8080); empty uses stdio")
	flag.Parse()

	// Cancel the server context on interrupt/termination so the transport
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

	if *httpAddr != "" {
		return serveHTTP(ctx, server, *httpAddr)
	}

	// Stdio is the default MCP transport for local tool integrations.
	if err := server.Run(ctx, &mcpsdk.StdioTransport{}); err != nil {
		return fmt.Errorf("server exited with error: %w", err)
	}
	return nil
}

// serveHTTP serves the MCP Streamable HTTP transport on addr until ctx is
// canceled, then shuts the listener down gracefully.
//
// NOTE: the endpoint is unauthenticated and exposes infrastructure-mutating
// tools. Bind it to localhost or place it behind an authenticating proxy.
func serveHTTP(ctx context.Context, server *mcp.Server, addr string) error {
	httpServer := &http.Server{Addr: addr, Handler: server.HTTPHandler()}

	go func() {
		<-ctx.Done()
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		_ = httpServer.Shutdown(shutdownCtx)
	}()

	log.Printf("massdriver-mcp-server listening on %s (streamable HTTP)", addr)
	if err := httpServer.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		return fmt.Errorf("http server: %w", err)
	}
	return nil
}
