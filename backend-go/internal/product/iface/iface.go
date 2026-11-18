package iface

import (
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

// ProductSource — where a product came from.
type ProductSource string

const (
	SourceERP    ProductSource = "ERP" // older dev-seed rows
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
	Barcodes         []BarcodeDto    `json:"barcodes"`
	Variants         []VariantDto    `json:"variants"`
	UpdatedAt        time.Time       `json:"updatedAt"`
}

type VariantSnapshotDto struct {
	ID    uuid.UUID       `json:"id"`
	Name  string          `json:"name"`
	Price decimal.Decimal `json:"price"`
}

// ProductSnapshotDto is the compact product shape carried in the event snapshot.
type ProductSnapshotDto struct {
	ID       uuid.UUID            `json:"id"`
	SKU      string               `json:"sku"`
	Name     string               `json:"name"`
	Brand    *string              `json:"brand,omitempty"`
	Price    decimal.Decimal      `json:"price"`
	ImageURL *string              `json:"imageUrl,omitempty"`
	Barcodes []string             `json:"barcodes"`
	Variants []VariantSnapshotDto `json:"variants"`
}

// SaleProduct carries the server-authoritative prices used when selling.
type SaleProduct struct {
	ID       uuid.UUID
	SKU      string
	Name     string
	Price    decimal.Decimal
	Variants []SaleVariant
	// BhoomyVariantID is the default variant's Produk Toko variant (products without options).
	BhoomyVariantID *uuid.UUID
}

// VariantLink maps a Produk Toko variant to what the cashier app sells (VariantID nil = the
// product itself, no options).
type VariantLink struct {
	ProductID       uuid.UUID
	VariantID       *uuid.UUID
	BhoomyVariantID uuid.UUID
}

// SaleVariant.Price is the effective price (override or product price).
type SaleVariant struct {
	ID              uuid.UUID
	Name            string
	Price           decimal.Decimal
	Active          bool
	BhoomyVariantID *uuid.UUID
}

// Products is the product module's contract for other modules.
type Products interface {
	// Find returns nil when the product does not exist.
	Find(id uuid.UUID) (*SaleProduct, error)
	// FindAt is Find with the location's prices.
	FindAt(locationID, id uuid.UUID) (*SaleProduct, error)
	// VariantLinks lists the Produk Toko variants a location sells (6c stock levels).
	VariantLinks(locationID uuid.UUID) ([]VariantLink, error)
	// All returns the store's products ordered by SKU, without variants.
	All(storeID uuid.UUID) ([]SaleProduct, error)
	Categories(storeID uuid.UUID) ([]Category, error)
	// RequireItems answers 400 unless every id is an item of the store.
	RequireItems(storeID uuid.UUID, ids []uuid.UUID) error
	FindAvailableByBarcode(storeID, eventID uuid.UUID, barcode string) (*ProductResponse, error)
}
