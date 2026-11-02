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
}

func NewService(db *gorm.DB, products productiface.Products) *Service {
	return &Service{db: db, products: products}
}

func (s *Service) ListForEvent(eventID uuid.UUID) ([]iface.PromotionResponse, error) {
	var promos []repository.Promotion
	if err := s.db.Where("event_id = ?", eventID).Find(&promos).Error; err != nil {
		return nil, err
	}
	sort.SliceStable(promos, func(i, j int) bool { return promos[i].Priority > promos[j].Priority })
	out := make([]iface.PromotionResponse, 0, len(promos))
	for i := range promos {
		out = append(out, promos[i].ToResponse())
	}
	return out, nil
}

func (s *Service) activeForEvent(eventID uuid.UUID) ([]repository.Promotion, error) {
	var promos []repository.Promotion
	err := s.db.Where("event_id = ? AND active = ?", eventID, true).Find(&promos).Error
	return promos, err
}

// BestFor evaluates the active promotions for an event against a cart and returns the single best
// applied promotion (or nil). Shared by the quote endpoint and order submission.
func (s *Service) BestFor(eventID uuid.UUID, cart []iface.CartLine, now time.Time) (*iface.Applied, error) {
	promos, err := s.activeForEvent(eventID)
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
	promos, err := s.activeForEvent(eventID)
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
// snapshot version. Deterministic given the data.
func (s *Service) Fingerprint(eventID uuid.UUID) (string, error) {
	var promos []repository.Promotion
	if err := s.db.Where("event_id = ?", eventID).Find(&promos).Error; err != nil {
		return "", err
	}
	sort.Slice(promos, func(i, j int) bool { return promos[i].ID.String() < promos[j].ID.String() })
	parts := make([]string, 0, len(promos))
	for i := range promos {
		p := &promos[i]
		parts = append(parts, strings.Join([]string{
			p.ID.String(), p.Name, string(p.Type), string(p.Scope), csvStr(p.ProductIdsCsv),
			decStr(p.PercentOff), decStr(p.AmountOff), intStr(p.MinQty), decStr(p.MinSubtotal),
			strconv.FormatBool(p.Active), timeStr(p.StartsAt), timeStr(p.EndsAt), strconv.Itoa(p.Priority),
		}, ":"))
	}
	return strings.Join(parts, ","), nil
}

func (s *Service) Create(eventID uuid.UUID, req iface.PromotionRequest) (*iface.PromotionResponse, error) {
	if err := validate(req); err != nil {
		return nil, err
	}
	promo := repository.Promotion{EventID: eventID, Name: strings.TrimSpace(req.Name), Type: req.Type}
	applyReq(&promo, req)
	if err := s.db.Create(&promo).Error; err != nil {
		return nil, err
	}
	resp := promo.ToResponse()
	return &resp, nil
}

func (s *Service) Update(id uuid.UUID, req iface.PromotionRequest) (*iface.PromotionResponse, error) {
	if err := validate(req); err != nil {
		return nil, err
	}
	var promo repository.Promotion
	err := s.db.First(&promo, "id = ?", id).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, common.NewAppError(http.StatusNotFound, "repository.Promotion not found")
	}
	if err != nil {
		return nil, err
	}
	promo.Name = strings.TrimSpace(req.Name)
	promo.Type = req.Type
	applyReq(&promo, req)
	promo.UpdatedAt = time.Now().UTC()
	if err := s.db.Save(&promo).Error; err != nil {
		return nil, err
	}
	resp := promo.ToResponse()
	return &resp, nil
}

func (s *Service) Delete(id uuid.UUID) error {
	var count int64
	if err := s.db.Model(&repository.Promotion{}).Where("id = ?", id).Count(&count).Error; err != nil {
		return err
	}
	if count == 0 {
		return common.NewAppError(http.StatusNotFound, "repository.Promotion not found")
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
		p, err := s.products.Find(line.ProductID)
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
	subtotal = subtotal.Round(2)

	promos, err := s.activeForEvent(eventID)
	if err != nil {
		return nil, err
	}
	applied := BestPromotion(cart, promos, time.Now())
	discount := decimal.Zero.Round(2)
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
	resp.Total = total.Round(2)
	return &resp, nil
}

// ---- validation + apply ----

func validate(req iface.PromotionRequest) error {
	pos := func(d *decimal.Decimal) bool { return d != nil && d.GreaterThan(decimal.Zero) }
	switch req.Type {
	case iface.TypePercentage:
		if !pos(req.PercentOff) {
			return common.NewAppError(http.StatusBadRequest, "percentOff required for a percentage promotion")
		}
	case iface.TypeFixed:
		if !pos(req.AmountOff) {
			return common.NewAppError(http.StatusBadRequest, "amountOff required for a fixed promotion")
		}
	case iface.TypeBundle:
		if !pos(req.AmountOff) {
			return common.NewAppError(http.StatusBadRequest, "amountOff required for a bundle promotion")
		}
		if req.MinQty == nil || *req.MinQty < 1 {
			return common.NewAppError(http.StatusBadRequest, "minQty (bundle size) required for a bundle promotion")
		}
	default:
		return common.NewAppError(http.StatusBadRequest, "type must be one of PERCENTAGE, FIXED, BUNDLE")
	}
	scope := iface.ScopeAll
	if req.Scope != nil {
		scope = *req.Scope
	}
	if scope == iface.ScopeSpecific && len(req.ProductIDs) == 0 {
		return common.NewAppError(http.StatusBadRequest, "SPECIFIC scope needs at least one product")
	}
	return nil
}

func applyReq(promo *repository.Promotion, req iface.PromotionRequest) {
	promo.PercentOff = req.PercentOff
	promo.AmountOff = req.AmountOff
	promo.Scope = iface.ScopeAll
	if req.Scope != nil {
		promo.Scope = *req.Scope
	}
	promo.SetProductIDs(req.ProductIDs)
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
}

func decStr(d *decimal.Decimal) string {
	if d == nil {
		return "null"
	}
	return d.String()
}

func csvStr(s *string) string {
	if s == nil {
		return ""
	}
	return *s
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
