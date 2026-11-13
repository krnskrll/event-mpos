package service

import (
	"net/http"
	"sort"
	"strings"
	"unicode/utf8"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	"gorm.io/gorm"

	catalogiface "github.com/krnskrll/event-mpos/internal/catalog/iface"
	"github.com/krnskrll/event-mpos/internal/platform/common"
	"github.com/krnskrll/event-mpos/internal/product/iface"
	"github.com/krnskrll/event-mpos/internal/product/repository"
)

// Katalog POS: items added from Produk Toko. POS owns the name in POS, photo, prices and where
// they sell; Produk Toko owns the product, its variants, SKUs and barcodes.

var _ iface.Catalogue = (*Service)(nil)

func (s *Service) SetSource(src catalogiface.Source) { s.source = src }

// CatalogueSource lists Produk Toko products not yet in Katalog POS that match q.
func (s *Service) CatalogueSource(storeID uuid.UUID, q string) ([]iface.SourceOption, error) {
	products, err := s.source.Products(storeID, nil)
	if err != nil {
		return nil, err
	}
	linked, err := s.linkedProducts(s.db, storeID)
	if err != nil {
		return nil, err
	}
	q = strings.ToLower(strings.TrimSpace(q))
	out := []iface.SourceOption{}
	for _, p := range products {
		if p.Archived || linked[p.ID] || !matches(p, q) {
			continue
		}
		out = append(out, iface.SourceOption{ProductID: p.ID, Name: p.Name, Code: p.Code, VariantCount: max(len(activeOptions(p)), 1)})
	}
	return out, nil
}

// AddFromProdukToko adds products to Katalog POS as "Tidak dijual" without a price.
func (s *Service) AddFromProdukToko(storeID uuid.UUID, productIDs []uuid.UUID) ([]uuid.UUID, error) {
	products, err := s.source.Products(storeID, productIDs)
	if err != nil {
		return nil, err
	}
	if len(products) != len(unique(productIDs)) {
		return nil, common.NewAppError(http.StatusNotFound, "Produk Toko product not found")
	}
	ids := make([]uuid.UUID, 0, len(products))
	err = s.db.Transaction(func(tx *gorm.DB) error {
		linked, err := s.linkedProducts(tx, storeID)
		if err != nil {
			return err
		}
		for _, p := range products {
			if linked[p.ID] {
				return common.NewAppError(http.StatusConflict, p.Name+" is already in Katalog POS")
			}
			item, err := s.newItem(tx, storeID, p)
			if err != nil {
				return err
			}
			if err := tx.Create(item).Error; err != nil {
				return err
			}
			ids = append(ids, item.ID)
		}
		return nil
	})
	return ids, err
}

func (s *Service) newItem(tx *gorm.DB, storeID uuid.UUID, p catalogiface.Product) (*repository.Product, error) {
	if taken, err := exists(tx.Model(&repository.Product{}).Where("store_id = ? AND sku = ?", storeID, p.Code)); err != nil || taken {
		return nil, conflictOr(err, "Code "+p.Code+" is already used by an older POS product")
	}
	item := &repository.Product{ID: uuid.New(), StoreID: storeID, SKU: p.Code, DisplayName: shorten(p.Name),
		Source: iface.SourceBhoomy, Status: repository.StatusArchived, BhoomyProductID: &p.ID}
	def := repository.ProductVariant{ItemID: item.ID, SKU: p.Code, SellInPOS: true, IsDefault: true}
	var skus []string
	for _, v := range p.Variants {
		switch {
		case v.IsDefault:
			def.BhoomyVariantID = &v.BhoomyVariantID
		case v.Status == "active":
			skus = append(skus, v.SKU)
			item.Variants = append(item.Variants, repository.ProductVariant{ItemID: item.ID, Name: v.VariantName, SKU: v.SKU,
				SellInPOS: true, BhoomyVariantID: &v.BhoomyVariantID})
		}
	}
	if len(skus) > 0 {
		if taken, err := exists(tx.Model(&repository.ProductVariant{}).Where("store_id = ? AND sku IN ? AND NOT is_default", storeID, skus)); err != nil || taken {
			return nil, conflictOr(err, "A variant SKU of "+p.Name+" is already used by an older POS product")
		}
	}
	item.Variants = append([]repository.ProductVariant{def}, item.Variants...)
	barcodes, err := s.freeBarcodes(tx, storeID, item.ID, p)
	item.Barcodes = barcodes
	return item, err
}

