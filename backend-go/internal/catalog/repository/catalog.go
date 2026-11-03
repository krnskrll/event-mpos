package repository

import (
	"time"

	"github.com/google/uuid"
)

const SourceProdukToko = "produk_toko"

type CachedVariant struct {
	BhoomyVariantID uuid.UUID  `gorm:"column:bhoomy_variant_id;type:uuid;primaryKey"`
	StoreID         uuid.UUID  `gorm:"column:store_id;type:uuid;not null"`
	BhoomyProductID uuid.UUID  `gorm:"column:bhoomy_product_id;type:uuid;not null"`
	ProductName     string     `gorm:"column:product_name;not null"`
	VariantName     string     `gorm:"column:variant_name;not null"`
	SKU             string     `gorm:"column:sku;not null"`
	Barcode         *string    `gorm:"column:barcode"`
	IsDefault       bool       `gorm:"column:is_default;not null"`
	BrandName       *string    `gorm:"column:brand_name"`
	CategoryID      *uuid.UUID `gorm:"column:category_id;type:uuid"`
	CategoryName    *string    `gorm:"column:category_name"`
	ImageURL        *string    `gorm:"column:image_url"`
	Status          string     `gorm:"column:status;not null"`
	ChangedAt       time.Time  `gorm:"column:changed_at;not null"`
	SyncedAt        time.Time  `gorm:"column:synced_at;not null"`
}

func (CachedVariant) TableName() string { return "pos_catalog_cache" }

type SyncCursor struct {
	StoreID   uuid.UUID  `gorm:"column:store_id;type:uuid;primaryKey"`
	Source    string     `gorm:"column:source;primaryKey"`
	CursorAt  *time.Time `gorm:"column:cursor_at"`
	CursorID  *uuid.UUID `gorm:"column:cursor_id;type:uuid"`
	LastError *string    `gorm:"column:last_error"`
	UpdatedAt time.Time  `gorm:"column:updated_at;not null"`
}

func (SyncCursor) TableName() string { return "pos_sync_cursors" }
