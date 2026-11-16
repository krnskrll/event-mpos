// Package bhoomy is POS's connection to the Bhoomy back office. Methods are added in the phase that
// first needs them; selling never depends on Bhoomy being reachable.
package bhoomy

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
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
	// Stock is what a POS location may sell per Produk Toko variant (6c).
	Stock(ctx context.Context, bhoomyStoreID, locationID uuid.UUID) (*Stock, error)
	// PostSale and CancelSale answer the sale's stock status; both are safe to send twice.
	PostSale(ctx context.Context, bhoomyStoreID uuid.UUID, sale Sale) (string, error)
	CancelSale(ctx context.Context, bhoomyStoreID, saleID uuid.UUID, at time.Time) (string, error)
}

type Stock struct {
	Tracked bool         `json:"tracked"`
	PlaceID *uuid.UUID   `json:"place_id"`
	Levels  []StockLevel `json:"levels"`
}

type StockLevel struct {
	VariantID uuid.UUID `json:"variant_id"`
	Available int       `json:"available"`
}

type Sale struct {
	PosSaleID       uuid.UUID  `json:"pos_sale_id"`
	PosLocationID   uuid.UUID  `json:"pos_location_id"`
	PosLocationName string     `json:"pos_location_name"`
	PlaceID         *uuid.UUID `json:"place_id"`
	ReceiptNo       string     `json:"receipt_no"`
	SoldBy          string     `json:"sold_by"`
	Total           int64      `json:"total"`
	SoldAt          time.Time  `json:"sold_at"`
	Lines           []SaleLine `json:"lines"`
}

type SaleLine struct {
	PosLineID uuid.UUID  `json:"pos_line_id"`
	VariantID *uuid.UUID `json:"variant_id"`
	SKU       string     `json:"sku"`
	Name      string     `json:"name"`
	Qty       int        `json:"qty"`
	UnitPrice int64      `json:"unit_price"`
	NetAmount int64      `json:"net_amount"`
}

// RefusedError is Bhoomy saying no for good (4xx): sending again will not help.
type RefusedError struct {
	Status int
	Body   string
}

func (e *RefusedError) Error() string {
	return fmt.Sprintf("bhoomy refused (%d): %s", e.Status, e.Body)
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

// Mock answers locally from Variants (by Bhoomy store) and Stocks (by location); set Err to
// simulate an unreachable Bhoomy. Sales and Cancels record what was sent; SaleStatus is the answer.
type Mock struct {
	Err        error
	Variants   map[uuid.UUID][]catalogiface.Variant
	Stocks     map[uuid.UUID]*Stock
	SaleStatus string
	Sales      []Sale
	Cancels    []uuid.UUID
}

func (m *Mock) Ping(context.Context) error { return m.Err }

func (m *Mock) Stock(_ context.Context, _, location uuid.UUID) (*Stock, error) {
	if m.Err != nil {
		return nil, m.Err
	}
	if st := m.Stocks[location]; st != nil {
		return st, nil
	}
	return &Stock{}, nil
}

func (m *Mock) PostSale(_ context.Context, _ uuid.UUID, sale Sale) (string, error) {
	if m.Err != nil {
		return "", m.Err
	}
	m.Sales = append(m.Sales, sale)
	if m.SaleStatus != "" {
		return m.SaleStatus, nil
	}
	if sale.PlaceID == nil {
		return "untracked", nil
	}
	return "posted", nil
}

func (m *Mock) CancelSale(_ context.Context, _, saleID uuid.UUID, _ time.Time) (string, error) {
	if m.Err != nil {
		return "", m.Err
	}
	m.Cancels = append(m.Cancels, saleID)
	return "cancelled", nil
}

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

func (c *HTTPClient) Stock(ctx context.Context, store, location uuid.UUID) (*Stock, error) {
	var st Stock
	return &st, c.call(ctx, http.MethodGet, "/stores/"+store.String()+"/pos-locations/"+location.String()+"/stock", nil, &st)
}

func (c *HTTPClient) PostSale(ctx context.Context, store uuid.UUID, sale Sale) (string, error) {
	var res struct {
		StockStatus string `json:"stock_status"`
	}
	err := c.call(ctx, http.MethodPost, "/stores/"+store.String()+"/sales", sale, &res)
	return res.StockStatus, err
}

func (c *HTTPClient) CancelSale(ctx context.Context, store, saleID uuid.UUID, at time.Time) (string, error) {
	var res struct {
		StockStatus string `json:"stock_status"`
	}
	err := c.call(ctx, http.MethodPost, "/stores/"+store.String()+"/sales/"+saleID.String()+"/cancel", map[string]time.Time{"cancelled_at": at}, &res)
	return res.StockStatus, err
}

// call sends one machine request under /api/v1/internal/pos and decodes its data. A 4xx other
// than 404 and 429 is a RefusedError.
func (c *HTTPClient) call(ctx context.Context, method, path string, in, out any) error {
	var body io.Reader
	if in != nil {
		b, err := json.Marshal(in)
		if err != nil {
			return err
		}
		body = bytes.NewReader(b)
	}
	req, err := http.NewRequestWithContext(ctx, method, c.baseURL+"/api/v1/internal/pos"+path, body)
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", "Bearer "+c.token)
	req.Header.Set("Content-Type", "application/json")
	resp, err := c.http.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		msg, _ := io.ReadAll(io.LimitReader(resp.Body, 500))
		if resp.StatusCode >= 400 && resp.StatusCode < 500 && resp.StatusCode != http.StatusNotFound && resp.StatusCode != http.StatusTooManyRequests {
			return &RefusedError{Status: resp.StatusCode, Body: string(msg)}
		}
		return fmt.Errorf("bhoomy: %s %s returned %d", method, path, resp.StatusCode)
	}
	wrap := struct {
		Data any `json:"data"`
	}{Data: out}
	return json.NewDecoder(resp.Body).Decode(&wrap)
}
