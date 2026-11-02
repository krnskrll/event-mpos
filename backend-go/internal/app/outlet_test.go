package app_test

import (
	"net/http"
	"testing"

	"github.com/krnskrll/event-mpos/internal/platform/tenant"
	"github.com/krnskrll/event-mpos/internal/testutil"
	useriface "github.com/krnskrll/event-mpos/internal/user/iface"
)

func TestOutletRules(t *testing.T) {
	a := testutil.NewApp(t)
	admin := a.CreateUser(useriface.RoleAdmin, "Passw0rd!")
	tok := login(t, a, admin.Email, "Passw0rd!").AccessToken
	create := func(body map[string]any) (int, map[string]any) {
		body["code"], body["name"] = testutil.Unique("O"), "Outlet"
		rec := a.Do(http.MethodPost, "/api/admin/events", body, tok)
		if rec.Code != http.StatusCreated {
			return rec.Code, nil
		}
		return rec.Code, testutil.Decode[map[string]any](t, rec)
	}

	for name, body := range map[string]map[string]any{
		"store without address": {"type": "store"},
		"store with end date":   {"type": "store", "address": "Jl. Kemang 1", "endDate": "2026-12-31"},
		"unknown type":          {"type": "kiosk"},
	} {
		if code, _ := create(body); code != http.StatusBadRequest {
			t.Errorf("%s: got %d, want 400", name, code)
		}
	}

	code, ev := create(map[string]any{})
	if code != http.StatusCreated || ev["type"] != "event" || ev["storeId"] != tenant.DefaultStoreID.String() {
		t.Fatalf("default outlet: %d %v", code, ev)
	}
	code, shop := create(map[string]any{"type": "store", "address": " Jl. Kemang 1 "})
	if code != http.StatusCreated || shop["type"] != "store" || shop["address"] != "Jl. Kemang 1" {
		t.Fatalf("shop: %d %v", code, shop)
	}

	// Stock: an active event is locked; an active shop keeps restocking until it closes.
	p := a.Product("5000")
	for _, o := range []map[string]any{ev, shop} {
		if rec := a.Do(http.MethodPatch, "/api/admin/events/"+o["id"].(string)+"/status", map[string]string{"status": "ACTIVE"}, tok); rec.Code != http.StatusOK {
			t.Fatalf("activate: %d", rec.Code)
		}
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
	a.Do(http.MethodPatch, "/api/admin/events/"+shop["id"].(string)+"/status", map[string]string{"status": "CLOSED"}, tok)
	if got := allocate(shop["id"].(string)); got != http.StatusConflict {
		t.Errorf("closed shop allocation: got %d, want 409", got)
	}
}
