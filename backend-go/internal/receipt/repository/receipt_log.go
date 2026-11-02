package repository

import (
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
