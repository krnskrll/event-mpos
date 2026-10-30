package app_test

import (
	"net/http"
	"sync"
	"sync/atomic"
	"testing"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/krnskrll/event-mpos/internal/stock"
	"github.com/krnskrll/event-mpos/internal/testutil"
)

// Ported from the Kotlin StockConcurrencyTest — against real Postgres row locks this time.

// race runs fn from n goroutines released at the same instant and counts successes and 409s.
func race(t *testing.T, n int, fn func() error) (ok, conflict int64) {
	t.Helper()
	var wg sync.WaitGroup
	start := make(chan struct{})
	var unexpected atomic.Value
	for i := 0; i < n; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			<-start
			err := fn()
			switch {
			case err == nil:
				atomic.AddInt64(&ok, 1)
			case testutil.StatusOf(err) == http.StatusConflict:
				atomic.AddInt64(&conflict, 1)
			default:
				unexpected.Store(err)
			}
		}()
	}
	close(start)
	wg.Wait()
	if err := unexpected.Load(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	return ok, conflict
}

func TestConcurrentDecrementsNeverOversell(t *testing.T) {
	a := testutil.NewApp(t)
	ev, p := a.ActiveEvent(), a.Product("10.00")
	const allocated, attempts = 100, 300
	a.Allocate(ev.ID, p.ID, allocated)

	ok, conflict := race(t, attempts, func() error {
		return a.DB.Transaction(func(tx *gorm.DB) error { return a.Stock.Decrement(tx, ev.ID, p.ID, 1) })
	})

	if ok != allocated || conflict != attempts-allocated {
		t.Errorf("ok=%d conflict=%d, want %d/%d", ok, conflict, allocated, attempts-allocated)
	}
	var alloc stock.EventStockAllocation
	if err := a.DB.Where("event_id = ? AND product_id = ?", ev.ID, p.ID).First(&alloc).Error; err != nil {
		t.Fatal(err)
	}
	if alloc.SoldQty != allocated || alloc.Remaining() < 0 {
		t.Errorf("soldQty=%d remaining=%d — oversold", alloc.SoldQty, alloc.Remaining())
	}
}

func TestConcurrentOrdersForLastUnitsNeverOversell(t *testing.T) {
	a := testutil.NewApp(t)
	ev, p := a.ActiveEvent(), a.Product("10.00")
	const allocated, buyers = 8, 40
	a.Allocate(ev.ID, p.ID, allocated)

	ok, conflict := race(t, buyers, func() error {
		_, err := submit(a, ev.ID, uuid.NewString(), line(p.ID, 1))
		return err
	})

	if ok != allocated || conflict != buyers-allocated {
		t.Errorf("ok=%d conflict=%d, want %d/%d", ok, conflict, allocated, buyers-allocated)
	}
	if got := a.Remaining(ev.ID, p.ID); got != 0 {
		t.Errorf("remaining = %d, want 0", got)
	}
	// Every successful order got a distinct daily number 1..allocated.
	var numbers []int
	a.DB.Table("orders").Where("event_id = ?", ev.ID).Order("order_number").Pluck("order_number", &numbers)
	for i, n := range numbers {
		if n != i+1 {
			t.Fatalf("order numbers = %v, want 1..%d with no gaps or duplicates", numbers, allocated)
		}
	}
}
