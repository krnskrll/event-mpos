package app_test

import (
	"net/http"
	"testing"
	"time"

	"github.com/google/uuid"

	auditiface "github.com/krnskrll/event-mpos/internal/audit/iface"
	orderiface "github.com/krnskrll/event-mpos/internal/order/iface"
	reportiface "github.com/krnskrll/event-mpos/internal/report/iface"
	"github.com/krnskrll/event-mpos/internal/testutil"
	useriface "github.com/krnskrll/event-mpos/internal/user/iface"
)

// EPIC-12 C4: cancel restores stock, writes ORDER_CANCELLED, follows per-role rules, and leaves the report.
func TestCancelOrder(t *testing.T) {
	a := testutil.NewApp(t)
	ev, p := a.ActiveEvent(), a.Product("10.00")
	a.Allocate(ev.ID, p.ID, 10)
	owner, other := a.CreateUser(useriface.RoleBA, "Ba123456!"), a.CreateUser(useriface.RoleBA, "Ba123456!")
	wh := a.CreateUser(useriface.RoleWarehouse, "Wh123456!")
	place := func(qty int, status orderiface.OrderStatus) uuid.UUID {
		conf, err := a.Orders.Submit(ev.ID, owner.ID, orderiface.SubmitOrderRequest{IdempotencyKey: uuid.NewString(),
			Customer: orderiface.CustomerInfo{Name: "C"}, Lines: []orderiface.OrderLineRequest{{ProductID: p.ID, Qty: qty}}})
		if err != nil {
			t.Fatal(err)
		}
		for st := orderiface.StatusNew; st != status; {
			next := map[orderiface.OrderStatus]orderiface.OrderStatus{orderiface.StatusNew: orderiface.StatusPacking, orderiface.StatusPacking: orderiface.StatusReady}[st]
			if _, err := a.Orders.AdvanceStatus(conf.ID, next, &wh.ID); err != nil {
				t.Fatal(err)
			}
			st = next
		}
		return conf.ID
	}
	reason := "customer changed mind"
	newOrder, packing, ready := place(3, orderiface.StatusNew), place(2, orderiface.StatusPacking), place(1, orderiface.StatusReady)

	cases := []struct {
		name  string
		order uuid.UUID
		actor *useriface.UserResponse
		want  int
	}{
		{"BA cannot cancel another BA's order", newOrder, other, http.StatusNotFound},
		{"BA cannot cancel once packing", packing, owner, http.StatusConflict},
		{"warehouse cannot cancel a READY order", ready, wh, http.StatusConflict},
		{"warehouse cancels a PACKING order", packing, wh, 0},
		{"BA cancels own NEW order", newOrder, owner, 0},
		{"cancelling twice is a conflict", newOrder, wh, http.StatusConflict},
	}
	for _, c := range cases {
		_, err := a.Orders.Cancel(c.order, c.actor.ID, c.actor.Role, &reason)
		if got := testutil.StatusOf(err); got != c.want || (c.want == 0 && err != nil) {
			t.Errorf("%s: got %v (status %d), want %d", c.name, err, got, c.want)
		}
	}

	if got := a.Remaining(ev.ID, p.ID); got != 9 { // only the READY order (1 unit) is still sold
		t.Errorf("remaining = %d, want 9", got)
	}
	if _, err := a.Orders.AdvanceStatus(packing, orderiface.StatusReady, &wh.ID); testutil.StatusOf(err) != http.StatusConflict {
		t.Errorf("advancing a cancelled order: %v, want 409", err)
	}
	trail, _ := a.Audit.ForOrder(newOrder)
	if last := trail[len(trail)-1]; last.Action != auditiface.ActionOrderCancelled || *last.Detail != "Order cancelled: "+reason {
		t.Errorf("last audit entry = %+v", last)
	}
	// The BA endpoint goes through the same rules: the order is already cancelled → 409.
	tok := login(t, a, owner.Email, "Ba123456!").AccessToken
	if rec := a.Do(http.MethodPost, "/api/me/orders/"+newOrder.String()+"/cancel", nil, tok); rec.Code != http.StatusConflict {
		t.Errorf("BA cancel endpoint on a cancelled order: %d, want 409", rec.Code)
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
