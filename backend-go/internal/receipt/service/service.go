// Package service records that a receipt was printed for an order (printing happens on the device;
// this is the backend's minimal audit counter — one row per order, with a reprint counter).
package service

import (
	"net/http"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	auditiface "github.com/krnskrll/event-mpos/internal/audit/iface"
	orderiface "github.com/krnskrll/event-mpos/internal/order/iface"
	"github.com/krnskrll/event-mpos/internal/platform/common"
	"github.com/krnskrll/event-mpos/internal/receipt/iface"
	"github.com/krnskrll/event-mpos/internal/receipt/repository"
)

var _ iface.Receipts = (*Service)(nil)

type Service struct {
	db     *gorm.DB
	orders orderiface.Orders
	audit  auditiface.Audit
}

func NewService(db *gorm.DB, orders orderiface.Orders, aud auditiface.Audit) *Service {
	return &Service{db: db, orders: orders, audit: aud}
}

// LogPrint records a print/reprint of an order sold at the location. In one transaction it bumps the
// receipt_logs counter (atomic upsert, safe under concurrent prints) and appends a RECEIPT_PRINTED
// audit entry. An unknown order, or one from another location, is a 404.
func (s *Service) LogPrint(orderID, locationID, staffID uuid.UUID) error {
	return s.db.Transaction(func(tx *gorm.DB) error {
		o, err := s.orders.ForReceipt(tx, orderID, locationID)
		if err != nil {
			return err
		}
		if o == nil {
			return common.NewAppError(http.StatusNotFound, "Order not found")
		}

		now := time.Now().UTC()
		row := repository.ReceiptLog{OrderID: orderID, PrintCount: 1, FirstPrintedAt: now, LastPrintedAt: now}
		if err := tx.Clauses(clause.OnConflict{
			Columns: []clause.Column{{Name: "order_id"}},
			DoUpdates: clause.Assignments(map[string]any{
				"print_count":     gorm.Expr("receipt_logs.print_count + 1"),
				"last_printed_at": now,
			}),
		}).Create(&row).Error; err != nil {
			return err
		}
		var printCount int
		if err := tx.Model(&repository.ReceiptLog{}).Where("order_id = ?", orderID).Pluck("print_count", &printCount).Error; err != nil {
			return err
		}
		return s.audit.RecordReceiptPrinted(tx, o.EventID, o.ID, o.ReferenceID, &staffID, printCount)
	})
}
