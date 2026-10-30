package order

import (
	"errors"
	"fmt"
	"net/http"
	"sort"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"github.com/krnskrll/event-mpos/internal/audit"
	"github.com/krnskrll/event-mpos/internal/common"
	"github.com/krnskrll/event-mpos/internal/event"
	"github.com/krnskrll/event-mpos/internal/product"
	"github.com/krnskrll/event-mpos/internal/promotion"
	"github.com/krnskrll/event-mpos/internal/stock"
	"github.com/krnskrll/event-mpos/internal/user"
)

// Service creates orders atomically and drives the warehouse board — the Go analogue of Kotlin
// OrderService.
type Service struct {
	db         *gorm.DB
	stock      *stock.Service
	promotions *promotion.Service
	audit      *audit.Service
	users      *user.Service
	jakarta    *time.Location
	// broadcast pushes an order change to the warehouse board after commit; wired in Phase 8 (nil = no-op).
	broadcast func(eventID uuid.UUID, view WarehouseOrderView)
}

func NewService(db *gorm.DB, stk *stock.Service, promotions *promotion.Service, aud *audit.Service, users *user.Service) *Service {
	loc, err := time.LoadLocation("Asia/Jakarta")
	if err != nil {
		loc = time.FixedZone("WIB", 7*3600)
	}
	return &Service{db: db, stock: stk, promotions: promotions, audit: aud, users: users, jakarta: loc}
}

// SetBroadcaster wires the warehouse-board push (Phase 8).
func (s *Service) SetBroadcaster(fn func(eventID uuid.UUID, view WarehouseOrderView)) {
	s.broadcast = fn
}

type draftItem struct {
	productID   uuid.UUID
	variantID   *uuid.UUID
	variantName *string
	sku, name   string
	unitPrice   decimal.Decimal
	qty         int
	lineTotal   decimal.Decimal
}

