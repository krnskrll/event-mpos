package order

import (
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	"gorm.io/gorm"
)

// OrderStatus — warehouse drives NEW→PACKING→READY→COLLECTED, strictly one step forward.
type OrderStatus string

const (
	StatusNew       OrderStatus = "NEW"
	StatusPacking   OrderStatus = "PACKING"
	StatusReady     OrderStatus = "READY"
	StatusCollected OrderStatus = "COLLECTED"
)

var statusOrder = []OrderStatus{StatusNew, StatusPacking, StatusReady, StatusCollected}

func ParseStatus(s string) (OrderStatus, bool) {
	for _, st := range statusOrder {
		if st == OrderStatus(s) {
			return st, true
		}
	}
	return "", false
}

func (s OrderStatus) next() (OrderStatus, bool) {
	for i, st := range statusOrder {
		if st == s && i+1 < len(statusOrder) {
			return statusOrder[i+1], true
		}
	}
	return "", false
}

// CanAdvanceTo reports whether target is exactly the next status forward.
func (s OrderStatus) CanAdvanceTo(target OrderStatus) bool {
	n, ok := s.next()
	return ok && n == target
}

// Order is a confirmed order, created atomically with stock decrement + promotion application.
type Order struct {
	ID             uuid.UUID       `gorm:"type:uuid;primaryKey"`
	EventID        uuid.UUID       `gorm:"column:event_id;type:uuid;not null;index:idx_order_event"`
	BaUserID       uuid.UUID       `gorm:"column:ba_user_id;type:uuid;not null"`
	IdempotencyKey string          `gorm:"column:idempotency_key;not null;uniqueIndex:idx_order_idem"`
	OrderNumber    int             `gorm:"column:order_number;not null"`
	ReferenceID    string          `gorm:"column:reference_id;not null"`
	CustomerName   string          `gorm:"column:customer_name;not null"`
	CustomerQr     *string         `gorm:"column:customer_qr"`
	CustomerTier   *string         `gorm:"column:customer_tier"`
	Subtotal       decimal.Decimal `gorm:"type:numeric(12,2);not null"`
	Discount       decimal.Decimal `gorm:"type:numeric(12,2);not null"`
	Total          decimal.Decimal `gorm:"type:numeric(12,2);not null"`
	PromotionID    *uuid.UUID      `gorm:"column:promotion_id;type:uuid"`
	PromotionName  *string         `gorm:"column:promotion_name"`
	Status         OrderStatus     `gorm:"type:varchar(20);not null"`
	Items          []OrderItem     `gorm:"foreignKey:OrderID;constraint:OnDelete:CASCADE"`
	CreatedAt      time.Time       `gorm:"column:created_at;not null"`
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

// ---- request DTOs ----

type CustomerInfo struct {
	Name   string  `json:"name" binding:"required"`
	QRCode *string `json:"qrCode"`
	Tier   *string `json:"tier"`
}

type OrderLineRequest struct {
	ProductID uuid.UUID  `json:"productId"`
	Qty       int        `json:"qty"`
	VariantID *uuid.UUID `json:"variantId"`
}

type SubmitOrderRequest struct {
	IdempotencyKey string             `json:"idempotencyKey" binding:"required"`
	Customer       CustomerInfo       `json:"customer"`
	Lines          []OrderLineRequest `json:"lines"`
}

// ---- response DTOs ----

type OrderItemView struct {
	ProductID   uuid.UUID       `json:"productId"`
	SKU         string          `json:"sku"`
	Name        string          `json:"name"`
	UnitPrice   decimal.Decimal `json:"unitPrice"`
	Qty         int             `json:"qty"`
	LineTotal   decimal.Decimal `json:"lineTotal"`
	VariantName *string         `json:"variantName,omitempty"`
}

type OrderConfirmation struct {
	ID                uuid.UUID       `json:"id"`
	OrderNumber       int             `json:"orderNumber"`
	OrderNumberDisplay string         `json:"orderNumberDisplay"`
	ReferenceID       string          `json:"referenceId"`
	Status            OrderStatus     `json:"status"`
	EventCode         string          `json:"eventCode"`
	EventName         string          `json:"eventName"`
	CustomerName      string          `json:"customerName"`
	CustomerTier      *string         `json:"customerTier,omitempty"`
	Items             []OrderItemView `json:"items"`
	Subtotal          decimal.Decimal `json:"subtotal"`
	Discount          decimal.Decimal `json:"discount"`
	Total             decimal.Decimal `json:"total"`
	PromotionName     *string         `json:"promotionName,omitempty"`
	CreatedAt         time.Time       `json:"createdAt"`
	CashierNote       string          `json:"cashierNote"`
	Duplicate         bool            `json:"duplicate"`
}

type OrderSummary struct {
	ID                 uuid.UUID       `json:"id"`
	OrderNumberDisplay string          `json:"orderNumberDisplay"`
	ReferenceID        string          `json:"referenceId"`
	Status             OrderStatus     `json:"status"`
	CustomerName       string          `json:"customerName"`
	Total              decimal.Decimal `json:"total"`
	ItemCount          int             `json:"itemCount"`
	CreatedAt          time.Time       `json:"createdAt"`
}

// ---- warehouse board ----

type WarehouseItem struct {
	Name string `json:"name"`
	SKU  string `json:"sku"`
	Qty  int    `json:"qty"`
}

type WarehouseOrderView struct {
	ID                 uuid.UUID       `json:"id"`
	OrderNumber        int             `json:"orderNumber"`
	OrderNumberDisplay string          `json:"orderNumberDisplay"`
	ReferenceID        string          `json:"referenceId"`
	CustomerName       string          `json:"customerName"`
	BaName             string          `json:"baName"`
	Items              []WarehouseItem `json:"items"`
	Status             OrderStatus     `json:"status"`
	Total              decimal.Decimal `json:"total"`
	CreatedAt          time.Time       `json:"createdAt"`
}

type AdvanceStatusRequest struct {
	Status OrderStatus `json:"status" binding:"required"`
}
