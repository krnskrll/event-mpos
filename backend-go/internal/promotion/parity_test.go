package promotion

import (
	"encoding/json"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

// TestPromotionParityFixture runs the shared fixture at testdata/promotion-cases.json (repo root).
// The Android PromotionParityTest runs the same file against the client engine, so a change to only
// one of the two engines fails CI.

type fixtureLine struct {
	ProductID string `json:"productId"`
	UnitPrice string `json:"unitPrice"`
	Qty       int    `json:"qty"`
}

type fixturePromotion struct {
	Name        string   `json:"name"`
	Type        string   `json:"type"`
	PercentOff  *string  `json:"percentOff"`
	AmountOff   *string  `json:"amountOff"`
	Scope       string   `json:"scope"`
	ProductIDs  []string `json:"productIds"`
	MinQty      *int     `json:"minQty"`
	MinSubtotal *string  `json:"minSubtotal"`
	StartsAt    *string  `json:"startsAt"`
	EndsAt      *string  `json:"endsAt"`
	Priority    int      `json:"priority"`
}

type promotionFixture struct {
	Now   string                   `json:"now"`
	Carts map[string][]fixtureLine `json:"carts"`
	Cases []struct {
		Name       string             `json:"name"`
		Cart       string             `json:"cart"`
		Promotions []fixturePromotion `json:"promotions"`
		Expected   struct {
			Promotion *string `json:"promotion"`
			Discount  string  `json:"discount"`
		} `json:"expected"`
	} `json:"cases"`
}

func optDecimal(s *string) *decimal.Decimal {
	if s == nil {
		return nil
	}
	d := decimal.RequireFromString(*s)
	return &d
}

func optTime(t *testing.T, s *string) *time.Time {
	if s == nil {
		return nil
	}
	v, err := time.Parse(time.RFC3339, *s)
	if err != nil {
		t.Fatalf("bad time %q: %v", *s, err)
	}
	return &v
}

func (f fixturePromotion) toPromotion(t *testing.T) Promotion {
	scope := ScopeAll
	if f.Scope != "" {
		scope = PromotionScope(f.Scope)
	}
	var ids *string
	if len(f.ProductIDs) > 0 {
		csv := strings.Join(f.ProductIDs, ",")
		ids = &csv
	}
	return Promotion{
		ID: uuid.New(), Name: f.Name, Type: PromotionType(f.Type),
		PercentOff: optDecimal(f.PercentOff), AmountOff: optDecimal(f.AmountOff),
		Scope: scope, ProductIdsCsv: ids, MinQty: f.MinQty, MinSubtotal: optDecimal(f.MinSubtotal),
		StartsAt: optTime(t, f.StartsAt), EndsAt: optTime(t, f.EndsAt), Active: true, Priority: f.Priority,
	}
}

func TestPromotionParityFixture(t *testing.T) {
	raw, err := os.ReadFile("../../../testdata/promotion-cases.json")
	if err != nil {
		t.Fatalf("read fixture: %v", err)
	}
	var fx promotionFixture
	if err := json.Unmarshal(raw, &fx); err != nil {
		t.Fatalf("parse fixture: %v", err)
	}
	now, err := time.Parse(time.RFC3339, fx.Now)
	if err != nil {
		t.Fatal(err)
	}
	if len(fx.Cases) == 0 {
		t.Fatal("fixture has no cases")
	}

	for _, c := range fx.Cases {
		t.Run(c.Name, func(t *testing.T) {
			cartLines, ok := fx.Carts[c.Cart]
			if !ok {
				t.Fatalf("unknown cart %q", c.Cart)
			}
			lines := make([]CartLine, 0, len(cartLines))
			for _, l := range cartLines {
				lines = append(lines, CartLine{
					ProductID: uuid.MustParse(l.ProductID), UnitPrice: decimal.RequireFromString(l.UnitPrice), Qty: l.Qty,
				})
			}
			promos := make([]Promotion, 0, len(c.Promotions))
			for _, p := range c.Promotions {
				promos = append(promos, p.toPromotion(t))
			}

			best := BestPromotion(lines, promos, now)
			wantDiscount := decimal.RequireFromString(c.Expected.Discount)
			switch {
			case c.Expected.Promotion == nil && best != nil:
				t.Errorf("chose %q (%s), want none", best.Promotion.Name, best.Discount)
			case c.Expected.Promotion != nil && best == nil:
				t.Errorf("chose none, want %q", *c.Expected.Promotion)
			case best != nil && (best.Promotion.Name != *c.Expected.Promotion || !best.Discount.Equal(wantDiscount)):
				t.Errorf("chose %q (%s), want %q (%s)", best.Promotion.Name, best.Discount, *c.Expected.Promotion, wantDiscount)
			}
		})
	}
}
