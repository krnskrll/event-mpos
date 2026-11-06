package iface

import (
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"

	"github.com/krnskrll/event-mpos/internal/platform/database"
)

// ProductSource — where a product came from.
type ProductSource string

const (
	SourceERP    ProductSource = "ERP"
	SourceManual ProductSource = "MANUAL"
)

type BarcodeDto struct {
	Barcode string  `json:"barcode"`
	Label   *string `json:"label,omitempty"`
}

type VariantDto struct {
	ID     uuid.UUID       `json:"id"`
	Name   string          `json:"name"`
	SKU    string          `json:"sku"`
	Price  decimal.Decimal `json:"price"`
	Active bool            `json:"active"`
}

type ProductResponse struct {
	ID               uuid.UUID       `json:"id"`
	DefaultVariantID uuid.UUID       `json:"defaultVariantId"`
	SKU              string          `json:"sku"`
	Name             string          `json:"name"`
	Brand            *string         `json:"brand,omitempty"`
	Price            decimal.Decimal `json:"price"`
	ImageURL         *string         `json:"imageUrl,omitempty"`
	Source           ProductSource   `json:"source"`
	Active           bool            `json:"active"`
	WarehouseStock   int             `json:"warehouseStock"`
	Barcodes         []BarcodeDto    `json:"barcodes"`
	Variants         []VariantDto    `json:"variants"`
	UpdatedAt        time.Time       `json:"updatedAt"`
}

type BarcodeInput struct {
	Barcode string  `json:"barcode"`
	Label   *string `json:"label"`
}

type CreateProductRequest struct {
	SKU      string          `json:"sku" binding:"required"`
	Name     string          `json:"name" binding:"required"`
	Brand    *string         `json:"brand"`
	Price    decimal.Decimal `json:"price"`
	ImageURL *string         `json:"imageUrl"`
	Barcodes []BarcodeInput  `json:"barcodes"`
}

type UpdateProductRequest struct {
	Name     *string          `json:"name"`
	Brand    *string          `json:"brand"`
	Price    *decimal.Decimal `json:"price"`
	ImageURL *string          `json:"imageUrl"`
	Active   *bool            `json:"active"`
	Barcodes *[]BarcodeInput  `json:"barcodes"`
}

type EventProductResponse struct {
	Product   ProductResponse      `json:"product"`
	Available bool                 `json:"available"`
	Variants  []LocationVariantDto `json:"variants"`
}

// LocationVariantDto is one variant at a location, the default one included.
type LocationVariantDto struct {
	ID            uuid.UUID        `json:"id"`
	Name          string           `json:"name"`
	IsDefault     bool             `json:"isDefault"`
	Available     bool             `json:"available"`
	LocationPrice *decimal.Decimal `json:"locationPrice"`
}

type SetAvailabilityRequest struct {
	Available *bool `json:"available" binding:"required"`
}

// SetLocationVariantRequest sets a variant's state at a location; an empty price clears it.
type SetLocationVariantRequest struct {
	Available *bool            `json:"available" binding:"required"`
	Price     *decimal.Decimal `json:"price"`
}

type VariantSnapshotDto struct {
	ID    uuid.UUID       `json:"id"`
	Name  string          `json:"name"`
	Price decimal.Decimal `json:"price"`
}

// ProductSnapshotDto is the compact product shape carried in the event snapshot.
type ProductSnapshotDto struct {
	ID             uuid.UUID            `json:"id"`
	SKU            string               `json:"sku"`
	Name           string               `json:"name"`
	Brand          *string              `json:"brand,omitempty"`
	Price          decimal.Decimal      `json:"price"`
	ImageURL       *string              `json:"imageUrl,omitempty"`
	Barcodes       []string             `json:"barcodes"`
	Variants       []VariantSnapshotDto `json:"variants"`
	RemainingStock int                  `json:"remainingStock"`
}

// SaleProduct carries the server-authoritative prices used when selling.
type SaleProduct struct {
	ID             uuid.UUID
	SKU            string
	Name           string
	Price          decimal.Decimal
	WarehouseStock int
	Variants       []SaleVariant
}

// SaleVariant.Price is the effective price (override or product price).
type SaleVariant struct {
	ID     uuid.UUID
	Name   string
	Price  decimal.Decimal
	Active bool
}

// SourceProduct is a product as delivered by the ERP.
type SourceProduct struct {
	SKU           string
	Name          string
	Brand         *string
	Price         decimal.Decimal
	ImageURL      *string
	StockQuantity int
	Barcodes      []BarcodeInput
}

// Products is the product module's contract for other modules.
type Products interface {
	// Find returns nil when the product does not exist.
	Find(id uuid.UUID) (*SaleProduct, error)
	// FindAt is Find with the location's prices.
	FindAt(locationID, id uuid.UUID) (*SaleProduct, error)
	// All returns the store's products ordered by SKU, without variants.
	All(storeID uuid.UUID) ([]SaleProduct, error)
	Categories(storeID uuid.UUID) ([]Category, error)
	// RequireItems answers 400 unless every id is an item of the store.
	RequireItems(storeID uuid.UUID, ids []uuid.UUID) error
	FindAvailableByBarcode(storeID, eventID uuid.UUID, barcode string) (*ProductResponse, error)
	// Upsert creates or updates by SKU inside tx; created reports a new row.
	Upsert(tx database.Tx, storeID uuid.UUID, p SourceProduct) (created bool, err error)
}
