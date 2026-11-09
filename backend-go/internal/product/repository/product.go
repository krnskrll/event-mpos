package repository

import (
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	"gorm.io/gorm"

	"github.com/krnskrll/event-mpos/internal/product/iface"
)

// Product is a Katalog POS item (pos_items). Its base price lives on the default variant.
type Product struct {
	ID             uuid.UUID           `gorm:"type:uuid;primaryKey"`
	StoreID        uuid.UUID           `gorm:"column:store_id;type:uuid;not null"`
	SKU            string              `gorm:"column:sku;not null"`
	DisplayName    string              `gorm:"column:display_name;not null"`
	Brand          *string             ``
	PhotoURL       *string             `gorm:"column:photo_url"`
	WarehouseStock int                 `gorm:"column:warehouse_stock;not null"`
	Source         iface.ProductSource `gorm:"type:varchar(20);not null"`
	Status         string              `gorm:"type:varchar(16);not null"`
	// BhoomyProductID links an item added from Produk Toko.
	BhoomyProductID *uuid.UUID       `gorm:"column:bhoomy_product_id;type:uuid"`
	Barcodes        []ProductBarcode `gorm:"foreignKey:ItemID;constraint:OnDelete:CASCADE"`
	Variants        []ProductVariant `gorm:"foreignKey:ItemID;constraint:OnDelete:CASCADE"`
	CreatedAt       time.Time        `gorm:"column:created_at;not null"`
	UpdatedAt       time.Time        `gorm:"column:updated_at;not null"`
}

const (
	StatusActive   = "active"
	StatusArchived = "archived"
)

func (Product) TableName() string { return "pos_items" }

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

func (p *Product) Active() bool { return p.Status == StatusActive }

// Default is the variant that carries the base price; every item has exactly one.
func (p *Product) Default() *ProductVariant {
	for i := range p.Variants {
		if p.Variants[i].IsDefault {
			return &p.Variants[i]
		}
	}
	return nil
}

func (p *Product) Price() decimal.Decimal {
	if d := p.Default(); d != nil && d.Price != nil {
		return *d.Price
	}
	return decimal.Zero
}

func (p *Product) SetPrice(price decimal.Decimal) {
	if d := p.Default(); d != nil {
		d.Price = &price
		return
	}
	p.Variants = append(p.Variants, ProductVariant{ItemID: p.ID, Name: "", SKU: p.SKU, Price: &price, SellInPOS: true, IsDefault: true})
}

// Options are the real variants (shade, size), without the default one.
func (p *Product) Options() []ProductVariant {
	out := make([]ProductVariant, 0, len(p.Variants))
	for _, v := range p.Variants {
		if !v.IsDefault {
			out = append(out, v)
		}
	}
	return out
}

// ProductBarcode — one scannable barcode for an item, unique within its store.
type ProductBarcode struct {
	ID      uuid.UUID `gorm:"type:uuid;primaryKey"`
	ItemID  uuid.UUID `gorm:"column:item_id;type:uuid;not null"`
	StoreID uuid.UUID `gorm:"column:store_id;type:uuid;not null"`
	Barcode string    `gorm:"not null"`
	Label   *string   ``
	Source  string    `gorm:"not null;default:pos"`
}

// Barcode sources: Produk Toko barcodes follow Bhoomy; POS barcodes are extras edited in Katalog POS.
const (
	BarcodeProdukToko = "produk_toko"
	BarcodePOS        = "pos"
)

func (ProductBarcode) TableName() string { return "pos_item_barcodes" }

func (b *ProductBarcode) BeforeCreate(tx *gorm.DB) error {
	if b.ID == uuid.Nil {
		b.ID = uuid.New()
	}
	return storeOfItem(tx, b.ItemID, &b.StoreID)
}

// ProductVariant — a sellable variation (shade/size). Shares the item's stock; a nil price follows
// the default variant.
type ProductVariant struct {
	ID        uuid.UUID        `gorm:"type:uuid;primaryKey"`
	ItemID    uuid.UUID        `gorm:"column:item_id;type:uuid;not null"`
	StoreID   uuid.UUID        `gorm:"column:store_id;type:uuid;not null"`
	Name      string           `gorm:"not null"`
	SKU       string           `gorm:"column:sku;not null"`
	Price     *decimal.Decimal `gorm:"column:price;type:bigint"`
	SellInPOS bool             `gorm:"column:sell_in_pos;not null"`
	IsDefault bool             `gorm:"column:is_default;not null"`
	// BhoomyVariantID links the variant to its Produk Toko variant.
	BhoomyVariantID *uuid.UUID `gorm:"column:bhoomy_variant_id;type:uuid"`
}

func (ProductVariant) TableName() string { return "pos_item_variants" }

func (v *ProductVariant) BeforeCreate(tx *gorm.DB) error {
	if v.ID == uuid.Nil {
		v.ID = uuid.New()
	}
	return storeOfItem(tx, v.ItemID, &v.StoreID)
}

