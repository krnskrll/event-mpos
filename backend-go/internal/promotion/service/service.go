package service

import (
	"errors"
	"net/http"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	"gorm.io/gorm"

	eventiface "github.com/krnskrll/event-mpos/internal/event/iface"
	"github.com/krnskrll/event-mpos/internal/platform/common"
	productiface "github.com/krnskrll/event-mpos/internal/product/iface"
	"github.com/krnskrll/event-mpos/internal/promotion/iface"
	"github.com/krnskrll/event-mpos/internal/promotion/repository"
)

// Service holds promotion CRUD + the authoritative quote — the Go analogue of Kotlin PromotionService.
var _ iface.Promotions = (*Service)(nil)

type Service struct {
	db       *gorm.DB
	products productiface.Products
	events   eventiface.Events
}

func NewService(db *gorm.DB, products productiface.Products, events eventiface.Events) *Service {
	return &Service{db: db, products: products, events: events}
}

// List is the store's promotions; with a location, only the ones that run there.
func (s *Service) List(storeID uuid.UUID, locationID *uuid.UUID) ([]iface.PromotionResponse, error) {
	q := s.db.Preload("Items").Where("store_id = ?", storeID)
	if locationID != nil {
		q = q.Where("event_id = ? OR event_id IS NULL", *locationID)
	}
	var promos []repository.Promotion
	if err := q.Order("active DESC, priority DESC, name").Find(&promos).Error; err != nil {
		return nil, err
	}
	out := make([]iface.PromotionResponse, 0, len(promos))
	for i := range promos {
		out = append(out, promos[i].ToResponse())
	}
	return out, nil
}

func (s *Service) Get(storeID, id uuid.UUID) (*iface.PromotionResponse, error) {
	p, err := s.find(storeID, id)
	if err != nil {
		return nil, err
	}
	resp := p.ToResponse()
	return &resp, nil
}

func (s *Service) find(storeID, id uuid.UUID) (*repository.Promotion, error) {
	var p repository.Promotion
	err := s.db.Preload("Items").First(&p, "id = ? AND store_id = ?", id, storeID).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, common.NewAppError(http.StatusNotFound, "Promotion not found")
	}
	return &p, err
}

// atLocation is what runs at a location: its own and the store-wide promotions, with each
// category expanded into its items.
func (s *Service) atLocation(eventID uuid.UUID, activeOnly bool) ([]repository.Promotion, error) {
	ev, err := s.events.Get(eventID)
	if err != nil {
		return nil, err
	}
	q := s.db.Preload("Items").Where("store_id = ? AND (event_id = ? OR event_id IS NULL)", ev.StoreID, eventID)
	if activeOnly {
		q = q.Where("active")
	}
	var promos []repository.Promotion
	if err := q.Find(&promos).Error; err != nil {
		return nil, err
	}
	var cats map[uuid.UUID][]uuid.UUID
	for i := range promos {
		p := &promos[i]
		if p.CategoryID == nil {
			p.SetProductIDs(p.ItemIDs())
			continue
		}
		if cats == nil {
			if cats, err = s.categories(ev.StoreID); err != nil {
				return nil, err
			}
		}
		p.SetProductIDs(cats[*p.CategoryID])
	}
	return promos, nil
}

func (s *Service) categories(storeID uuid.UUID) (map[uuid.UUID][]uuid.UUID, error) {
	cs, err := s.products.Categories(storeID)
	out := make(map[uuid.UUID][]uuid.UUID, len(cs))
	for _, c := range cs {
		out[c.ID] = c.ItemIDs
	}
	return out, err
}

// BestFor evaluates the active promotions for an event against a cart and returns the single best
// applied promotion (or nil). Shared by the quote endpoint and order submission.
func (s *Service) BestFor(eventID uuid.UUID, cart []iface.CartLine, now time.Time) (*iface.Applied, error) {
	promos, err := s.atLocation(eventID, true)
	if err != nil {
		return nil, err
	}
	best := BestPromotion(cart, promos, now)
	if best == nil {
		return nil, nil
	}
	return &iface.Applied{ID: best.Promotion.ID, Name: best.Promotion.Name, Discount: best.Discount}, nil
}

// SnapshotPromotions returns active promotions as compact snapshot DTOs.
func (s *Service) SnapshotPromotions(eventID uuid.UUID) ([]any, error) {
	promos, err := s.atLocation(eventID, true)
	if err != nil {
		return nil, err
	}
	out := make([]any, 0, len(promos))
	for i := range promos {
		out = append(out, promos[i].ToSnapshotDto())
	}
	return out, nil
}