// Submit creates an order ATOMICALLY: idempotency → validate → decrement stock (row locks) → apply
// best promotion → allocate the daily number → persist + audit. Any failure rolls the whole thing
// back. A retried submit with the same idempotency key returns the original order.
func (s *Service) Submit(eventID, baUserID uuid.UUID, req SubmitOrderRequest) (*OrderConfirmation, error) {
	// 1) idempotency fast-path
	if existing, err := s.findByIdempotencyKey(req.IdempotencyKey); err != nil {
		return nil, err
	} else if existing != nil {
		return s.toConfirmation(existing, true)
	}

	// event must be ACTIVE
	var ev event.Event
	if err := s.db.First(&ev, "id = ?", eventID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, common.NewAppError(http.StatusNotFound, "Event not found")
		}
		return nil, err
	}
	if ev.Status != event.StatusActive {
		return nil, common.NewAppError(http.StatusConflict,
			"Event "+ev.Code+" is "+string(ev.Status)+" — cannot accept orders")
	}

	// 2) build line items from the catalogue (server-authoritative prices)
	var lines []OrderLineRequest
	for _, l := range req.Lines {
		if l.Qty > 0 {
			lines = append(lines, l)
		}
	}
	if len(lines) == 0 {
		return nil, common.NewAppError(http.StatusBadRequest, "Cart is empty")
	}

	cartLines := make([]promotion.CartLine, 0, len(lines))
	drafts := make([]draftItem, 0, len(lines))
	for _, line := range lines {
		var p product.Product
		if err := s.db.Preload("Variants").First(&p, "id = ?", line.ProductID).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return nil, common.NewAppError(http.StatusBadRequest, "Unknown product "+line.ProductID.String())
			}
			return nil, err
		}
		unitPrice := p.Price
		var variantID *uuid.UUID
		var variantName *string
		if line.VariantID != nil {
			var found *product.ProductVariant
			for i := range p.Variants {
				if p.Variants[i].ID == *line.VariantID && p.Variants[i].Active {
					found = &p.Variants[i]
					break
				}
			}
			if found == nil {
				return nil, common.NewAppError(http.StatusBadRequest,
					"Unknown variant "+line.VariantID.String()+" for product "+p.SKU)
			}
			unitPrice = found.EffectivePrice(p.Price)
			variantID = &found.ID
			name := found.Name
			variantName = &name
		}
		lineTotal := unitPrice.Mul(decimal.NewFromInt(int64(line.Qty)))
		cartLines = append(cartLines, promotion.CartLine{ProductID: p.ID, UnitPrice: unitPrice, Qty: line.Qty})
		drafts = append(drafts, draftItem{
			productID: p.ID, variantID: variantID, variantName: variantName, sku: p.SKU, name: p.Name,
			unitPrice: unitPrice, qty: line.Qty, lineTotal: lineTotal,
		})
	}

	// 4) pricing (single best promotion — same engine as the quote endpoint)
	subtotal := decimal.Zero
	for _, d := range drafts {
		subtotal = subtotal.Add(d.lineTotal)
	}
	subtotal = subtotal.Round(2)
	applied, err := s.promotions.BestFor(eventID, cartLines, time.Now())
	if err != nil {
		return nil, err
	}
	discount := decimal.Zero.Round(2)
	var promoID *uuid.UUID
	var promoName *string
	if applied != nil {
		discount = applied.Discount
		promoID = &applied.Promotion.ID
		promoName = &applied.Promotion.Name
	}
	total := subtotal.Sub(discount)
	if total.IsNegative() {
		total = decimal.Zero
	}
	total = total.Round(2)

	dateKey := time.Now().In(s.jakarta).Format("20060102")

	order := Order{
		EventID: eventID, BaUserID: baUserID, IdempotencyKey: req.IdempotencyKey,
		ReferenceID: "", CustomerName: strings.TrimSpace(req.Customer.Name), CustomerQr: req.Customer.QRCode,
		CustomerTier: req.Customer.Tier, Subtotal: subtotal, Discount: discount, Total: total,
		PromotionID: promoID, PromotionName: promoName, Status: StatusNew,
	}

	// 3+5+6) atomic tx: stock decrement (row locks) → daily number → persist + audit
	// Lock allocation rows in a fixed (product id) order so two multi-line orders can't each hold
	// one row and wait on the other.
	lockOrder := append([]OrderLineRequest(nil), lines...)
	sort.Slice(lockOrder, func(i, j int) bool { return lockOrder[i].ProductID.String() < lockOrder[j].ProductID.String() })
	txErr := s.db.Transaction(func(tx *gorm.DB) error {
		for _, line := range lockOrder {
			if err := s.stock.Decrement(tx, eventID, line.ProductID, line.Qty); err != nil {
				return err
			}
		}
		number, err := s.nextOrderNumber(tx, eventID, dateKey)
		if err != nil {
			return err
		}
		order.OrderNumber = number
		order.ReferenceID = fmt.Sprintf("ORD-%s-%s-%03d", dateKey, ev.Code, number)
		for _, d := range drafts {
			order.Items = append(order.Items, OrderItem{
				ProductID: d.productID, VariantID: d.variantID, VariantName: d.variantName,
				SKU: d.sku, Name: d.name, UnitPrice: d.unitPrice, Qty: d.qty, LineTotal: d.lineTotal,
			})
		}
		if err := tx.Create(&order).Error; err != nil {
			return err
		}
		return s.audit.RecordOrderCreated(tx, eventID, order.ID, order.ReferenceID, &baUserID)
	})
	if txErr != nil {
		// concurrent submit with the same idempotency key won the race — return its order
		if errors.Is(txErr, gorm.ErrDuplicatedKey) {
			if existing, e := s.findByIdempotencyKey(req.IdempotencyKey); e == nil && existing != nil {
				return s.toConfirmation(existing, true)
			}
		}
		return nil, txErr
	}

	s.stock.InvalidateSummary(eventID)
	if s.broadcast != nil {
		if view, err := s.toWarehouseView(&order); err == nil {
			s.broadcast(eventID, *view)
		}
	}
	return s.toConfirmation(&order, false)
}

// ---- warehouse board ----

