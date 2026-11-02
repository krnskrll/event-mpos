package service

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

	eventiface "github.com/krnskrll/event-mpos/internal/event/iface"
	"github.com/krnskrll/event-mpos/internal/platform/common"
	"github.com/krnskrll/event-mpos/internal/platform/database"
	productiface "github.com/krnskrll/event-mpos/internal/product/iface"
	"github.com/krnskrll/event-mpos/internal/stock/iface"
	"github.com/krnskrll/event-mpos/internal/stock/repository"
)

type Service struct {
	events   eventiface.Events
	products productiface.Products
	db       *gorm.DB
	summary  *summaryCache
}

var _ iface.Stock = (*Service)(nil)

func NewService(db *gorm.DB, events eventiface.Events, products productiface.Products) *Service {
	return &Service{db: db, events: events, products: products, summary: newSummaryCache(SummaryTTL)}
}

// InvalidateSummary drops the cached stock summaries for an event. Call it after any committed
// change to the event's stock (order submit/cancel, allocation edit).
func (s *Service) InvalidateSummary(eventID uuid.UUID) { s.summary.invalidate(eventID) }

// ---- admin allocation ----

func (s *Service) AllocationsForEvent(eventID uuid.UUID) ([]iface.AllocationRow, error) {
	ev, err := s.events.Get(eventID)
	if err != nil {
		return nil, err
	}
	products, err := s.products.All(ev.StoreID)
	if err != nil {
		return nil, err
	}
	byProduct, err := s.allocationsByProduct(eventID)
	if err != nil {
		return nil, err
	}
	rows := make([]iface.AllocationRow, 0, len(products))
	for i := range products {
		p := &products[i]
		a := byProduct[p.ID]
		row := iface.AllocationRow{ProductID: p.ID, SKU: p.SKU, Name: p.Name, WarehouseStock: p.WarehouseStock}
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
func (s *Service) SetAllocation(eventID, productID uuid.UUID, allocatedQty int) (*iface.AllocationRow, error) {
	ev, err := s.events.Get(eventID)
	if err != nil {
		return nil, err
	}
	if ev.Type == eventiface.TypeStore && ev.Status == eventiface.StatusClosed {
		return nil, common.NewAppError(http.StatusConflict, "Allocations are locked once a store is CLOSED")
	}
	if ev.Type != eventiface.TypeStore && ev.Status != eventiface.StatusDraft {
		return nil, common.NewAppError(http.StatusConflict,
			"Allocations are locked once an event leaves DRAFT (status: "+string(ev.Status)+")")
	}
	p, err := s.products.Find(productID)
	if err != nil {
		return nil, err
	}
	if p == nil {
		return nil, common.NewAppError(http.StatusNotFound, "Product not found")
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
		alloc = &repository.EventStockAllocation{EventID: eventID, ProductID: productID}
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
	row := iface.AllocationRow{ProductID: p.ID, SKU: p.SKU, Name: p.Name, WarehouseStock: p.WarehouseStock,
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
		alloc = &repository.EventStockAllocation{EventID: eventID, ProductID: productID}
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
func (s *Service) LevelsForEvent(eventID uuid.UUID) ([]iface.StockLevelRow, error) {
	key := summaryKey{eventID, kindWarehouse}
	v, ok, gen := s.summary.get(key)
	if ok {
		return v.([]iface.StockLevelRow), nil
	}
	rows, err := s.levelsForEvent(eventID)
	if err == nil {
		s.summary.put(key, rows, gen)
	}
	return rows, err
}

// BaLevelsForEvent is the compact BA stock summary (served from the summary cache).
func (s *Service) BaLevelsForEvent(eventID uuid.UUID) ([]iface.BaStockLevel, error) {
	key := summaryKey{eventID, kindBa}
	v, ok, gen := s.summary.get(key)
	if ok {
		return v.([]iface.BaStockLevel), nil
	}
	rows, err := s.baLevelsForEvent(eventID)
	if err == nil {
		s.summary.put(key, rows, gen)
	}
	return rows, err
}

func (s *Service) levelsForEvent(eventID uuid.UUID) ([]iface.StockLevelRow, error) {
	ev, err := s.events.Get(eventID)
	if err != nil {
		return nil, err
	}
	products, err := s.productsByID(ev.StoreID)
	if err != nil {
		return nil, err
	}
	var allocs []repository.EventStockAllocation
	if err := s.db.Where("event_id = ?", eventID).Find(&allocs).Error; err != nil {
		return nil, err
	}
	rows := make([]iface.StockLevelRow, 0, len(allocs))
	for i := range allocs {
		a := &allocs[i]
		p, ok := products[a.ProductID]
		if !ok {
			continue
		}
		rows = append(rows, iface.StockLevelRow{
			ProductID: p.ID, SKU: p.SKU, Name: p.Name, AllocatedQty: a.AllocatedQty,
			SoldQty: a.SoldQty, Remaining: a.Remaining(), LowStock: lowStock(a.Remaining()),
		})
	}
	sort.Slice(rows, func(i, j int) bool { return rows[i].SKU < rows[j].SKU })
	return rows, nil
}

func (s *Service) baLevelsForEvent(eventID uuid.UUID) ([]iface.BaStockLevel, error) {
	var allocs []repository.EventStockAllocation
	if err := s.db.Where("event_id = ?", eventID).Find(&allocs).Error; err != nil {
		return nil, err
	}
	rows := make([]iface.BaStockLevel, 0, len(allocs))
	for i := range allocs {
		a := &allocs[i]
		rows = append(rows, iface.BaStockLevel{ProductID: a.ProductID, Remaining: a.Remaining(), LowStock: lowStock(a.Remaining())})
	}
	return rows, nil
}

// Remaining is the remaining quantity for a single product (enriches the snapshot).
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
// the snapshot version.
func (s *Service) AllocationFingerprint(eventID uuid.UUID) (string, error) {
	var allocs []repository.EventStockAllocation
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
func (s *Service) Decrement(tx database.Tx, eventID, productID uuid.UUID, qty int) error {
	if qty <= 0 {
		return common.NewAppError(http.StatusBadRequest, "qty must be positive")
	}
	var alloc repository.EventStockAllocation
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

// Reserve is Decrement for orders a device already confirmed offline: it never rejects. It reports
// whether the sale oversold the allocation; remaining then goes negative so the shortfall stays
// visible to the warehouse instead of the order being lost.
func (s *Service) Reserve(tx database.Tx, eventID, productID uuid.UUID, qty int) (oversold bool, err error) {
	var alloc repository.EventStockAllocation
	err = tx.Clauses(clause.Locking{Strength: "UPDATE"}).
		Where("event_id = ? AND product_id = ?", eventID, productID).First(&alloc).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return true, nil // nothing allocated at all: flag it, nothing to decrement
	}
	if err != nil {
		return false, err
	}
	oversold = alloc.Remaining() < qty
	alloc.SoldQty += qty
	return oversold, tx.Save(&alloc).Error
}

// Restore returns qty of a product to the event's stock (order cancelled), under the same row lock
// as Decrement. Must run inside the caller's transaction.
func (s *Service) Restore(tx database.Tx, eventID, productID uuid.UUID, qty int) error {
	var alloc repository.EventStockAllocation
	if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
		Where("event_id = ? AND product_id = ?", eventID, productID).First(&alloc).Error; err != nil {
		return err
	}
	alloc.SoldQty = max(alloc.SoldQty-qty, 0)
	return tx.Save(&alloc).Error
}

// ---- helpers ----

func (s *Service) findAllocation(tx *gorm.DB, eventID, productID uuid.UUID) (*repository.EventStockAllocation, error) {
	var a repository.EventStockAllocation
	err := tx.Where("event_id = ? AND product_id = ?", eventID, productID).First(&a).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &a, nil
}

func (s *Service) Allocations(eventID uuid.UUID) ([]iface.Allocation, error) {
	var allocs []repository.EventStockAllocation
	if err := s.db.Where("event_id = ?", eventID).Find(&allocs).Error; err != nil {
		return nil, err
	}
	out := make([]iface.Allocation, 0, len(allocs))
	for _, a := range allocs {
		out = append(out, iface.Allocation{ProductID: a.ProductID, AllocatedQty: a.AllocatedQty, SoldQty: a.SoldQty, Remaining: a.Remaining()})
	}
	return out, nil
}

func (s *Service) allocationsByProduct(eventID uuid.UUID) (map[uuid.UUID]*repository.EventStockAllocation, error) {
	var allocs []repository.EventStockAllocation
	if err := s.db.Where("event_id = ?", eventID).Find(&allocs).Error; err != nil {
		return nil, err
	}
	m := make(map[uuid.UUID]*repository.EventStockAllocation, len(allocs))
	for i := range allocs {
		m[allocs[i].ProductID] = &allocs[i]
	}
	return m, nil
}

func (s *Service) productsByID(storeID uuid.UUID) (map[uuid.UUID]productiface.SaleProduct, error) {
	products, err := s.products.All(storeID)
	if err != nil {
		return nil, err
	}
	m := make(map[uuid.UUID]productiface.SaleProduct, len(products))
	for _, p := range products {
		m[p.ID] = p
	}
	return m, nil
}

// LowStockThreshold — remaining in 1..threshold flags a low-stock warning (0 = out, not "low").
const LowStockThreshold = 10

func lowStock(remaining int) bool { return remaining >= 1 && remaining <= LowStockThreshold }