// Fingerprint is the promotion config fingerprint folded into the (opaque, backend-relative)
// snapshot version. Deterministic given the data, category membership included.
func (s *Service) Fingerprint(eventID uuid.UUID) (string, error) {
	promos, err := s.atLocation(eventID, false)
	if err != nil {
		return "", err
	}
	sort.Slice(promos, func(i, j int) bool { return promos[i].ID.String() < promos[j].ID.String() })
	parts := make([]string, 0, len(promos))
	for i := range promos {
		p := &promos[i]
		ids := make([]string, 0, len(p.ProductIDs()))
		for _, id := range p.ProductIDs() {
			ids = append(ids, id.String())
		}
		sort.Strings(ids)
		parts = append(parts, strings.Join([]string{
			p.ID.String(), p.Name, string(p.Type), string(p.Scope), strings.Join(ids, "|"),
			decStr(p.PercentOff), decStr(p.AmountOff), intStr(p.MinQty), decStr(p.MinSubtotal),
			strconv.FormatBool(p.Active), timeStr(p.StartsAt), timeStr(p.EndsAt), strconv.Itoa(p.Priority),
		}, ":"))
	}
	return strings.Join(parts, ","), nil
}

func (s *Service) Create(storeID uuid.UUID, req iface.PromotionRequest) (*iface.PromotionResponse, error) {
	return s.save(&repository.Promotion{StoreID: storeID}, req)
}

func (s *Service) Update(storeID, id uuid.UUID, req iface.PromotionRequest) (*iface.PromotionResponse, error) {
	promo, err := s.find(storeID, id)
	if err != nil {
		return nil, err
	}
	return s.save(promo, req)
}

func (s *Service) save(promo *repository.Promotion, req iface.PromotionRequest) (*iface.PromotionResponse, error) {
	if req.Scope == nil || *req.Scope == iface.ScopeAll {
		req.CategoryID, req.ProductIDs = nil, nil
	}
	if err := s.validate(promo.StoreID, req); err != nil {
		return nil, err
	}
	applyReq(promo, req)
	err := s.db.Transaction(func(tx *gorm.DB) error {
		items := promo.Items
		if err := tx.Omit("Items").Save(promo).Error; err != nil {
			return err
		}
		if err := tx.Where("promotion_id = ?", promo.ID).Delete(&repository.PromotionItem{}).Error; err != nil {
			return err
		}
		for i := range items {
			items[i].PromotionID = promo.ID
		}
		if len(items) == 0 {
			return nil
		}
		return tx.Create(&items).Error
	})
	if err != nil {
		return nil, err
	}
	resp := promo.ToResponse()
	return &resp, nil
}

func (s *Service) Delete(storeID, id uuid.UUID) error {
	if _, err := s.find(storeID, id); err != nil {
		return err
	}
	return s.db.Delete(&repository.Promotion{}, "id = ?", id).Error
}

// Quote is the authoritative cart quote: DB prices + the shared engine picks the single best promo.
func (s *Service) Quote(eventID uuid.UUID, lines []iface.QuoteLine) (*iface.QuoteResponse, error) {
	cart := make([]iface.CartLine, 0, len(lines))
	for _, line := range lines {
		if line.Qty <= 0 {
			continue
		}
		p, err := s.products.FindAt(eventID, line.ProductID)
		if err != nil {
			return nil, err
		}
		if p == nil {
			continue
		}
		cart = append(cart, iface.CartLine{ProductID: p.ID, UnitPrice: p.Price, Qty: line.Qty})
	}
	subtotal := decimal.Zero
	for _, l := range cart {
		subtotal = subtotal.Add(l.UnitPrice.Mul(decimal.NewFromInt(int64(l.Qty))))
	}
	subtotal = common.Rupiah(subtotal)

	promos, err := s.atLocation(eventID, true)
	if err != nil {
		return nil, err
	}
	applied := BestPromotion(cart, promos, time.Now())
	discount := decimal.Zero
	resp := iface.QuoteResponse{Subtotal: subtotal}
	if applied != nil {
		discount = applied.Discount
		id := applied.Promotion.ID
		name := applied.Promotion.Name
		resp.AppliedPromotionID = &id
		resp.AppliedPromotionName = &name
	}
	resp.Discount = discount
	total := subtotal.Sub(discount)
	if total.IsNegative() {
		total = decimal.Zero
	}
	resp.Total = common.Rupiah(total)
	return &resp, nil
}

