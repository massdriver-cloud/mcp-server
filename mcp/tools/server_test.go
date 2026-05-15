package tools

import (
	"context"
	"strings"
	"testing"

	"github.com/massdriver-cloud/massdriver-sdk-go/massdriver/platform/server"
)

type stubServer struct {
	getFn func(context.Context) (*server.Server, error)
}

func (s *stubServer) Get(ctx context.Context) (*server.Server, error) {
	return s.getFn(ctx)
}

func TestHandleGetServer(t *testing.T) {
	tests := []struct {
		name     string
		stub     *stubServer
		wantText string
	}{
		{
			name: "success",
			stub: &stubServer{
				getFn: func(context.Context) (*server.Server, error) {
					return &server.Server{
						AppURL:  "https://app.massdriver.cloud",
						Version: "1.2.3",
						Mode:    "cloud",
					}, nil
				},
			},
			wantText: "1.2.3",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &Client{Server: tt.stub}
			handler := HandleGetServer(c)
			result, _, err := handler(context.Background(), nil, GetServerInput{})
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if !strings.Contains(resultText(t, result), tt.wantText) {
				t.Errorf("expected %q in result, got: %s", tt.wantText, resultText(t, result))
			}
		})
	}
}
