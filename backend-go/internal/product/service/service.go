package service

import (
	"errors"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/krnskrll/event-mpos/internal/platform/common"
	"github.com/krnskrll/event-mpos/internal/platform/database"
	"github.com/krnskrll/event-mpos/internal/product/iface"
	"github.com/krnskrll/event-mpos/internal/product/repository"
)

type Service struct {
	db *gorm.DB
	// stockRemaining enriches snapshot products; nil leaves remainingStock at 0.
	stockRemaining func(eventID, productID uuid.UUID) (int, error)
}

var _ iface.Products = (*Service)(nil)

func NewService(db *gorm.DB) *Service { return &Service{db: db} }

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
	sku := strings.TrimSpace(req.SKU)
	var count int64
	if err := s.db.Model(&repository.Product{}).Where("sku = ?", sku).Count(&count).Error; err != nil {
		return nil, err
	}
	if count > 0 {
		return nil, common.NewAppError(http.StatusConflict, "A product with SKU '"+req.SKU+"' already exists")
	}
	for _, bc := range req.Barcodes {
		if free, err := s.barcodeFree(strings.TrimSpace(bc.Barcode)); err != nil {
			return nil, err
		} else if !free {
			return nil, common.NewAppError(http.StatusConflict, "Barcode '"+bc.Barcode+"' is already in use")
		}
	}
	p := repository.Product{
		StoreID:  storeID,
		SKU:      sku,
		Name:     strings.TrimSpace(req.Name),
		Brand:    trimPtr(req.Brand),
		Price:    req.Price,
		ImageURL: trimPtr(req.ImageURL),
		Source:   iface.SourceManual,
		Active:   true,
		Barcodes: buildBarcodes(req.Barcodes),
	}
	if err := s.db.Create(&p).Error; err != nil {
		return nil, err
	}
	// Return the in-memory entity (like Kotlin's save().toResponse()) so the price serializes exactly
	// as supplied (e.g. 12000, not the DB round-trip 12000.00).
	resp := p.ToResponse()
	return &resp, nil
}

