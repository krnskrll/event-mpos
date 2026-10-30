package app_test

import (
	"net/http"
	"testing"

	"github.com/google/uuid"

	"github.com/krnskrll/event-mpos/internal/audit"
	"github.com/krnskrll/event-mpos/internal/order"
	"github.com/krnskrll/event-mpos/internal/testutil"
	"github.com/krnskrll/event-mpos/internal/user"
)

// EPIC-12 C4: cancel restores stock, writes ORDER_CANCELLED, follows per-role rules, and leaves the report.
func TestCancelOrder(t *testing.T) {
	a := testutil.NewApp(t)
	ev, p := a.ActiveEvent(), a.Product("10.00")
	a.Allocate(ev.ID, p.ID, 10)
	owner, other := a.CreateUser(user.RoleBA, "Ba123456!"), a.CreateUser(user.RoleBA, "Ba123456!")
	wh := a.CreateUser(user.RoleWarehouse, "Wh123456!")
	place := func(qty int, status order.OrderStatus) uuid.UUID {
		conf, err := a.Orders.Submit(ev.ID, owner.ID, order.SubmitOrderRequest{IdempotencyKey: uuid.NewString(),
			Customer: order.CustomerInfo{Name: "C"}, Lines: []order.OrderLineRequest{{ProductID: p.ID, Qty: qty}}})
		if err != nil {
			t.Fatal(err)
		}
		for st := order.StatusNew; st != status; {
			next := map[order.OrderStatus]order.OrderStatus{order.StatusNew: order.StatusPacking, order.StatusPacking: order.StatusReady}[st]
			if _, err := a.Orders.AdvanceStatus(conf.ID, next, &wh.ID); err != nil {
				t.Fatal(err)
			}
			st = next
		}
		return conf.ID
	}
	reason := "customer changed mind"
	newOrder, packing, ready := place(3, order.StatusNew), place(2, order.StatusPacking), place(1, order.StatusReady)

	cases := []struct {
		name  string
		order uuid.UUID
		actor *user.UserResponse
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
	if _, err := a.Orders.AdvanceStatus(packing, order.StatusReady, &wh.ID); testutil.StatusOf(err) != http.StatusConflict {
		t.Errorf("advancing a cancelled order: %v, want 409", err)
	}
	trail, _ := a.Audit.ForOrder(newOrder)
	if last := trail[len(trail)-1]; last.Action != audit.ActionOrderCancelled || *last.Detail != "Order cancelled: "+reason {
		t.Errorf("last audit entry = %+v", last)
	}
	// The BA endpoint goes through the same rules: the order is already cancelled → 409.
	tok := login(t, a, owner.Email, "Ba123456!").AccessToken
	if rec := a.Do(http.MethodPost, "/api/me/orders/"+newOrder.String()+"/cancel", nil, tok); rec.Code != http.StatusConflict {
		t.Errorf("BA cancel endpoint on a cancelled order: %d, want 409", rec.Code)
	}

	closeEvent(t, a, ev.ID)
	r, err := a.Reports.EndOfEvent(ev.ID)
	if err != nil {
		t.Fatal(err)
	}
	if r.TotalOrders != 1 || r.CancelledOrders != 2 || r.TotalItems != 1 || !r.TotalRevenue.Equal(dec("10")) {
		t.Errorf("report orders/cancelled/items/revenue = %d/%d/%d/%s, want 1/2/1/10", r.TotalOrders, r.CancelledOrders, r.TotalItems, r.TotalRevenue)
	}
}
