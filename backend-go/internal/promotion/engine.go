package promotion

import (
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

// CartLine is a normalized cart line for promotion evaluation.
type CartLine struct {
	ProductID uuid.UUID
	UnitPrice decimal.Decimal
	Qty       int
}

// AppliedPromotion is the single promotion chosen for a cart, plus the discount it yields.
type AppliedPromotion struct {
	Promotion *Promotion
	Discount  decimal.Decimal
}

// BestPromotion evaluates every active+in-window promotion and returns the one with the largest
// discount (ties broken by higher priority, then higher name). One promotion per order. Mirrors the
// Kotlin PromotionEngine — keep the Android copy identical.
func BestPromotion(lines []CartLine, promotions []Promotion, now time.Time) *AppliedPromotion {
	var best *AppliedPromotion
	for i := range promotions {
		promo := &promotions[i]
		discount, ok := DiscountFor(promo, lines, now)
		if !ok || !discount.GreaterThan(decimal.Zero) {
			continue
		}
		cand := &AppliedPromotion{Promotion: promo, Discount: discount}
		if best == nil || greater(cand, best) {
			best = cand
		}
	}
	return best
}

// greater reports whether a beats b by (discount, priority, name) ascending — i.e. a is the max.
func greater(a, b *AppliedPromotion) bool {
	if c := a.Discount.Cmp(b.Discount); c != 0 {
		return c > 0
	}
	if a.Promotion.Priority != b.Promotion.Priority {
		return a.Promotion.Priority > b.Promotion.Priority
	}
	return a.Promotion.Name > b.Promotion.Name
}

// DiscountFor returns the discount this promotion yields for the cart, and whether it applies.
func DiscountFor(promo *Promotion, lines []CartLine, now time.Time) (decimal.Decimal, bool) {
	if !promo.Active {
		return decimal.Zero, false
	}
	if promo.StartsAt != nil && now.Before(*promo.StartsAt) {
		return decimal.Zero, false
	}
	if promo.EndsAt != nil && now.After(*promo.EndsAt) {
		return decimal.Zero, false
	}

	var applicable []CartLine
	switch promo.Scope {
	case ScopeSpecific:
		ids := map[uuid.UUID]bool{}
		for _, id := range promo.ProductIDs() {
			ids[id] = true
		}
		for _, l := range lines {
			if ids[l.ProductID] {
				applicable = append(applicable, l)
			}
		}
	default: // ScopeAll
		applicable = lines
	}
	if len(applicable) == 0 {
		return decimal.Zero, false
	}

	qty := 0
	subtotal := decimal.Zero
	for _, l := range applicable {
		qty += l.Qty
		subtotal = subtotal.Add(l.UnitPrice.Mul(decimal.NewFromInt(int64(l.Qty))))
	}

	if promo.MinQty != nil && qty < *promo.MinQty {
		return decimal.Zero, false
	}
	if promo.MinSubtotal != nil && subtotal.LessThan(*promo.MinSubtotal) {
		return decimal.Zero, false
	}

	var raw decimal.Decimal
	switch promo.Type {
	case TypePercentage:
		if promo.PercentOff == nil {
			return decimal.Zero, false
		}
		raw = subtotal.Mul(*promo.PercentOff).Div(decimal.NewFromInt(100))
	case TypeFixed:
		if promo.AmountOff == nil {
			return decimal.Zero, false
		}
		raw = *promo.AmountOff
	case TypeBundle:
		if promo.MinQty == nil {
			return decimal.Zero, false
		}
		if qty < *promo.MinQty {
			return decimal.Zero, false
		}
		if promo.AmountOff == nil {
			return decimal.Zero, false
		}
		raw = *promo.AmountOff
	default:
		return decimal.Zero, false
	}

	// never discount more than the applicable subtotal, never negative
	capped := decimal.Min(raw, subtotal)
	if capped.IsNegative() {
		capped = decimal.Zero
	}
	return capped.Round(2), true
}
