package bhoomy

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
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