func (s *Service) WarehouseOrders(eventID uuid.UUID, status *OrderStatus) ([]WarehouseOrderView, error) {
	q := s.db.Preload("Items").Where("event_id = ?", eventID)
	if status != nil {
		q = q.Where("status = ?", *status)
	}
	var orders []Order
	if err := q.Order("created_at DESC").Find(&orders).Error; err != nil {
		return nil, err
	}
	views := make([]WarehouseOrderView, 0, len(orders))
	for i := range orders {
		v, err := s.toWarehouseView(&orders[i])
		if err != nil {
			return nil, err
		}
		views = append(views, *v)
	}
	return views, nil
}

// AdvanceStatus moves an order one step forward, auditing + broadcasting the change. The order row is
// locked so a concurrent cancel can't be overtaken by a stale advance.
func (s *Service) AdvanceStatus(orderID uuid.UUID, target OrderStatus, actorUserID *uuid.UUID) (*WarehouseOrderView, error) {
	var o Order
	err := s.db.Transaction(func(tx *gorm.DB) error {
		if err := lockOrder(tx, orderID, &o); err != nil {
			return err
		}
		if !o.Status.CanAdvanceTo(target) {
			return common.NewAppError(http.StatusConflict, "Cannot move order from "+string(o.Status)+" to "+string(target))
		}
		from := o.Status
		if err := tx.Model(&o).Update("status", target).Error; err != nil {
			return err
		}
		o.Status = target
		return s.audit.RecordStatusChange(tx, o.EventID, o.ID, o.ReferenceID, string(from), string(target), actorUserID)
	})
	if err != nil {
		return nil, err
	}
	return s.publish(&o)
}

// Cancel cancels an order and returns its stock, in one transaction with the audit entry. A BA may
// cancel only their own order while it is NEW (others' orders are a 404); warehouse and admin may
// cancel any order that is NEW or PACKING. Anything else is a 409.
func (s *Service) Cancel(orderID, actorID uuid.UUID, actorRole user.Role, reason *string) (*WarehouseOrderView, error) {
	var o Order
	err := s.db.Transaction(func(tx *gorm.DB) error {
		if err := lockOrder(tx, orderID, &o); err != nil {
			return err
		}
		isBa := actorRole == user.RoleBA
		if isBa && o.BaUserID != actorID {
			return common.NewAppError(http.StatusNotFound, "Order not found")
		}
		if !o.Status.CanCancel() || (isBa && o.Status != StatusNew) {
			return common.NewAppError(http.StatusConflict, "Order "+o.ReferenceID+" is "+string(o.Status)+" and can no longer be cancelled")
		}
		for _, it := range o.Items {
			if err := s.stock.Restore(tx, o.EventID, it.ProductID, it.Qty); err != nil {
				return err
			}
		}
		from := o.Status
		if err := tx.Model(&o).Update("status", StatusCancelled).Error; err != nil {
			return err
		}
		o.Status = StatusCancelled
		return s.audit.RecordCancelled(tx, o.EventID, o.ID, o.ReferenceID, string(from), &actorID, reason)
	})
	if err != nil {
		return nil, err
	}
	s.stock.InvalidateSummary(o.EventID)
	return s.publish(&o)
}

// lockOrder loads an order with its items, holding a row lock until the transaction ends.
func lockOrder(tx *gorm.DB, orderID uuid.UUID, o *Order) error {
	err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).First(o, "id = ?", orderID).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return common.NewAppError(http.StatusNotFound, "Order not found")
	}
	if err != nil {
		return err
	}
	return tx.Where("order_id = ?", orderID).Find(&o.Items).Error
}

// publish pushes the order's current board view to the warehouse WebSocket after commit.
func (s *Service) publish(o *Order) (*WarehouseOrderView, error) {
	view, err := s.toWarehouseView(o)
	if err != nil {
		return nil, err
	}
	if s.broadcast != nil {
		s.broadcast(o.EventID, *view)
	}
	return view, nil
}

// ---- BA history / detail ----

