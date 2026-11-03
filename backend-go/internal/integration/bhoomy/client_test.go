package bhoomy

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/google/uuid"

	catalogiface "github.com/krnskrll/event-mpos/internal/catalog/iface"
)

func TestNew(t *testing.T) {
	cases := []struct {
		mode, url, token string
		wantErr          bool
	}{
		{"", "", "", false},
		{"mock", "", "", false},
		{"http", "http://bhoomy", "secret", false},
		{"http", "", "secret", true},
		{"http", "http://bhoomy", "", true},
		{"grpc", "", "", true},
	}
	for _, c := range cases {
		if _, err := New(c.mode, c.url, c.token); (err != nil) != c.wantErr {
			t.Errorf("New(%q, %q, %q) err = %v, wantErr %v", c.mode, c.url, c.token, err, c.wantErr)
		}
	}
}

func TestHTTPPing(t *testing.T) {
	var gotAuth, gotPath string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotAuth, gotPath = r.Header.Get("Authorization"), r.URL.Path
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	c, _ := New("http", srv.URL+"/", "secret")
	if err := c.Ping(context.Background()); err != nil {
		t.Fatal(err)
	}
	if gotAuth != "Bearer secret" || gotPath != "/api/health" {
		t.Errorf("auth %q path %q", gotAuth, gotPath)
	}
}

func TestHTTPPingFailures(t *testing.T) {
	down := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusServiceUnavailable)
	}))
	defer down.Close()
	slow := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		time.Sleep(200 * time.Millisecond)
	}))
	defer slow.Close()

	for name, c := range map[string]*HTTPClient{
		"non-200": {baseURL: down.URL, token: "t", http: http.DefaultClient},
		"timeout": {baseURL: slow.URL, token: "t", http: &http.Client{Timeout: 50 * time.Millisecond}},
	} {
		if err := c.Ping(context.Background()); err == nil {
			t.Errorf("%s: expected an error", name)
		}
	}
}

func TestHTTPVariantChanges(t *testing.T) {
	store, after := uuid.New(), uuid.New()
	since := time.Date(2026, 11, 3, 1, 2, 3, 456000, time.UTC)
	var got *http.Request
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		got = r
		w.Write([]byte(`{"errorStatus":false,"data":{"items":[{"variant_id":"` + after.String() + `","product_name":"Serum","sku":"S-1","is_default":true,"status":"active","changed_at":"2026-11-03T01:02:03Z"}],"next":null}}`))
	}))
	defer srv.Close()

	c, _ := New("http", srv.URL, "secret")
	page, err := c.VariantChanges(context.Background(), store, &catalogiface.Cursor{Since: since, After: after}, 50)
	if err != nil {
		t.Fatal(err)
	}
	if got.URL.Path != "/api/v1/internal/pos/stores/"+store.String()+"/variants" || got.Header.Get("Authorization") != "Bearer secret" {
		t.Errorf("request %s %v", got.URL.Path, got.Header)
	}
	if q := got.URL.Query(); q.Get("limit") != "50" || q.Get("after") != after.String() || q.Get("since") != since.Format(time.RFC3339Nano) {
		t.Errorf("query %v", q)
	}
	if len(page.Items) != 1 || page.Items[0].ProductName != "Serum" || page.Next != nil {
		t.Errorf("page %+v", page)
	}
}
