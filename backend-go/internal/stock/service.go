package stock

import (
	"errors"
	"fmt"
	"net/http"
	"sort"
	"strconv"
	"strings"

	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"github.com/krnskrll/event-mpos/internal/common"
	"github.com/krnskrll/event-mpos/internal/event"
	"github.com/krnskrll/event-mpos/internal/product"
)

// Service holds stock allocation + the oversell-proof decrement — the Go analogue of Kotlin StockService.
type Service struct {
	db      *gorm.DB
	summary *summaryCache
}

func NewService(db *gorm.DB) *Service { return &Service{db: db, summary: newSummaryCache(SummaryTTL)} }

// InvalidateSummary drops the cached stock summaries for an event. Call it after any committed
// change to the event's stock (order submit/cancel, allocation edit).
func (s *Service) InvalidateSummary(eventID uuid.UUID) { s.summary.invalidate(eventID) }

// ---- admin allocation ----

func (s *Service) AllocationsForEvent(eventID uuid.UUID) ([]AllocationRow, error) {
	var products []product.Product
	if err := s.db.Order("sku").Find(&products).Error; err != nil {
		return nil, err
	}
	byProduct, err := s.allocationsByProduct(eventID)
	if err != nil {
		return nil, err
	}
	rows := make([]AllocationRow, 0, len(products))
	for i := range products {
		p := &products[i]
		a := byProduct[p.ID]
		row := AllocationRow{ProductID: p.ID, SKU: p.SKU, Name: p.Name, WarehouseStock: p.WarehouseStock}
		if a != nil {
			row.AllocatedQty = a.AllocatedQty
			row.SoldQty = a.SoldQty
			row.Remaining = a.Remaining()
		}
		rows = append(rows, row)
	}
	return rows, nil
}

// SetAllocation sets how much of a product is allocated to an event. Allowed only while DRAFT.
func (s *Service) SetAllocation(eventID, productID uuid.UUID, allocatedQty int) (*AllocationRow, error) {
	var ev event.Event
	if err := s.db.First(&ev, "id = ?", eventID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, common.NewAppError(http.StatusNotFound, "Event not found")
		}
		return nil, err
	}
	if ev.Status != event.StatusDraft {
		return nil, common.NewAppError(http.StatusConflict,
			"Allocations are locked once an event leaves DRAFT (status: "+string(ev.Status)+")")
	}
	var p product.Product
	if err := s.db.First(&p, "id = ?", productID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, common.NewAppError(http.StatusNotFound, "Product not found")
		}
		return nil, err
	}
	if allocatedQty > p.WarehouseStock {
		return nil, common.NewAppError(http.StatusBadRequest,
			fmt.Sprintf("Only %d in warehouse for %s", p.WarehouseStock, p.SKU))
	}
	alloc, err := s.findAllocation(s.db, eventID, productID)
	if err != nil {
		return nil, err
	}
	if alloc == nil {
		alloc = &EventStockAllocation{EventID: eventID, ProductID: productID}
	}
	if allocatedQty < alloc.SoldQty {
		return nil, common.NewAppError(http.StatusBadRequest,
			fmt.Sprintf("Cannot allocate below already-sold quantity (%d)", alloc.SoldQty))
	}
	alloc.AllocatedQty = allocatedQty
	if err := s.db.Save(alloc).Error; err != nil {
		return nil, err
	}
	s.InvalidateSummary(eventID)
	row := AllocationRow{ProductID: p.ID, SKU: p.SKU, Name: p.Name, WarehouseStock: p.WarehouseStock,
		AllocatedQty: alloc.AllocatedQty, SoldQty: alloc.SoldQty, Remaining: alloc.Remaining()}
	return &row, nil
}

// ForceAllocate sets an allocation directly, bypassing the DRAFT lock (seed/bootstrap).
func (s *Service) ForceAllocate(eventID, productID uuid.UUID, allocatedQty int) error {
	alloc, err := s.findAllocation(s.db, eventID, productID)
	if err != nil {
		return err
	}
	if alloc == nil {
		alloc = &EventStockAllocation{EventID: eventID, ProductID: productID}
	}
	alloc.AllocatedQty = allocatedQty
	if err := s.db.Save(alloc).Error; err != nil {
		return err
	}
	s.InvalidateSummary(eventID)
	return nil
}

// ---- views ----

// LevelsForEvent is the warehouse stock summary (served from the summary cache).
func (s *Service) LevelsForEvent(eventID uuid.UUID) ([]StockLevelRow, error) {
	key := summaryKey{eventID, kindWarehouse}
	v, ok, gen := s.summary.get(key)
	if ok {
		return v.([]StockLevelRow), nil
	}
	rows, err := s.levelsForEvent(eventID)
	if err == nil {
		s.summary.put(key, rows, gen)
	}
	return rows, err
}

// BaLevelsForEvent is the compact BA stock summary (served from the summary cache).
func (s *Service) BaLevelsForEvent(eventID uuid.UUID) ([]BaStockLevel, error) {
	key := summaryKey{eventID, kindBa}
	v, ok, gen := s.summary.get(key)
	if ok {
		return v.([]BaStockLevel), nil
	}
	rows, err := s.baLevelsForEvent(eventID)
	if err == nil {
		s.summary.put(key, rows, gen)
	}
	return rows, err
}

