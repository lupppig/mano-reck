// Package healthcheck implements the shell-free probe used by the backend
// container to verify process and dependency readiness.
package healthcheck

import (
	"context"
	"fmt"
	"io"
	"net/http"
)

const maxResponseBytes = 4 << 10

// Client is the minimal HTTP behavior required by the readiness probe.
type Client interface {
	Do(*http.Request) (*http.Response, error)
}

// Check returns nil only when the readiness endpoint answers successfully.
func Check(ctx context.Context, client Client, endpoint string) error {
	if client == nil {
		return fmt.Errorf("health check HTTP client is required")
	}
	request, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return fmt.Errorf("create health check request: %w", err)
	}
	response, err := client.Do(request)
	if err != nil {
		return fmt.Errorf("request readiness: %w", err)
	}
	defer response.Body.Close()
	_, _ = io.Copy(io.Discard, io.LimitReader(response.Body, maxResponseBytes))
	if response.StatusCode != http.StatusOK {
		return fmt.Errorf("readiness returned HTTP %d", response.StatusCode)
	}
	return nil
}
