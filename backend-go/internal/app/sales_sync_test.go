package app_test

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"

	"github.com/krnskrll/event-mpos/internal/integration/bhoomy"
	orderiface "github.com/krnskrll/event-mpos/internal/order/iface"
	productrepo "github.com/krnskrll/event-mpos/internal/product/repository"
	"github.com/krnskrll/event-mpos/internal/testutil"
)

// linkedSale is a store on Bhoomy with one event selling one product (Produk Toko variant bv).
func linkedSale(t *testing.T) (*testutil.TestApp, *bhoomy.Mock, uuid.UUID, *productrepo.Product, uuid.UUID) {
	t.Helper()
	a := testutil.NewApp(t)
	mock := a.Bhoomy.(*bhoomy.Mock)
	store := a.NewStore("Linked")
	a.DB.Exec("UPDATE pos_stores SET bhoomy_store_id = ? WHERE id = ?", uuid.New(), store)
	ev, p, bv := a.ActiveEventIn(store), a.ProductIn(store, "100.00"), uuid.New()
	a.DB.Exec("UPDATE pos_item_variants SET bhoomy_variant_id = ? WHERE item_id = ?", bv, p.ID)
	if err := a.DB.Create(&productrepo.LocationItem{LocationID: ev.ID, ItemVariantID: p.Default().ID, Active: true}).Error; err != nil {
		t.Fatal(err)
	}
	return a, mock, ev.ID, p, bv
}

func sellAt(t *testing.T, a *testutil.TestApp, ev, product uuid.UUID, place *uuid.UUID) *orderiface.OrderConfirmation {
	t.Helper()
	conf, err := a.Orders.Submit(ev, uuid.New(), orderiface.SubmitOrderRequest{
		IdempotencyKey: uuid.NewString(), Customer: orderiface.CustomerInfo{Name: "Aisha"},
		Lines: []orderiface.OrderLineRequest{{ProductID: product, Qty: 2}}, StockPlaceID: place, StockKnown: true,
	})
	if err != nil {
		t.Fatal(err)
	}
	return conf
}

// drain sends everything due, other tests' orders included (one shared database).
func drain(a *testutil.TestApp) {
	for a.SalesSync.RunOnce(context.Background()) == 100 {
	}
}

type syncRow struct {
	SyncState     string
	SyncAttempts  int
	StockStatus   *string
	StockConflict bool
	BhoomyPlaceID *uuid.UUID
}

func syncOf(a *testutil.TestApp, id uuid.UUID) syncRow {
	var r syncRow
	a.DB.Raw("SELECT sync_state, sync_attempts, stock_status, stock_conflict, bhoomy_place_id FROM orders WHERE id = ?", id).Scan(&r)
	return r
}

func TestSalesAndCancelsGoToBhoomy(t *testing.T) {
	a, mock, ev, p, bv := linkedSale(t)
	place := uuid.New()
	o := sellAt(t, a, ev, p.ID, &place)

	// Bhoomy down: the order waits and is tried again later.
	mock.Err = errors.New("connection refused")
	drain(a)
	drain(a)
	if r := syncOf(a, o.ID); r.SyncState != "pending" || r.SyncAttempts != 1 {
		t.Fatalf("after a failed send (and no retry before its backoff): %+v", r)
	}

	mock.Err, mock.SaleStatus = nil, "conflict"
	a.DB.Exec("UPDATE orders SET next_sync_at = now() WHERE id = ?", o.ID)
	drain(a)
	if len(mock.Sales) != 1 {
		t.Fatalf("sales sent: %d", len(mock.Sales))
	}
	s := mock.Sales[0]
	if s.PosSaleID != o.ID || s.PosLocationID != ev || *s.PlaceID != place || *s.Lines[0].VariantID != bv || s.Lines[0].Qty != 2 || s.ReceiptNo != o.ReferenceID {
		t.Fatalf("sale = %+v", s)
	}
	if r := syncOf(a, o.ID); r.SyncState != "sent" || *r.StockStatus != "conflict" || !r.StockConflict {
		t.Fatalf("after Bhoomy's answer: %+v", r)
	}
	if drain(a); len(mock.Sales) != 1 {
		t.Fatal("a sent sale was sent again")
	}

	// A cancel goes next, on its own.
	if _, err := a.Orders.Cancel(o.ID, uuid.New(), "Sari", nil, false); err != nil {
		t.Fatal(err)
	}
	drain(a)
	if len(mock.Cancels) != 1 || mock.Cancels[0] != o.ID || len(mock.Sales) != 1 {
		t.Fatalf("cancels %v, sales %d", mock.Cancels, len(mock.Sales))
	}
	if r := syncOf(a, o.ID); r.SyncState != "sent" || *r.StockStatus != "cancelled" {
		t.Fatalf("after cancel: %+v", r)
	}
}

func TestSalesThatCannotReachBhoomy(t *testing.T) {
	a, mock, ev, p, _ := linkedSale(t)

	refused := sellAt(t, a, ev, p.ID, nil)
	mock.Err = &bhoomy.RefusedError{Status: 422, Body: "VARIANT_NOT_FOUND"}
	drain(a)
	if r := syncOf(a, refused.ID); r.SyncState != "failed" {
		t.Errorf("refused sale: %+v, want failed", r)
	}

	mock.Err = nil
	other := a.ActiveEvent() // the default store is not on Bhoomy
	o := sellAt(t, a, other.ID, a.Product("10.00").ID, nil)
	drain(a)
	if r := syncOf(a, o.ID); r.SyncState != "skipped" || len(mock.Sales) != 0 {
		t.Errorf("store not on Bhoomy: %+v, sales %d", r, len(mock.Sales))
	}
}

func TestStockLevelsAndTheOldAppsPlace(t *testing.T) {
	a, mock, ev, p, bv := linkedSale(t)
	ctx, place := context.Background(), uuid.New()
	mock.Stocks = map[uuid.UUID]*bhoomy.Stock{ev: {Tracked: true, PlaceID: &place, Levels: []bhoomy.StockLevel{{VariantID: bv, Available: 4}}}}

	lv, err := a.StockReader.Levels(ctx, ev)
	if err != nil || !lv.Tracked || *lv.PlaceID != place || len(lv.Levels) != 1 {
		t.Fatalf("%v %+v", err, lv)
	}
	if l := lv.Levels[0]; l.ProductID != p.ID || l.VariantID != nil || l.Available != 4 {
		t.Fatalf("level %+v", l)
	}
	// 30 s in memory: Bhoomy's change shows after that, and Bhoomy down keeps the app on its numbers.
	mock.Stocks[ev] = &bhoomy.Stock{Tracked: true, PlaceID: &place, Levels: []bhoomy.StockLevel{{VariantID: bv, Available: 0}}}
	if lv, _ := a.StockReader.Levels(ctx, ev); lv.Levels[0].Available != 4 {
		t.Errorf("cache not used: %+v", lv.Levels[0])
	}

	// An app from before 6c sends no place: the sale takes the location's current one.
	conf, err := a.Orders.Submit(ev, uuid.New(), orderiface.SubmitOrderRequest{
		IdempotencyKey: uuid.NewString(), Customer: orderiface.CustomerInfo{Name: "Old app"},
		Lines: []orderiface.OrderLineRequest{{ProductID: p.ID, Qty: 1}},
	})
	if err != nil {
		t.Fatal(err)
	}
	if r := syncOf(a, conf.ID); r.BhoomyPlaceID == nil || *r.BhoomyPlaceID != place {
		t.Errorf("old app's sale place = %v, want %s", r.BhoomyPlaceID, place)
	}
}
