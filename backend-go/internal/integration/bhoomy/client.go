// Package bhoomy is POS's connection to the Bhoomy back office. Methods are added in the phase that
// first needs them; selling never depends on Bhoomy being reachable.
package bhoomy

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"

	catalogiface "github.com/krnskrll/event-mpos/internal/catalog/iface"
)

type Client interface {
	Ping(ctx context.Context) error
	// VariantChanges returns one page of a Bhoomy store's Produk Toko changes after from (nil = all).
	VariantChanges(ctx context.Context, bhoomyStoreID uuid.UUID, from *catalogiface.Cursor, limit int) (*catalogiface.Page, error)
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

// Mock answers locally from Variants (by Bhoomy store); set Err to simulate an unreachable Bhoomy.
type Mock struct {
	Err      error
	Variants map[uuid.UUID][]catalogiface.Variant
}

func (m *Mock) Ping(context.Context) error { return m.Err }

func (m *Mock) VariantChanges(_ context.Context, store uuid.UUID, from *catalogiface.Cursor, limit int) (*catalogiface.Page, error) {
	if m.Err != nil {
		return nil, m.Err
	}
	all := append([]catalogiface.Variant(nil), m.Variants[store]...)
	sort.Slice(all, func(i, j int) bool {
		return before(all[i].ChangedAt, all[i].BhoomyVariantID, all[j].ChangedAt, all[j].BhoomyVariantID)
	})
	page := &catalogiface.Page{Items: []catalogiface.Variant{}}
	for _, v := range all {
		if from != nil && !before(from.Since, from.After, v.ChangedAt, v.BhoomyVariantID) {
			continue
		}
		if len(page.Items) == limit {
			last := page.Items[limit-1]
			page.Next = &catalogiface.Cursor{Since: last.ChangedAt, After: last.BhoomyVariantID}
			break
		}
		page.Items = append(page.Items, v)
	}
	return page, nil
}

func before(at time.Time, id uuid.UUID, bt time.Time, bid uuid.UUID) bool {
	if !at.Equal(bt) {
		return at.Before(bt)
	}
	return id.String() < bid.String()
}

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

func (c *HTTPClient) VariantChanges(ctx context.Context, store uuid.UUID, from *catalogiface.Cursor, limit int) (*catalogiface.Page, error) {
	q := url.Values{"limit": {strconv.Itoa(limit)}}
	if from != nil {
		q.Set("since", from.Since.Format(time.RFC3339Nano))
		q.Set("after", from.After.String())
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.baseURL+"/api/v1/internal/pos/stores/"+store.String()+"/variants?"+q.Encode(), nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+c.token)
	resp, err := c.http.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("bhoomy: variant changes returned %d", resp.StatusCode)
	}
	var body struct {
		Data catalogiface.Page `json:"data"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		return nil, fmt.Errorf("bhoomy: variant changes: %w", err)
	}
	return &body.Data, nil
}
