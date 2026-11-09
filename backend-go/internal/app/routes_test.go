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

// The old admin surface that Bhoomy's back office replaced is gone; what Bhoomy does not own yet stays.
func TestRetiredAdminRoutes(t *testing.T) {
	a := testutil.NewApp(t)
	tok := adminToken(t, a)
	ev, id := a.ActiveEvent().ID.String(), uuid.NewString()
	for _, r := range []string{
		"POST /events", "PUT /events/" + ev, "PATCH /events/" + ev + "/status",
		"GET /products", "POST /products", "GET /products/" + id, "PUT /products/" + id, "POST /products/" + id + "/photo",
		"POST /erp/sync", "GET /erp/status", "GET /members",
		"GET /events/" + ev + "/products", "PUT /events/" + ev + "/products/" + id, "PUT /events/" + ev + "/variants/" + id,
		"GET /events/" + ev + "/report", "GET /events/" + ev + "/report.csv", "GET /events/" + ev + "/audit",
		"GET /events/" + ev + "/promotions", "POST /events/" + ev + "/promotions", "PUT /promotions/" + id, "DELETE /promotions/" + id,
		// Phase 5: accounts, cashier assignment and member sync moved to Bhoomy (Staf POS, Member POS).
		"GET /users", "POST /users", "PATCH /users/" + id, "POST /assignments",
		"POST /members/sync", "GET /members/sync/status", "GET /sync", "PUT /sync/MEMBER",
	} {
		method, path, _ := strings.Cut(r, " ")
		if rec := a.Do(method, "/api/admin"+path, map[string]any{}, tok); rec.Code != http.StatusNotFound {
			t.Errorf("%s: got %d, want 404", r, rec.Code)
		}
	}
	for _, path := range []string{"/events", "/events/" + ev + "/stock", "/catalog/sync"} {
		if rec := a.Do(http.MethodGet, "/api/admin"+path, nil, tok); rec.Code != http.StatusOK {
			t.Errorf("GET %s: got %d, want 200", path, rec.Code)
		}
	}
	if rec := a.Do(http.MethodGet, "/api/me", nil, tok); rec.Code != http.StatusNotFound {
		t.Errorf("GET /api/me: got %d, want 404 (GET /api/device replaces it)", rec.Code)
	}
}
