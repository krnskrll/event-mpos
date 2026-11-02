package iface

import (
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

type PromotionType string

const (
	TypePercentage PromotionType = "PERCENTAGE"
	TypeFixed      PromotionType = "FIXED"
	TypeBundle     PromotionType = "BUNDLE"
)

func ParseType(s string) (PromotionType, bool) {
	switch PromotionType(s) {
	case TypePercentage, TypeFixed, TypeBundle:
		return PromotionType(s), true
	}
	return "", false
}

type PromotionScope string

const (
	ScopeAll      PromotionScope = "ALL"
	ScopeSpecific PromotionScope = "SPECIFIC"
)

type PromotionResponse struct {
	ID          uuid.UUID        `json:"id"`
	EventID     uuid.UUID        `json:"eventId"`
	Name        string           `json:"name"`
	Type        PromotionType    `json:"type"`
	PercentOff  *decimal.Decimal `json:"percentOff,omitempty"`
	AmountOff   *decimal.Decimal `json:"amountOff,omitempty"`
	Scope       PromotionScope   `json:"scope"`
	ProductIDs  []uuid.UUID      `json:"productIds"`
	MinQty      *int             `json:"minQty,omitempty"`
	MinSubtotal *decimal.Decimal `json:"minSubtotal,omitempty"`
	StartsAt    *time.Time       `json:"startsAt,omitempty"`
	EndsAt      *time.Time       `json:"endsAt,omitempty"`
	Active      bool             `json:"active"`
	Priority    int              `json:"priority"`
}

type PromotionRequest struct {
	Name        string           `json:"name" binding:"required"`
	Type        PromotionType    `json:"type" binding:"required"`
	PercentOff  *decimal.Decimal `json:"percentOff"`
	AmountOff   *decimal.Decimal `json:"amountOff"`
	Scope       *PromotionScope  `json:"scope"`
	ProductIDs  []uuid.UUID      `json:"productIds"`
	MinQty      *int             `json:"minQty"`
	MinSubtotal *decimal.Decimal `json:"minSubtotal"`
	StartsAt    *time.Time       `json:"startsAt"`
	EndsAt      *time.Time       `json:"endsAt"`
	Active      *bool            `json:"active"`
	Priority    *int             `json:"priority"`
}

type PromotionSnapshotDto struct {
	ID          uuid.UUID        `json:"id"`
	Name        string           `json:"name"`
	Type        PromotionType    `json:"type"`
	PercentOff  *decimal.Decimal `json:"percentOff,omitempty"`
	AmountOff   *decimal.Decimal `json:"amountOff,omitempty"`
	Scope       PromotionScope   `json:"scope"`
	ProductIDs  []uuid.UUID      `json:"productIds"`
	MinQty      *int             `json:"minQty,omitempty"`
	MinSubtotal *decimal.Decimal `json:"minSubtotal,omitempty"`
	StartsAt    *time.Time       `json:"startsAt,omitempty"`
	EndsAt      *time.Time       `json:"endsAt,omitempty"`
	// Priority breaks ties between equal discounts; the device engine needs it to match the server.
	Priority int `json:"priority"`
}

type QuoteLine struct {
	ProductID uuid.UUID `json:"productId"`
	Qty       int       `json:"qty"`
}

type QuoteRequest struct {
	Lines []QuoteLine `json:"lines"`
}

type QuoteResponse struct {
	Subtotal             decimal.Decimal `json:"subtotal"`
	Discount             decimal.Decimal `json:"discount"`
	Total                decimal.Decimal `json:"total"`
	AppliedPromotionID   *uuid.UUID      `json:"appliedPromotionId,omitempty"`
	AppliedPromotionName *string         `json:"appliedPromotionName,omitempty"`
}

// CartLine is a normalized cart line for promotion evaluation.
type CartLine struct {
	ProductID uuid.UUID
	UnitPrice decimal.Decimal
	Qty       int
}

// Applied is the single promotion chosen for a cart and the discount it gives.
type Applied struct {
	ID       uuid.UUID
	Name     string
	Discount decimal.Decimal
}

type Promotions interface {
	BestFor(eventID uuid.UUID, cart []CartLine, now time.Time) (*Applied, error)
	Quote(eventID uuid.UUID, lines []QuoteLine) (*QuoteResponse, error)
	SnapshotPromotions(eventID uuid.UUID) ([]any, error)
	Fingerprint(eventID uuid.UUID) (string, error)
}
