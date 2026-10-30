// Package receipt records that a receipt was printed for an order (printing happens on the device;
// this is the backend's minimal audit counter — one row per order, with a reprint counter).
package receipt

import (
	"errors"
	"net/http"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"github.com/krnskrll/event-mpos/internal/audit"
	"github.com/krnskrll/event-mpos/internal/common"
	"github.com/krnskrll/event-mpos/internal/order"
)

type ReceiptLog struct {
	ID             uuid.UUID `gorm:"type:uuid;primaryKey"`
	OrderID        uuid.UUID `gorm:"column:order_id;uniqueIndex;not null"`
	PrintCount     int       `gorm:"column:print_count;not null"`
	FirstPrintedAt time.Time `gorm:"column:first_printed_at;not null"`
	LastPrintedAt  time.Time `gorm:"column:last_printed_at;not null"`
}

func (ReceiptLog) TableName() string { return "receipt_logs" }

func (r *ReceiptLog) BeforeCreate(_ *gorm.DB) error {
	if r.ID == uuid.Nil {
		r.ID = uuid.New()
	}
	return nil
}

type Service struct {
	db    *gorm.DB
	audit *audit.Service
}

func NewService(db *gorm.DB, aud *audit.Service) *Service { return &Service{db: db, audit: aud} }

// LogPrint records a print/reprint of one of the BA's own orders. In one transaction it bumps the
// receipt_logs counter (atomic upsert, safe under concurrent prints) and appends a RECEIPT_PRINTED
// audit entry. An unknown order, or another BA's order, is a 404.
func (s *Service) LogPrint(orderID, baUserID uuid.UUID) error {
	return s.db.Transaction(func(tx *gorm.DB) error {
		var o order.Order
		err := tx.Select("id", "event_id", "reference_id", "ba_user_id").
			Where("id = ? AND ba_user_id = ?", orderID, baUserID).Take(&o).Error
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return common.NewAppError(http.StatusNotFound, "Order not found")
		}
		if err != nil {
			return err
		}

		now := time.Now().UTC()
		row := ReceiptLog{OrderID: orderID, PrintCount: 1, FirstPrintedAt: now, LastPrintedAt: now}
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
		if err := tx.Model(&ReceiptLog{}).Where("order_id = ?", orderID).Pluck("print_count", &printCount).Error; err != nil {
			return err
		}
		return s.audit.RecordReceiptPrinted(tx, o.EventID, o.ID, o.ReferenceID, &baUserID, printCount)
	})
}