// CatalogueItems lists Katalog POS (items added from Produk Toko only).
func (s *Service) CatalogueItems(storeID uuid.UUID) ([]iface.CatalogueItem, error) {
	var items []repository.Product
	if err := s.preloaded().Where("store_id = ? AND bhoomy_product_id IS NOT NULL", storeID).Order("display_name").Find(&items).Error; err != nil {
		return nil, err
	}
	src, err := s.sourceOf(storeID, items)
	if err != nil {
		return nil, err
	}
	var sold []struct{ ItemID, LocationID uuid.UUID }
	if err := s.db.Table("pos_location_items li").Select("DISTINCT v.item_id, li.location_id").
		Joins("JOIN pos_item_variants v ON v.id = li.item_variant_id").
		Where("li.active AND v.store_id = ?", storeID).Scan(&sold).Error; err != nil {
		return nil, err
	}
	locations := map[uuid.UUID][]uuid.UUID{}
	for _, r := range sold {
		locations[r.ItemID] = append(locations[r.ItemID], r.LocationID)
	}
	out := make([]iface.CatalogueItem, 0, len(items))
	for i := range items {
		p := &items[i]
		sp := src[*p.BhoomyProductID]
		row := iface.CatalogueItem{ID: p.ID, DisplayName: p.DisplayName, PhotoURL: p.PhotoURL, ProductName: sp.Name, Code: sp.Code,
			CategoryName: sp.Category, ImageURL: sp.ImageURL, Status: p.Status, LocationIDs: locations[p.ID]}
		if row.LocationIDs == nil {
			row.LocationIDs = []uuid.UUID{}
		}
		row.Variants, row.VariantsSold, row.PriceMin, row.PriceMax = priceRange(p)
		out = append(out, row)
	}
	return out, nil
}

// CatalogueItem returns the edit form; locs are the store's locations, marked where the item sells.
func (s *Service) CatalogueItem(storeID, id uuid.UUID, locs []iface.ItemLocation) (*iface.CatalogueItemDetail, error) {
	p, err := s.findLinked(storeID, id)
	if err != nil {
		return nil, err
	}
	src, err := s.sourceOf(storeID, []repository.Product{*p})
	if err != nil {
		return nil, err
	}
	sp := src[*p.BhoomyProductID]
	d := &iface.CatalogueItemDetail{ID: p.ID, DisplayName: p.DisplayName, PhotoURL: p.PhotoURL, Status: p.Status,
		BasePrice: p.Default().Price, Locations: locs, Variants: []iface.CatalogueVariant{},
		Source: iface.SourceFacts{ProductName: sp.Name, Code: sp.Code, Brand: sp.Brand, CategoryName: sp.Category,
			ImageURL: sp.ImageURL, Archived: sp.Archived}}
	for _, v := range sp.Variants {
		if v.IsDefault {
			d.Source.Barcode = v.Barcode
		}
	}
	for _, v := range p.Options() {
		d.Variants = append(d.Variants, iface.CatalogueVariant{ID: v.ID, Name: v.Name, SKU: v.SKU, SellInPOS: v.SellInPOS, Price: v.Price})
	}
	ids := make([]uuid.UUID, len(p.Variants))
	for i, v := range p.Variants {
		ids[i] = v.ID
	}
	var rows []repository.LocationItem
	if err := s.db.Where("item_variant_id IN ?", ids).Find(&rows).Error; err != nil {
		return nil, err
	}
	for i := range d.Locations {
		for _, r := range rows {
			if r.LocationID == d.Locations[i].ID {
				d.Locations[i].Sold = d.Locations[i].Sold || r.Active
				d.Locations[i].OwnPrice = d.Locations[i].OwnPrice || r.PriceOverride != nil
			}
		}
	}
	return d, nil
}

