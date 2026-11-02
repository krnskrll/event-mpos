package service

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

	auditiface "github.com/krnskrll/event-mpos/internal/audit/iface"
	eventiface "github.com/krnskrll/event-mpos/internal/event/iface"
	"github.com/krnskrll/event-mpos/internal/order/iface"
	"github.com/krnskrll/event-mpos/internal/order/repository"
	"github.com/krnskrll/event-mpos/internal/platform/common"
	"github.com/krnskrll/event-mpos/internal/platform/database"
	productiface "github.com/krnskrll/event-mpos/internal/product/iface"
	promotioniface "github.com/krnskrll/event-mpos/internal/promotion/iface"
	stockiface "github.com/krnskrll/event-mpos/internal/stock/iface"
	useriface "github.com/krnskrll/event-mpos/internal/user/iface"
)

var _ iface.Orders = (*Service)(nil)

type Service struct {
	db         *gorm.DB
	events     eventiface.Events
	products   productiface.Products
	stock      stockiface.Stock
	promotions promotioniface.Promotions
	audit      auditiface.Audit
	users      useriface.Users
	jakarta    *time.Location
	// broadcast pushes an order change to the warehouse board after commit; nil is a no-op.
	broadcast func(eventID uuid.UUID, view iface.WarehouseOrderView)
}

func NewService(db *gorm.DB, events eventiface.Events, products productiface.Products, stk stockiface.Stock, promotions promotioniface.Promotions, aud auditiface.Audit, users useriface.Users) *Service {
	loc, err := time.LoadLocation("Asia/Jakarta")
	if err != nil {
		loc = time.FixedZone("WIB", 7*3600)
	}
	return &Service{db: db, events: events, products: products, stock: stk, promotions: promotions, audit: aud, users: users, jakarta: loc}
}

