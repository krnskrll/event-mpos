package app_test

import (
	"net/http"
	"testing"

	"github.com/google/uuid"

	"github.com/krnskrll/event-mpos/internal/testutil"
	useriface "github.com/krnskrll/event-mpos/internal/user/iface"
)

func TestOutletRules(t *testing.T) {
	a := testutil.NewApp(t)
	bhoomyStore := uuid.New()
	store := posStore(t, a, bhoomyStore)
	admin := a.CreateUserIn(store, useriface.RoleAdmin, "Passw0rd!")
	tok := login(t, a, admin.Email, "Passw0rd!").AccessToken
	if code := a.Back(http.MethodPost, "/locations", J{"type": "kiosk", "name": "Kios"}, bhoomyStore).Code; code != http.StatusBadRequest {
		t.Errorf("unknown type: got %d, want 400", code)
	}
	ev := ok[J](t, a.Back(http.MethodPost, "/locations", J{"type": "event", "name": "Bazaar", "startDate": "2027-01-01", "endDate": "2027-01-07"}, bhoomyStore), http.StatusCreated)
	if ev["storeId"] != store.String() {
		t.Fatalf("event: %v", ev)
	}

	// The outlet comes with the Cabang; editing it still needs an address and no end date.
	shop := outletOf(t, a, bhoomyStore)
	put := func(body J) int {
		return a.Back(http.MethodPut, "/locations/"+shop["id"].(string), body, bhoomyStore).Code
	}
	if got := put(J{"endDate": "2026-12-31", "address": "Jl. Kemang 1"}); got != http.StatusBadRequest {
		t.Errorf("outlet with end date: got %d, want 400", got)
	}
	if got := put(J{"address": " Jl. Kemang 1 "}); got != http.StatusOK {
		t.Fatalf("outlet address: got %d", got)
	}
	if shop = outletOf(t, a, bhoomyStore); shop["address"] != "Jl. Kemang 1" {
		t.Errorf("address = %v", shop["address"])
	}

	// Stock: an active event is locked; an active outlet keeps restocking until it closes.
	p := a.ProductIn(store, "5000")
	status := func(id, s string) {
		t.Helper()
		if rec := a.Back(http.MethodPatch, "/locations/"+id+"/status", J{"status": s}, bhoomyStore); rec.Code != http.StatusOK {
			t.Fatalf("%s: %d", s, rec.Code)
		}
	}
	for _, o := range []J{ev, shop} {
		status(o["id"].(string), "ACTIVE")
	}
	allocate := func(id string) int {
		return a.Do(http.MethodPut, "/api/admin/events/"+id+"/stock/"+p.ID.String(), map[string]int{"allocatedQty": 3}, tok).Code
	}
	if got := allocate(ev["id"].(string)); got != http.StatusConflict {
		t.Errorf("active event allocation: got %d, want 409", got)
	}
	if got := allocate(shop["id"].(string)); got != http.StatusOK {
		t.Errorf("active shop allocation: got %d, want 200", got)
	}
	status(shop["id"].(string), "CLOSED")
	if got := allocate(shop["id"].(string)); got != http.StatusConflict {
		t.Errorf("closed shop allocation: got %d, want 409", got)
	}
}
