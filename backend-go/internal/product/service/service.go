package service

import (
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	"gorm.io/gorm"

	catalogiface "github.com/krnskrll/event-mpos/internal/catalog/iface"
	"github.com/krnskrll/event-mpos/internal/platform/common"
	"github.com/krnskrll/event-mpos/internal/platform/database"
	"github.com/krnskrll/event-mpos/internal/platform/media"
	"github.com/krnskrll/event-mpos/internal/product/iface"
	"github.com/krnskrll/event-mpos/internal/product/repository"
)

type Service struct {
	db     *gorm.DB
	media  *media.Store
	source catalogiface.Source
	// stockRemaining enriches snapshot products; nil leaves remainingStock at 0.
	stockRemaining func(eventID, productID uuid.UUID) (int, error)
}

var _ iface.Products = (*Service)(nil)

func NewService(db *gorm.DB, m *media.Store) *Service { return &Service{db: db, media: m} }

func (s *Service) SetStockRemaining(fn func(eventID, productID uuid.UUID) (int, error)) {
	s.stockRemaining = fn
}

// SetPhoto stores a POS photo for the product and points its image URL at it.
func (s *Service) SetPhoto(storeID, id uuid.UUID, data []byte) (*iface.ProductResponse, error) {
	p, err := s.findIn(storeID, id)
	if err != nil {
		return nil, err
	}
	url, err := s.media.Save(data)
	if err != nil {
		return nil, err
	}
	p.PhotoURL, p.UpdatedAt = &url, nowUTC()
	if err := s.db.Model(p).Select("photo_url", "updated_at").Updates(p).Error; err != nil {
		return nil, err
	}
	resp := p.ToResponse()
	return &resp, nil
}

func saveLocationItem(tx *gorm.DB, locationID, variantID uuid.UUID, active bool, price *decimal.Decimal, li *repository.LocationItem) error {
	li.LocationID, li.ItemVariantID, li.Active, li.PriceOverride = locationID, variantID, active, price
	return tx.Save(li).Error
}

func (s *Service) locationItems(locationID uuid.UUID) (repository.LocationItems, error) {
	var list []repository.LocationItem
	if err := s.db.Where("location_id = ?", locationID).Find(&list).Error; err != nil {
		return nil, err
	}
	rows := make(repository.LocationItems, len(list))
	for _, li := range list {
		rows[li.ItemVariantID] = li
	}
	return rows, nil
}

// availableSnapshotEntities returns active products the event sells, at its prices, ordered by sku.
func (s *Service) availableSnapshotEntities(eventID uuid.UUID) ([]repository.Product, error) {
	rows, err := s.locationItems(eventID)
	if err != nil {
		return nil, err
	}
	sold := s.db.Model(&repository.ProductVariant{}).Select("item_id").
		Where("id IN (?)", s.db.Model(&repository.LocationItem{}).Select("item_variant_id").Where("location_id = ? AND active", eventID))
	var products []repository.Product
	if err := s.preloaded().Where("id IN (?) AND status = ?", sold, repository.StatusActive).Order("sku").Find(&products).Error; err != nil {
		return nil, err
	}
	for i := range products {
		rows.Apply(&products[i])
	}
	return products, nil
}

func (s *Service) availableSnapshotProducts(eventID uuid.UUID) ([]iface.ProductSnapshotDto, error) {
	products, err := s.availableSnapshotEntities(eventID)
	if err != nil {
		return nil, err
	}
	out := make([]iface.ProductSnapshotDto, 0, len(products))
	for i := range products {
		dto := products[i].ToSnapshotDto()
		if s.stockRemaining != nil {
			rem, err := s.stockRemaining(eventID, products[i].ID)
			if err != nil {
				return nil, err
			}
			dto.RemainingStock = rem
		}
		out = append(out, dto)
	}
	return out, nil
}

func (s *Service) AvailableSnapshotProducts(eventID uuid.UUID) ([]any, error) {
	dtos, err := s.availableSnapshotProducts(eventID)
	if err != nil {
		return nil, err
	}
	out := make([]any, len(dtos))
	for i := range dtos {
		out[i] = dtos[i]
	}
	return out, nil
}

// AvailabilityFingerprint is a stable hash of the event's available products, folded into the
// snapshot version. Format: "sku:price:barcodes" per product (barcodes "+"-joined), comma-joined.
func (s *Service) AvailabilityFingerprint(eventID uuid.UUID) (string, error) {
	dtos, err := s.availableSnapshotProducts(eventID)
	if err != nil {
		return "", err
	}
	parts := make([]string, 0, len(dtos))
	for _, d := range dtos {
		part := d.SKU + ":" + d.Price.String() + ":" + strings.Join(d.Barcodes, "+")
		for _, v := range d.Variants {
			part += "|" + v.Price.String()
		}
		parts = append(parts, part)
	}
	return strings.Join(parts, ","), nil
}

func (s *Service) FindAvailableByBarcode(storeID, eventID uuid.UUID, barcode string) (*iface.ProductResponse, error) {
	productID, err := s.ownerOfBarcode(s.db, storeID, strings.TrimSpace(barcode))
	if err != nil {
		return nil, err
	}
	if productID == nil {
		return nil, common.NewAppError(http.StatusNotFound, "No product found for barcode '"+barcode+"'")
	}
	p, err := s.find(*productID)
	if err != nil {
		return nil, err
	}
	rows, err := s.locationItems(eventID)
	if err != nil {
		return nil, err
	}
	if !rows.Apply(p) || !p.Active() {
		return nil, common.NewAppError(http.StatusNotFound, "That product isn't available for this event")
	}
	resp := p.ToResponse()
	return &resp, nil
}