func (s *Service) SetBroadcaster(fn func(eventID uuid.UUID, view iface.WarehouseOrderView)) {
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
func (s *Service) Submit(eventID, baUserID uuid.UUID, req iface.SubmitOrderRequest) (*iface.OrderConfirmation, error) {
	// 1) idempotency fast-path
	if existing, err := s.findByIdempotencyKey(req.IdempotencyKey); err != nil {
		return nil, err
	} else if existing != nil {
		return s.toConfirmation(existing, true)
	}

	// event must be ACTIVE
	ev, err := s.events.Get(eventID)
	if err != nil {
		return nil, err
	}
	lateSynced, err := acceptsOrders(ev, req.ClientCreatedAt)
	if err != nil {
		return nil, err
	}

	// 2) build line items from the catalogue (server-authoritative prices)
	var lines []iface.OrderLineRequest
	for _, l := range req.Lines {
		if l.Qty > 0 {
			lines = append(lines, l)
		}
	}
	if len(lines) == 0 {
		return nil, common.NewAppError(http.StatusBadRequest, "Cart is empty")
	}

	cartLines := make([]promotioniface.CartLine, 0, len(lines))
	drafts := make([]draftItem, 0, len(lines))
	for _, line := range lines {
		p, err := s.products.Find(line.ProductID)
		if err != nil {
			return nil, err
		}
		if p == nil {
			return nil, common.NewAppError(http.StatusBadRequest, "Unknown product "+line.ProductID.String())
		}
		unitPrice := p.Price
		var variantID *uuid.UUID
		var variantName *string
		if line.VariantID != nil {
			var found *productiface.SaleVariant
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
			unitPrice = found.Price
			variantID = &found.ID
			name := found.Name
			variantName = &name
		}
		lineTotal := unitPrice.Mul(decimal.NewFromInt(int64(line.Qty)))
		cartLines = append(cartLines, promotioniface.CartLine{ProductID: p.ID, UnitPrice: unitPrice, Qty: line.Qty})
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
		promoID = &applied.ID
		promoName = &applied.Name
	}
	total := subtotal.Sub(discount)
	if total.IsNegative() {
		total = decimal.Zero
	}
	total = total.Round(2)

	dateKey := time.Now().In(s.jakarta).Format("20060102")

	order := repository.Order{
		EventID: eventID, BaUserID: baUserID, IdempotencyKey: req.IdempotencyKey,
		ReferenceID: "", CustomerName: strings.TrimSpace(req.Customer.Name), CustomerQr: req.Customer.QRCode,
		CustomerTier: req.Customer.Tier, Subtotal: subtotal, Discount: discount, Total: total,
		PromotionID: promoID, PromotionName: promoName, Status: iface.StatusNew,
		DeviceID: req.DeviceID, ClientCreatedAt: req.ClientCreatedAt, LateSynced: lateSynced,
	}

	// 3+5+6) atomic tx: stock decrement (row locks) → daily number → persist + audit
	// Lock allocation rows in a fixed (product id) order so two multi-line orders can't each hold
	// one row and wait on the other.
	lockOrder := append([]iface.OrderLineRequest(nil), lines...)
	sort.Slice(lockOrder, func(i, j int) bool { return lockOrder[i].ProductID.String() < lockOrder[j].ProductID.String() })
	txErr := s.db.Transaction(func(tx *gorm.DB) error {
		for _, line := range lockOrder {
			if req.OrderNumber != nil { // already confirmed on the device: never reject, flag oversells
				oversold, err := s.stock.Reserve(tx, eventID, line.ProductID, line.Qty)
				if err != nil {
					return err
				}
				order.StockConflict = order.StockConflict || oversold
			} else if err := s.stock.Decrement(tx, eventID, line.ProductID, line.Qty); err != nil {
				return err
			}
		}
		number, businessDate, err := s.orderNumber(tx, eventID, dateKey, req)
		if err != nil {
			return err
		}
		order.OrderNumber = number
		order.ReferenceID = fmt.Sprintf("ORD-%s-%s-%03d", businessDate, ev.Code, number)
		for _, d := range drafts {
			order.Items = append(order.Items, repository.OrderItem{
				ProductID: d.productID, VariantID: d.variantID, VariantName: d.variantName,
				SKU: d.sku, Name: d.name, UnitPrice: d.unitPrice, Qty: d.qty, LineTotal: d.lineTotal,
			})
		}
		if err := tx.Create(&order).Error; err != nil {
			return err
		}
		return s.audit.RecordOrderCreated(tx, eventID, order.ID, order.ReferenceID, &baUserID, createdDetail(&order))
	})
	if txErr != nil {
		// concurrent submit with the same idempotency key won the race — return its order
		if errors.Is(txErr, gorm.ErrDuplicatedKey) {
			if existing, e := s.findByIdempotencyKey(req.IdempotencyKey); e == nil && existing != nil {
				return s.toConfirmation(existing, true)
			}
			// Otherwise the reference (event + date + number) is taken: a device reused a number.
			return nil, common.NewAppError(http.StatusConflict, "Order number already used")
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

func (s *Service) WarehouseOrders(eventID uuid.UUID, status *iface.OrderStatus) ([]iface.WarehouseOrderView, error) {
	q := s.db.Preload("Items").Where("event_id = ?", eventID)
	if status != nil {
		q = q.Where("status = ?", *status)
	}
	var orders []repository.Order
	if err := q.Order("created_at DESC").Find(&orders).Error; err != nil {
		return nil, err
	}
	views := make([]iface.WarehouseOrderView, 0, len(orders))
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
func (s *Service) AdvanceStatus(orderID uuid.UUID, target iface.OrderStatus, actorUserID *uuid.UUID) (*iface.WarehouseOrderView, error) {
	var o repository.Order
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
func (s *Service) Cancel(orderID, actorID uuid.UUID, actorRole useriface.Role, reason *string) (*iface.WarehouseOrderView, error) {
	var o repository.Order
	err := s.db.Transaction(func(tx *gorm.DB) error {
		if err := lockOrder(tx, orderID, &o); err != nil {
			return err
		}
		isBa := actorRole == useriface.RoleBA
		if isBa && o.BaUserID != actorID {
			return common.NewAppError(http.StatusNotFound, "Order not found")
		}
		if !o.Status.CanCancel() || (isBa && o.Status != iface.StatusNew) {
			return common.NewAppError(http.StatusConflict, "Order "+o.ReferenceID+" is "+string(o.Status)+" and can no longer be cancelled")
		}
		for _, it := range o.Items {
			if err := s.stock.Restore(tx, o.EventID, it.ProductID, it.Qty); err != nil {
				return err
			}
		}
		from := o.Status
		if err := tx.Model(&o).Update("status", iface.StatusCancelled).Error; err != nil {
			return err
		}
		o.Status = iface.StatusCancelled
		return s.audit.RecordCancelled(tx, o.EventID, o.ID, o.ReferenceID, string(from), &actorID, reason)
	})
	if err != nil {
		return nil, err
	}
	s.stock.InvalidateSummary(o.EventID)
	return s.publish(&o)
}

// lockOrder loads an order with its items, holding a row lock until the transaction ends.
func lockOrder(tx *gorm.DB, orderID uuid.UUID, o *repository.Order) error {
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
func (s *Service) publish(o *repository.Order) (*iface.WarehouseOrderView, error) {
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

func (s *Service) History(eventID, baUserID uuid.UUID) ([]iface.OrderSummary, error) {
	var orders []repository.Order
	if err := s.db.Preload("Items").
		Where("event_id = ? AND ba_user_id = ?", eventID, baUserID).
		Order("created_at DESC").Find(&orders).Error; err != nil {
		return nil, err
	}
	out := make([]iface.OrderSummary, 0, len(orders))
	for i := range orders {
		o := &orders[i]
		count := 0
		for _, it := range o.Items {
			count += it.Qty
		}
		out = append(out, iface.OrderSummary{
			ID: o.ID, OrderNumberDisplay: fmt.Sprintf("#%03d", o.OrderNumber), ReferenceID: o.ReferenceID,
			Status: o.Status, CustomerName: o.CustomerName, Total: o.Total, ItemCount: count, CreatedAt: o.CreatedAt,
		})
	}
	return out, nil
}

func (s *Service) Get(orderID uuid.UUID) (*iface.OrderConfirmation, error) {
	var o repository.Order
	if err := s.db.Preload("Items").First(&o, "id = ?", orderID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, common.NewAppError(http.StatusNotFound, "Order not found")
		}
		return nil, err
	}
	return s.toConfirmation(&o, false)
}

// ---- helpers ----

func (s *Service) findByIdempotencyKey(key string) (*repository.Order, error) {
	var o repository.Order
	err := s.db.Preload("Items").Where("idempotency_key = ?", key).First(&o).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &o, nil
}

// createdDetail is the ORDER_CREATED audit text; it carries the late-sync / stock-conflict markers the
// report's audit trail shows per order.
func createdDetail(o *repository.Order) string {
	d := "Order created"
	if o.LateSynced {
		d += " · late-synced after event close"
	}
	if o.StockConflict {
		d += " · stock conflict"
	}
	return d
}

// lateSyncGrace is how long after closing an event still takes orders confirmed before the close (D2).
const lateSyncGrace = 24 * time.Hour

// acceptsOrders applies decision D2: an ACTIVE event takes orders; a CLOSED one only takes orders the
// device created before closed_at and uploads within 24h of closing — those are marked late-synced.
func acceptsOrders(ev *eventiface.EventResponse, clientCreatedAt *time.Time) (lateSynced bool, err error) {
	if ev.Status == eventiface.StatusActive {
		return false, nil
	}
	if ev.Status == eventiface.StatusClosed && ev.ClosedAt != nil {
		if clientCreatedAt != nil && clientCreatedAt.Before(*ev.ClosedAt) && time.Since(*ev.ClosedAt) <= lateSyncGrace {
			return true, nil
		}
		return false, common.NewAppError(http.StatusConflict, "Event "+ev.Code+" closed at "+ev.ClosedAt.Format(time.RFC3339)+
			"; only orders created before closing and uploaded within 24 hours are accepted")
	}
	return false, common.NewAppError(http.StatusConflict, "Event "+ev.Code+" is "+string(ev.Status)+" — cannot accept orders")
}

// orderNumber returns the order's daily number and business date: the device's own number when it
// confirmed the order itself (it must lie in one of the device's blocks), else the next server number.
func (s *Service) orderNumber(tx *gorm.DB, eventID uuid.UUID, today string, req iface.SubmitOrderRequest) (int, string, error) {
	if req.OrderNumber == nil {
		n, err := s.nextOrderNumber(tx, eventID, today)
		return n, today, err
	}
	n, deviceID := *req.OrderNumber, ""
	if req.DeviceID != nil {
		deviceID = *req.DeviceID
	}
	var b repository.NumberBlock
	err := tx.Where("device_id = ? AND event_id = ? AND range_start <= ? AND range_end >= ?", deviceID, eventID, n, n).Take(&b).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return 0, "", common.NewAppError(http.StatusBadRequest, fmt.Sprintf("Order number %d is not in a block issued to this device", n))
	}
	if err != nil {
		return 0, "", err
	}
	err = tx.Model(&b).Update("next_unused", gorm.Expr("GREATEST(next_unused, ?)", n+1)).Error
	return n, b.BusinessDate, err
}

// IssueNumberBlock reserves the next NumberBlockSize numbers of today's sequence for a device. The
// range comes off the same locked sequence row as server-assigned numbers, so they never collide.
func (s *Service) IssueNumberBlock(eventID uuid.UUID, deviceID string) (*iface.NumberBlockResponse, error) {
	ev, err := s.events.Get(eventID)
	if err != nil {
		return nil, common.NewAppError(http.StatusNotFound, "Event not found")
	}
	if ev.Status != eventiface.StatusActive {
		return nil, common.NewAppError(http.StatusConflict, "Event "+ev.Code+" is "+string(ev.Status)+" — no new order numbers")
	}
	today := time.Now().In(s.jakarta).Format("20060102")
	var b repository.NumberBlock
	err = s.db.Transaction(func(tx *gorm.DB) error {
		end, err := s.advanceSequence(tx, eventID, today, iface.NumberBlockSize)
		if err != nil {
			return err
		}
		start := end - iface.NumberBlockSize + 1
		b = repository.NumberBlock{DeviceID: deviceID, EventID: eventID, BusinessDate: today, RangeStart: start, RangeEnd: end, NextUnused: start, CreatedAt: time.Now().UTC()}
		return tx.Create(&b).Error
	})
	if err != nil {
		return nil, err
	}
	return &iface.NumberBlockResponse{EventID: eventID, BusinessDate: today, RangeStart: b.RangeStart, RangeEnd: b.RangeEnd}, nil
}

func (s *Service) nextOrderNumber(tx *gorm.DB, eventID uuid.UUID, dateKey string) (int, error) {
	return s.advanceSequence(tx, eventID, dateKey, 1)
}

// advanceSequence moves the event's daily sequence forward by n under its row lock and returns the
// new last number.
func (s *Service) advanceSequence(tx *gorm.DB, eventID uuid.UUID, dateKey string, n int) (int, error) {
	// Ensure the row exists without aborting the tx on a concurrent create.
	if err := tx.Clauses(clause.OnConflict{DoNothing: true}).
		Create(&repository.OrderDailySequence{EventID: eventID, DateKey: dateKey, LastNumber: 0}).Error; err != nil {
		return 0, err
	}
	var seq repository.OrderDailySequence
	if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
		Where("event_id = ? AND date_key = ?", eventID, dateKey).First(&seq).Error; err != nil {
		return 0, err
	}
	seq.LastNumber += n
	if err := tx.Save(&seq).Error; err != nil {
		return 0, err
	}
	return seq.LastNumber, nil
}

func (s *Service) toWarehouseView(o *repository.Order) (*iface.WarehouseOrderView, error) {
	baName := "BA"
	if u, err := s.users.FindByID(o.BaUserID); err == nil && u != nil {
		baName = u.FullName
	}
	items := make([]iface.WarehouseItem, 0, len(o.Items))
	for _, it := range o.Items {
		items = append(items, iface.WarehouseItem{Name: it.Name, SKU: it.SKU, Qty: it.Qty})
	}
	return &iface.WarehouseOrderView{
		ID: o.ID, OrderNumber: o.OrderNumber, OrderNumberDisplay: fmt.Sprintf("#%03d", o.OrderNumber),
		ReferenceID: o.ReferenceID, CustomerName: o.CustomerName, BaName: baName, Items: items,
		Status: o.Status, Total: o.Total, CreatedAt: o.CreatedAt, StockConflict: o.StockConflict, LateSynced: o.LateSynced,
	}, nil
}

func (s *Service) toConfirmation(o *repository.Order, duplicate bool) (*iface.OrderConfirmation, error) {
	code, name := "", ""
	if ev, err := s.events.Get(o.EventID); err == nil {
		code, name = ev.Code, ev.Name
	}
	items := make([]iface.OrderItemView, 0, len(o.Items))
	for _, it := range o.Items {
		items = append(items, iface.OrderItemView{
			ProductID: it.ProductID, SKU: it.SKU, Name: it.Name, UnitPrice: it.UnitPrice,
			Qty: it.Qty, LineTotal: it.LineTotal, VariantName: it.VariantName,
		})
	}
	return &iface.OrderConfirmation{
		ID: o.ID, OrderNumber: o.OrderNumber, OrderNumberDisplay: fmt.Sprintf("#%03d", o.OrderNumber),
		ReferenceID: o.ReferenceID, Status: o.Status, EventCode: code, EventName: name,
		CustomerName: o.CustomerName, CustomerTier: o.CustomerTier, Items: items,
		Subtotal: o.Subtotal, Discount: o.Discount, Total: o.Total, PromotionName: o.PromotionName,
		CreatedAt: o.CreatedAt, CashierNote: "Please proceed to the cashier to complete payment.",
		Duplicate: duplicate, StockConflict: o.StockConflict, LateSynced: o.LateSynced,
	}, nil
}

func (s *Service) ForReceipt(tx database.Tx, orderID, baUserID uuid.UUID) (*iface.ReceiptOrder, error) {
	var o repository.Order
	err := tx.Select("id", "event_id", "reference_id", "ba_user_id").
		Where("id = ? AND ba_user_id = ?", orderID, baUserID).Take(&o).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &iface.ReceiptOrder{ID: o.ID, EventID: o.EventID, ReferenceID: o.ReferenceID}, nil
}

func (s *Service) ForReport(eventID uuid.UUID) ([]iface.SoldOrder, int, error) {
	var orders []repository.Order
	if err := s.db.Preload("Items").Where("event_id = ? AND status <> ?", eventID, iface.StatusCancelled).
		Order("created_at DESC").Find(&orders).Error; err != nil {
		return nil, 0, err
	}
	var cancelled int64
	if err := s.db.Model(&repository.Order{}).Where("event_id = ? AND status = ?", eventID, iface.StatusCancelled).
		Count(&cancelled).Error; err != nil {
		return nil, 0, err
	}
	out := make([]iface.SoldOrder, 0, len(orders))
	for _, o := range orders {
		items := make([]iface.SoldItem, 0, len(o.Items))
		for _, it := range o.Items {
			items = append(items, iface.SoldItem{ProductID: it.ProductID, SKU: it.SKU, Name: it.Name, Qty: it.Qty, LineTotal: it.LineTotal})
		}
		out = append(out, iface.SoldOrder{Total: o.Total, Discount: o.Discount, LateSynced: o.LateSynced, StockConflict: o.StockConflict, Items: items})
	}
	return out, int(cancelled), nil
}
