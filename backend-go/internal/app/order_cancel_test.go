package app_test

import (
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"

	auditiface "github.com/krnskrll/event-mpos/internal/audit/iface"
	orderiface "github.com/krnskrll/event-mpos/internal/order/iface"
	reportiface "github.com/krnskrll/event-mpos/internal/report/iface"
	staffiface "github.com/krnskrll/event-mpos/internal/staff/iface"
	"github.com/krnskrll/event-mpos/internal/testutil"
)

// Cancelling restores stock, writes ORDER_CANCELLED naming cashier and supervisor, needs a
// supervisor of the location with the right PIN, and leaves the report.
func TestCancelOrder(t *testing.T) {
	a := testutil.NewApp(t)
	ev, elsewhere, p := a.ActiveEvent(), a.ActiveEvent(), a.Product("10.00")
	a.Allocate(ev.ID, p.ID, 10)
	kasir, sari := a.NewStaff(ev.StoreID, staffiface.RoleKasir, ev.ID), a.NewStaff(ev.StoreID, staffiface.RoleSupervisor, ev.ID)
	away := a.NewStaff(ev.StoreID, staffiface.RoleSupervisor, elsewhere.ID)
	dev := a.PairDevice(ev.ID)
	place := func(qty int, status orderiface.OrderStatus) uuid.UUID {
		conf, err := a.Orders.Submit(ev.ID, kasir.ID, orderiface.SubmitOrderRequest{IdempotencyKey: uuid.NewString(),
			Customer: orderiface.CustomerInfo{Name: "C"}, Lines: []orderiface.OrderLineRequest{{ProductID: p.ID, Qty: qty}}})
		if err != nil {
			t.Fatal(err)
		}
		for st := orderiface.StatusNew; st != status; {
			next := map[orderiface.OrderStatus]orderiface.OrderStatus{orderiface.StatusNew: orderiface.StatusPacking, orderiface.StatusPacking: orderiface.StatusReady}[st]
			if _, err := a.Orders.AdvanceStatus(conf.ID, next, &kasir.ID); err != nil {
				t.Fatal(err)
			}
			st = next
		}
		return conf.ID
	}
	newOrder, packing, ready := place(3, orderiface.StatusNew), place(2, orderiface.StatusPacking), place(1, orderiface.StatusReady)
	cashier := func(id uuid.UUID) string { return "/api/me/orders/" + id.String() + "/cancel" }
	board := func(id uuid.UUID) string { return "/api/warehouse/orders/" + id.String() + "/cancel" }
	approve := func(by uuid.UUID, pin string) J {
		return J{"reason": "customer changed mind", "supervisorId": by, "supervisorPin": pin}
	}

	cases := []struct {
		name, path string
		body       J
		want       int
		message    string
	}{
		{"wrong PIN", cashier(newOrder), approve(sari.ID, "0000"), http.StatusForbidden, "Wrong supervisor PIN"},
		{"a kasir cannot approve", cashier(newOrder), approve(kasir.ID, testutil.TestPin), http.StatusForbidden, "supervisor"},
		{"another location's supervisor", cashier(newOrder), approve(away.ID, testutil.TestPin), http.StatusForbidden, "supervisor"},
		{"no supervisor", cashier(newOrder), J{"reason": "x"}, http.StatusBadRequest, "supervisorId"},
		{"cashier history cannot cancel once packing", cashier(packing), approve(sari.ID, testutil.TestPin), http.StatusConflict, ""},
		{"board cannot cancel a READY order", board(ready), approve(sari.ID, testutil.TestPin), http.StatusConflict, ""},
		{"board cancels a PACKING order", board(packing), approve(sari.ID, testutil.TestPin), http.StatusOK, ""},
		{"cashier cancels a NEW order", cashier(newOrder), approve(sari.ID, testutil.TestPin), http.StatusOK, ""},
		{"cancelling twice is a conflict", board(newOrder), approve(sari.ID, testutil.TestPin), http.StatusConflict, ""},
	}
	for _, c := range cases {
		rec := a.DeviceDo(http.MethodPost, c.path, c.body, dev, kasir.ID)
		if rec.Code != c.want || !strings.Contains(rec.Body.String(), c.message) {
			t.Errorf("%s: %d %s, want %d %q", c.name, rec.Code, rec.Body.String(), c.want, c.message)
		}
	}
	if rec := a.DeviceDo(http.MethodPost, cashier(ready), approve(sari.ID, testutil.TestPin), dev, uuid.Nil); rec.Code != http.StatusBadRequest {
		t.Errorf("cancel without X-Staff-Id: %d, want 400", rec.Code)
	}

	if got := a.Remaining(ev.ID, p.ID); got != 9 { // only the READY order (1 unit) is still sold
		t.Errorf("remaining = %d, want 9", got)
	}
	if _, err := a.Orders.AdvanceStatus(packing, orderiface.StatusReady, &kasir.ID); testutil.StatusOf(err) != http.StatusConflict {
		t.Errorf("advancing a cancelled order: %v, want 409", err)
	}
	trail, _ := a.Audit.ForOrder(newOrder)
	last := trail[len(trail)-1]
	if last.Action != auditiface.ActionOrderCancelled || last.ActorName != kasir.Name ||
		*last.Detail != "Order cancelled: customer changed mind · approved by "+sari.Name {
		t.Errorf("last audit entry = %+v (%s)", last, *last.Detail)
	}

	closeEvent(t, a, ev.ID)
	if r := salesSoFar(t, a, ev.ID); r.TotalOrders != 1 || r.CancelledOrders != 2 || r.TotalItems != 1 || !r.TotalRevenue.Equal(dec("10")) {
		t.Errorf("report orders/cancelled/items/revenue = %d/%d/%d/%s, want 1/2/1/10", r.TotalOrders, r.CancelledOrders, r.TotalItems, r.TotalRevenue)
	}
}

// salesSoFar is the location's sales report over every day up to now.
func salesSoFar(t *testing.T, a *testutil.TestApp, eventID uuid.UUID) *reportiface.SalesReport {
	t.Helper()
	r, err := a.Reports.Sales(eventID, time.Date(2000, 1, 1, 0, 0, 0, 0, time.UTC), time.Now(), time.Now())
	if err != nil {
		t.Fatal(err)
	}
	return r
}
