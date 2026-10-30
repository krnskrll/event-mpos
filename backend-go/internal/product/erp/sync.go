package erp

import (
	"errors"
	"log"
	"strconv"
	"time"

	"gorm.io/gorm"

	"github.com/krnskrll/event-mpos/internal/product"
)

// SyncStatus is a single-row record (id=1) of the most recent ERP sync.
type SyncStatus struct {
	ID          int        `gorm:"primaryKey" json:"id"`
	LastSyncAt  *time.Time `gorm:"column:last_sync_at" json:"lastSyncAt,omitempty"`
	LastResult  *string    `gorm:"column:last_result" json:"lastResult,omitempty"`
	ProductCount int       `gorm:"column:product_count" json:"productCount"`
	Source      *string    `json:"source,omitempty"`
	Message     *string    `json:"message,omitempty"`
}

func (SyncStatus) TableName() string { return "erp_sync_status" }

// SyncResult is returned by POST /api/admin/erp/sync.
type SyncResult struct {
	Ok      bool    `json:"ok"`
	Created int     `json:"created"`
	Updated int     `json:"updated"`
	Total   int     `json:"total"`
	Source  string  `json:"source"`
	Message *string `json:"message,omitempty"`
}

// SyncService pulls products from the active Source and upserts them by SKU (idempotent). Barcodes
// are rewritten only when they actually changed, to avoid unique-constraint churn.
type SyncService struct {
	source Source
	db     *gorm.DB
}

func NewSyncService(source Source, db *gorm.DB) *SyncService {
	return &SyncService{source: source, db: db}
}

func (s *SyncService) Sync() SyncResult {
	log.Printf("ERP sync starting from source: %s", s.source.Description())
	products, err := s.source.FetchProducts()
	if err != nil {
		log.Printf("ERP sync failed: %v", err)
		msg := err.Error()
		result := SyncResult{Ok: false, Source: s.source.Description(), Message: &msg}
		s.saveStatus(result)
		return result
	}

	created, updated := 0, 0
	txErr := s.db.Transaction(func(tx *gorm.DB) error {
		for _, erp := range products {
			var existing product.Product
			err := tx.Preload("Barcodes").Where("sku = ?", erp.SKU).First(&existing).Error
			if errors.Is(err, gorm.ErrRecordNotFound) {
				p := product.Product{
					SKU: erp.SKU, Name: erp.Name, Brand: erp.Brand, Price: erp.Price,
					ImageURL: erp.ImageURL, WarehouseStock: erp.StockQuantity, Source: product.SourceERP,
					Active: true, Barcodes: toBarcodes(erp.Barcodes),
				}
				if err := tx.Create(&p).Error; err != nil {
					return err
				}
				created++
				continue
			}
			if err != nil {
				return err
			}
			existing.Name = erp.Name
			existing.Brand = erp.Brand
			existing.Price = erp.Price
			existing.ImageURL = erp.ImageURL
			existing.WarehouseStock = erp.StockQuantity
			existing.Source = product.SourceERP
			existing.UpdatedAt = time.Now().UTC()
			if barcodesChanged(existing.Barcodes, erp.Barcodes) {
				if err := tx.Where("product_id = ?", existing.ID).Delete(&product.ProductBarcode{}).Error; err != nil {
					return err
				}
				for _, b := range toBarcodes(erp.Barcodes) {
					b.ProductID = existing.ID
					if err := tx.Create(&b).Error; err != nil {
						return err
					}
				}
			}
			if err := tx.Model(&existing).Omit("Barcodes", "Variants").Save(&existing).Error; err != nil {
				return err
			}
			updated++
		}
		return nil
	})
	if txErr != nil {
		log.Printf("ERP sync failed: %v", txErr)
		msg := txErr.Error()
		result := SyncResult{Ok: false, Source: s.source.Description(), Message: &msg}
		s.saveStatus(result)
		return result
	}

	result := SyncResult{Ok: true, Created: created, Updated: updated, Total: len(products), Source: s.source.Description()}
	s.saveStatus(result)
	log.Printf("ERP sync done: %d created, %d updated, %d total", created, updated, len(products))
	return result
}

func (s *SyncService) Status() SyncStatus {
	var st SyncStatus
	if err := s.db.First(&st, 1).Error; err != nil {
		return SyncStatus{ID: 1}
	}
	return st
}

func (s *SyncService) saveStatus(result SyncResult) {
	var st SyncStatus
	if err := s.db.First(&st, 1).Error; err != nil {
		st = SyncStatus{ID: 1}
	}
	now := time.Now().UTC()
	st.LastSyncAt = &now
	res := "SUCCESS"
	if !result.Ok {
		res = "FAILED"
	}
	st.LastResult = &res
	st.ProductCount = result.Total
	src := result.Source
	st.Source = &src
	if result.Message != nil {
		st.Message = result.Message
	} else {
		msg := "created=" + strconv.Itoa(result.Created) + ", updated=" + strconv.Itoa(result.Updated) + ", total=" + strconv.Itoa(result.Total)
		st.Message = &msg
	}
	s.db.Save(&st)
}

func toBarcodes(bs []Barcode) []product.ProductBarcode {
	out := make([]product.ProductBarcode, 0, len(bs))
	for _, b := range bs {
		out = append(out, product.ProductBarcode{Barcode: b.Code, Label: b.Label})
	}
	return out
}

func barcodesChanged(current []product.ProductBarcode, incoming []Barcode) bool {
	cur := make(map[string]bool, len(current))
	for _, c := range current {
		cur[c.Barcode] = true
	}
	inc := make(map[string]bool, len(incoming))
	for _, i := range incoming {
		inc[i.Code] = true
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
