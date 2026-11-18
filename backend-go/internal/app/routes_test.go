package app_test

import (
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"testing"

	"github.com/google/uuid"

	"github.com/krnskrll/event-mpos/internal/testutil"
)

// Set UPDATE_GOLDEN=1 to rewrite the golden file after an intended route change.
func TestRoutesUnchanged(t *testing.T) {
	a := testutil.NewApp(t)
	var routes []string
	for _, r := range a.Router.Routes() {
		routes = append(routes, r.Method+" "+r.Path)
	}
	sort.Strings(routes)
	got := strings.Join(routes, "\n") + "\n"

	golden := filepath.Join("testdata", "routes.golden")
	if os.Getenv("UPDATE_GOLDEN") != "" {
		if err := os.WriteFile(golden, []byte(got), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	want, err := os.ReadFile(golden)
	if err != nil {
		t.Fatal(err)
	}
	if got != string(want) {
		t.Errorf("routes changed:\n--- want\n%s--- got\n%s", want, got)
	}
}

// The old dashboard is gone (6f): its login, every /api/admin route and the stock allocation
// routes. So are the device routes the offline-first app never calls.
func TestRetiredRoutes(t *testing.T) {
	a := testutil.NewApp(t)
	ev, id := a.ActiveEvent().ID.String(), uuid.NewString()
	for _, r := range []string{
		"POST /api/auth/login", "POST /api/auth/refresh", "POST /api/auth/logout",
		"GET /api/admin/events", "GET /api/admin/catalog/sync", "POST /api/admin/catalog/sync",
		"GET /api/admin/events/" + ev + "/stock", "PUT /api/admin/events/" + ev + "/stock/" + id,
		"POST /api/admin/events", "GET /api/admin/products", "POST /api/admin/erp/sync", "GET /api/admin/users", "POST /api/admin/assignments",
		"GET /api/me", "GET /api/me/events", "GET /api/me/members", "GET /api/me/orders/" + id,
		"GET /api/me/events/" + ev + "/stock", "GET /api/me/events/" + ev + "/orders",
		"GET /api/me/events/" + ev + "/products/by-barcode/899", "POST /api/me/events/" + ev + "/promotions/quote",
		"GET /api/warehouse/events", "GET /api/warehouse/events/" + ev + "/stock",
	} {
		method, path, _ := strings.Cut(r, " ")
		if rec := a.Do(method, path, map[string]any{}, ""); rec.Code != http.StatusNotFound {
			t.Errorf("%s: got %d, want 404", r, rec.Code)
		}
	}
}
