// Package receipt records that a receipt was printed for an order (printing happens on the device;
// this is the backend's minimal audit counter — one row per order, with a reprint counter).
package receipt

import (
	"errors"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
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
	db *gorm.DB
}

func NewService(db *gorm.DB) *Service { return &Service{db: db} }

// LogPrint records a print/reprint: first print inserts the row; reprints bump the counter.
func (s *Service) LogPrint(orderID uuid.UUID) error {
	var existing ReceiptLog
	err := s.db.Where("order_id = ?", orderID).First(&existing).Error
	now := time.Now().UTC()
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return s.db.Create(&ReceiptLog{OrderID: orderID, PrintCount: 1, FirstPrintedAt: now, LastPrintedAt: now}).Error
	}
	if err != nil {
		return err
	}
	existing.PrintCount++
	existing.LastPrintedAt = now
	return s.db.Save(&existing).Error
}
