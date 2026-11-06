package iface

import (
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"

	eventiface "github.com/krnskrll/event-mpos/internal/event/iface"
)

type ProductReportRow struct {
	ProductID uuid.UUID       `json:"productId"`
	SKU       string          `json:"sku"`
	Name      string          `json:"name"`
	UnitsSold int             `json:"unitsSold"`
	Revenue   decimal.Decimal `json:"revenue"`
	Allocated int             `json:"allocated"`
	Sold      int             `json:"sold"`
	Remaining int             `json:"remaining"`
}

type EndOfEventReport struct {
	EventID     uuid.UUID              `json:"eventId"`
	EventCode   string                 `json:"eventCode"`
	EventName   string                 `json:"eventName"`
	Status      eventiface.EventStatus `json:"status"`
	GeneratedAt time.Time              `json:"generatedAt"`
	TotalOrders int                    `json:"totalOrders"`
	// CancelledOrders are excluded from every other total (orders, items, revenue, discount).
	CancelledOrders int `json:"cancelledOrders"`
	// Orders uploaded after the event closed (D2) and orders that sold past the allocation (D1).
	LateSyncedOrders    int                `json:"lateSyncedOrders"`
	StockConflictOrders int                `json:"stockConflictOrders"`
	TotalItems          int                `json:"totalItems"`
	TotalRevenue        decimal.Decimal    `json:"totalRevenue"`
	TotalDiscount       decimal.Decimal    `json:"totalDiscount"`
	Products            []ProductReportRow `json:"products"`
}

// SalesRow is one item (and variant) sold in the period; revenue is before the order discount.
type SalesRow struct {
	ProductID uuid.UUID       `json:"productId"`
	SKU       string          `json:"sku"`
	Name      string          `json:"name"`
	UnitsSold int             `json:"unitsSold"`
	Revenue   decimal.Decimal `json:"revenue"`
}

// SalesReport is a location's sales between two WIB dates, both included. It is final once no
// more orders can arrive: the event is closed or the period ended over a day ago.
type SalesReport struct {
	LocationID       uuid.UUID       `json:"locationId"`
	From             string          `json:"from"`
	To               string          `json:"to"`
	Final            bool            `json:"final"`
	TotalOrders      int             `json:"totalOrders"`
	CancelledOrders  int             `json:"cancelledOrders"`
	LateSyncedOrders int             `json:"lateSyncedOrders"`
	TotalItems       int             `json:"totalItems"`
	ProductCount     int             `json:"productCount"`
	PromotionsUsed   int             `json:"promotionsUsed"`
	TotalRevenue     decimal.Decimal `json:"totalRevenue"`
	TotalDiscount    decimal.Decimal `json:"totalDiscount"`
	Products         []SalesRow      `json:"products"`
}
