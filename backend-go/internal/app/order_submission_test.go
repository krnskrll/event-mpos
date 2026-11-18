package app_test

import (
	"strings"
	"testing"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"

	orderiface "github.com/krnskrll/event-mpos/internal/order/iface"
	promotioniface "github.com/krnskrll/event-mpos/internal/promotion/iface"
	"github.com/krnskrll/event-mpos/internal/testutil"
)

// Ported from the Kotlin OrderSubmissionTest.

func submit(a *testutil.TestApp, eventID uuid.UUID, key string, lines ...orderiface.OrderLineRequest) (*orderiface.OrderConfirmation, error) {
	return a.Orders.Submit(eventID, uuid.New(), orderiface.SubmitOrderRequest{
		IdempotencyKey: key, Customer: orderiface.CustomerInfo{Name: "Aisha"}, Lines: lines,
	})
}

func line(productID uuid.UUID, qty int) orderiface.OrderLineRequest {
	return orderiface.OrderLineRequest{ProductID: productID, Qty: qty}
}

func dec(s string) decimal.Decimal { return decimal.RequireFromString(s) }

func TestSubmitCreatesNumberedOrderForBhoomy(t *testing.T) {
	a := testutil.NewApp(t)
	ev, p := a.ActiveEvent(), a.Product("100.00")

	conf, err := submit(a, ev.ID, uuid.NewString(), line(p.ID, 3))
	if err != nil {
		t.Fatal(err)
	}
	if conf.OrderNumber != 1 {
		t.Errorf("orderNumber = %d, want 1", conf.OrderNumber)
	}
	if !strings.HasPrefix(conf.ReferenceID, "ORD-") || !strings.HasSuffix(conf.ReferenceID, "-"+ev.Code+"-001") {
		t.Errorf("referenceId = %s", conf.ReferenceID)
	}
	if !conf.Subtotal.Equal(dec("300")) {
		t.Errorf("subtotal = %s, want 300", conf.Subtotal)
	}
	// 6c: POS no longer counts stock; the order waits for the job that sends it to Bhoomy.
	var state string
	a.DB.Raw("SELECT sync_state FROM orders WHERE id = ?", conf.ID).Scan(&state)
	if state != "pending" {
		t.Errorf("sync_state = %q, want pending", state)
	}
}

func TestIdempotentSubmitReturnsOriginalOrder(t *testing.T) {
	a := testutil.NewApp(t)
	ev, p := a.ActiveEvent(), a.Product("100.00")

	key := uuid.NewString()
	first, err := submit(a, ev.ID, key, line(p.ID, 2))
	if err != nil {
		t.Fatal(err)
	}
	second, err := submit(a, ev.ID, key, line(p.ID, 2)) // double tap / retry
	if err != nil {
		t.Fatal(err)
	}
	if first.ID != second.ID || !second.Duplicate {
		t.Errorf("retry created a new order: first=%s second=%s duplicate=%v", first.ID, second.ID, second.Duplicate)
	}
	var n int64
	a.DB.Raw("SELECT count(*) FROM orders WHERE idempotency_key = ?", key).Scan(&n)
	if n != 1 {
		t.Errorf("orders = %d, want 1", n)
	}
}

func TestDailyNumberIncrementsPerEvent(t *testing.T) {
	a := testutil.NewApp(t)
	ev, other, p := a.ActiveEvent(), a.ActiveEvent(), a.Product("100.00")

	for want := 1; want <= 3; want++ {
		conf, err := submit(a, ev.ID, uuid.NewString(), line(p.ID, 1))
		if err != nil {
			t.Fatal(err)
		}
		if conf.OrderNumber != want {
			t.Errorf("order %d got number %d", want, conf.OrderNumber)
		}
	}
	// Another event keeps its own sequence.
	conf, err := submit(a, other.ID, uuid.NewString(), line(p.ID, 1))
	if err != nil {
		t.Fatal(err)
	}
	if conf.OrderNumber != 1 {
		t.Errorf("other event first number = %d, want 1", conf.OrderNumber)
	}
}

// 6c: stock is Bhoomy's. POS never refuses a sale for stock; a shortfall becomes a Konflik stok there.
func TestStockNeverRefusesASale(t *testing.T) {
	a := testutil.NewApp(t)
	ev, p := a.ActiveEvent(), a.Product("100.00")
	conf, err := submit(a, ev.ID, uuid.NewString(), line(p.ID, 5))
	if err != nil || conf.OrderNumber != 1 {
		t.Fatalf("got %+v, %v", conf, err)
	}
}

func TestBestPromotionIsAppliedAtSubmit(t *testing.T) {
	a := testutil.NewApp(t)
	ev, p := a.ActiveEvent(), a.Product("100.00")
	pct := dec("10")
	if _, err := a.Promotions.Create(ev.StoreID, promotioniface.PromotionRequest{LocationID: &ev.ID, Name: "10% off", Type: promotioniface.TypePercentage, PercentOff: &pct}); err != nil {
		t.Fatal(err)
	}

	conf, err := submit(a, ev.ID, uuid.NewString(), line(p.ID, 2))
	if err != nil {
		t.Fatal(err)
	}
	if !conf.Subtotal.Equal(dec("200")) || !conf.Discount.Equal(dec("20")) || !conf.Total.Equal(dec("180")) {
		t.Errorf("subtotal/discount/total = %s/%s/%s, want 200/20/180", conf.Subtotal, conf.Discount, conf.Total)
	}
	if conf.PromotionName == nil || *conf.PromotionName != "10% off" {
		t.Errorf("promotionName = %v", conf.PromotionName)
	}
}