// UpdateCatalogueItem saves the edit form. A location switched on sells every variant; one
// switched off stops them all; unchanged locations keep their per-variant choices.
func (s *Service) UpdateCatalogueItem(storeID, id uuid.UUID, req iface.UpdateItemRequest) error {
	name := strings.TrimSpace(req.DisplayName)
	if name == "" || utf8.RuneCountInString(name) > iface.MaxDisplayName {
		return common.NewAppError(http.StatusBadRequest, "displayName: 1 to 40 characters")
	}
	if err := checkPrice("basePrice", req.BasePrice); err != nil {
		return err
	}
	if req.Status == repository.StatusActive && req.BasePrice == nil {
		return common.NewAppError(http.StatusBadRequest, "basePrice: set a price before selling in POS")
	}
	p, err := s.findLinked(storeID, id)
	if err != nil {
		return err
	}
	byID := map[uuid.UUID]*repository.ProductVariant{}
	for i := range p.Variants {
		byID[p.Variants[i].ID] = &p.Variants[i]
	}
	p.Default().Price = req.BasePrice
	for _, in := range req.Variants {
		v := byID[in.ID]
		if v == nil || v.IsDefault {
			return common.NewAppError(http.StatusNotFound, "Variant not found")
		}
		if err := checkPrice("price", in.Price); err != nil {
			return err
		}
		v.SellInPOS, v.Price = in.SellInPOS, in.Price
	}
	p.DisplayName, p.Status, p.UpdatedAt = name, req.Status, nowUTC()
	return s.db.Transaction(func(tx *gorm.DB) error {
		for i := range p.Variants {
			if err := tx.Save(&p.Variants[i]).Error; err != nil {
				return err
			}
		}
		if err := tx.Model(p).Select("display_name", "status", "updated_at").Updates(p).Error; err != nil {
			return err
		}
		for _, loc := range req.Locations {
			rows, err := locationItemsIn(tx, loc.ID)
			if err != nil {
				return err
			}
			sold := false
			for _, v := range p.Variants {
				sold = sold || rows[v.ID].Active
			}
			if sold == loc.Sold {
				continue
			}
			for _, v := range p.Variants {
				li := rows[v.ID]
				if err := saveLocationItem(tx, loc.ID, v.ID, loc.Sold, li.PriceOverride, &li); err != nil {
					return err
				}
			}
		}
		return nil
	})
}

// LocationRows lists the Katalog POS variants a location has rows for, sold or not.
func (s *Service) LocationRows(storeID, locationID uuid.UUID) ([]iface.LocationRow, error) {
	rows, err := s.locationItems(locationID)
	if err != nil {
		return nil, err
	}
	var items []repository.Product
	err = s.preloaded().Where("store_id = ? AND bhoomy_product_id IS NOT NULL AND id IN (?)", storeID,
		s.db.Model(&repository.ProductVariant{}).Select("item_id").Where("id IN (?)",
			s.db.Model(&repository.LocationItem{}).Select("item_variant_id").Where("location_id = ?", locationID))).
		Order("display_name").Find(&items).Error
	if err != nil {
		return nil, err
	}
	out := []iface.LocationRow{}
	for i := range items {
		p := &items[i]
		options := p.Options()
		sort.SliceStable(options, func(a, b int) bool { return options[a].Name < options[b].Name })
		for _, v := range append([]repository.ProductVariant{*p.Default()}, options...) {
			if !v.SellInPOS {
				continue
			}
			li := rows[v.ID]
			catalog := p.Default().Price
			if v.Price != nil {
				catalog = v.Price
			}
			out = append(out, iface.LocationRow{ItemID: p.ID, VariantID: v.ID, ItemName: p.DisplayName, VariantName: v.Name,
				CatalogPrice: catalog, Sold: li.Active, Price: li.PriceOverride, BhoomyVariantID: v.BhoomyVariantID})
		}
	}
	return out, nil
}

