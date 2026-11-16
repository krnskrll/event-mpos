package service

import (
	"time"

	"github.com/google/uuid"

	"github.com/krnskrll/event-mpos/internal/order/iface"
	"github.com/krnskrll/event-mpos/internal/order/repository"
)

var _ iface.Sync = (*Service)(nil)

// DueForSync is the oldest orders waiting to be sent to Bhoomy (6c), with their lines.
func (s *Service) DueForSync(limit int) ([]iface.SyncOrder, error) {
	var orders []repository.Order
	err := s.db.Preload("Items").Where("sync_state = ? AND next_sync_at <= ?", repository.SyncPending, time.Now().UTC()).
		Order("created_at").Limit(limit).Find(&orders).Error
	if err != nil {
		return nil, err
	}
	out := make([]iface.SyncOrder, 0, len(orders))
	for _, o := range orders {
		so := iface.SyncOrder{ID: o.ID, EventID: o.EventID, ReferenceID: o.ReferenceID, PlaceID: o.BhoomyPlaceID, Total: o.Total,
			SoldAt: o.CreatedAt, Cancelled: o.Status == iface.StatusCancelled, Sent: o.SyncedAt != nil, Attempts: o.SyncAttempts}
		if o.ClientCreatedAt != nil {
			so.SoldAt = *o.ClientCreatedAt
		}
		if o.StaffID != nil {
			so.SoldBy, _ = s.names.NameOf(s.db, *o.StaffID)
		}
		for _, it := range o.Items {
			name := it.Name
			if it.VariantName != nil {
				name += " · " + *it.VariantName
			}
			so.Lines = append(so.Lines, iface.SyncLine{ID: it.ID, BhoomyVariantID: it.BhoomyVariantID, SKU: it.SKU, Name: name,
				Qty: it.Qty, UnitPrice: it.UnitPrice, LineTotal: it.LineTotal})
		}
		out = append(out, so)
	}
	return out, nil
}

// MarkSynced records Bhoomy's answer. If the order was cancelled after this send began it stays
// pending, so its cancel goes next.
func (s *Service) MarkSynced(id uuid.UUID, cancelled bool, stockStatus string) error {
	state := "CASE WHEN status = 'CANCELLED' THEN 'pending' ELSE 'sent' END"
	if cancelled {
		state = "'sent'"
	}
	return s.db.Exec(`UPDATE orders SET sync_state = `+state+`, synced_at = COALESCE(synced_at, now()),
		stock_status = ?, stock_conflict = ?, sync_attempts = 0, sync_error = NULL WHERE id = ?`,
		stockStatus, stockStatus == "conflict", id).Error
}

// MarkSyncRetry waits longer after each failure: 15 s, 30 s, 1 min … capped at 1 hour.
func (s *Service) MarkSyncRetry(id uuid.UUID, attempts int, msg string) error {
	wait := 15 * time.Second << min(attempts, 8)
	return s.db.Model(&repository.Order{}).Where("id = ?", id).Updates(map[string]any{
		"sync_attempts": attempts + 1, "next_sync_at": time.Now().UTC().Add(min(wait, time.Hour)), "sync_error": msg,
	}).Error
}

func (s *Service) MarkSyncDone(id uuid.UUID, state, msg string) error {
	return s.db.Model(&repository.Order{}).Where("id = ?", id).Updates(map[string]any{"sync_state": state, "sync_error": msg}).Error
}
