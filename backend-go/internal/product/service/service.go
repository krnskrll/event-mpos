package service

import (
	"errors"
	"net/http"
	"strconv"
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

func (s *Service) List(storeID uuid.UUID) ([]iface.ProductResponse, error) {
	var products []repository.Product
	if err := s.preloaded().Where("store_id = ?", storeID).Order("sku").Find(&products).Error; err != nil {
		return nil, err
	}
	return mapResponses(products), nil
}

func (s *Service) Get(storeID, id uuid.UUID) (*iface.ProductResponse, error) {
	p, err := s.findIn(storeID, id)
	if err != nil {
		return nil, err
	}
	resp := p.ToResponse()
	return &resp, nil
}

func (s *Service) Create(storeID uuid.UUID, req iface.CreateProductRequest) (*iface.ProductResponse, error) {
	if err := common.CheckRupiah("price", &req.Price); err != nil {
		return nil, err
	}
	sku := strings.TrimSpace(req.SKU)
	var count int64
	if err := s.db.Model(&repository.Product{}).Where("store_id = ? AND sku = ?", storeID, sku).Count(&count).Error; err != nil {
		return nil, err
	}
	if count > 0 {
		return nil, common.NewAppError(http.StatusConflict, "A product with SKU '"+req.SKU+"' already exists")
	}
	for _, bc := range req.Barcodes {
		if free, err := s.barcodeFree(storeID, strings.TrimSpace(bc.Barcode)); err != nil {
			return nil, err
		} else if !free {
			return nil, common.NewAppError(http.StatusConflict, "Barcode '"+bc.Barcode+"' is already in use")
		}
	}
	p := repository.Product{
		StoreID:     storeID,
		SKU:         sku,
		DisplayName: strings.TrimSpace(req.Name),
		Brand:       trimPtr(req.Brand),
		PhotoURL:    trimPtr(req.ImageURL),
		Source:      iface.SourceManual,
		Status:      repository.StatusActive,
		Barcodes:    buildBarcodes(req.Barcodes),
	}
	p.SetPrice(req.Price)
	if err := s.db.Create(&p).Error; err != nil {
		return nil, err
	}
	// Return the in-memory entity (like Kotlin's save().toResponse()) so the price serializes exactly
	// as supplied (e.g. 12000, not the DB round-trip 12000.00).
	resp := p.ToResponse()
	return &resp, nil
}

func (s *Service) Update(storeID, id uuid.UUID, req iface.UpdateProductRequest) (*iface.ProductResponse, error) {
	if err := common.CheckRupiah("price", req.Price); err != nil {
		return nil, err
	}
	p, err := s.findIn(storeID, id)
	if err != nil {
		return nil, err
	}
	if req.Name != nil {
		p.DisplayName = strings.TrimSpace(*req.Name)
	}
	if req.Brand != nil {
		p.Brand = trimPtr(req.Brand)
	}
	if req.Price != nil {
		p.SetPrice(*req.Price)
	}
	if req.ImageURL != nil {
		p.PhotoURL = trimPtr(req.ImageURL)
	}
	if req.Active != nil {
		if *req.Active && p.Default().Price == nil {
			return nil, common.NewAppError(http.StatusConflict, "Set a price before selling this product")
		}
		p.SetActive(*req.Active)
	}
	err = s.db.Transaction(func(tx *gorm.DB) error {
		if req.Barcodes != nil {
			for _, bc := range *req.Barcodes {
				owner, err := s.ownerOfBarcode(tx, storeID, strings.TrimSpace(bc.Barcode))
				if err != nil {
					return err
				}
				if owner != nil && *owner != p.ID {
					return common.NewAppError(http.StatusConflict, "Barcode '"+bc.Barcode+"' belongs to another product")
				}
			}
			if err := tx.Where("item_id = ?", p.ID).Delete(&repository.ProductBarcode{}).Error; err != nil {
				return err
			}
			p.Barcodes = buildBarcodes(*req.Barcodes)
		}
		p.UpdatedAt = nowUTC()
		return tx.Session(&gorm.Session{FullSaveAssociations: true}).Save(p).Error
	})
	if err != nil {
		return nil, err
	}
	resp := p.ToResponse()
	return &resp, nil
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

func (s *Service) EventProducts(storeID, eventID uuid.UUID) ([]iface.EventProductResponse, error) {
	rows, err := s.locationItems(eventID)
	if err != nil {
		return nil, err
	}
	var products []repository.Product
	if err := s.preloaded().Where("store_id = ?", storeID).Order("sku").Find(&products).Error; err != nil {
		return nil, err
	}
	out := make([]iface.EventProductResponse, 0, len(products))
	for i := range products {
		out = append(out, eventProduct(&products[i], rows))
	}
	return out, nil
}

// SetAvailability switches every variant of a product at a location, keeping location prices.
func (s *Service) SetAvailability(storeID, eventID, productID uuid.UUID, available bool) (*iface.EventProductResponse, error) {
	p, err := s.findIn(storeID, productID)
	if err != nil {
		return nil, err
	}
	rows, err := s.locationItems(eventID)
	if err != nil {
		return nil, err
	}
	err = s.db.Transaction(func(tx *gorm.DB) error {
		for _, v := range p.Variants {
			li := rows[v.ID]
			if err := saveLocationItem(tx, eventID, v.ID, available, li.PriceOverride, &li); err != nil {
				return err
			}
			rows[v.ID] = li
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	resp := eventProduct(p, rows)
	return &resp, nil
}

// SetLocationVariant sets one variant's availability and price at a location.
func (s *Service) SetLocationVariant(storeID, eventID, variantID uuid.UUID, req iface.SetLocationVariantRequest) (*iface.EventProductResponse, error) {
	if err := common.CheckRupiah("price", req.Price); err != nil {
		return nil, err
	}
	if req.Price != nil && req.Price.IsNegative() {
		return nil, common.NewAppError(http.StatusBadRequest, "price: must not be negative")
	}
	var v repository.ProductVariant
	err := s.db.Where("id = ? AND store_id = ?", variantID, storeID).First(&v).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, common.NewAppError(http.StatusNotFound, "Variant not found")
	}
	if err != nil {
		return nil, err
	}
	p, err := s.findIn(storeID, v.ItemID)
	if err != nil {
		return nil, err
	}
	rows, err := s.locationItems(eventID)
	if err != nil {
		return nil, err
	}
	li := rows[variantID]
	if err := saveLocationItem(s.db, eventID, variantID, *req.Available, req.Price, &li); err != nil {
		return nil, err
	}
	rows[variantID] = li
	resp := eventProduct(p, rows)
	return &resp, nil
}

func saveLocationItem(tx *gorm.DB, locationID, variantID uuid.UUID, active bool, price *decimal.Decimal, li *repository.LocationItem) error {
	li.LocationID, li.ItemVariantID, li.Active, li.PriceOverride = locationID, variantID, active, price
	return tx.Save(li).Error
}

func eventProduct(p *repository.Product, rows repository.LocationItems) iface.EventProductResponse {
	variants := make([]iface.LocationVariantDto, 0, len(p.Variants))
	available := false
	for _, v := range p.Variants {
		li, ok := rows[v.ID]
		active := ok && li.Active
		available = available || active
		variants = append(variants, iface.LocationVariantDto{ID: v.ID, Name: v.Name, IsDefault: v.IsDefault, Available: active, LocationPrice: li.PriceOverride})
	}
	return iface.EventProductResponse{Product: p.ToResponse(), Available: available, Variants: variants}
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
// snapshot version. Format: "sku:price:barcodeCount" per product, comma-joined (matches Kotlin).
func (s *Service) AvailabilityFingerprint(eventID uuid.UUID) (string, error) {
	dtos, err := s.availableSnapshotProducts(eventID)
	if err != nil {
		return "", err
	}
	parts := make([]string, 0, len(dtos))
	for _, d := range dtos {
		part := d.SKU + ":" + d.Price.String() + ":" + strconv.Itoa(len(d.Barcodes))
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

// sourceBarcodes keeps ERP values as delivered (no trimming, unlike admin input).
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

func (s *Service) barcodeFree(storeID uuid.UUID, barcode string) (bool, error) {
	var count int64
	if err := s.db.Model(&repository.ProductBarcode{}).Where("store_id = ? AND barcode = ?", storeID, barcode).Count(&count).Error; err != nil {
		return false, err
	}
	return count == 0, nil
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

func mapResponses(products []repository.Product) []iface.ProductResponse {
	out := make([]iface.ProductResponse, 0, len(products))
	for i := range products {
		out = append(out, products[i].ToResponse())
	}
	return out
}

func buildBarcodes(inputs []iface.BarcodeInput) []repository.ProductBarcode {
	out := make([]repository.ProductBarcode, 0, len(inputs))
	for _, in := range inputs {
		out = append(out, repository.ProductBarcode{Barcode: strings.TrimSpace(in.Barcode), Label: trimPtr(in.Label)})
	}
	return out
}

func trimPtr(s *string) *string {
	if s == nil {
		return nil
	}
	t := strings.TrimSpace(*s)
	return &t
}

func nowUTC() time.Time { return time.Now().UTC() }
