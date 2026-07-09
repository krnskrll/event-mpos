package stock

import (
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// LowStockThreshold — remaining in 1..threshold flags a low-stock warning (0 = out, not "low").
const LowStockThreshold = 10

// EventStockAllocation is stock allocated to an event for a product. Once allocated, the platform
// owns this count. remaining = allocatedQty - soldQty and must never go negative.
type EventStockAllocation struct {
	ID          uuid.UUID `gorm:"type:uuid;primaryKey"`
	EventID     uuid.UUID `gorm:"column:event_id;type:uuid;not null;uniqueIndex:uk_event_stock;index:idx_event_stock_event"`
	ProductID   uuid.UUID `gorm:"column:product_id;type:uuid;not null;uniqueIndex:uk_event_stock"`
	AllocatedQty int      `gorm:"column:allocated_qty;not null"`
	SoldQty     int       `gorm:"column:sold_qty;not null"`
	// Optimistic-lock column kept for schema parity; the real oversell guard is the pessimistic
	// FOR UPDATE row lock in Decrement.
	Version int64 `gorm:"not null"`
}

func (EventStockAllocation) TableName() string { return "event_stock_allocations" }

func (a *EventStockAllocation) BeforeCreate(_ *gorm.DB) error {
	if a.ID == uuid.Nil {
		a.ID = uuid.New()
	}
	return nil
}

func (a *EventStockAllocation) Remaining() int { return a.AllocatedQty - a.SoldQty }

// ---- DTOs ----

type AllocationRow struct {
	ProductID      uuid.UUID `json:"productId"`
	SKU            string    `json:"sku"`
	Name           string    `json:"name"`
	WarehouseStock int       `json:"warehouseStock"`
	AllocatedQty   int       `json:"allocatedQty"`
	SoldQty        int       `json:"soldQty"`
	Remaining      int       `json:"remaining"`
}

type SetAllocationRequest struct {
	AllocatedQty int `json:"allocatedQty"`
}

type StockLevelRow struct {
	ProductID    uuid.UUID `json:"productId"`
	SKU          string    `json:"sku"`
	Name         string    `json:"name"`
	AllocatedQty int       `json:"allocatedQty"`
	SoldQty      int       `json:"soldQty"`
	Remaining    int       `json:"remaining"`
	LowStock     bool      `json:"lowStock"`
}

type BaStockLevel struct {
	ProductID uuid.UUID `json:"productId"`
	Remaining int       `json:"remaining"`
	LowStock  bool      `json:"lowStock"`
}

func lowStock(remaining int) bool { return remaining >= 1 && remaining <= LowStockThreshold }
