package app_test

import (
	"net/http"
	"net/url"
	"testing"

	"github.com/google/uuid"

	"github.com/krnskrll/event-mpos/internal/testutil"
)

// Phase 6a: the outlet is the Cabang. One open outlet per POS store, created on the first visit
// and named after the Cabang; the back office cannot add another.

func backAs(a *testutil.TestApp, method, path string, body any, bhoomyStore uuid.UUID, name, address string) int {
	return a.Send(method, "/api/backoffice"+path, body, map[string]string{
		"Authorization": "Bearer " + testutil.ServiceToken, "X-Bhoomy-Store-Id": bhoomyStore.String(),
		"X-Bhoomy-Store-Name": url.QueryEscape(name), "X-Bhoomy-Store-Address": url.QueryEscape(address),
	}).Code
}

// outletOf returns the store's outlet (created on the first visit).
func outletOf(t *testing.T, a *testutil.TestApp, bhoomyStore uuid.UUID) J {
	t.Helper()
	for _, l := range ok[[]J](t, a.Back(http.MethodGet, "/locations", nil, bhoomyStore), http.StatusOK) {
		if l["type"] == "outlet" && l["status"] != "CLOSED" {
			return l
		}
	}
	t.Fatal("no open outlet")
	return nil
}

func TestOutletIsTheCabang(t *testing.T) {
	a := testutil.NewApp(t)
	store := uuid.New()

	if code := backAs(a, http.MethodGet, "/locations", nil, store, "Glow Mart Bandung", "Jl. Dago 12, Bandung"); code != http.StatusOK {
		t.Fatalf("first visit: %d", code)
	}
	outlets := func() []J {
		var out []J
		for _, l := range ok[[]J](t, a.Back(http.MethodGet, "/locations", nil, store), http.StatusOK) {
			if l["type"] == "outlet" {
				out = append(out, l)
			}
		}
		return out
	}
	got := outlets()
	if len(got) != 1 || got[0]["name"] != "Glow Mart Bandung" || got[0]["code"] != "OUT-GMB" || got[0]["status"] != "DRAFT" ||
		got[0]["address"] != "Jl. Dago 12, Bandung" {
		t.Fatalf("want one draft outlet named and addressed after the Cabang, got %v", got)
	}

	t.Run("GET /outlet returns it", func(t *testing.T) {
		if o := ok[J](t, a.Back(http.MethodGet, "/outlet", nil, store), http.StatusOK); o["id"] != got[0]["id"] {
			t.Fatalf("outlet = %v", o)
		}
	})

	t.Run("visits do not add more", func(t *testing.T) {
		a.Back(http.MethodGet, "/locations", nil, store)
		if n := len(outlets()); n != 1 {
			t.Fatalf("outlets = %d, want 1", n)
		}
	})

	t.Run("the back office cannot add an outlet", func(t *testing.T) {
		body := J{"type": "outlet", "name": "Outlet Kedua", "address": "Jl. Dago 1"}
		if code := a.Back(http.MethodPost, "/locations", body, store).Code; code != http.StatusConflict {
			t.Fatalf("got %d, want 409", code)
		}
	})

	t.Run("events are added as before", func(t *testing.T) {
		body := J{"type": "event", "name": "Bazaar Dago", "startDate": "2027-01-01", "endDate": "2027-01-03"}
		ok[J](t, a.Back(http.MethodPost, "/locations", body, store), http.StatusCreated)
	})

	t.Run("the outlet pauses instead of closing", func(t *testing.T) {
		id := got[0]["id"].(string)
		status := func(s string) int {
			return a.Back(http.MethodPatch, "/locations/"+id+"/status", J{"status": s}, store).Code
		}
		if code := status("CLOSED"); code != http.StatusConflict {
			t.Fatalf("draft → closed: %d, want 409", code)
		}
		for _, s := range []string{"ACTIVE", "DRAFT", "ACTIVE"} {
			if code := status(s); code != http.StatusOK {
				t.Fatalf("→ %s: %d", s, code)
			}
		}
		if code := status("CLOSED"); code != http.StatusConflict {
			t.Fatalf("active → closed: %d, want 409", code)
		}
		if o := outletOf(t, a, store); o["id"] != id {
			t.Fatalf("the same outlet stays, got %v", o)
		}
	})

	t.Run("the store takes the Cabang name once Bhoomy sends it", func(t *testing.T) {
		other := uuid.New()
		a.Back(http.MethodGet, "/locations", nil, other)
		backAs(a, http.MethodGet, "/locations", nil, other, "Toko Bumi 2", "")
		var name string
		a.DB.Raw("SELECT name FROM pos_stores WHERE bhoomy_store_id = ?", other).Scan(&name)
		if name != "Toko Bumi 2" {
			t.Fatalf("store name = %q", name)
		}
	})

	t.Run("the database allows one open outlet per store", func(t *testing.T) {
		err := a.DB.Exec(`INSERT INTO pos_locations (id, store_id, type, code, name, status, created_at, updated_at)
			SELECT gen_random_uuid(), store_id, 'outlet', 'OUT-DUP', 'Dup', 'DRAFT', now(), now()
			  FROM pos_locations WHERE type = 'outlet' AND status <> 'CLOSED' LIMIT 1`).Error
		if err == nil {
			t.Fatal("a second open outlet was accepted")
		}
	})
}