func (s *Service) Update(storeID, id uuid.UUID, req iface.UpdateProductRequest) (*iface.ProductResponse, error) {
	p, err := s.findIn(storeID, id)
	if err != nil {
		return nil, err
	}
	if req.Name != nil {
		p.Name = strings.TrimSpace(*req.Name)
	}
	if req.Brand != nil {
		p.Brand = trimPtr(req.Brand)
	}
	if req.Price != nil {
		p.Price = *req.Price
	}
	if req.ImageURL != nil {
		p.ImageURL = trimPtr(req.ImageURL)
	}
	if req.Active != nil {
		p.Active = *req.Active
	}
	err = s.db.Transaction(func(tx *gorm.DB) error {
		if req.Barcodes != nil {
			for _, bc := range *req.Barcodes {
				owner, err := s.ownerOfBarcode(tx, strings.TrimSpace(bc.Barcode))
				if err != nil {
					return err
				}
				if owner != nil && *owner != p.ID {
					return common.NewAppError(http.StatusConflict, "Barcode '"+bc.Barcode+"' belongs to another product")
				}
			}
			if err := tx.Where("product_id = ?", p.ID).Delete(&repository.ProductBarcode{}).Error; err != nil {
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

func (s *Service) EventProducts(storeID, eventID uuid.UUID) ([]iface.EventProductResponse, error) {
	var eps []repository.EventProduct
	if err := s.db.Where("event_id = ? AND available = ?", eventID, true).Find(&eps).Error; err != nil {
		return nil, err
	}
	availableIDs := make(map[uuid.UUID]bool, len(eps))
	for _, ep := range eps {
		availableIDs[ep.ProductID] = true
	}
	var products []repository.Product
	if err := s.preloaded().Where("store_id = ?", storeID).Order("sku").Find(&products).Error; err != nil {
		return nil, err
	}
	out := make([]iface.EventProductResponse, 0, len(products))
	for i := range products {
		out = append(out, iface.EventProductResponse{Product: products[i].ToResponse(), Available: availableIDs[products[i].ID]})
	}
	return out, nil
}

func (s *Service) SetAvailability(storeID, eventID, productID uuid.UUID, available bool) (*iface.EventProductResponse, error) {
	p, err := s.findIn(storeID, productID)
	if err != nil {
		return nil, err
	}
	var link repository.EventProduct
	err = s.db.Where("event_id = ? AND product_id = ?", eventID, productID).First(&link).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		link = repository.EventProduct{EventID: eventID, ProductID: productID}
	} else if err != nil {
		return nil, err
	}
	link.Available = available
	if err := s.db.Save(&link).Error; err != nil {
		return nil, err
	}
	resp := iface.EventProductResponse{Product: p.ToResponse(), Available: available}
	return &resp, nil
}

// availableSnapshotEntities returns active products available for the event, ordered by sku.
func (s *Service) availableSnapshotEntities(eventID uuid.UUID) ([]repository.Product, error) {
	var eps []repository.EventProduct
	if err := s.db.Where("event_id = ? AND available = ?", eventID, true).Find(&eps).Error; err != nil {
		return nil, err
	}
	if len(eps) == 0 {
		return nil, nil
	}
	ids := make([]uuid.UUID, 0, len(eps))
	for _, ep := range eps {
		ids = append(ids, ep.ProductID)
	}
	var products []repository.Product
	if err := s.preloaded().Where("id IN ? AND active = ?", ids, true).Order("sku").Find(&products).Error; err != nil {
		return nil, err
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
		parts = append(parts, d.SKU+":"+d.Price.String()+":"+strconv.Itoa(len(d.Barcodes)))
	}
	return strings.Join(parts, ","), nil
}

func (s *Service) FindAvailableByBarcode(eventID uuid.UUID, barcode string) (*iface.ProductResponse, error) {
	productID, err := s.ownerOfBarcode(s.db, strings.TrimSpace(barcode))
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
	var link repository.EventProduct
	err = s.db.Where("event_id = ? AND product_id = ?", eventID, *productID).First(&link).Error
	notLinked := errors.Is(err, gorm.ErrRecordNotFound)
	if err != nil && !notLinked {
		return nil, err
	}
	if notLinked || !link.Available || !p.Active {
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

func (s *Service) All(storeID uuid.UUID) ([]iface.SaleProduct, error) {
	var products []repository.Product
	if err := s.db.Where("store_id = ?", storeID).Order("sku").Find(&products).Error; err != nil {
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
	err := tx.Preload("Barcodes").Where("sku = ?", in.SKU).First(&p).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		p = repository.Product{
			StoreID: storeID, SKU: in.SKU, Name: in.Name, Brand: in.Brand, Price: in.Price, ImageURL: in.ImageURL,
			WarehouseStock: in.StockQuantity, Source: iface.SourceERP, Active: true, Barcodes: sourceBarcodes(in.Barcodes),
		}
		return true, tx.Create(&p).Error
	}
	if err != nil {
		return false, err
	}
	p.Name, p.Brand, p.Price, p.ImageURL = in.Name, in.Brand, in.Price, in.ImageURL
	p.WarehouseStock, p.Source, p.UpdatedAt = in.StockQuantity, iface.SourceERP, nowUTC()
	if barcodesChanged(p.Barcodes, in.Barcodes) {
		if err := tx.Where("product_id = ?", p.ID).Delete(&repository.ProductBarcode{}).Error; err != nil {
			return false, err
		}
		for _, b := range sourceBarcodes(in.Barcodes) {
			b.ProductID = p.ID
			if err := tx.Create(&b).Error; err != nil {
				return false, err
			}
		}
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

func (s *Service) barcodeFree(barcode string) (bool, error) {
	var count int64
	if err := s.db.Model(&repository.ProductBarcode{}).Where("barcode = ?", barcode).Count(&count).Error; err != nil {
		return false, err
	}
	return count == 0, nil
}

// ownerOfBarcode returns the product id owning a barcode, or nil if the barcode is unused.
func (s *Service) ownerOfBarcode(tx *gorm.DB, barcode string) (*uuid.UUID, error) {
	var bc repository.ProductBarcode
	err := tx.Where("barcode = ?", barcode).First(&bc).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &bc.ProductID, nil
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
