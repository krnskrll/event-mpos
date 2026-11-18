package app_test

import (
	"net/http"
	"strings"
	"testing"

	"github.com/google/uuid"

	orderiface "github.com/krnskrll/event-mpos/internal/order/iface"
	promotioniface "github.com/krnskrll/event-mpos/internal/promotion/iface"
	staffiface "github.com/krnskrll/event-mpos/internal/staff/iface"
	"github.com/krnskrll/event-mpos/internal/testutil"
)

// Written from the attacker's side: store B (its devices, or Bhoomy acting for it) naming store-A ids
// must get 404 (never a hint the row exists), and store-A rows must not appear in store-B lists.
func TestStoreIsolation(t *testing.T) {
	a := testutil.NewApp(t)
	bhoomyA, bhoomyB := uuid.New(), uuid.New()
	storeA, storeB := posStore(t, a, bhoomyA), posStore(t, a, bhoomyB)

	// Store A with one of everything.
	evA, pA := a.ActiveEventIn(storeA), a.ProductIn(storeA, "10000")
	staffA, devA := a.NewStaff(storeA, staffiface.RoleKasir, evA.ID), a.PairDevice(evA.ID)
	orderA, err := a.Orders.Submit(evA.ID, staffA.ID, orderiface.SubmitOrderRequest{
		IdempotencyKey: testutil.Unique("iso-"), Customer: orderiface.CustomerInfo{Name: "A"},
		Lines: []orderiface.OrderLineRequest{{ProductID: pA.ID, Qty: 1}},
	})
	if err != nil {
		t.Fatal(err)
	}
	pct := dec("10")
	promo, err := a.Promotions.Create(storeA, promotioniface.PromotionRequest{LocationID: &evA.ID, Name: "A10", Type: promotioniface.TypePercentage, PercentOff: &pct})
	if err != nil {
		t.Fatal(err)
	}
	qrA, codeA := testutil.Unique("QR-A-"), testutil.Unique("M-A-")
	a.DB.Exec(`INSERT INTO members (id, member_code, qr_code, name, tier, active, updated_at, store_id)
		VALUES (?, ?, ?, 'Member A', 'GOLD', true, now(), ?)`, uuid.New(), codeA, qrA, storeA)

	// Store B with a paired device and its staff.
	evB := a.ActiveEventIn(storeB)
	devB, staffB := a.PairDevice(evB.ID), a.NewStaff(storeB, staffiface.RoleSupervisor, evB.ID).ID

	ev, p, ord, pr := evA.ID.String(), pA.ID.String(), orderA.ID.String(), promo.ID.String()
	// Store B's device: another location's id is 403, another store's order or member 404.
	cancel := J{"supervisorId": staffB, "supervisorPin": testutil.TestPin}
	for _, r := range []struct {
		name, method, path string
		body               any
		want               int
	}{
		{"snapshot", http.MethodGet, "/api/me/events/" + ev + "/snapshot", nil, http.StatusForbidden},
		{"submit order", http.MethodPost, "/api/me/events/" + ev + "/orders", J{"idempotencyKey": "k", "lines": []J{{"productId": p, "qty": 1}}}, http.StatusForbidden},
		{"board orders", http.MethodGet, "/api/warehouse/events/" + ev + "/orders", nil, http.StatusForbidden},
		{"advance order", http.MethodPatch, "/api/warehouse/orders/" + ord + "/status", J{"status": "PACKING"}, http.StatusNotFound},
		{"board cancel", http.MethodPost, "/api/warehouse/orders/" + ord + "/cancel", cancel, http.StatusNotFound},
		{"cashier cancel", http.MethodPost, "/api/me/orders/" + ord + "/cancel", cancel, http.StatusNotFound},
		{"member by QR", http.MethodGet, "/api/me/members/by-qr/" + qrA, nil, http.StatusNotFound},
		{"order board socket", http.MethodGet, "/ws/events/" + ev + "/orders?token=" + devB.Token, nil, http.StatusForbidden},
	} {
		if rec := a.DeviceDo(r.method, r.path, r.body, devB, staffB); rec.Code != r.want {
			t.Errorf("device %s: store B got %d, want %d (%s)", r.name, rec.Code, r.want, rec.Body.String())
		}
	}

	backRefused := []struct {
		name, method, path string
		body               any
	}{
		{"update location", http.MethodPut, "/locations/" + ev, J{"name": "x"}},
		{"change location status", http.MethodPatch, "/locations/" + ev + "/status", J{"status": "CLOSED"}},
		{"location items", http.MethodGet, "/locations/" + ev + "/items", nil},
		{"save location items", http.MethodPut, "/locations/" + ev + "/items", J{"rows": []J{}}},
		{"get promotion", http.MethodGet, "/promotions/" + pr, nil},
		{"update promotion", http.MethodPut, "/promotions/" + pr, J{"name": "x", "type": "FIXED", "amountOff": 1}},
		{"delete promotion", http.MethodDelete, "/promotions/" + pr, nil},
		{"report", http.MethodGet, "/locations/" + ev + "/report", nil},
		{"audit", http.MethodGet, "/locations/" + ev + "/audit", nil},
		{"location devices", http.MethodGet, "/locations/" + ev + "/devices", nil},
		{"pairing code", http.MethodPost, "/locations/" + ev + "/pairing-code", nil},
		{"staff", http.MethodGet, "/staff/" + staffA.ID.String(), nil},
		{"update staff", http.MethodPut, "/staff/" + staffA.ID.String(), J{"name": "x", "role": "KASIR", "locationIds": []string{}}},
		{"rename device", http.MethodPatch, "/devices/" + devA.ID.String(), J{"name": "x"}},
		{"remove device", http.MethodDelete, "/devices/" + devA.ID.String(), nil},
	}
	for _, r := range backRefused {
		if rec := a.Back(r.method, r.path, r.body, bhoomyB); rec.Code != http.StatusNotFound {
			t.Errorf("back office %s: store B got %d, want 404 (%s)", r.name, rec.Code, rec.Body.String())
		}
	}

	for path, id := range map[string]string{"/api/device/staff": staffA.ID.String()} {
		rec := a.DeviceDo(http.MethodGet, path, nil, devB, uuid.Nil)
		if rec.Code != http.StatusOK || strings.Contains(rec.Body.String(), id) {
			t.Errorf("device %s: store B sees store A data (%d)", path, rec.Code)
		}
	}
	for path, id := range map[string]string{"/locations": ev, "/promotions": pr, "/members": codeA, "/staff": staffA.ID.String()} {
		rec := a.Back(http.MethodGet, path, nil, bhoomyB)
		if rec.Code != http.StatusOK || strings.Contains(rec.Body.String(), id) {
			t.Errorf("back office %s: store B sees store A data (%d)", path, rec.Code)
		}
	}

	// Store A still reaches its own data, so the 404s above are about the store, not the route.
	for _, path := range []string{"/locations/" + ev + "/audit", "/locations/" + ev + "/report", "/promotions/" + pr} {
		if rec := a.Back(http.MethodGet, path, nil, bhoomyA); rec.Code != http.StatusOK {
			t.Errorf("store A %s: %d", path, rec.Code)
		}
	}
}
