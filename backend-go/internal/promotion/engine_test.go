package promotion

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

func d(s string) decimal.Decimal  { return decimal.RequireFromString(s) }
func dp(s string) *decimal.Decimal { v := d(s); return &v }
func ip(i int) *int               { return &i }

func TestPercentage(t *testing.T) {
	p := Promotion{Type: TypePercentage, Scope: ScopeAll, PercentOff: dp("10"), Active: true}
	line := CartLine{ProductID: uuid.New(), UnitPrice: d("79000.00"), Qty: 2} // subtotal 158000
	got, ok := DiscountFor(&p, []CartLine{line}, time.Now())
	if !ok || !got.Equal(d("15800.00")) {
		t.Fatalf("percentage: got %s ok=%v, want 15800.00", got, ok)
	}
}

func TestFixedWithMinSubtotal(t *testing.T) {
	p := Promotion{Type: TypeFixed, Scope: ScopeAll, AmountOff: dp("20000"), MinSubtotal: dp("200000"), Active: true}
	line := CartLine{ProductID: uuid.New(), UnitPrice: d("99000.00"), Qty: 1} // subtotal 99000 < 200000
	if _, ok := DiscountFor(&p, []CartLine{line}, time.Now()); ok {
		t.Error("fixed should not apply below minSubtotal")
	}
	line.Qty = 3 // 297000 >= 200000
	got, ok := DiscountFor(&p, []CartLine{line}, time.Now())
	if !ok || !got.Equal(d("20000.00")) {
		t.Fatalf("fixed: got %s ok=%v, want 20000.00", got, ok)
	}
}

func TestBundle(t *testing.T) {
	pid := uuid.New()
	p := Promotion{Type: TypeBundle, Scope: ScopeSpecific, AmountOff: dp("30000"), MinQty: ip(3), Active: true}
	p.SetProductIDs([]uuid.UUID{pid})
	line := CartLine{ProductID: pid, UnitPrice: d("79000.00"), Qty: 2}
	if _, ok := DiscountFor(&p, []CartLine{line}, time.Now()); ok {
		t.Error("bundle should not apply below bundle size")
	}
	line.Qty = 3
	got, ok := DiscountFor(&p, []CartLine{line}, time.Now())
	if !ok || !got.Equal(d("30000.00")) {
		t.Fatalf("bundle: got %s ok=%v, want 30000.00", got, ok)
	}
}

func TestDiscountCappedAtSubtotal(t *testing.T) {
	p := Promotion{Type: TypeFixed, Scope: ScopeAll, AmountOff: dp("100000"), Active: true}
	line := CartLine{ProductID: uuid.New(), UnitPrice: d("55000.00"), Qty: 1} // subtotal 55000 < 100000
	got, _ := DiscountFor(&p, []CartLine{line}, time.Now())
	if !got.Equal(d("55000.00")) {
		t.Fatalf("cap: got %s, want capped at 55000.00", got)
	}
}

func TestBestPromotionPicksLargest(t *testing.T) {
	line := CartLine{ProductID: uuid.New(), UnitPrice: d("100000.00"), Qty: 3} // subtotal 300000
	small := Promotion{Name: "10pct", Type: TypePercentage, Scope: ScopeAll, PercentOff: dp("10"), Active: true, Priority: 5}
	big := Promotion{Name: "fixed50k", Type: TypeFixed, Scope: ScopeAll, AmountOff: dp("50000"), Active: true, Priority: 1}
	best := BestPromotion([]CartLine{line}, []Promotion{small, big}, time.Now())
	if best == nil || best.Promotion.Name != "fixed50k" { // 50000 > 30000
		t.Fatalf("best: got %v, want fixed50k", best)
	}
}

func TestInactiveIgnored(t *testing.T) {
	p := Promotion{Type: TypePercentage, Scope: ScopeAll, PercentOff: dp("10"), Active: false}
	line := CartLine{ProductID: uuid.New(), UnitPrice: d("79000.00"), Qty: 1}
	if _, ok := DiscountFor(&p, []CartLine{line}, time.Now()); ok {
		t.Error("inactive promo should not apply")
	}
}