// SaveLocationRows saves what a location sells and its prices, per variant.
func (s *Service) SaveLocationRows(storeID, locationID uuid.UUID, in []iface.LocationRowInput) error {
	return s.db.Transaction(func(tx *gorm.DB) error {
		rows, err := locationItemsIn(tx, locationID)
		if err != nil {
			return err
		}
		for _, r := range in {
			if err := checkPrice("price", r.Price); err != nil {
				return err
			}
			if ok, err := exists(tx.Model(&repository.ProductVariant{}).Where("id = ? AND store_id = ?", r.VariantID, storeID)); err != nil || !ok {
				return notFoundOr(err, "Variant not found")
			}
			li := rows[r.VariantID]
			if err := saveLocationItem(tx, locationID, r.VariantID, r.Sold, r.Price, &li); err != nil {
				return err
			}
		}
		return nil
	})
}

// AddToLocation starts selling every variant of the items at a location, keeping its prices.
func (s *Service) AddToLocation(storeID, locationID uuid.UUID, itemIDs []uuid.UUID) error {
	return s.db.Transaction(func(tx *gorm.DB) error {
		rows, err := locationItemsIn(tx, locationID)
		if err != nil {
			return err
		}
		for _, id := range unique(itemIDs) {
			p, err := s.findLinked(storeID, id)
			if err != nil {
				return err
			}
			for _, v := range p.Variants {
				li := rows[v.ID]
				if err := saveLocationItem(tx, locationID, v.ID, true, li.PriceOverride, &li); err != nil {
					return err
				}
			}
		}
		return nil
	})
}

// LocationItemCounts is how many Katalog POS items each location sells.
func (s *Service) LocationItemCounts(storeID uuid.UUID) (map[uuid.UUID]int, error) {
	var rows []struct {
		LocationID uuid.UUID
		N          int
	}
	err := s.db.Table("pos_location_items li").Select("li.location_id, COUNT(DISTINCT v.item_id) AS n").
		Joins("JOIN pos_item_variants v ON v.id = li.item_variant_id").
		Joins("JOIN pos_items i ON i.id = v.item_id").
		Where("li.active AND i.store_id = ? AND i.bhoomy_product_id IS NOT NULL", storeID).
		Group("li.location_id").Scan(&rows).Error
	out := make(map[uuid.UUID]int, len(rows))
	for _, r := range rows {
		out[r.LocationID] = r.N
	}
	return out, err
}

// Reconcile makes linked items follow Produk Toko: variant names and SKUs, barcodes, new
// variants (added switched off) and archived ones (switched off).
func (s *Service) Reconcile(storeID uuid.UUID, productIDs []uuid.UUID) error {
	var items []repository.Product
	if err := s.preloaded().Where("store_id = ? AND bhoomy_product_id IN ?", storeID, productIDs).Find(&items).Error; err != nil {
		return err
	}
	if len(items) == 0 {
		return nil
	}
	src, err := s.sourceOf(storeID, items)
	if err != nil {
		return err
	}
	return s.db.Transaction(func(tx *gorm.DB) error {
		for i := range items {
			if err := s.reconcile(tx, &items[i], src[*items[i].BhoomyProductID]); err != nil {
				return err
			}
		}
		return nil
	})
}

