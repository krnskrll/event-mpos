package iface

import (
	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

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
