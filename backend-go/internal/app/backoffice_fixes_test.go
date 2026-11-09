package app_test

import (
	"fmt"
	"net/http"
	"sync"
	"testing"
	"time"

	"github.com/google/uuid"

	auditiface "github.com/krnskrll/event-mpos/internal/audit/iface"
	auditrepo "github.com/krnskrll/event-mpos/internal/audit/repository"
	eventiface "github.com/krnskrll/event-mpos/internal/event/iface"
	orderiface "github.com/krnskrll/event-mpos/internal/order/iface"
	"github.com/krnskrll/event-mpos/internal/platform/tenant"
	receiptrepo "github.com/krnskrll/event-mpos/internal/receipt/repository"
	reportiface "github.com/krnskrll/event-mpos/internal/report/iface"
	staffiface "github.com/krnskrll/event-mpos/internal/staff/iface"
	stockrepo "github.com/krnskrll/event-mpos/internal/stock/repository"
	"github.com/krnskrll/event-mpos/internal/testutil"
	useriface "github.com/krnskrll/event-mpos/internal/user/iface"
)

// EPIC-12 Milestone D: back-office fixes.

func closeEvent(t *testing.T, a *testutil.TestApp, eventID uuid.UUID) {
	t.Helper()
	if _, err := a.Events.ChangeStatus(tenant.DefaultStoreID, eventID, eventiface.StatusClosed); err != nil {
		t.Fatalf("close event: %v", err)
	}
}

func adminToken(t *testing.T, a *testutil.TestApp) string {
	t.Helper()
	admin := a.CreateUser(useriface.RoleAdmin, "Admin123!")
	return login(t, a, admin.Email, "Admin123!").AccessToken
}

// The location report is final once the location is closed: no more orders can arrive.
func TestReportIsFinalOnceClosed(t *testing.T) {
	a := testutil.NewApp(t)
	bhoomyStore := uuid.New()
	store := posStore(t, a, bhoomyStore)
	ev, p := a.ActiveEventIn(store), a.ProductIn(store, "50.00")
	a.Allocate(ev.ID, p.ID, 10)
	if _, err := submit(a, ev.ID, uuid.NewString(), line(p.ID, 1)); err != nil {
		t.Fatal(err)
	}
	report := func() reportiface.SalesReport {
		return ok[reportiface.SalesReport](t, a.Back(http.MethodGet, "/locations/"+ev.ID.String()+"/report", nil, bhoomyStore), http.StatusOK)
	}
	if r := report(); r.Final || r.TotalOrders != 1 {
		t.Errorf("ACTIVE location: final=%v orders=%d, want not final with 1 order", r.Final, r.TotalOrders)
	}
	if _, err := a.Events.ChangeStatus(store, ev.ID, eventiface.StatusClosed); err != nil {
		t.Fatal(err)
	}
	if r := report(); !r.Final || r.TotalOrders != 1 {
		t.Errorf("CLOSED location: final=%v orders=%d, want final with 1 order", r.Final, r.TotalOrders)
	}
}

