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

var (
	_ iface.Catalog = (*Service)(nil)
	_ iface.Source  = (*Service)(nil)
)

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

// StoreFor returns the POS store linked to a Bhoomy store, creating it on the first visit.
func (s *Service) StoreFor(bhoomyStoreID uuid.UUID, name string) (id uuid.UUID, created bool, err error) {
	if name == "" {
		name = "Bhoomy " + bhoomyStoreID.String()[:8]
	}
	row := struct {
		ID       uuid.UUID
		Inserted bool
	}{}
	err = s.db.Raw(`INSERT INTO pos_stores (id, name, bhoomy_store_id) VALUES (?, ?, ?)
		ON CONFLICT (bhoomy_store_id) DO UPDATE SET name = pos_stores.name
		RETURNING id, (xmax = 0) AS inserted`, uuid.New(), name, bhoomyStoreID).Scan(&row).Error
	return row.ID, row.Inserted, err
}

func (s *Service) Products(storeID uuid.UUID, ids []uuid.UUID) ([]iface.Product, error) {
	q := s.db.Where("store_id = ?", storeID)
	if ids != nil {
		q = q.Where("bhoomy_product_id IN ?", ids)
	}
	var rows []repository.CachedVariant
	if err := q.Order("product_name, bhoomy_product_id, NOT is_default, variant_name").Find(&rows).Error; err != nil {
		return nil, err
	}
	byID := map[uuid.UUID]*iface.Product{}
	var out []*iface.Product
	for _, r := range rows {
		p := byID[r.BhoomyProductID]
		if p == nil {
			p = &iface.Product{ID: r.BhoomyProductID, Name: r.ProductName, Code: r.SKU, Brand: r.BrandName,
				Category: r.CategoryName, ImageURL: r.ImageURL, Archived: true}
			byID[r.BhoomyProductID], out = p, append(out, p)
		}
		p.Archived = p.Archived && r.Status != "active"
		p.Variants = append(p.Variants, iface.Variant{BhoomyVariantID: r.BhoomyVariantID, BhoomyProductID: r.BhoomyProductID,
			ProductName: r.ProductName, VariantName: r.VariantName, SKU: r.SKU, Barcode: r.Barcode, IsDefault: r.IsDefault,
			BrandName: r.BrandName, CategoryID: r.CategoryID, CategoryName: r.CategoryName, ImageURL: r.ImageURL,
			Status: r.Status, ChangedAt: r.ChangedAt})
	}
	res := make([]iface.Product, len(out))
	for i, p := range out {
		res[i] = *p
	}
	return res, nil
}

func (s *Service) Categories(storeID uuid.UUID) ([]iface.Category, error) {
	var rows []repository.CachedVariant
	if err := s.db.Where("store_id = ? AND is_default AND category_id IS NOT NULL", storeID).
		Order("category_name, category_id").Find(&rows).Error; err != nil {
		return nil, err
	}
	var out []iface.Category
	for _, r := range rows {
		if n := len(out); n == 0 || out[n-1].ID != *r.CategoryID {
			out = append(out, iface.Category{ID: *r.CategoryID})
			if r.CategoryName != nil {
				out[n].Name = *r.CategoryName
			}
		}
		c := &out[len(out)-1]
		c.ProductIDs = append(c.ProductIDs, r.BhoomyProductID)
	}
	return out, nil
}
