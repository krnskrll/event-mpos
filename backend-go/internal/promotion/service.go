package promotion

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

	"github.com/krnskrll/event-mpos/internal/common"
	"github.com/krnskrll/event-mpos/internal/product"
)

// Service holds promotion CRUD + the authoritative quote — the Go analogue of Kotlin PromotionService.
type Service struct {
	db *gorm.DB
}

func NewService(db *gorm.DB) *Service { return &Service{db: db} }

func (s *Service) ListForEvent(eventID uuid.UUID) ([]PromotionResponse, error) {
	var promos []Promotion
	if err := s.db.Where("event_id = ?", eventID).Find(&promos).Error; err != nil {
		return nil, err
	}
	sort.SliceStable(promos, func(i, j int) bool { return promos[i].Priority > promos[j].Priority })
	out := make([]PromotionResponse, 0, len(promos))
	for i := range promos {
		out = append(out, promos[i].ToResponse())
	}
	return out, nil
}

func (s *Service) activeForEvent(eventID uuid.UUID) ([]Promotion, error) {
	var promos []Promotion
	err := s.db.Where("event_id = ? AND active = ?", eventID, true).Find(&promos).Error
	return promos, err
}

// BestFor evaluates the active promotions for an event against a cart and returns the single best
// applied promotion (or nil). Shared by the quote endpoint and order submission.
func (s *Service) BestFor(eventID uuid.UUID, cart []CartLine, now time.Time) (*AppliedPromotion, error) {
	promos, err := s.activeForEvent(eventID)
	if err != nil {
		return nil, err
	}
	return BestPromotion(cart, promos, now), nil
}

// SnapshotPromotions returns active promotions as compact snapshot DTOs (event.PromotionProvider).
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
	var promos []Promotion
	if err := s.db.Where("event_id = ?", eventID).Find(&promos).Error; err != nil {
		return "", err
	}
	sort.Slice(promos, func(i, j int) bool { return promos[i].ID.String() < promos[j].ID.String() })
	parts := make([]string, 0, len(promos))
	for i := range promos {
		p := &promos[i]
		parts = append(parts, strings.Join([]string{
			p.ID.String(), string(p.Type), decStr(p.PercentOff), decStr(p.AmountOff),
			intStr(p.MinQty), decStr(p.MinSubtotal), strconv.FormatBool(p.Active),
			timeStr(p.StartsAt), timeStr(p.EndsAt),
		}, ":"))
	}
	return strings.Join(parts, ","), nil
}

func (s *Service) Create(eventID uuid.UUID, req PromotionRequest) (*PromotionResponse, error) {
	if err := validate(req); err != nil {
		return nil, err
	}
	promo := Promotion{EventID: eventID, Name: strings.TrimSpace(req.Name), Type: req.Type}
	applyReq(&promo, req)
	if err := s.db.Create(&promo).Error; err != nil {
		return nil, err
	}
	resp := promo.ToResponse()
	return &resp, nil
}

func (s *Service) Update(id uuid.UUID, req PromotionRequest) (*PromotionResponse, error) {
	if err := validate(req); err != nil {
		return nil, err
	}
	var promo Promotion
	err := s.db.First(&promo, "id = ?", id).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, common.NewAppError(http.StatusNotFound, "Promotion not found")
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
	if err := s.db.Model(&Promotion{}).Where("id = ?", id).Count(&count).Error; err != nil {
		return err
	}
	if count == 0 {
		return common.NewAppError(http.StatusNotFound, "Promotion not found")
	}
	return s.db.Delete(&Promotion{}, "id = ?", id).Error
}

// Quote is the authoritative cart quote: DB prices + the shared engine picks the single best promo.
func (s *Service) Quote(eventID uuid.UUID, lines []QuoteLine) (*QuoteResponse, error) {
	cart := make([]CartLine, 0, len(lines))
	for _, line := range lines {
		if line.Qty <= 0 {
			continue
		}
		var p product.Product
		if err := s.db.First(&p, "id = ?", line.ProductID).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				continue
			}
			return nil, err
		}
		cart = append(cart, CartLine{ProductID: p.ID, UnitPrice: p.Price, Qty: line.Qty})
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
	resp := QuoteResponse{Subtotal: subtotal}
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

func validate(req PromotionRequest) error {
	pos := func(d *decimal.Decimal) bool { return d != nil && d.GreaterThan(decimal.Zero) }
	switch req.Type {
	case TypePercentage:
		if !pos(req.PercentOff) {
			return common.NewAppError(http.StatusBadRequest, "percentOff required for a percentage promotion")
		}
	case TypeFixed:
		if !pos(req.AmountOff) {
			return common.NewAppError(http.StatusBadRequest, "amountOff required for a fixed promotion")
		}
	case TypeBundle:
		if !pos(req.AmountOff) {
			return common.NewAppError(http.StatusBadRequest, "amountOff required for a bundle promotion")
		}
		if req.MinQty == nil || *req.MinQty < 1 {
			return common.NewAppError(http.StatusBadRequest, "minQty (bundle size) required for a bundle promotion")
		}
	default:
		return common.NewAppError(http.StatusBadRequest, "type must be one of PERCENTAGE, FIXED, BUNDLE")
	}
	scope := ScopeAll
	if req.Scope != nil {
		scope = *req.Scope
	}
	if scope == ScopeSpecific && len(req.ProductIDs) == 0 {
		return common.NewAppError(http.StatusBadRequest, "SPECIFIC scope needs at least one product")
	}
	return nil
}

func applyReq(promo *Promotion, req PromotionRequest) {
	promo.PercentOff = req.PercentOff
	promo.AmountOff = req.AmountOff
	promo.Scope = ScopeAll
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
