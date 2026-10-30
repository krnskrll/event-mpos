package app_test

import (
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/krnskrll/event-mpos/internal/event"
	"github.com/krnskrll/event-mpos/internal/order"
	"github.com/krnskrll/event-mpos/internal/testutil"
	"github.com/krnskrll/event-mpos/internal/user"
)

// EPIC-12 C1/C2: device number blocks (D1) and the late-sync window after close (D2).
func TestOfflineOrdersAndLateSync(t *testing.T) {
	a := testutil.NewApp(t)
	ev, p := a.ActiveEvent(), a.Product("10.00")
	a.Allocate(ev.ID, p.ID, 2)
	ba := a.CreateUser(user.RoleBA, "Ba123456!")
	device := "device-" + uuid.NewString()
	deviceOrder := func(n int, createdAt time.Time) (*order.OrderConfirmation, error) {
		return a.Orders.Submit(ev.ID, ba.ID, order.SubmitOrderRequest{
			IdempotencyKey: uuid.NewString(), Customer: order.CustomerInfo{Name: "Offline"},
			Lines:    []order.OrderLineRequest{{ProductID: p.ID, Qty: 2}},
			DeviceID: &device, OrderNumber: &n, ClientCreatedAt: &createdAt,
		})
	}

	// Blocks and server numbers come off one sequence, so they never overlap.
	b1, err := a.Orders.IssueNumberBlock(ev.ID, device)
	if err != nil || b1.RangeStart != 1 || b1.RangeEnd != order.NumberBlockSize {
		t.Fatalf("first block = %+v, %v", b1, err)
	}
	server, err := submit(a, ev.ID, uuid.NewString(), line(p.ID, 1))
	if err != nil || server.OrderNumber != order.NumberBlockSize+1 {
		t.Fatalf("server number after a block = %v, %v; want %d", server, err, order.NumberBlockSize+1)
	}

	// A device-numbered order keeps its number and is never rejected for stock: 1 left, 2 sold.
	conf, err := deviceOrder(3, time.Now())
	if err != nil {
		t.Fatal(err)
	}
	if conf.OrderNumber != 3 || !strings.HasSuffix(conf.ReferenceID, "-003") || !conf.StockConflict {
		t.Errorf("offline order = #%d %s conflict=%v; want #3, -003, conflict", conf.OrderNumber, conf.ReferenceID, conf.StockConflict)
	}
	if got := a.Remaining(ev.ID, p.ID); got != -1 {
		t.Errorf("remaining = %d, want -1 (the shortfall stays visible)", got)
	}
	for _, n := range []int{order.NumberBlockSize + 1, 999} { // outside this device's blocks
		if _, err := deviceOrder(n, time.Now()); testutil.StatusOf(err) != http.StatusBadRequest {
			t.Errorf("number %d outside the block: %v, want 400", n, err)
		}
	}
	if _, err := deviceOrder(3, time.Now()); testutil.StatusOf(err) != http.StatusConflict {
		t.Errorf("reused number: %v, want 409", err)
	}

	// D2: after close, only orders created before closed_at and uploaded within 24h are accepted.
	beforeClose := time.Now().Add(-time.Minute)
	closeEvent(t, a, ev.ID)
	late, err := deviceOrder(4, beforeClose)
	if err != nil || !late.LateSynced {
		t.Fatalf("queued order after close: %+v, %v; want accepted + lateSynced", late, err)
	}
	if _, err := deviceOrder(5, time.Now().Add(time.Minute)); testutil.StatusOf(err) != http.StatusConflict {
		t.Errorf("order created after close: %v, want 409", err)
	}
	a.DB.Model(&event.Event{}).Where("id = ?", ev.ID).Update("closed_at", time.Now().Add(-25*time.Hour))
	if _, err := deviceOrder(6, time.Now().Add(-26*time.Hour)); testutil.StatusOf(err) != http.StatusConflict {
		t.Errorf("upload more than 24h after close: %v, want 409", err)
	}
	if _, err := a.Orders.IssueNumberBlock(ev.ID, device); testutil.StatusOf(err) != http.StatusConflict {
		t.Errorf("number block for a closed event: %v, want 409", err)
	}

	r, err := a.Reports.EndOfEvent(ev.ID)
	if err != nil || r.LateSyncedOrders != 1 || r.StockConflictOrders != 2 {
		t.Errorf("report late/conflict = %d/%d (%v), want 1/2", r.LateSyncedOrders, r.StockConflictOrders, err)
	}
}