// ---- validation + apply ----

func (s *Service) validate(storeID uuid.UUID, req iface.PromotionRequest) error {
	if err := validate(req); err != nil {
		return err
	}
	if req.LocationID != nil {
		if err := s.events.RequireInStore(storeID, *req.LocationID); err != nil {
			return err
		}
	}
	if req.CategoryID != nil {
		cats, err := s.categories(storeID)
		if err != nil {
			return err
		}
		if _, ok := cats[*req.CategoryID]; !ok {
			return common.NewAppError(http.StatusBadRequest, "Unknown category")
		}
	}
	return s.products.RequireItems(storeID, req.ProductIDs)
}

func validate(req iface.PromotionRequest) error {
	pos := func(d *decimal.Decimal) bool { return d != nil && d.GreaterThan(decimal.Zero) }
	bad := func(msg string) error { return common.NewAppError(http.StatusBadRequest, msg) }
	if strings.TrimSpace(req.Name) == "" {
		return bad("name required")
	}
	if err := common.CheckRupiah("amountOff", req.AmountOff); err != nil {
		return err
	}
	if err := common.CheckRupiah("minSubtotal", req.MinSubtotal); err != nil {
		return err
	}
	scope := iface.ScopeAll
	if req.Scope != nil {
		scope = *req.Scope
	}
	switch req.Type {
	case iface.TypePercentage:
		if !pos(req.PercentOff) || req.PercentOff.GreaterThan(decimal.NewFromInt(100)) {
			return bad("percentOff must be above 0 and at most 100")
		}
	case iface.TypeFixed:
		if !pos(req.AmountOff) {
			return bad("amountOff required for a fixed promotion")
		}
	case iface.TypeBundle:
		if !pos(req.AmountOff) {
			return bad("amountOff required for a bundle promotion")
		}
		if req.MinQty == nil || *req.MinQty < 1 {
			return bad("minQty (bundle size) required for a bundle promotion")
		}
		if scope != iface.ScopeSpecific {
			return bad("a bundle promotion needs a category or chosen products")
		}
	default:
		return bad("type must be one of PERCENTAGE, FIXED, BUNDLE")
	}
	if req.StartsAt != nil && req.EndsAt != nil && !req.EndsAt.After(*req.StartsAt) {
		return bad("endsAt must be after startsAt")
	}
	byCategory, byItems := req.CategoryID != nil, len(req.ProductIDs) > 0
	switch {
	case scope == iface.ScopeSpecific && byCategory == byItems:
		return bad("SPECIFIC scope needs either a category or at least one product")
	case scope != iface.ScopeAll && scope != iface.ScopeSpecific:
		return bad("scope must be ALL or SPECIFIC")
	}
	return nil
}

func applyReq(promo *repository.Promotion, req iface.PromotionRequest) {
	promo.Name = strings.TrimSpace(req.Name)
	promo.Type = req.Type
	promo.EventID = req.LocationID
	promo.CategoryID = req.CategoryID
	promo.PercentOff = req.PercentOff
	promo.AmountOff = req.AmountOff
	promo.Scope = iface.ScopeAll
	if req.Scope != nil {
		promo.Scope = *req.Scope
	}
	promo.Items = nil
	seen := map[uuid.UUID]bool{}
	for _, id := range req.ProductIDs {
		if !seen[id] {
			seen[id] = true
			promo.Items = append(promo.Items, repository.PromotionItem{ItemID: id})
		}
	}
	promo.MinQty = req.MinQty
	promo.MinSubtotal = req.MinSubtotal
	promo.StartsAt = req.StartsAt
	promo.EndsAt = req.EndsAt
	promo.Active = true
	if req.Active != nil {
		promo.Active = *req.Active
	}
	promo.Priority = 0
	if req.Priority != nil {
		promo.Priority = *req.Priority
	}
	promo.UpdatedAt = time.Now().UTC()
}

func decStr(d *decimal.Decimal) string {
	if d == nil {
		return "null"
	}
	return d.String()
}

func intStr(i *int) string {
	if i == nil {
		return "null"
	}
	return strconv.Itoa(*i)
}

func timeStr(t *time.Time) string {
	if t == nil {
		return "null"
	}
	return t.UTC().Format(time.RFC3339)
}