func (s *Service) Find(id uuid.UUID) (*iface.SaleProduct, error) {
	var p repository.Product
	err := s.db.Preload("Variants").First(&p, "id = ?", id).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	sale := p.ToSale()
	return &sale, nil
}

func (s *Service) FindAt(locationID, id uuid.UUID) (*iface.SaleProduct, error) {
	var p repository.Product
	err := s.db.Preload("Variants").First(&p, "id = ?", id).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	rows, err := s.locationItems(locationID)
	if err != nil {
		return nil, err
	}
	rows.Price(&p)
	sale := p.ToSale()
	return &sale, nil
}

func (s *Service) All(storeID uuid.UUID) ([]iface.SaleProduct, error) {
	var products []repository.Product
	if err := s.db.Preload("Variants").Where("store_id = ?", storeID).Order("sku").Find(&products).Error; err != nil {
		return nil, err
	}
	out := make([]iface.SaleProduct, 0, len(products))
	for i := range products {
		out = append(out, products[i].ToSale())
	}
	return out, nil
}

// Upsert rewrites barcodes only when they changed, to avoid unique-constraint churn.
func (s *Service) Upsert(tx database.Tx, storeID uuid.UUID, in iface.SourceProduct) (bool, error) {
	var p repository.Product
	err := tx.Preload("Barcodes").Preload("Variants").Where("store_id = ? AND sku = ?", storeID, in.SKU).First(&p).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		p = repository.Product{
			StoreID: storeID, SKU: in.SKU, DisplayName: in.Name, Brand: in.Brand, PhotoURL: in.ImageURL,
			WarehouseStock: in.StockQuantity, Source: iface.SourceERP, Status: repository.StatusActive, Barcodes: sourceBarcodes(in.Barcodes),
		}
		p.SetPrice(common.Rupiah(in.Price))
		return true, tx.Create(&p).Error
	}
	if err != nil {
		return false, err
	}
	p.DisplayName, p.Brand, p.PhotoURL = in.Name, in.Brand, in.ImageURL
	p.WarehouseStock, p.Source, p.UpdatedAt = in.StockQuantity, iface.SourceERP, nowUTC()
	p.SetPrice(common.Rupiah(in.Price))
	if barcodesChanged(p.Barcodes, in.Barcodes) {
		if err := tx.Where("item_id = ?", p.ID).Delete(&repository.ProductBarcode{}).Error; err != nil {
			return false, err
		}
		for _, b := range sourceBarcodes(in.Barcodes) {
			b.ItemID = p.ID
			if err := tx.Create(&b).Error; err != nil {
				return false, err
			}
		}
	}
	if err := tx.Save(p.Default()).Error; err != nil {
		return false, err
	}
	return false, tx.Model(&p).Omit("Barcodes", "Variants").Save(&p).Error
}

// sourceBarcodes keeps ERP values as delivered (no trimming).
func sourceBarcodes(in []iface.BarcodeInput) []repository.ProductBarcode {
	out := make([]repository.ProductBarcode, 0, len(in))
	for _, b := range in {
		out = append(out, repository.ProductBarcode{Barcode: b.Barcode, Label: b.Label})
	}
	return out
}

func barcodesChanged(current []repository.ProductBarcode, incoming []iface.BarcodeInput) bool {
	cur := make(map[string]bool, len(current))
	for _, c := range current {
		cur[c.Barcode] = true
	}
	inc := make(map[string]bool, len(incoming))
	for _, i := range incoming {
		inc[i.Barcode] = true
	}
	if len(cur) != len(inc) {
		return true
	}
	for k := range inc {
		if !cur[k] {
			return true
		}
	}
	return false
}

func (s *Service) preloaded() *gorm.DB {
	return s.db.
		Preload("Barcodes", func(db *gorm.DB) *gorm.DB { return db.Order("barcode") }).
		Preload("Variants", func(db *gorm.DB) *gorm.DB { return db.Order("sku") })
}

func (s *Service) find(id uuid.UUID) (*repository.Product, error) {
	var p repository.Product
	err := s.preloaded().First(&p, "id = ?", id).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, common.NewAppError(http.StatusNotFound, "Product not found")
	}
	if err != nil {
		return nil, err
	}
	return &p, nil
}

func (s *Service) findIn(storeID, id uuid.UUID) (*repository.Product, error) {
	p, err := s.find(id)
	if err != nil {
		return nil, err
	}
	if p.StoreID != storeID {
		return nil, common.NewAppError(http.StatusNotFound, "Product not found")
	}
	return p, nil
}

// ownerOfBarcode returns the store's product id owning a barcode, or nil if the barcode is unused.
func (s *Service) ownerOfBarcode(tx *gorm.DB, storeID uuid.UUID, barcode string) (*uuid.UUID, error) {
	var bc repository.ProductBarcode
	err := tx.Where("store_id = ? AND barcode = ?", storeID, barcode).First(&bc).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &bc.ItemID, nil
}

func nowUTC() time.Time { return time.Now().UTC() }