func (s *Service) reconcile(tx *gorm.DB, p *repository.Product, sp catalogiface.Product) error {
	byBhoomy := map[uuid.UUID]*repository.ProductVariant{}
	for i := range p.Variants {
		if id := p.Variants[i].BhoomyVariantID; id != nil {
			byBhoomy[*id] = &p.Variants[i]
		}
	}
	for _, sv := range sp.Variants {
		v := byBhoomy[sv.BhoomyVariantID]
		switch {
		case v == nil && sv.IsDefault:
			p.Default().BhoomyVariantID = &sv.BhoomyVariantID
			v = p.Default()
		case v == nil && sv.Status == "active":
			nv := repository.ProductVariant{ItemID: p.ID, Name: sv.VariantName, SKU: sv.SKU, BhoomyVariantID: &sv.BhoomyVariantID}
			if err := tx.Create(&nv).Error; err != nil {
				return err
			}
			continue
		case v == nil:
			continue
		case !v.IsDefault:
			v.Name, v.SKU = sv.VariantName, sv.SKU
			v.SellInPOS = v.SellInPOS && sv.Status == "active"
		}
		if err := tx.Save(v).Error; err != nil {
			return err
		}
	}
	barcodes, err := s.freeBarcodes(tx, p.StoreID, p.ID, sp)
	if err != nil {
		return err
	}
	if err := tx.Where("item_id = ?", p.ID).Delete(&repository.ProductBarcode{}).Error; err != nil {
		return err
	}
	for i := range barcodes {
		if err := tx.Create(&barcodes[i]).Error; err != nil {
			return err
		}
	}
	return tx.Model(&repository.Product{}).Where("id = ?", p.ID).Update("updated_at", nowUTC()).Error
}

// freeBarcodes are the product's active barcodes not used by another item of the store.
func (s *Service) freeBarcodes(tx *gorm.DB, storeID, itemID uuid.UUID, p catalogiface.Product) ([]repository.ProductBarcode, error) {
	var out []repository.ProductBarcode
	seen := map[string]bool{}
	for _, v := range p.Variants {
		if v.Barcode == nil || *v.Barcode == "" || v.Status != "active" || seen[*v.Barcode] {
			continue
		}
		seen[*v.Barcode] = true
		owner, err := s.ownerOfBarcode(tx, storeID, *v.Barcode)
		if err != nil {
			return nil, err
		}
		if owner != nil && *owner != itemID {
			continue
		}
		b := repository.ProductBarcode{ItemID: itemID, StoreID: storeID, Barcode: *v.Barcode}
		if !v.IsDefault {
			b.Label = &v.VariantName
		}
		out = append(out, b)
	}
	return out, nil
}

func (s *Service) Categories(storeID uuid.UUID) ([]iface.Category, error) {
	cats, err := s.source.Categories(storeID)
	if err != nil {
		return nil, err
	}
	var items []repository.Product
	if err := s.db.Select("id, bhoomy_product_id").Where("store_id = ? AND bhoomy_product_id IS NOT NULL", storeID).
		Find(&items).Error; err != nil {
		return nil, err
	}
	itemOf := make(map[uuid.UUID]uuid.UUID, len(items))
	for _, p := range items {
		itemOf[*p.BhoomyProductID] = p.ID
	}
	out := make([]iface.Category, 0, len(cats))
	for _, c := range cats {
		cat := iface.Category{ID: c.ID, Name: c.Name, ItemIDs: []uuid.UUID{}}
		for _, pid := range c.ProductIDs {
			if id, ok := itemOf[pid]; ok {
				cat.ItemIDs = append(cat.ItemIDs, id)
			}
		}
		out = append(out, cat)
	}
	return out, nil
}

func (s *Service) RequireItems(storeID uuid.UUID, ids []uuid.UUID) error {
	ids = unique(ids)
	var n int64
	if err := s.db.Model(&repository.Product{}).Where("store_id = ? AND id IN ?", storeID, ids).Count(&n).Error; err != nil {
		return err
	}
	if int(n) != len(ids) {
		return common.NewAppError(http.StatusBadRequest, "Some products are not in Katalog POS")
	}
	return nil
}

func (s *Service) findLinked(storeID, id uuid.UUID) (*repository.Product, error) {
	p, err := s.findIn(storeID, id)
	if err != nil {
		return nil, err
	}
	if p.BhoomyProductID == nil {
		return nil, common.NewAppError(http.StatusNotFound, "Product not found")
	}
	return p, nil
}

