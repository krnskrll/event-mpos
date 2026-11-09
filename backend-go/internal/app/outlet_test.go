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
	create := func(body J) (int, J) {
		body["name"] = "Outlet Kemang"
		rec := a.Back(http.MethodPost, "/locations", body, bhoomyStore)
		if rec.Code != http.StatusCreated {
			return rec.Code, nil
		}
		return rec.Code, testutil.Decode[J](t, rec)
	}

	for name, body := range map[string]J{
		"outlet without address": {"type": "outlet"},
		"outlet with end date":   {"type": "outlet", "address": "Jl. Kemang 1", "endDate": "2026-12-31"},
		"unknown type":           {"type": "kiosk"},
	} {
		if code, _ := create(body); code != http.StatusBadRequest {
			t.Errorf("%s: got %d, want 400", name, code)
		}
	}

	code, ev := create(J{"type": "event", "startDate": "2027-01-01", "endDate": "2027-01-07"})
	if code != http.StatusCreated || ev["type"] != "event" || ev["storeId"] != store.String() {
		t.Fatalf("event: %d %v", code, ev)
	}
	code, shop := create(J{"type": "outlet", "address": " Jl. Kemang 1 "})
	if code != http.StatusCreated || shop["type"] != "outlet" || shop["address"] != "Jl. Kemang 1" {
		t.Fatalf("shop: %d %v", code, shop)
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
