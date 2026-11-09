package iface

import (
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"

	"github.com/krnskrll/event-mpos/internal/platform/database"
)

type OrderStatus string

const (
	StatusNew       OrderStatus = "NEW"
	StatusPacking   OrderStatus = "PACKING"
	StatusReady     OrderStatus = "READY"
	StatusCollected OrderStatus = "COLLECTED"
	// StatusCancelled is terminal and off the forward flow; reachable only via Service.Cancel.
	StatusCancelled OrderStatus = "CANCELLED"
)

var statusOrder = []OrderStatus{StatusNew, StatusPacking, StatusReady, StatusCollected}

func ParseStatus(s string) (OrderStatus, bool) {
	for _, st := range append(statusOrder, StatusCancelled) {
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

// CanCancel reports whether the order may still be cancelled (nothing has left the warehouse yet).
func (s OrderStatus) CanCancel() bool { return s == StatusNew || s == StatusPacking }

// CanAdvanceTo reports whether target is exactly the next status forward.
func (s OrderStatus) CanAdvanceTo(target OrderStatus) bool {
	n, ok := s.next()
	return ok && n == target
}

// NumberBlockSize is how many order numbers a device gets per block (it asks for more below 10 left).
const NumberBlockSize = 50

// NumberBlockResponse: POST /api/me/events/{id}/number-blocks, issued to the calling device.
type NumberBlockResponse struct {
	EventID      uuid.UUID `json:"eventId"`
	BusinessDate string    `json:"businessDate"`
	RangeStart   int       `json:"rangeStart"`
	RangeEnd     int       `json:"rangeEnd"`
}

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
	// Offline-confirmed orders carry the device's final number (from its block) and creation time.
	// The server sets DeviceID to the paired device; a value sent by the app is ignored.
	DeviceID        *string    `json:"deviceId"`
	OrderNumber     *int       `json:"orderNumber"`
	ClientCreatedAt *time.Time `json:"clientCreatedAt"`
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
	ID                 uuid.UUID       `json:"id"`
	OrderNumber        int             `json:"orderNumber"`
	OrderNumberDisplay string          `json:"orderNumberDisplay"`
	ReferenceID        string          `json:"referenceId"`
	Status             OrderStatus     `json:"status"`
	EventCode          string          `json:"eventCode"`
	EventName          string          `json:"eventName"`
	CustomerName       string          `json:"customerName"`
	CustomerTier       *string         `json:"customerTier,omitempty"`
	Items              []OrderItemView `json:"items"`
	Subtotal           decimal.Decimal `json:"subtotal"`
	Discount           decimal.Decimal `json:"discount"`
	Total              decimal.Decimal `json:"total"`
	PromotionName      *string         `json:"promotionName,omitempty"`
	CreatedAt          time.Time       `json:"createdAt"`
	CashierNote        string          `json:"cashierNote"`
	Duplicate          bool            `json:"duplicate"`
	StockConflict      bool            `json:"stockConflict"`
	LateSynced         bool            `json:"lateSynced"`
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
	StockConflict      bool            `json:"stockConflict"`
	LateSynced         bool            `json:"lateSynced"`
}

// CancelOrderRequest: a supervisor of the location approves with their PIN.
type CancelOrderRequest struct {
	Reason        *string   `json:"reason"`
	SupervisorID  uuid.UUID `json:"supervisorId"`
	SupervisorPin string    `json:"supervisorPin"`
}

type AdvanceStatusRequest struct {
	Status OrderStatus `json:"status" binding:"required"`
}

// ReceiptOrder is what the receipt module needs to log a print.
type ReceiptOrder struct {
	ID          uuid.UUID
	EventID     uuid.UUID
	ReferenceID string
}

// SoldOrder is a non-cancelled order as reports see it.
type SoldOrder struct {
	PromotionID *uuid.UUID
	Total       decimal.Decimal
	Discount    decimal.Decimal
	LateSynced  bool
	Items       []SoldItem
}

type SoldItem struct {
	ProductID   uuid.UUID
	SKU         string
	Name        string
	VariantName *string
	Qty         int
	LineTotal   decimal.Decimal
}

type Orders interface {
	Submit(eventID, staffID uuid.UUID, req SubmitOrderRequest) (*OrderConfirmation, error)
	IssueNumberBlock(eventID uuid.UUID, deviceID string) (*NumberBlockResponse, error)
	// History is the orders a device sold at the location, newest first.
	History(eventID uuid.UUID, deviceID string) ([]OrderSummary, error)
	Get(orderID uuid.UUID) (*OrderConfirmation, error)
	// LocationOf is the location an order was sold at (404 when it does not exist).
	LocationOf(orderID uuid.UUID) (uuid.UUID, error)
	// ForReceipt returns nil when the order does not exist or was sold at another location.
	ForReceipt(tx database.Tx, orderID, locationID uuid.UUID) (*ReceiptOrder, error)
	// ForReport returns non-cancelled orders sold in [from, to) newest first, plus the cancelled
	// count. A zero bound is open; the sale time is the device's for offline orders.
	ForReport(eventID uuid.UUID, from, to time.Time) ([]SoldOrder, int, error)
}
