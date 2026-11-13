package iface

import (
	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

// SourceBhoomy marks items added from Produk Toko (Bhoomy) through Katalog POS.
const SourceBhoomy ProductSource = "BHOOMY"

// MaxDisplayName is the longest name in POS (product button, cart and receipt).
const MaxDisplayName = 40

// SourceOption is a Produk Toko product that can still be added to Katalog POS.
type SourceOption struct {
	ProductID    uuid.UUID `json:"productId"`
	Name         string    `json:"name"`
	Code         string    `json:"code"`
	VariantCount int       `json:"variantCount"`
}

// CatalogueItem is one Katalog POS row.
type CatalogueItem struct {
	ID           uuid.UUID        `json:"id"`
	DisplayName  string           `json:"displayName"`
	PhotoURL     *string          `json:"photoUrl"`
	ProductName  string           `json:"productName"`
	Code         string           `json:"code"`
	CategoryName *string          `json:"categoryName"`
	ImageURL     *string          `json:"imageUrl"`
	Variants     int              `json:"variants"`
	VariantsSold int              `json:"variantsSold"`
	PriceMin     *decimal.Decimal `json:"priceMin"`
	PriceMax     *decimal.Decimal `json:"priceMax"`
	Status       string           `json:"status"`
	LocationIDs  []uuid.UUID      `json:"locationIds"`
}

type SourceFacts struct {
	ProductName  string  `json:"productName"`
	Code         string  `json:"code"`
	Brand        *string `json:"brand"`
	CategoryName *string `json:"categoryName"`
	Barcode      *string `json:"barcode"`
	ImageURL     *string `json:"imageUrl"`
	Archived     bool    `json:"archived"`
}

type CatalogueVariant struct {
	ID        uuid.UUID        `json:"id"`
	Name      string           `json:"name"`
	SKU       string           `json:"sku"`
	SellInPOS bool             `json:"sellInPos"`
	Price     *decimal.Decimal `json:"price"`
}

// ItemLocation is a location as seen from one item.
type ItemLocation struct {
	ID       uuid.UUID `json:"id"`
	Name     string    `json:"name"`
	Type     string    `json:"type"`
	Status   string    `json:"status"`
	Sold     bool      `json:"sold"`
	OwnPrice bool      `json:"ownPrice"`
}

type CatalogueItemDetail struct {
	ID          uuid.UUID          `json:"id"`
	DisplayName string             `json:"displayName"`
	PhotoURL    *string            `json:"photoUrl"`
	Status      string             `json:"status"`
	BasePrice   *decimal.Decimal   `json:"basePrice"`
	Source      SourceFacts        `json:"source"`
	Variants    []CatalogueVariant `json:"variants"`
	Locations   []ItemLocation     `json:"locations"`
}

type AddItemsRequest struct {
	ProductIDs []uuid.UUID `json:"productIds" binding:"required,min=1"`
}

type AddItemsResponse struct {
	IDs []uuid.UUID `json:"ids"`
}

type VariantInput struct {
	ID        uuid.UUID        `json:"id"`
	SellInPOS bool             `json:"sellInPos"`
	Price     *decimal.Decimal `json:"price"`
}

type LocationInput struct {
	ID   uuid.UUID `json:"id"`
	Sold bool      `json:"sold"`
}

// UpdateItemRequest is the whole edit form; an empty variant price follows the base price.
type UpdateItemRequest struct {
	DisplayName string           `json:"displayName" binding:"required"`
	Status      string           `json:"status" binding:"required,oneof=active archived"`
	BasePrice   *decimal.Decimal `json:"basePrice"`
	Variants    []VariantInput   `json:"variants"`
	Locations   []LocationInput  `json:"locations"`
}

// LocationRow is one item variant at a location.
type LocationRow struct {
	ItemID       uuid.UUID        `json:"itemId"`
	VariantID    uuid.UUID        `json:"variantId"`
	ItemName     string           `json:"itemName"`
	VariantName  string           `json:"variantName"`
	CatalogPrice *decimal.Decimal `json:"catalogPrice"`
	Sold         bool             `json:"sold"`
	Price        *decimal.Decimal `json:"price"`
	// BhoomyVariantID is the Produk Toko variant, so Bhoomy stock can be matched to the row (6b).
	BhoomyVariantID *uuid.UUID `json:"bhoomyVariantId"`
}

type LocationRowInput struct {
	VariantID uuid.UUID        `json:"variantId"`
	Sold      bool             `json:"sold"`
	Price     *decimal.Decimal `json:"price"`
}

type SaveLocationRowsRequest struct {
	Rows []LocationRowInput `json:"rows"`
}

type AddLocationItemsRequest struct {
	ItemIDs []uuid.UUID `json:"itemIds" binding:"required,min=1"`
}

// Category is a Produk Toko category with the store's Katalog POS items in it.
type Category struct {
	ID      uuid.UUID   `json:"id"`
	Name    string      `json:"name"`
	ItemIDs []uuid.UUID `json:"itemIds"`
}

// Catalogue is what the Bhoomy sync needs: linked items follow Produk Toko changes.
type Catalogue interface {
	Reconcile(storeID uuid.UUID, productIDs []uuid.UUID) error
}
