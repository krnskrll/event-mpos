package repository

import (
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	"gorm.io/gorm"

	"github.com/krnskrll/event-mpos/internal/order/iface"
)

// Order is a confirmed order, created atomically with stock decrement + promotion application.
type Order struct {
	ID             uuid.UUID         `gorm:"type:uuid;primaryKey"`
	EventID        uuid.UUID         `gorm:"column:event_id;type:uuid;not null;index:idx_order_event"`
	BaUserID       uuid.UUID         `gorm:"column:ba_user_id;type:uuid;not null"`
	IdempotencyKey string            `gorm:"column:idempotency_key;not null;uniqueIndex:idx_order_idem"`
	OrderNumber    int               `gorm:"column:order_number;not null"`
	ReferenceID    string            `gorm:"column:reference_id;not null"`
	CustomerName   string            `gorm:"column:customer_name;not null"`
	CustomerQr     *string           `gorm:"column:customer_qr"`
	CustomerTier   *string           `gorm:"column:customer_tier"`
	Subtotal       decimal.Decimal   `gorm:"type:numeric(12,2);not null"`
	Discount       decimal.Decimal   `gorm:"type:numeric(12,2);not null"`
	Total          decimal.Decimal   `gorm:"type:numeric(12,2);not null"`
	PromotionID    *uuid.UUID        `gorm:"column:promotion_id;type:uuid"`
	PromotionName  *string           `gorm:"column:promotion_name"`
	Status         iface.OrderStatus `gorm:"type:varchar(20);not null"`
	Items          []OrderItem       `gorm:"foreignKey:OrderID;constraint:OnDelete:CASCADE"`
	// Set for orders confirmed on a device with a number from its block (EPIC-12 C1/C2).
	DeviceID        *string    `gorm:"column:device_id"`
	ClientCreatedAt *time.Time `gorm:"column:client_created_at"`
	StockConflict   bool       `gorm:"column:stock_conflict;not null;default:false"` // sold past the allocation
	LateSynced      bool       `gorm:"column:late_synced;not null;default:false"`    // uploaded after the event closed
	CreatedAt       time.Time  `gorm:"column:created_at;not null"`
}

func (Order) TableName() string { return "orders" }

func (o *Order) BeforeCreate(_ *gorm.DB) error {
	if o.ID == uuid.Nil {
		o.ID = uuid.New()
	}
	if o.CreatedAt.IsZero() {
		o.CreatedAt = time.Now().UTC()
	}
	return nil
}

// OrderItem is a line on an order — snapshots name/price/variant at purchase time.
type OrderItem struct {
	ID          uuid.UUID       `gorm:"type:uuid;primaryKey"`
	OrderID     uuid.UUID       `gorm:"column:order_id;type:uuid;not null;index:idx_order_item_order"`
	ProductID   uuid.UUID       `gorm:"column:product_id;type:uuid;not null"`
	VariantID   *uuid.UUID      `gorm:"column:variant_id;type:uuid"`
	VariantName *string         `gorm:"column:variant_name"`
	SKU         string          `gorm:"column:sku;not null"`
	Name        string          `gorm:"not null"`
	UnitPrice   decimal.Decimal `gorm:"column:unit_price;type:numeric(12,2);not null"`
	Qty         int             `gorm:"not null"`
	LineTotal   decimal.Decimal `gorm:"column:line_total;type:numeric(12,2);not null"`
}

func (OrderItem) TableName() string { return "order_items" }

func (i *OrderItem) BeforeCreate(_ *gorm.DB) error {
	if i.ID == uuid.Nil {
		i.ID = uuid.New()
	}
	return nil
}

// NumberBlock is a range of one day's order numbers reserved for a device, so it can confirm orders
// offline with final numbers. Carved from OrderDailySequence under its row lock; never reused.
type NumberBlock struct {
	ID           uuid.UUID `gorm:"type:uuid;primaryKey"`
	DeviceID     string    `gorm:"column:device_id;not null;index:idx_number_block_device,priority:1"`
	EventID      uuid.UUID `gorm:"column:event_id;type:uuid;not null;index:idx_number_block_device,priority:2"`
	BusinessDate string    `gorm:"column:business_date;not null"`
	RangeStart   int       `gorm:"column:range_start;not null"`
	RangeEnd     int       `gorm:"column:range_end;not null"`
	NextUnused   int       `gorm:"column:next_unused;not null"` // highest uploaded number + 1 (informational)
	CreatedAt    time.Time `gorm:"column:created_at;not null"`
}

func (NumberBlock) TableName() string { return "device_number_block" }

func (b *NumberBlock) BeforeCreate(_ *gorm.DB) error {
	if b.ID == uuid.Nil {
		b.ID = uuid.New()
	}
	return nil
}

// OrderDailySequence is a per-event, per-day counter for the human order number (#042).
type OrderDailySequence struct {
	ID         uuid.UUID `gorm:"type:uuid;primaryKey"`
	EventID    uuid.UUID `gorm:"column:event_id;type:uuid;not null;uniqueIndex:uk_seq_event_date"`
	DateKey    string    `gorm:"column:date_key;not null;uniqueIndex:uk_seq_event_date"`
	LastNumber int       `gorm:"column:last_number;not null"`
	Version    int64     `gorm:"not null"`
}

func (OrderDailySequence) TableName() string { return "order_daily_sequences" }

func (s *OrderDailySequence) BeforeCreate(_ *gorm.DB) error {
	if s.ID == uuid.Nil {
		s.ID = uuid.New()
	}
	return nil
}
