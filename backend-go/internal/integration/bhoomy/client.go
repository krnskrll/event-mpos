// Package bhoomy is POS's connection to the Bhoomy back office. Methods are added in the phase that
// first needs them; selling never depends on Bhoomy being reachable.
package bhoomy

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"
)

type Client interface {
	Ping(ctx context.Context) error
}

// New returns the client for mode "mock" (default) or "http".
func New(mode, baseURL, token string) (Client, error) {
	switch mode {
	case "", "mock":
		return &Mock{}, nil
	case "http":
		if baseURL == "" || token == "" {
			return nil, errors.New("bhoomy: BHOOMY_URL and BHOOMY_SERVICE_TOKEN are required when BHOOMY_MODE=http")
		}
		return &HTTPClient{baseURL: strings.TrimRight(baseURL, "/"), token: token, http: &http.Client{Timeout: 5 * time.Second}}, nil
	}
	return nil, fmt.Errorf("bhoomy: unknown BHOOMY_MODE %q", mode)
}

// Mock answers locally; set Err to simulate an unreachable Bhoomy.
type Mock struct{ Err error }

func (m *Mock) Ping(context.Context) error { return m.Err }

type HTTPClient struct {
	baseURL string
	token   string
	http    *http.Client
}

func (c *HTTPClient) Ping(ctx context.Context) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.baseURL+"/api/health", nil)
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", "Bearer "+c.token)
	resp, err := c.http.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("bhoomy: ping returned %d", resp.StatusCode)
	}
	return nil
}
