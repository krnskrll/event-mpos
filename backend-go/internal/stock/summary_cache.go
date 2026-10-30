package stock

import (
	"sync"
	"time"

	"github.com/google/uuid"
)

// SummaryTTL is how long a computed stock summary is served from memory. Every BA device and every
// warehouse board polls the summary, so this caps the database load to one query per event per TTL.
// Writes that change stock invalidate the event's entry immediately, so the TTL only bounds staleness
// from changes this process didn't make.
const SummaryTTL = 15 * time.Second

type summaryKind int

const (
	kindWarehouse summaryKind = iota
	kindBa
)

type summaryKey struct {
	eventID uuid.UUID
	kind    summaryKind
}

type summaryEntry struct {
	value   any
	expires time.Time
}

// summaryCache is a small per-event TTL cache. A per-event generation counter makes invalidation
// race-safe: a read that started before an invalidation never stores its (now stale) result.
type summaryCache struct {
	mu      sync.Mutex
	ttl     time.Duration
	now     func() time.Time
	entries map[summaryKey]summaryEntry
	gen     map[uuid.UUID]uint64
}

func newSummaryCache(ttl time.Duration) *summaryCache {
	return &summaryCache{ttl: ttl, now: time.Now, entries: map[summaryKey]summaryEntry{}, gen: map[uuid.UUID]uint64{}}
}

// get returns a fresh cached value, or the event's current generation for a later put.
func (c *summaryCache) get(k summaryKey) (any, bool, uint64) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if e, ok := c.entries[k]; ok && c.now().Before(e.expires) {
		return e.value, true, 0
	}
	return nil, false, c.gen[k.eventID]
}

// put stores v unless the event was invalidated since the matching get.
func (c *summaryCache) put(k summaryKey, v any, gen uint64) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.gen[k.eventID] != gen {
		return
	}
	c.entries[k] = summaryEntry{value: v, expires: c.now().Add(c.ttl)}
}

func (c *summaryCache) invalidate(eventID uuid.UUID) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.gen[eventID]++
	delete(c.entries, summaryKey{eventID, kindWarehouse})
	delete(c.entries, summaryKey{eventID, kindBa})
}
