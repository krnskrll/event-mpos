package product

import (
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	"gorm.io/gorm"
)

// ProductSource — where a product came from.
type ProductSource string

const (
	SourceERP    ProductSource = "ERP"
	SourceManual ProductSource = "MANUAL"
)

// Product is a catalogue product. Per-event availability lives in EventProduct.
type Product struct {
	ID             uuid.UUID       `gorm:"type:uuid;primaryKey"`
	SKU            string          `gorm:"column:sku;uniqueIndex;not null"`
	Name           string          `gorm:"not null"`
	Brand          *string         ``
	Price          decimal.Decimal `gorm:"type:numeric(12,2);not null"`
	ImageURL       *string         `gorm:"column:image_url"`
	WarehouseStock int             `gorm:"column:warehouse_stock;not null"`
	Source         ProductSource   `gorm:"type:varchar(20);not null"`
	Active         bool            `gorm:"not null;default:true"`
	Barcodes       []ProductBarcode `gorm:"foreignKey:ProductID;constraint:OnDelete:CASCADE"`
	Variants       []ProductVariant `gorm:"foreignKey:ProductID;constraint:OnDelete:CASCADE"`
	CreatedAt      time.Time       `gorm:"column:created_at;not null"`
	UpdatedAt      time.Time       `gorm:"column:updated_at;not null"`
}

func (Product) TableName() string { return "products" }

func (p *Product) BeforeCreate(_ *gorm.DB) error {
	if p.ID == uuid.Nil {
		p.ID = uuid.New()
	}
	now := time.Now().UTC()
	if p.CreatedAt.IsZero() {
		p.CreatedAt = now
	}
	if p.UpdatedAt.IsZero() {
		p.UpdatedAt = now
	}
	return nil
}

// ProductBarcode — one scannable, globally-unique barcode for a product.
type ProductBarcode struct {
	ID        uuid.UUID `gorm:"type:uuid;primaryKey"`
	ProductID uuid.UUID `gorm:"column:product_id;type:uuid;not null"`
	Barcode   string    `gorm:"uniqueIndex;not null"`
	Label     *string   ``
}

func (ProductBarcode) TableName() string { return "product_barcodes" }

func (b *ProductBarcode) BeforeCreate(_ *gorm.DB) error {
	if b.ID == uuid.Nil {
		b.ID = uuid.New()
	}
	return nil
}

// ProductVariant — a sellable variation (shade/size). Shares the product's stock; may override price.
type ProductVariant struct {
	ID            uuid.UUID        `gorm:"type:uuid;primaryKey"`
	ProductID     uuid.UUID        `gorm:"column:product_id;type:uuid;not null"`
	Name          string           `gorm:"not null"`
	SKU           string           `gorm:"column:sku;uniqueIndex;not null"`
	PriceOverride *decimal.Decimal `gorm:"column:price_override;type:numeric(12,2)"`
	Active        bool             `gorm:"not null;default:true"`
}

func (ProductVariant) TableName() string { return "product_variants" }

func (v *ProductVariant) BeforeCreate(_ *gorm.DB) error {
	if v.ID == uuid.Nil {
		v.ID = uuid.New()
	}
	return nil
}

// EffectivePrice is the price actually charged for this variant.
func (v *ProductVariant) EffectivePrice(productPrice decimal.Decimal) decimal.Decimal {
	if v.PriceOverride != nil {
		return *v.PriceOverride
	}
	return productPrice
}

// EventProduct marks a catalogue product available for a specific event.
type EventProduct struct {
	ID        uuid.UUID `gorm:"type:uuid;primaryKey"`
	EventID   uuid.UUID `gorm:"column:event_id;type:uuid;not null;uniqueIndex:uk_event_product;index:idx_event_product_event"`
	ProductID uuid.UUID `gorm:"column:product_id;type:uuid;not null;uniqueIndex:uk_event_product"`
	// No default:true — GORM omits a zero-value bool on INSERT when a default is set, which would
	// silently persist `true` when creating a row with available=false. We always set it explicitly.
	Available bool `gorm:"not null"`
}

func (EventProduct) TableName() string { return "event_products" }

func (ep *EventProduct) BeforeCreate(_ *gorm.DB) error {
	if ep.ID == uuid.Nil {
		ep.ID = uuid.New()
	}
	return nil
}

// ---- DTOs ----

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
	ID             uuid.UUID       `json:"id"`
	SKU            string          `json:"sku"`
	Name           string          `json:"name"`
	Brand          *string         `json:"brand,omitempty"`
	Price          decimal.Decimal `json:"price"`
	ImageURL       *string         `json:"imageUrl,omitempty"`
	Source         ProductSource   `json:"source"`
	Active         bool            `json:"active"`
	WarehouseStock int             `json:"warehouseStock"`
	Barcodes       []BarcodeDto    `json:"barcodes"`
	Variants       []VariantDto    `json:"variants"`
	UpdatedAt      time.Time       `json:"updatedAt"`
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
	Product   ProductResponse `json:"product"`
	Available bool            `json:"available"`
}

type SetAvailabilityRequest struct {
	Available *bool `json:"available" binding:"required"`
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

func (p *Product) ToResponse() ProductResponse {
	barcodes := make([]BarcodeDto, 0, len(p.Barcodes))
	for i := range p.Barcodes {
		barcodes = append(barcodes, BarcodeDto{Barcode: p.Barcodes[i].Barcode, Label: p.Barcodes[i].Label})
	}
	variants := make([]VariantDto, 0, len(p.Variants))
	for i := range p.Variants {
		v := &p.Variants[i]
		variants = append(variants, VariantDto{ID: v.ID, Name: v.Name, SKU: v.SKU, Price: v.EffectivePrice(p.Price), Active: v.Active})
	}
	return ProductResponse{
		ID:             p.ID,
		SKU:            p.SKU,
		Name:           p.Name,
		Brand:          p.Brand,
		Price:          p.Price,
		ImageURL:       p.ImageURL,
		Source:         p.Source,
		Active:         p.Active,
		WarehouseStock: p.WarehouseStock,
		Barcodes:       barcodes,
		Variants:       variants,
		UpdatedAt:      p.UpdatedAt,
	}
}

func (p *Product) ToSnapshotDto() ProductSnapshotDto {
	barcodes := make([]string, 0, len(p.Barcodes))
	for i := range p.Barcodes {
		barcodes = append(barcodes, p.Barcodes[i].Barcode)
	}
	variants := make([]VariantSnapshotDto, 0, len(p.Variants))
	for i := range p.Variants {
		v := &p.Variants[i]
		if v.Active {
			variants = append(variants, VariantSnapshotDto{ID: v.ID, Name: v.Name, Price: v.EffectivePrice(p.Price)})
		}
	}
	return ProductSnapshotDto{
		ID:       p.ID,
		SKU:      p.SKU,
		Name:     p.Name,
		Brand:    p.Brand,
		Price:    p.Price,
		ImageURL: p.ImageURL,
		Barcodes: barcodes,
		Variants: variants,
	}
}
