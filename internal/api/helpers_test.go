package api

import (
	"testing"

	"github.com/Khan/genqlient/graphql"
	mdclient "github.com/massdriver-cloud/massdriver-sdk-go/massdriver/client"
)

// newTestClient creates a *client.Client wired to the given mock GQL transport.
func newTestClient(t *testing.T, gqlClient graphql.Client) *mdclient.Client {
	t.Helper()
	return &mdclient.Client{GQL: gqlClient}
}
