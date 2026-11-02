package erp

import (
	"log"
	"strconv"
	"time"

	"gorm.io/gorm"

	productiface "github.com/krnskrll/event-mpos/internal/product/iface"
	syncconfigiface "github.com/krnskrll/event-mpos/internal/syncconfig/iface"
)

// SyncStatus is a single-row record (id=1) of the most recent ERP sync.
type SyncStatus struct {
	ID           int        `gorm:"primaryKey" json:"id"`
	LastSyncAt   *time.Time `gorm:"column:last_sync_at" json:"lastSyncAt,omitempty"`
	LastResult   *string    `gorm:"column:last_result" json:"lastResult,omitempty"`
	ProductCount int        `gorm:"column:product_count" json:"productCount"`
	Source       *string    `json:"source,omitempty"`
	Message      *string    `json:"message,omitempty"`
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
	source   Source
	db       *gorm.DB
	products productiface.Products
}

func NewSyncService(source Source, db *gorm.DB, products productiface.Products) *SyncService {
	return &SyncService{source: source, db: db, products: products}
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
			isNew, err := s.products.Upsert(tx, erp.toSource())
			if err != nil {
				return err
			}
			if isNew {
				created++
			} else {
				updated++
			}
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

func (p Product) toSource() productiface.SourceProduct {
	barcodes := make([]productiface.BarcodeInput, 0, len(p.Barcodes))
	for _, b := range p.Barcodes {
		barcodes = append(barcodes, productiface.BarcodeInput{Barcode: b.Code, Label: b.Label})
	}
	return productiface.SourceProduct{
		SKU: p.SKU, Name: p.Name, Brand: p.Brand, Price: p.Price, ImageURL: p.ImageURL,
		StockQuantity: p.StockQuantity, Barcodes: barcodes,
	}
}

func (s *SyncService) LastRun() syncconfigiface.LastRun {
	st := s.Status()
	return syncconfigiface.LastRun{At: st.LastSyncAt, Result: st.LastResult, Count: st.ProductCount, Source: st.Source, Message: st.Message}
}
