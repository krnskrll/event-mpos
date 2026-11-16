package bhoomy

import (
	"context"
	"sync"
	"time"

	"github.com/google/uuid"

	catalogiface "github.com/krnskrll/event-mpos/internal/catalog/iface"
	eventiface "github.com/krnskrll/event-mpos/internal/event/iface"
	productiface "github.com/krnskrll/event-mpos/internal/product/iface"
)

// stockTTL is how long a location's numbers are reused: no POS stock table, only this memory.
const stockTTL = 30 * time.Second

// Levels is what the cashier app may sell (6c), keyed by what it sells. VariantID nil = the
// product itself. Untracked: no limits at all.
type Levels struct {
	Tracked bool         `json:"tracked"`
	PlaceID *uuid.UUID   `json:"placeId"`
	Levels  []LevelEntry `json:"levels"`
}

type LevelEntry struct {
	ProductID uuid.UUID  `json:"productId"`
	VariantID *uuid.UUID `json:"variantId"`
	Available int        `json:"available"`
}

// StockReader reads a location's stock from Bhoomy and keeps it 30 s in memory.
type StockReader struct {
	client   Client
	catalog  catalogiface.Catalog
	events   eventiface.Events
	products productiface.Products

	mu    sync.Mutex
	cache map[uuid.UUID]cachedStock
}

type cachedStock struct {
	at    time.Time
	stock *Stock
}

func NewStockReader(client Client, catalog catalogiface.Catalog, events eventiface.Events, products productiface.Products) *StockReader {
	return &StockReader{client: client, catalog: catalog, events: events, products: products, cache: map[uuid.UUID]cachedStock{}}
}

// Levels maps Bhoomy's numbers onto the location's products; a variant Bhoomy has no row for
// has 0. Errors mean Bhoomy did not answer: the app keeps its last numbers.
func (r *StockReader) Levels(ctx context.Context, locationID uuid.UUID) (*Levels, error) {
	st, err := r.read(ctx, locationID)
	if err != nil {
		return nil, err
	}
	out := &Levels{Tracked: st.Tracked, PlaceID: st.PlaceID, Levels: []LevelEntry{}}
	if !st.Tracked {
		return out, nil
	}
	avail := make(map[uuid.UUID]int, len(st.Levels))
	for _, l := range st.Levels {
		avail[l.VariantID] = l.Available
	}
	links, err := r.products.VariantLinks(locationID)
	if err != nil {
		return nil, err
	}
	for _, l := range links {
		out.Levels = append(out.Levels, LevelEntry{ProductID: l.ProductID, VariantID: l.VariantID, Available: avail[l.BhoomyVariantID]})
	}
	return out, nil
}

// PlaceOf is the location's linked place now, for an app from before 6c; nil when untracked or
// Bhoomy did not answer.
func (r *StockReader) PlaceOf(locationID uuid.UUID) *uuid.UUID {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	st, err := r.read(ctx, locationID)
	if err != nil || !st.Tracked {
		return nil
	}
	return st.PlaceID
}

func (r *StockReader) read(ctx context.Context, locationID uuid.UUID) (*Stock, error) {
	r.mu.Lock()
	c, ok := r.cache[locationID]
	r.mu.Unlock()
	if ok && time.Since(c.at) < stockTTL {
		return c.stock, nil
	}
	ev, err := r.events.Get(locationID)
	if err != nil {
		return nil, err
	}
	store, err := bhoomyStore(r.catalog, ev.StoreID)
	if err != nil {
		return nil, err
	}
	st := &Stock{}
	if store != uuid.Nil {
		if st, err = r.client.Stock(ctx, store, locationID); err != nil {
			return nil, err
		}
	}
	r.mu.Lock()
	r.cache[locationID] = cachedStock{at: time.Now(), stock: st}
	r.mu.Unlock()
	return st, nil
}

// bhoomyStore is the Bhoomy store of a POS store; uuid.Nil when it is not on Bhoomy.
func bhoomyStore(catalog catalogiface.Catalog, storeID uuid.UUID) (uuid.UUID, error) {
	links, err := catalog.Links()
	if err != nil {
		return uuid.Nil, err
	}
	for _, l := range links {
		if l.StoreID == storeID {
			return l.BhoomyStoreID, nil
		}
	}
	return uuid.Nil, nil
}
