package app_test

import (
	"net/http"
	"strings"
	"testing"

	"github.com/google/uuid"

	orderiface "github.com/krnskrll/event-mpos/internal/order/iface"
	"github.com/krnskrll/event-mpos/internal/testutil"
	useriface "github.com/krnskrll/event-mpos/internal/user/iface"
)

// Written from the attacker's side: a store-B user naming store-A ids must get 404 (never a hint the
// row exists), and store-A rows must not appear in store-B lists.
func TestStoreIsolation(t *testing.T) {
	a := testutil.NewApp(t)
	const pw = "Passw0rd!"

	// Store A (default) with one of everything.
	adminA := a.CreateUser(useriface.RoleAdmin, pw)
	baA := a.CreateUser(useriface.RoleBA, pw)
	evA, pA := a.ActiveEvent(), a.Product("10000")
	a.Allocate(evA.ID, pA.ID, 10)
	orderA, err := a.Orders.Submit(evA.ID, baA.ID, orderiface.SubmitOrderRequest{
		IdempotencyKey: testutil.Unique("iso-"), Customer: orderiface.CustomerInfo{Name: "A"},
		Lines: []orderiface.OrderLineRequest{{ProductID: pA.ID, Qty: 1}},
	})
	if err != nil {
		t.Fatal(err)
	}
	tokA := login(t, a, adminA.Email, pw).AccessToken
	promoRec := a.Do(http.MethodPost, "/api/admin/events/"+evA.ID.String()+"/promotions",
		map[string]any{"name": "A10", "type": "PERCENTAGE", "percentOff": 10, "scope": "ALL", "active": true, "priority": 1}, tokA)
	promoA := testutil.Decode[map[string]any](t, promoRec)["id"].(string)
	qrA := testutil.Unique("QR-A-")
	a.DB.Exec(`INSERT INTO members (id, member_code, qr_code, name, tier, active, updated_at, store_id)
		VALUES (?, ?, ?, 'Member A', 'GOLD', true, now(), ?)`, uuid.New(), testutil.Unique("M-A-"), qrA, evA.StoreID)

	// Store B with its own users and event.
	storeB := a.NewStore("Store B")
	adminB, whB, baB := a.CreateUserIn(storeB, useriface.RoleAdmin, pw), a.CreateUserIn(storeB, useriface.RoleWarehouse, pw), a.CreateUserIn(storeB, useriface.RoleBA, pw)
	evB := a.ActiveEventIn(storeB)
	tokB, tokWhB, tokBaB := login(t, a, adminB.Email, pw).AccessToken, login(t, a, whB.Email, pw).AccessToken, login(t, a, baB.Email, pw).AccessToken

	ev, p, ord := evA.ID.String(), pA.ID.String(), orderA.ID.String()
	refused := []struct {
		name, method, path, token string
		body                      any
	}{
		{"update event", http.MethodPut, "/api/admin/events/" + ev, tokB, map[string]string{"name": "x"}},
		{"change event status", http.MethodPatch, "/api/admin/events/" + ev + "/status", tokB, map[string]string{"status": "CLOSED"}},
		{"assign own BA to A's event", http.MethodPost, "/api/admin/assignments", tokB, map[string]string{"userId": baB.ID.String(), "eventId": ev}},
		{"assign A's BA to own event", http.MethodPost, "/api/admin/assignments", tokB, map[string]string{"userId": baA.ID.String(), "eventId": evB.ID.String()}},
		{"get product", http.MethodGet, "/api/admin/products/" + p, tokB, nil},
		{"update product", http.MethodPut, "/api/admin/products/" + p, tokB, map[string]string{"name": "x"}},
		{"event products", http.MethodGet, "/api/admin/events/" + ev + "/products", tokB, nil},
		{"set availability", http.MethodPut, "/api/admin/events/" + ev + "/products/" + p, tokB, map[string]bool{"available": false}},
		{"update user", http.MethodPatch, "/api/admin/users/" + adminA.ID.String(), tokB, map[string]bool{"active": false}},
		{"allocations", http.MethodGet, "/api/admin/events/" + ev + "/stock", tokB, nil},
		{"set allocation", http.MethodPut, "/api/admin/events/" + ev + "/stock/" + p, tokB, map[string]int{"allocatedQty": 1}},
		{"list promotions", http.MethodGet, "/api/admin/events/" + ev + "/promotions", tokB, nil},
		{"create promotion", http.MethodPost, "/api/admin/events/" + ev + "/promotions", tokB, map[string]any{"name": "x", "type": "FIXED", "amountOff": 1}},
		{"update promotion", http.MethodPut, "/api/admin/promotions/" + promoA, tokB, map[string]any{"name": "x", "type": "FIXED", "amountOff": 1}},
		{"delete promotion", http.MethodDelete, "/api/admin/promotions/" + promoA, tokB, nil},
		{"report", http.MethodGet, "/api/admin/events/" + ev + "/report", tokB, nil},
		{"report csv", http.MethodGet, "/api/admin/events/" + ev + "/report.csv", tokB, nil},
		{"audit", http.MethodGet, "/api/admin/events/" + ev + "/audit", tokB, nil},
		{"warehouse stock", http.MethodGet, "/api/warehouse/events/" + ev + "/stock", tokWhB, nil},
		{"warehouse orders", http.MethodGet, "/api/warehouse/events/" + ev + "/orders", tokWhB, nil},
		{"advance order", http.MethodPatch, "/api/warehouse/orders/" + ord + "/status", tokWhB, map[string]string{"status": "PACKING"}},
		{"cancel order", http.MethodPost, "/api/warehouse/orders/" + ord + "/cancel", tokWhB, nil},
		{"member by QR", http.MethodGet, "/api/me/members/by-qr/" + qrA, tokBaB, nil},
		{"order board socket", http.MethodGet, "/ws/events/" + ev + "/orders?token=" + tokWhB, "", nil},
	}
	for _, r := range refused {
		if rec := a.Do(r.method, r.path, r.body, r.token); rec.Code != http.StatusNotFound {
			t.Errorf("%s: store B got %d, want 404 (%s)", r.name, rec.Code, rec.Body.String())
		}
	}

	hidden := []struct{ path, token, id string }{
		{"/api/admin/events", tokB, ev},
		{"/api/admin/products", tokB, p},
		{"/api/admin/users", tokB, adminA.ID.String()},
		{"/api/admin/members", tokB, qrA},
		{"/api/warehouse/events", tokWhB, ev},
		{"/api/me/members", tokBaB, qrA},
	}
	for _, h := range hidden {
		rec := a.Do(http.MethodGet, h.path, nil, h.token)
		if rec.Code != http.StatusOK || strings.Contains(rec.Body.String(), h.id) {
			t.Errorf("%s: store B sees store A data (%d)", h.path, rec.Code)
		}
	}

	// Store A still reaches its own data, so the 404s above are about the store, not the route.
	for _, path := range []string{"/api/admin/products/" + p, "/api/admin/events/" + ev + "/stock", "/api/admin/events/" + ev + "/audit"} {
		if rec := a.Do(http.MethodGet, path, nil, tokA); rec.Code != http.StatusOK {
			t.Errorf("store A %s: %d", path, rec.Code)
		}
	}
}
