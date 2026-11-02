package iface

import (
	"github.com/google/uuid"

	"github.com/krnskrll/event-mpos/internal/platform/database"
)

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

// Allocation is the stock allocated to an event for one product.
type Allocation struct {
	ProductID    uuid.UUID
	AllocatedQty int
	SoldQty      int
	Remaining    int
}

// Stock is the stock module's contract. Decrement, Reserve and Restore run inside the caller's tx.
type Stock interface {
	Decrement(tx database.Tx, eventID, productID uuid.UUID, qty int) error
	Reserve(tx database.Tx, eventID, productID uuid.UUID, qty int) (oversold bool, err error)
	Restore(tx database.Tx, eventID, productID uuid.UUID, qty int) error
	InvalidateSummary(eventID uuid.UUID)
	Remaining(eventID, productID uuid.UUID) (int, error)
	AllocationFingerprint(eventID uuid.UUID) (string, error)
	BaLevelsForEvent(eventID uuid.UUID) ([]BaStockLevel, error)
	Allocations(eventID uuid.UUID) ([]Allocation, error)
}
