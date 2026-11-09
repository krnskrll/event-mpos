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
	} {
		method, path, _ := strings.Cut(r, " ")
		if rec := a.Do(method, "/api/admin"+path, map[string]any{}, tok); rec.Code != http.StatusNotFound {
			t.Errorf("%s: got %d, want 404", r, rec.Code)
		}
	}
	for _, path := range []string{"/users", "/events", "/events/" + ev + "/stock", "/catalog/sync", "/members/sync/status"} {
		if rec := a.Do(http.MethodGet, "/api/admin"+path, nil, tok); rec.Code != http.StatusOK {
			t.Errorf("GET %s: got %d, want 200", path, rec.Code)
		}
	}

	overview := testutil.Decode[map[string]any](t, a.Do(http.MethodGet, "/api/admin/sync", nil, tok))
	if _, erp := overview["erp"]; erp || overview["member"] == nil {
		t.Errorf("sync overview %v, want the member sync only", overview)
	}
	rec := a.Do(http.MethodPut, "/api/admin/sync/ERP", map[string]any{"enabled": true, "cron": "0 0 * * * *"}, tok)
	if rec.Code != http.StatusBadRequest || !strings.Contains(rec.Body.String(), "retired") {
		t.Errorf("ERP schedule: %d %s, want 400 retired", rec.Code, rec.Body.String())
	}
}