func (s *Service) History(eventID, baUserID uuid.UUID) ([]OrderSummary, error) {
	var orders []Order
	if err := s.db.Preload("Items").
		Where("event_id = ? AND ba_user_id = ?", eventID, baUserID).
		Order("created_at DESC").Find(&orders).Error; err != nil {
		return nil, err
	}
	out := make([]OrderSummary, 0, len(orders))
	for i := range orders {
		o := &orders[i]
		count := 0
		for _, it := range o.Items {
			count += it.Qty
		}
		out = append(out, OrderSummary{
			ID: o.ID, OrderNumberDisplay: fmt.Sprintf("#%03d", o.OrderNumber), ReferenceID: o.ReferenceID,
			Status: o.Status, CustomerName: o.CustomerName, Total: o.Total, ItemCount: count, CreatedAt: o.CreatedAt,
		})
	}
	return out, nil
}

func (s *Service) Get(orderID uuid.UUID) (*OrderConfirmation, error) {
	var o Order
	if err := s.db.Preload("Items").First(&o, "id = ?", orderID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, common.NewAppError(http.StatusNotFound, "Order not found")
		}
		return nil, err
	}
	return s.toConfirmation(&o, false)
}

// ---- helpers ----

func (s *Service) findByIdempotencyKey(key string) (*Order, error) {
	var o Order
	err := s.db.Preload("Items").Where("idempotency_key = ?", key).First(&o).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &o, nil
}

func (s *Service) nextOrderNumber(tx *gorm.DB, eventID uuid.UUID, dateKey string) (int, error) {
	// Ensure the row exists without aborting the tx on a concurrent create.
	if err := tx.Clauses(clause.OnConflict{DoNothing: true}).
		Create(&OrderDailySequence{EventID: eventID, DateKey: dateKey, LastNumber: 0}).Error; err != nil {
		return 0, err
	}
	var seq OrderDailySequence
	if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
		Where("event_id = ? AND date_key = ?", eventID, dateKey).First(&seq).Error; err != nil {
		return 0, err
	}
	seq.LastNumber++
	if err := tx.Save(&seq).Error; err != nil {
		return 0, err
	}
	return seq.LastNumber, nil
}

func (s *Service) toWarehouseView(o *Order) (*WarehouseOrderView, error) {
	baName := "BA"
	if u, err := s.users.FindByID(o.BaUserID); err == nil && u != nil {
		baName = u.FullName
	}
	items := make([]WarehouseItem, 0, len(o.Items))
	for _, it := range o.Items {
		items = append(items, WarehouseItem{Name: it.Name, SKU: it.SKU, Qty: it.Qty})
	}
	return &WarehouseOrderView{
		ID: o.ID, OrderNumber: o.OrderNumber, OrderNumberDisplay: fmt.Sprintf("#%03d", o.OrderNumber),
		ReferenceID: o.ReferenceID, CustomerName: o.CustomerName, BaName: baName, Items: items,
		Status: o.Status, Total: o.Total, CreatedAt: o.CreatedAt,
	}, nil
}

func (s *Service) toConfirmation(o *Order, duplicate bool) (*OrderConfirmation, error) {
	code, name := "", ""
	var ev event.Event
	if err := s.db.First(&ev, "id = ?", o.EventID).Error; err == nil {
		code, name = ev.Code, ev.Name
	}
	items := make([]OrderItemView, 0, len(o.Items))
	for _, it := range o.Items {
		items = append(items, OrderItemView{
			ProductID: it.ProductID, SKU: it.SKU, Name: it.Name, UnitPrice: it.UnitPrice,
			Qty: it.Qty, LineTotal: it.LineTotal, VariantName: it.VariantName,
		})
	}
	return &OrderConfirmation{
		ID: o.ID, OrderNumber: o.OrderNumber, OrderNumberDisplay: fmt.Sprintf("#%03d", o.OrderNumber),
		ReferenceID: o.ReferenceID, Status: o.Status, EventCode: code, EventName: name,
		CustomerName: o.CustomerName, CustomerTier: o.CustomerTier, Items: items,
		Subtotal: o.Subtotal, Discount: o.Discount, Total: o.Total, PromotionName: o.PromotionName,
		CreatedAt: o.CreatedAt, CashierNote: "Please proceed to the cashier to complete payment.",
		Duplicate: duplicate,
	}, nil
}
