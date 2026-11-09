package app_test

import (
	"net/http"
	"testing"
	"time"

	"github.com/google/uuid"

	auditiface "github.com/krnskrll/event-mpos/internal/audit/iface"
	devicerepo "github.com/krnskrll/event-mpos/internal/device/repository"
	orderiface "github.com/krnskrll/event-mpos/internal/order/iface"
	orderrepo "github.com/krnskrll/event-mpos/internal/order/repository"
	staffiface "github.com/krnskrll/event-mpos/internal/staff/iface"
	"github.com/krnskrll/event-mpos/internal/testutil"
)

func orderBody(productID uuid.UUID) J {
	return J{"idempotencyKey": uuid.NewString(), "customer": J{"name": "Dewi"}, "lines": []J{{"productId": productID, "qty": 1}}}
}

// The cashier API runs on a paired device for an active staff member of its location.
func TestDeviceAuth(t *testing.T) {
	a := testutil.NewApp(t)
	ev, other, p := a.ActiveEvent(), a.ActiveEvent(), a.Product("10000")
	a.Allocate(ev.ID, p.ID, 100)
	a.Allocate(other.ID, p.ID, 100)
	dev, otherDev := a.PairDevice(ev.ID), a.PairDevice(other.ID)
	kasir := a.NewStaff(ev.StoreID, staffiface.RoleKasir, ev.ID)
	elsewhere := a.NewStaff(ev.StoreID, staffiface.RoleKasir, other.ID)
	foreign := a.NewStaff(a.NewStore("Store X"), staffiface.RoleKasir)
	orders := "/api/me/events/" + ev.ID.String() + "/orders"

	for name, c := range map[string]struct {
		dev   testutil.Device
		path  string
		staff uuid.UUID
		want  int
	}{
		"no staff":                   {dev, orders, uuid.Nil, http.StatusBadRequest},
		"staff of another location":  {dev, orders, elsewhere.ID, http.StatusForbidden},
		"staff of another store":     {dev, orders, foreign.ID, http.StatusForbidden},
		"another location's eventId": {otherDev, orders, elsewhere.ID, http.StatusForbidden},
	} {
		if rec := a.DeviceDo(http.MethodPost, c.path, orderBody(p.ID), c.dev, c.staff); rec.Code != c.want {
			t.Errorf("%s: %d, want %d (%s)", name, rec.Code, c.want, rec.Body.String())
		}
	}

	// Orders record the staff member and the device; history is the device's own.
	conf := ok[orderiface.OrderConfirmation](t, a.DeviceDo(http.MethodPost, orders, orderBody(p.ID), dev, kasir.ID), http.StatusOK)
	var row orderrepo.Order
	a.DB.Take(&row, "id = ?", conf.ID)
	if row.StaffID == nil || *row.StaffID != kasir.ID || row.DeviceID == nil || *row.DeviceID != dev.ID.String() || row.BaUserID != nil {
		t.Errorf("order row staff=%v device=%v ba=%v", row.StaffID, row.DeviceID, row.BaUserID)
	}
	second := a.PairDevice(ev.ID)
	ok[orderiface.OrderConfirmation](t, a.DeviceDo(http.MethodPost, orders, orderBody(p.ID), second, kasir.ID), http.StatusOK)
	if h := ok[[]orderiface.OrderSummary](t, a.DeviceDo(http.MethodGet, orders, nil, dev, uuid.Nil), http.StatusOK); len(h) != 1 || h[0].ID != conf.ID {
		t.Errorf("history %+v, want this device's one order", h)
	}
	if trail, _ := a.Audit.ForOrder(conf.ID); trail[0].ActorName != kasir.Name {
		t.Errorf("created by %q, want %q", trail[0].ActorName, kasir.Name)
	}
	if rec := a.DeviceDo(http.MethodGet, "/api/me/orders/"+conf.ID.String(), nil, otherDev, uuid.Nil); rec.Code != http.StatusNotFound {
		t.Errorf("another location's order: %d, want 404", rec.Code)
	}

	// Number blocks go to the paired device; its offline order is accepted even after the cashier
	// was deactivated, while a new online sale is refused.
	block := ok[orderiface.NumberBlockResponse](t, a.DeviceDo(http.MethodPost, "/api/me/events/"+ev.ID.String()+"/number-blocks", J{}, dev, uuid.Nil), http.StatusCreated)
	off := false
	if _, err := a.Staff.Update(ev.StoreID, kasir.ID, staffiface.StaffRequest{Name: kasir.Name, Role: kasir.Role, LocationIDs: kasir.LocationIDs, Active: &off}); err != nil {
		t.Fatal(err)
	}
	offline := orderBody(p.ID)
	offline["orderNumber"], offline["clientCreatedAt"], offline["deviceId"] = block.RangeStart, time.Now().Add(-time.Minute), "install-id"
	if rec := a.DeviceDo(http.MethodPost, orders, offline, dev, kasir.ID); rec.Code != http.StatusOK {
		t.Errorf("offline order by a since-deactivated cashier: %d %s, want 200", rec.Code, rec.Body.String())
	}
	if rec := a.DeviceDo(http.MethodPost, orders, orderBody(p.ID), dev, kasir.ID); rec.Code != http.StatusForbidden {
		t.Errorf("online order by an inactive cashier: %d, want 403", rec.Code)
	}

	// A removed device is locked out; user tokens no longer reach the cashier API.
	a.DB.Model(&devicerepo.Device{}).Where("id = ?", second.ID).Update("removed_at", time.Now())
	if rec := a.DeviceDo(http.MethodGet, "/api/me/events/"+ev.ID.String()+"/snapshot", nil, second, uuid.Nil); rec.Code != http.StatusUnauthorized {
		t.Errorf("removed device: %d, want 401", rec.Code)
	}
	if rec := a.Do(http.MethodGet, "/api/me/events", nil, adminToken(t, a)); rec.Code != http.StatusUnauthorized {
		t.Errorf("user token on /api/me: %d, want 401", rec.Code)
	}
	if rec := a.DeviceDo(http.MethodGet, "/api/me/events/"+ev.ID.String()+"/snapshot", nil, dev, uuid.Nil); rec.Code != http.StatusOK {
		t.Errorf("snapshot: %d", rec.Code)
	}
}