func (s *Service) linkedProducts(tx *gorm.DB, storeID uuid.UUID) (map[uuid.UUID]bool, error) {
	var ids []uuid.UUID
	err := tx.Model(&repository.Product{}).Where("store_id = ? AND bhoomy_product_id IS NOT NULL", storeID).Pluck("bhoomy_product_id", &ids).Error
	out := make(map[uuid.UUID]bool, len(ids))
	for _, id := range ids {
		out[id] = true
	}
	return out, err
}

func (s *Service) sourceOf(storeID uuid.UUID, items []repository.Product) (map[uuid.UUID]catalogiface.Product, error) {
	ids := make([]uuid.UUID, 0, len(items))
	for _, p := range items {
		ids = append(ids, *p.BhoomyProductID)
	}
	products, err := s.source.Products(storeID, ids)
	out := make(map[uuid.UUID]catalogiface.Product, len(products))
	for _, p := range products {
		out[p.ID] = p
	}
	return out, err
}

func locationItemsIn(tx *gorm.DB, locationID uuid.UUID) (repository.LocationItems, error) {
	var list []repository.LocationItem
	if err := tx.Where("location_id = ?", locationID).Find(&list).Error; err != nil {
		return nil, err
	}
	rows := make(repository.LocationItems, len(list))
	for _, li := range list {
		rows[li.ItemVariantID] = li
	}
	return rows, nil
}

// priceRange counts the variants (an item without options counts as one) and the prices it sells at.
func priceRange(p *repository.Product) (total, sold int, lo, hi *decimal.Decimal) {
	base := p.Default().Price
	options := p.Options()
	if len(options) == 0 {
		return 1, 1, base, base
	}
	for _, v := range options {
		total++
		if !v.SellInPOS {
			continue
		}
		sold++
		price := v.Price
		if price == nil {
			price = base
		}
		if price == nil {
			continue
		}
		if lo == nil || price.LessThan(*lo) {
			lo = price
		}
		if hi == nil || price.GreaterThan(*hi) {
			hi = price
		}
	}
	return total, sold, lo, hi
}

func activeOptions(p catalogiface.Product) []catalogiface.Variant {
	var out []catalogiface.Variant
	for _, v := range p.Variants {
		if !v.IsDefault && v.Status == "active" {
			out = append(out, v)
		}
	}
	return out
}

func matches(p catalogiface.Product, q string) bool {
	if q == "" || strings.Contains(strings.ToLower(p.Name+" "+p.Code), q) {
		return true
	}
	for _, v := range p.Variants {
		if strings.Contains(strings.ToLower(v.SKU), q) || (v.Barcode != nil && strings.Contains(*v.Barcode, q)) {
			return true
		}
	}
	return false
}

func checkPrice(field string, price *decimal.Decimal) error {
	if err := common.CheckRupiah(field, price); err != nil {
		return err
	}
	if price != nil && price.IsNegative() {
		return common.NewAppError(http.StatusBadRequest, field+": must not be negative")
	}
	return nil
}

func shorten(name string) string {
	name = strings.TrimSpace(name)
	if r := []rune(name); len(r) > iface.MaxDisplayName {
		return strings.TrimSpace(string(r[:iface.MaxDisplayName]))
	}
	return name
}

func unique(ids []uuid.UUID) []uuid.UUID {
	seen := map[uuid.UUID]bool{}
	out := make([]uuid.UUID, 0, len(ids))
	for _, id := range ids {
		if !seen[id] {
			seen[id] = true
			out = append(out, id)
		}
	}
	return out
}

func exists(q *gorm.DB) (bool, error) {
	var n int64
	err := q.Count(&n).Error
	return n > 0, err
}

func conflictOr(err error, msg string) error {
	if err != nil {
		return err
	}
	return common.NewAppError(http.StatusConflict, msg)
}

func notFoundOr(err error, msg string) error {
	if err != nil {
		return err
	}
	return common.NewAppError(http.StatusNotFound, msg)
}
