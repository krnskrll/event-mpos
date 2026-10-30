package app_test

import (
	"net/http"
	"strings"
	"testing"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"

	"github.com/krnskrll/event-mpos/internal/order"
	"github.com/krnskrll/event-mpos/internal/promotion"
	"github.com/krnskrll/event-mpos/internal/testutil"
)

// Ported from the Kotlin OrderSubmissionTest.

func submit(a *testutil.TestApp, eventID uuid.UUID, key string, lines ...order.OrderLineRequest) (*order.OrderConfirmation, error) {
	return a.Orders.Submit(eventID, uuid.New(), order.SubmitOrderRequest{
		IdempotencyKey: key, Customer: order.CustomerInfo{Name: "Aisha"}, Lines: lines,
	})
}

func line(productID uuid.UUID, qty int) order.OrderLineRequest {
	return order.OrderLineRequest{ProductID: productID, Qty: qty}
}

func dec(s string) decimal.Decimal { return decimal.RequireFromString(s) }

func TestSubmitCreatesNumberedOrderAndDecrementsStock(t *testing.T) {
	a := testutil.NewApp(t)
	ev, p := a.ActiveEvent(), a.Product("100.00")
	a.Allocate(ev.ID, p.ID, 10)

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
	if got := a.Remaining(ev.ID, p.ID); got != 7 {
		t.Errorf("remaining = %d, want 7", got)
	}
}

func TestIdempotentSubmitReturnsOriginalOrder(t *testing.T) {
	a := testutil.NewApp(t)
	ev, p := a.ActiveEvent(), a.Product("100.00")
	a.Allocate(ev.ID, p.ID, 10)

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
	if got := a.Remaining(ev.ID, p.ID); got != 8 {
		t.Errorf("remaining = %d, want 8 (decremented once)", got)
	}
}

func TestDailyNumberIncrementsPerEvent(t *testing.T) {
	a := testutil.NewApp(t)
	ev, other, p := a.ActiveEvent(), a.ActiveEvent(), a.Product("100.00")
	a.Allocate(ev.ID, p.ID, 100)
	a.Allocate(other.ID, p.ID, 100)

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

func TestOversellIsRejectedAndWholeOrderRollsBack(t *testing.T) {
	a := testutil.NewApp(t)
	ev, p1, p2 := a.ActiveEvent(), a.Product("100.00"), a.Product("50.00")
	a.Allocate(ev.ID, p1.ID, 10)
	a.Allocate(ev.ID, p2.ID, 2)

	// Line 1 fits, line 2 oversells: nothing may be decremented and no number consumed.
	_, err := submit(a, ev.ID, uuid.NewString(), line(p1.ID, 3), line(p2.ID, 5))
	if testutil.StatusOf(err) != http.StatusConflict {
		t.Fatalf("oversell: got %v, want 409", err)
	}
	if got := a.Remaining(ev.ID, p1.ID); got != 10 {
		t.Errorf("p1 remaining = %d, want 10 (rolled back)", got)
	}
	if got := a.Remaining(ev.ID, p2.ID); got != 2 {
		t.Errorf("p2 remaining = %d, want 2", got)
	}
	conf, err := submit(a, ev.ID, uuid.NewString(), line(p1.ID, 1))
	if err != nil {
		t.Fatal(err)
	}
	if conf.OrderNumber != 1 {
		t.Errorf("number after rollback = %d, want 1 (failed order must not consume one)", conf.OrderNumber)
	}
}

func TestBestPromotionIsAppliedAtSubmit(t *testing.T) {
	a := testutil.NewApp(t)
	ev, p := a.ActiveEvent(), a.Product("100.00")
	a.Allocate(ev.ID, p.ID, 100)
	pct := dec("10")
	if _, err := a.Promotions.Create(ev.ID, promotion.PromotionRequest{Name: "10% off", Type: promotion.TypePercentage, PercentOff: &pct}); err != nil {
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
