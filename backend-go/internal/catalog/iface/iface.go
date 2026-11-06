// Package iface is the catalog module's contract: POS's read-only copy of Produk Toko (Bhoomy).
package iface

import (
	"time"

	"github.com/google/uuid"
)

// Variant is one Produk Toko variant as Bhoomy sends it.
type Variant struct {
	BhoomyVariantID uuid.UUID  `json:"variant_id"`
	BhoomyProductID uuid.UUID  `json:"product_id"`
	ProductName     string     `json:"product_name"`
	VariantName     string     `json:"variant_name"`
	SKU             string     `json:"sku"`
	Barcode         *string    `json:"barcode"`
	IsDefault       bool       `json:"is_default"`
	BrandName       *string    `json:"brand_name"`
	CategoryID      *uuid.UUID `json:"category_id"`
	CategoryName    *string    `json:"category_name"`
	ImageURL        *string    `json:"image_url"`
	Status          string     `json:"status"`
	ChangedAt       time.Time  `json:"changed_at"`
}

// Cursor marks the last change pulled: changes after (Since, After) are still to come.
type Cursor struct {
	Since time.Time `json:"since"`
	After uuid.UUID `json:"after"`
}

type Page struct {
	Items []Variant `json:"items"`
	Next  *Cursor   `json:"next"`
}

// Link is a POS store connected to its Bhoomy store.
type Link struct {
	StoreID       uuid.UUID
	BhoomyStoreID uuid.UUID
}

type SyncStatus struct {
	Linked    bool       `json:"linked"`
	Cached    int64      `json:"cached"`
	CursorAt  *time.Time `json:"cursorAt"`
	LastError *string    `json:"lastError"`
	UpdatedAt *time.Time `json:"updatedAt"`
}

// Catalog is what the Bhoomy sync needs from the catalog module.
type Catalog interface {
	Links() ([]Link, error)
	Cursor(storeID uuid.UUID) (*Cursor, error)
	// SavePage stores a page of changes and moves the cursor in one transaction.
	SavePage(storeID uuid.UUID, page Page, at Cursor) error
	SaveError(storeID uuid.UUID, err error) error
}

// Product is one Produk Toko product from POS's copy, with all its variants.
type Product struct {
	ID       uuid.UUID
	Name     string
	Code     string
	Brand    *string
	Category *string
	ImageURL *string
	Archived bool
	Variants []Variant
}

// Source is what other modules read from POS's copy of Produk Toko.
type Source interface {
	// Products returns the store's cached products; nil ids means all of them.
	Products(storeID uuid.UUID, ids []uuid.UUID) ([]Product, error)
	// Categories groups the store's products by their default variant's category.
	Categories(storeID uuid.UUID) ([]Category, error)
}

type Category struct {
	ID         uuid.UUID
	Name       string
	ProductIDs []uuid.UUID
}