func TestAuditLogIsPaginatedAndFiltered(t *testing.T) {
	a := testutil.NewApp(t)
	bhoomyStore := uuid.New()
	ev := a.ActiveEventIn(posStore(t, a, bhoomyStore))
	ba, wh := uuid.New(), uuid.New()

	// 125 entries, one second apart (entry 0 is the newest). Every 5th is a receipt print;
	// even entries are by the cashier, odd ones by the packer.
	const total = 125
	targetOrder := uuid.New()
	base := time.Now().UTC()
	for i := 0; i < total; i++ {
		actor, action, orderID := wh, auditiface.ActionStatusChanged, uuid.New()
		if i%2 == 0 {
			actor = ba
		}
		if i%5 == 0 {
			action = auditiface.ActionReceiptPrinted
		}
		if i < 3 {
			orderID = targetOrder
		}
		row := auditrepo.OrderAuditLog{
			EventID: ev.ID, OrderID: orderID, OrderReference: fmt.Sprintf("ORD-%03d", i), Action: action,
			ActorUserID: &actor, ActorName: "x", CreatedAt: base.Add(-time.Duration(i) * time.Second),
		}
		if err := a.DB.Create(&row).Error; err != nil {
			t.Fatal(err)
		}
	}

	get := func(query string) auditiface.Page {
		t.Helper()
		rec := a.Back(http.MethodGet, "/locations/"+ev.ID.String()+"/audit"+query, nil, bhoomyStore)
		if rec.Code != http.StatusOK {
			t.Fatalf("audit%s: %d %s", query, rec.Code, rec.Body.String())
		}
		return testutil.Decode[auditiface.Page](t, rec)
	}

	first := get("")
	if first.Page != 0 || first.Size != 50 || first.Total != total || len(first.Items) != 50 {
		t.Fatalf("default page: page=%d size=%d total=%d items=%d", first.Page, first.Size, first.Total, len(first.Items))
	}
	if first.Items[0].OrderReference != "ORD-000" || first.Items[49].OrderReference != "ORD-049" {
		t.Errorf("not newest-first: %s … %s", first.Items[0].OrderReference, first.Items[49].OrderReference)
	}

	seen := map[uuid.UUID]bool{}
	for page := 0; page < 3; page++ {
		for _, e := range get(fmt.Sprintf("?page=%d", page)).Items {
			if seen[e.ID] {
				t.Fatalf("entry %s appears on two pages", e.ID)
			}
			seen[e.ID] = true
		}
	}
	if len(seen) != total {
		t.Errorf("paging through 3 pages saw %d entries, want %d", len(seen), total)
	}
	if last := get("?page=2"); len(last.Items) != 25 {
		t.Errorf("last page has %d items, want 25", len(last.Items))
	}
	if beyond := get("?page=9"); len(beyond.Items) != 0 || beyond.Total != total {
		t.Errorf("page past the end: %d items, total %d", len(beyond.Items), beyond.Total)
	}

	if big := get("?size=500"); big.Size != 200 || len(big.Items) != total {
		t.Errorf("size=500: size=%d items=%d, want capped at 200 with all %d", big.Size, len(big.Items), total)
	}
	if byActor := get("?actor=" + ba.String()); byActor.Total != 63 {
		t.Errorf("actor=BA total = %d, want 63", byActor.Total)
	}
	if byAction := get("?action=RECEIPT_PRINTED"); byAction.Total != 25 {
		t.Errorf("action=RECEIPT_PRINTED total = %d, want 25", byAction.Total)
	}
	// Receipt prints (every 5th) by the BA (even) → i = 0, 10, 20, … 120 → 13 entries.
	if both := get("?action=RECEIPT_PRINTED&actor=" + ba.String()); both.Total != 13 {
		t.Errorf("actor+action total = %d, want 13", both.Total)
	}
	if byOrder := get("?orderId=" + targetOrder.String()); byOrder.Total != 3 {
		t.Errorf("orderId total = %d, want 3", byOrder.Total)
	}

	for _, bad := range []string{"?action=NOPE", "?actor=not-a-uuid", "?page=-1", "?size=0", "?page=x"} {
		if rec := a.Back(http.MethodGet, "/locations/"+ev.ID.String()+"/audit"+bad, nil, bhoomyStore); rec.Code != http.StatusBadRequest {
			t.Errorf("audit%s: %d, want 400", bad, rec.Code)
		}
	}
}

