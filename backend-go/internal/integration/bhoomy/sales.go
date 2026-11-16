package bhoomy

import (
	"context"
	"errors"
	"log"
	"time"

	"github.com/google/uuid"

	catalogiface "github.com/krnskrll/event-mpos/internal/catalog/iface"
	eventiface "github.com/krnskrll/event-mpos/internal/event/iface"
	orderiface "github.com/krnskrll/event-mpos/internal/order/iface"
)

const salesBatch = 100

// SalesSync sends sales and cancels to Bhoomy (6c): unsent orders, oldest first, each retried
// with backoff until Bhoomy accepts. Selling never waits for it.
type SalesSync struct {
	client  Client
	orders  orderiface.Sync
	events  eventiface.Events
	catalog catalogiface.Catalog
}

func NewSalesSync(client Client, orders orderiface.Sync, events eventiface.Events, catalog catalogiface.Catalog) *SalesSync {
	return &SalesSync{client: client, orders: orders, events: events, catalog: catalog}
}

// RunOnce sends what is due and returns how many orders it handled.
func (s *SalesSync) RunOnce(ctx context.Context) int {
	due, err := s.orders.DueForSync(salesBatch)
	if err != nil {
		log.Printf("bhoomy sales sync: %v", err)
		return 0
	}
	locations := map[uuid.UUID]*eventiface.EventResponse{}
	for _, o := range due {
		if err := s.send(ctx, o, locations); err != nil {
			log.Printf("bhoomy sales sync %s: %v", o.ReferenceID, err)
		}
	}
	return len(due)
}

func (s *SalesSync) send(ctx context.Context, o orderiface.SyncOrder, locations map[uuid.UUID]*eventiface.EventResponse) error {
	loc, ok := locations[o.EventID]
	if !ok {
		var err error
		if loc, err = s.events.Get(o.EventID); err != nil {
			return s.orders.MarkSyncRetry(o.ID, o.Attempts, err.Error())
		}
		locations[o.EventID] = loc
	}
	store, err := bhoomyStore(s.catalog, loc.StoreID)
	if err != nil {
		return s.orders.MarkSyncRetry(o.ID, o.Attempts, err.Error())
	}
	if store == uuid.Nil {
		return s.orders.MarkSyncDone(o.ID, "skipped", "store is not on Bhoomy")
	}
	status := ""
	if !o.Sent {
		if status, err = s.client.PostSale(ctx, store, saleOf(o, loc.Name)); err != nil {
			return s.failed(o, err)
		}
	}
	if o.Cancelled {
		if status, err = s.client.CancelSale(ctx, store, o.ID, time.Now().UTC()); err != nil {
			return s.failed(o, err)
		}
	}
	return s.orders.MarkSynced(o.ID, o.Cancelled, status)
}

func (s *SalesSync) failed(o orderiface.SyncOrder, err error) error {
	var refused *RefusedError
	if errors.As(err, &refused) {
		return s.orders.MarkSyncDone(o.ID, "failed", err.Error())
	}
	return s.orders.MarkSyncRetry(o.ID, o.Attempts, err.Error())
}

func saleOf(o orderiface.SyncOrder, locationName string) Sale {
	sale := Sale{PosSaleID: o.ID, PosLocationID: o.EventID, PosLocationName: locationName, PlaceID: o.PlaceID,
		ReceiptNo: o.ReferenceID, SoldBy: o.SoldBy, Total: o.Total.IntPart(), SoldAt: o.SoldAt}
	for _, l := range o.Lines {
		sale.Lines = append(sale.Lines, SaleLine{PosLineID: l.ID, VariantID: l.BhoomyVariantID, SKU: l.SKU, Name: l.Name,
			Qty: l.Qty, UnitPrice: l.UnitPrice.IntPart(), NetAmount: l.LineTotal.IntPart()})
	}
	return sale
}