func (s *Service) levelsForEvent(eventID uuid.UUID) ([]StockLevelRow, error) {
	products, err := s.productsByID()
	if err != nil {
		return nil, err
	}
	var allocs []EventStockAllocation
	if err := s.db.Where("event_id = ?", eventID).Find(&allocs).Error; err != nil {
		return nil, err
	}
	rows := make([]StockLevelRow, 0, len(allocs))
	for i := range allocs {
		a := &allocs[i]
		p, ok := products[a.ProductID]
		if !ok {
			continue
		}
		rows = append(rows, StockLevelRow{
			ProductID: p.ID, SKU: p.SKU, Name: p.Name, AllocatedQty: a.AllocatedQty,
			SoldQty: a.SoldQty, Remaining: a.Remaining(), LowStock: lowStock(a.Remaining()),
		})
	}
	sort.Slice(rows, func(i, j int) bool { return rows[i].SKU < rows[j].SKU })
	return rows, nil
}

func (s *Service) baLevelsForEvent(eventID uuid.UUID) ([]BaStockLevel, error) {
	var allocs []EventStockAllocation
	if err := s.db.Where("event_id = ?", eventID).Find(&allocs).Error; err != nil {
		return nil, err
	}
	rows := make([]BaStockLevel, 0, len(allocs))
	for i := range allocs {
		a := &allocs[i]
		rows = append(rows, BaStockLevel{ProductID: a.ProductID, Remaining: a.Remaining(), LowStock: lowStock(a.Remaining())})
	}
	return rows, nil
}

// Remaining is the remaining quantity for a single product (enriches the snapshot). event.StockProvider.
func (s *Service) Remaining(eventID, productID uuid.UUID) (int, error) {
	a, err := s.findAllocation(s.db, eventID, productID)
	if err != nil {
		return 0, err
	}
	if a == nil {
		return 0, nil
	}
	return a.Remaining(), nil
}

// AllocationFingerprint is a fingerprint of allocation *configuration* (not soldQty), folded into
// the snapshot version. event.StockProvider.
func (s *Service) AllocationFingerprint(eventID uuid.UUID) (string, error) {
	var allocs []EventStockAllocation
	if err := s.db.Where("event_id = ?", eventID).Find(&allocs).Error; err != nil {
		return "", err
	}
	sort.Slice(allocs, func(i, j int) bool { return allocs[i].ProductID.String() < allocs[j].ProductID.String() })
	parts := make([]string, 0, len(allocs))
	for i := range allocs {
		parts = append(parts, allocs[i].ProductID.String()+":"+strconv.Itoa(allocs[i].AllocatedQty))
	}
	return strings.Join(parts, ","), nil
}

// ---- the atomic, oversell-proof decrement (called by order submission in Phase 7) ----

// Decrement atomically reserves qty of a product's event stock, holding a PESSIMISTIC_WRITE row lock
// (SELECT ... FOR UPDATE) so concurrent submissions serialize and overselling is impossible. Must be
// called inside a transaction; pass the order-creation tx so it commits/rolls back together.
func (s *Service) Decrement(tx *gorm.DB, eventID, productID uuid.UUID, qty int) error {
	if qty <= 0 {
		return common.NewAppError(http.StatusBadRequest, "qty must be positive")
	}
	var alloc EventStockAllocation
	err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
		Where("event_id = ? AND product_id = ?", eventID, productID).First(&alloc).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return common.NewAppError(http.StatusConflict, "No stock allocated for this product at this event")
	}
	if err != nil {
		return err
	}
	if alloc.Remaining() < qty {
		return common.NewAppError(http.StatusConflict,
			fmt.Sprintf("Insufficient stock: %d remaining, requested %d", alloc.Remaining(), qty))
	}
	alloc.SoldQty += qty
	return tx.Save(&alloc).Error
}

// Restore returns qty of a product to the event's stock (order cancelled), under the same row lock
// as Decrement. Must run inside the caller's transaction.
func (s *Service) Restore(tx *gorm.DB, eventID, productID uuid.UUID, qty int) error {
	var alloc EventStockAllocation
	if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
		Where("event_id = ? AND product_id = ?", eventID, productID).First(&alloc).Error; err != nil {
		return err
	}
	alloc.SoldQty = max(alloc.SoldQty-qty, 0)
	return tx.Save(&alloc).Error
}

// ---- helpers ----

func (s *Service) findAllocation(tx *gorm.DB, eventID, productID uuid.UUID) (*EventStockAllocation, error) {
	var a EventStockAllocation
	err := tx.Where("event_id = ? AND product_id = ?", eventID, productID).First(&a).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &a, nil
}

func (s *Service) allocationsByProduct(eventID uuid.UUID) (map[uuid.UUID]*EventStockAllocation, error) {
	var allocs []EventStockAllocation
	if err := s.db.Where("event_id = ?", eventID).Find(&allocs).Error; err != nil {
		return nil, err
	}
	m := make(map[uuid.UUID]*EventStockAllocation, len(allocs))
	for i := range allocs {
		m[allocs[i].ProductID] = &allocs[i]
	}
	return m, nil
}

func (s *Service) productsByID() (map[uuid.UUID]product.Product, error) {
	var products []product.Product
	if err := s.db.Find(&products).Error; err != nil {
		return nil, err
	}
	m := make(map[uuid.UUID]product.Product, len(products))
	for _, p := range products {
		m[p.ID] = p
	}
	return m, nil
}