// The pickup board shows only the device's location; any staff of it may move orders along.
func TestPickupBoard(t *testing.T) {
	a := testutil.NewApp(t)
	ev, other, p := a.ActiveEvent(), a.ActiveEvent(), a.Product("10000")
	a.Allocate(ev.ID, p.ID, 10)
	a.Allocate(other.ID, p.ID, 10)
	packer, elsewhere := a.NewStaff(ev.StoreID, staffiface.RoleKasir, ev.ID), a.NewStaff(ev.StoreID, staffiface.RoleKasir, other.ID)
	dev, otherDev := a.PairDevice(ev.ID), a.PairDevice(other.ID)
	mine := ok[orderiface.OrderConfirmation](t, a.DeviceDo(http.MethodPost, "/api/me/events/"+ev.ID.String()+"/orders", orderBody(p.ID), dev, packer.ID), http.StatusOK)
	ok[orderiface.OrderConfirmation](t, a.DeviceDo(http.MethodPost, "/api/me/events/"+other.ID.String()+"/orders", orderBody(p.ID), otherDev, elsewhere.ID), http.StatusOK)

	events := ok[[]J](t, a.DeviceDo(http.MethodGet, "/api/warehouse/events", nil, dev, uuid.Nil), http.StatusOK)
	if len(events) != 1 || events[0]["id"] != ev.ID.String() {
		t.Errorf("board events %v, want only the device's location", events)
	}
	board := ok[[]orderiface.WarehouseOrderView](t, a.DeviceDo(http.MethodGet, "/api/warehouse/events/"+ev.ID.String()+"/orders", nil, dev, uuid.Nil), http.StatusOK)
	if len(board) != 1 || board[0].ID != mine.ID || board[0].BaName != packer.Name {
		t.Errorf("board %+v, want the one order here by %s", board, packer.Name)
	}
	ok[[]J](t, a.DeviceDo(http.MethodGet, "/api/warehouse/events/"+ev.ID.String()+"/stock", nil, dev, uuid.Nil), http.StatusOK)

	advance := "/api/warehouse/orders/" + mine.ID.String() + "/status"
	for name, c := range map[string]struct {
		dev   testutil.Device
		staff uuid.UUID
		want  int
	}{
		"no staff":                  {dev, uuid.Nil, http.StatusBadRequest},
		"staff of another location": {dev, elsewhere.ID, http.StatusForbidden},
		"another location's device": {otherDev, elsewhere.ID, http.StatusNotFound},
	} {
		if rec := a.DeviceDo(http.MethodPatch, advance, J{"status": "PACKING"}, c.dev, c.staff); rec.Code != c.want {
			t.Errorf("advance, %s: %d, want %d", name, rec.Code, c.want)
		}
	}
	if v := ok[orderiface.WarehouseOrderView](t, a.DeviceDo(http.MethodPatch, advance, J{"status": "PACKING"}, dev, packer.ID), http.StatusOK); v.Status != orderiface.StatusPacking {
		t.Errorf("advanced to %s", v.Status)
	}
	if trail, _ := a.Audit.ForOrder(mine.ID); trail[1].Action != auditiface.ActionStatusChanged || trail[1].ActorName != packer.Name {
		t.Errorf("status change audit %+v", trail[1])
	}

	for token, want := range map[string]int{"nope": http.StatusUnauthorized, otherDev.Token: http.StatusForbidden} {
		if rec := a.Do(http.MethodGet, "/ws/events/"+ev.ID.String()+"/orders?token="+token, nil, ""); rec.Code != want {
			t.Errorf("board socket: %d, want %d", rec.Code, want)
		}
	}
}
