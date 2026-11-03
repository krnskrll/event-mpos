package service

import (
	"errors"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"github.com/krnskrll/event-mpos/internal/catalog/iface"
	"github.com/krnskrll/event-mpos/internal/catalog/repository"
)

type Service struct{ db *gorm.DB }

var _ iface.Catalog = (*Service)(nil)

func NewService(db *gorm.DB) *Service { return &Service{db: db} }

func (s *Service) Links() ([]iface.Link, error) {
	var links []iface.Link
	err := s.db.Table("pos_stores").Select("id AS store_id, bhoomy_store_id").Where("bhoomy_store_id IS NOT NULL").Scan(&links).Error
	return links, err
}

func (s *Service) Cursor(storeID uuid.UUID) (*iface.Cursor, error) {
	c, err := s.cursor(s.db, storeID)
	if err != nil || c == nil || c.CursorAt == nil || c.CursorID == nil {
		return nil, err
	}
	return &iface.Cursor{Since: *c.CursorAt, After: *c.CursorID}, nil
}

func (s *Service) SavePage(storeID uuid.UUID, page iface.Page, at iface.Cursor) error {
	now := time.Now().UTC()
	return s.db.Transaction(func(tx *gorm.DB) error {
		if len(page.Items) > 0 {
			rows := make([]repository.CachedVariant, 0, len(page.Items))
			for _, v := range page.Items {
				rows = append(rows, repository.CachedVariant{
					BhoomyVariantID: v.BhoomyVariantID, StoreID: storeID, BhoomyProductID: v.BhoomyProductID,
					ProductName: v.ProductName, VariantName: v.VariantName, SKU: v.SKU, Barcode: v.Barcode,
					IsDefault: v.IsDefault, BrandName: v.BrandName, CategoryID: v.CategoryID, CategoryName: v.CategoryName,
					ImageURL: v.ImageURL, Status: v.Status, ChangedAt: v.ChangedAt, SyncedAt: now,
				})
			}
			if err := tx.Clauses(clause.OnConflict{UpdateAll: true}).Create(&rows).Error; err != nil {
				return err
			}
		}
		return tx.Save(&repository.SyncCursor{StoreID: storeID, Source: repository.SourceProdukToko,
			CursorAt: &at.Since, CursorID: &at.After, UpdatedAt: now}).Error
	})
}

// SaveError records a failed pull and keeps the cursor where it was.
func (s *Service) SaveError(storeID uuid.UUID, syncErr error) error {
	c, err := s.cursor(s.db, storeID)
	if err != nil {
		return err
	}
	if c == nil {
		c = &repository.SyncCursor{StoreID: storeID, Source: repository.SourceProdukToko}
	}
	msg := syncErr.Error()
	c.LastError, c.UpdatedAt = &msg, time.Now().UTC()
	return s.db.Save(c).Error
}

func (s *Service) Status(storeID uuid.UUID) (*iface.SyncStatus, error) {
	st := &iface.SyncStatus{}
	if err := s.db.Table("pos_stores").Select("bhoomy_store_id IS NOT NULL").Where("id = ?", storeID).Scan(&st.Linked).Error; err != nil {
		return nil, err
	}
	if err := s.db.Model(&repository.CachedVariant{}).Where("store_id = ?", storeID).Count(&st.Cached).Error; err != nil {
		return nil, err
	}
	c, err := s.cursor(s.db, storeID)
	if err != nil {
		return nil, err
	}
	if c != nil {
		st.CursorAt, st.LastError, st.UpdatedAt = c.CursorAt, c.LastError, &c.UpdatedAt
	}
	return st, nil
}

func (s *Service) cursor(tx *gorm.DB, storeID uuid.UUID) (*repository.SyncCursor, error) {
	var c repository.SyncCursor
	err := tx.Where("store_id = ? AND source = ?", storeID, repository.SourceProdukToko).First(&c).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	return &c, err
}
