package api

import (
	sdkclient "github.com/massdriver-cloud/massdriver-sdk-go/massdriver/client"
)

// NewClient creates a *client.Client configured from environment variables.
// Required: MASSDRIVER_API_KEY, MASSDRIVER_ORGANIZATION_ID
// Optional: MASSDRIVER_URL (defaults to https://api.massdriver.cloud)
func NewClient() (*sdkclient.Client, error) {
	return sdkclient.New()
}
