package stock

import (
	"testing"
	"time"

	"github.com/google/uuid"
)

func TestSummaryCacheExpiresAndInvalidationWins(t *testing.T) {
	clock := time.Date(2026, 10, 30, 10, 0, 0, 0, time.UTC)
	c := newSummaryCache(15 * time.Second)
	c.now = func() time.Time { return clock }
	ev := uuid.New()
	k := summaryKey{ev, kindBa}

	_, ok, gen := c.get(k)
	if ok {
		t.Fatal("empty cache returned a value")
	}
	c.put(k, "v1", gen)
	if v, ok, _ := c.get(k); !ok || v != "v1" {
		t.Fatalf("got %v %v, want cached v1", v, ok)
	}

	clock = clock.Add(14 * time.Second)
	if _, ok, _ := c.get(k); !ok {
		t.Error("expired before the TTL")
	}
	clock = clock.Add(2 * time.Second)
	if _, ok, _ := c.get(k); ok {
		t.Error("still served after the TTL")
	}

	// A read that started before an invalidation must not cache its stale result.
	_, _, gen = c.get(k)
	c.invalidate(ev)
	c.put(k, "stale", gen)
	if v, ok, _ := c.get(k); ok {
		t.Errorf("stale value %v was cached after invalidation", v)
	}

	// Invalidation clears both summary kinds for the event, and only that event.
	other := uuid.New()
	_, _, g1 := c.get(summaryKey{ev, kindWarehouse})
	c.put(summaryKey{ev, kindWarehouse}, "w", g1)
	_, _, g2 := c.get(summaryKey{other, kindWarehouse})
	c.put(summaryKey{other, kindWarehouse}, "o", g2)
	c.invalidate(ev)
	if _, ok, _ := c.get(summaryKey{ev, kindWarehouse}); ok {
		t.Error("warehouse summary survived invalidation")
	}
	if _, ok, _ := c.get(summaryKey{other, kindWarehouse}); !ok {
		t.Error("another event's summary was invalidated")
	}
}