func TestReceiptPrintWritesLogAndAuditEntry(t *testing.T) {
	a := testutil.NewApp(t)
	ev, p := a.ActiveEvent(), a.Product("20.00")
	a.Allocate(ev.ID, p.ID, 10)
	ba := a.NewStaff(ev.StoreID, staffiface.RoleKasir, ev.ID)
	dev := a.PairDevice(ev.ID)
	conf, err := a.Orders.Submit(ev.ID, ba.ID, orderiface.SubmitOrderRequest{
		IdempotencyKey: uuid.NewString(), Customer: orderiface.CustomerInfo{Name: "Cara"},
		Lines: []orderiface.OrderLineRequest{{ProductID: p.ID, Qty: 1}},
	})
	if err != nil {
		t.Fatal(err)
	}
	path := "/api/me/orders/" + conf.ID.String() + "/receipts"

	for i := 0; i < 2; i++ {
		if rec := a.DeviceDo(http.MethodPost, path, nil, dev, ba.ID); rec.Code != http.StatusNoContent {
			t.Fatalf("print %d: %d %s", i+1, rec.Code, rec.Body.String())
		}
	}
	printCount := func() int {
		var n int
		a.DB.Model(&receiptrepo.ReceiptLog{}).Where("order_id = ?", conf.ID).Pluck("print_count", &n)
		return n
	}
	if n := printCount(); n != 2 {
		t.Errorf("print_count = %d, want 2", n)
	}
	trail, _ := a.Audit.ForOrder(conf.ID)
	if len(trail) != 3 || trail[1].Action != auditiface.ActionReceiptPrinted || trail[2].Action != auditiface.ActionReceiptPrinted {
		t.Fatalf("trail = %+v", trail)
	}
	if *trail[1].Detail != "Receipt printed" || *trail[2].Detail != "Receipt reprinted (print #2)" || trail[1].ActorName != ba.Name {
		t.Errorf("receipt entries = %+v / %+v", trail[1], trail[2])
	}

	// Another location's device, an unknown order, or no staff: refused, nothing changes.
	otherEv := a.ActiveEvent()
	if rec := a.DeviceDo(http.MethodPost, path, nil, a.PairDevice(otherEv.ID), a.NewStaff(ev.StoreID, staffiface.RoleKasir, otherEv.ID).ID); rec.Code != http.StatusNotFound {
		t.Errorf("other location's order: %d, want 404", rec.Code)
	}
	if rec := a.DeviceDo(http.MethodPost, path, nil, dev, uuid.Nil); rec.Code != http.StatusBadRequest {
		t.Errorf("print without X-Staff-Id: %d, want 400", rec.Code)
	}
	if rec := a.DeviceDo(http.MethodPost, "/api/me/orders/"+uuid.NewString()+"/receipts", nil, dev, ba.ID); rec.Code != http.StatusNotFound {
		t.Errorf("unknown order: %d, want 404", rec.Code)
	}

	// Concurrent prints never lose a count or an audit entry.
	var wg sync.WaitGroup
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			if err := a.Receipts.LogPrint(conf.ID, ev.ID, ba.ID); err != nil {
				t.Errorf("concurrent print: %v", err)
			}
		}()
	}
	wg.Wait()
	if n := printCount(); n != 12 {
		t.Errorf("print_count after concurrent prints = %d, want 12", n)
	}
	if trail, _ := a.Audit.ForOrder(conf.ID); len(trail) != 13 {
		t.Errorf("audit entries = %d, want 13 (1 created + 12 prints)", len(trail))
	}
}

func TestStockSummaryCacheIsInvalidatedByOrders(t *testing.T) {
	a := testutil.NewApp(t)
	ev, p := a.ActiveEvent(), a.Product("20.00")
	a.Allocate(ev.ID, p.ID, 10)

	remaining := func() (int, int) {
		t.Helper()
		ba, err := a.Stock.BaLevelsForEvent(ev.ID)
		if err != nil || len(ba) != 1 {
			t.Fatalf("ba levels: %v %v", ba, err)
		}
		wh, err := a.Stock.LevelsForEvent(ev.ID)
		if err != nil || len(wh) != 1 {
			t.Fatalf("warehouse levels: %v %v", wh, err)
		}
		return ba[0].Remaining, wh[0].Remaining
	}
	if b, w := remaining(); b != 10 || w != 10 {
		t.Fatalf("initial remaining %d/%d, want 10", b, w)
	}

	// A write that bypasses the service is not seen until the entry is invalidated or expires.
	a.DB.Model(&stockrepo.EventStockAllocation{}).Where("event_id = ?", ev.ID).Update("sold_qty", 1)
	if b, w := remaining(); b != 10 || w != 10 {
		t.Errorf("cache not used: remaining %d/%d, want the cached 10", b, w)
	}
	a.Stock.InvalidateSummary(ev.ID)
	if b, w := remaining(); b != 9 || w != 9 {
		t.Errorf("after invalidate: %d/%d, want 9", b, w)
	}

	// An order submit clears the event's cached summary immediately.
	if _, err := submit(a, ev.ID, uuid.NewString(), line(p.ID, 3)); err != nil {
		t.Fatal(err)
	}
	if b, w := remaining(); b != 6 || w != 6 {
		t.Errorf("after submit: %d/%d, want 6", b, w)
	}
}