// storeOfItem copies the parent item's store onto a child row; it never comes from a request.
func storeOfItem(tx *gorm.DB, itemID uuid.UUID, dst *uuid.UUID) error {
	return tx.Session(&gorm.Session{NewDB: true}).Model(&Product{}).Select("store_id").Where("id = ?", itemID).Row().Scan(dst)
}

// EffectivePrice is the price actually charged for this variant.
func (v *ProductVariant) EffectivePrice(basePrice decimal.Decimal) decimal.Decimal {
	if v.Price != nil {
		return *v.Price
	}
	return basePrice
}

// LocationItem is what a location sells, per variant, with an optional location price.
type LocationItem struct {
	ID            uuid.UUID        `gorm:"type:uuid;primaryKey"`
	LocationID    uuid.UUID        `gorm:"column:location_id;type:uuid;not null"`
	ItemVariantID uuid.UUID        `gorm:"column:item_variant_id;type:uuid;not null"`
	PriceOverride *decimal.Decimal `gorm:"column:price_override;type:bigint"`
	Active        bool             `gorm:"not null"`
}

func (LocationItem) TableName() string { return "pos_location_items" }

func (li *LocationItem) BeforeCreate(_ *gorm.DB) error {
	if li.ID == uuid.Nil {
		li.ID = uuid.New()
	}
	return nil
}

// LocationItems are one location's rows by variant id.
type LocationItems map[uuid.UUID]LocationItem

// Price sets the location's prices on p.
func (r LocationItems) Price(p *Product) {
	for i := range p.Variants {
		if li, ok := r[p.Variants[i].ID]; ok && li.PriceOverride != nil {
			p.Variants[i].Price = li.PriceOverride
		}
	}
}

// Apply prices p and keeps the default variant plus the real variants the location sells; sold
// reports whether it sells any variant of p.
func (r LocationItems) Apply(p *Product) (sold bool) {
	r.Price(p)
	kept := make([]ProductVariant, 0, len(p.Variants))
	for _, v := range p.Variants {
		li, ok := r[v.ID]
		active := ok && li.Active
		sold = sold || active
		if v.IsDefault || active {
			kept = append(kept, v)
		}
	}
	p.Variants = kept
	return sold
}

func (p *Product) ToResponse() iface.ProductResponse {
	barcodes := make([]iface.BarcodeDto, 0, len(p.Barcodes))
	for i := range p.Barcodes {
		barcodes = append(barcodes, iface.BarcodeDto{Barcode: p.Barcodes[i].Barcode, Label: p.Barcodes[i].Label})
	}
	options := p.Options()
	variants := make([]iface.VariantDto, 0, len(options))
	for _, v := range options {
		variants = append(variants, iface.VariantDto{ID: v.ID, Name: v.Name, SKU: v.SKU, Price: v.EffectivePrice(p.Price()), Active: v.SellInPOS})
	}
	return iface.ProductResponse{
		ID:               p.ID,
		DefaultVariantID: defaultID(p),
		SKU:              p.SKU,
		Name:             p.DisplayName,
		Brand:            p.Brand,
		Price:            p.Price(),
		ImageURL:         p.PhotoURL,
		Source:           p.Source,
		Active:           p.Active(),
		WarehouseStock:   p.WarehouseStock,
		Barcodes:         barcodes,
		Variants:         variants,
		UpdatedAt:        p.UpdatedAt,
	}
}

func (p *Product) ToSnapshotDto() iface.ProductSnapshotDto {
	barcodes := make([]string, 0, len(p.Barcodes))
	for i := range p.Barcodes {
		barcodes = append(barcodes, p.Barcodes[i].Barcode)
	}
	variants := make([]iface.VariantSnapshotDto, 0, len(p.Variants))
	for _, v := range p.Options() {
		if v.SellInPOS {
			variants = append(variants, iface.VariantSnapshotDto{ID: v.ID, Name: v.Name, Price: v.EffectivePrice(p.Price())})
		}
	}
	return iface.ProductSnapshotDto{
		ID:       p.ID,
		SKU:      p.SKU,
		Name:     p.DisplayName,
		Brand:    p.Brand,
		Price:    p.Price(),
		ImageURL: p.PhotoURL,
		Barcodes: barcodes,
		Variants: variants,
	}
}

func (p *Product) ToSale() iface.SaleProduct {
	options := p.Options()
	variants := make([]iface.SaleVariant, 0, len(options))
	for _, v := range options {
		variants = append(variants, iface.SaleVariant{ID: v.ID, Name: v.Name, Price: v.EffectivePrice(p.Price()), Active: v.SellInPOS})
	}
	return iface.SaleProduct{ID: p.ID, SKU: p.SKU, Name: p.DisplayName, Price: p.Price(), WarehouseStock: p.WarehouseStock, Variants: variants}
}

func defaultID(p *Product) uuid.UUID {
	if d := p.Default(); d != nil {
		return d.ID
	}
	return uuid.Nil
}
