package promotion

import (
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	"gorm.io/gorm"
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

// Promotion is an event-scoped promotion evaluated by the shared PromotionEngine.
type Promotion struct {
	ID            uuid.UUID        `gorm:"type:uuid;primaryKey"`
	EventID       uuid.UUID        `gorm:"column:event_id;type:uuid;not null;index"`
	Name          string           `gorm:"not null"`
	Type          PromotionType    `gorm:"type:varchar(20);not null"`
	PercentOff    *decimal.Decimal `gorm:"column:percent_off;type:numeric(5,2)"`
	AmountOff     *decimal.Decimal `gorm:"column:amount_off;type:numeric(12,2)"`
	Scope         PromotionScope   `gorm:"type:varchar(20);not null"`
	ProductIdsCsv *string          `gorm:"column:product_ids;type:varchar(4000)"`
	MinQty        *int             `gorm:"column:min_qty"`
	MinSubtotal   *decimal.Decimal `gorm:"column:min_subtotal;type:numeric(12,2)"`
	StartsAt      *time.Time       `gorm:"column:starts_at"`
	EndsAt        *time.Time       `gorm:"column:ends_at"`
	Active        bool             `gorm:"not null"`
	Priority      int              `gorm:"not null"`
	UpdatedAt     time.Time        `gorm:"column:updated_at;not null"`
}

func (Promotion) TableName() string { return "promotions" }

func (p *Promotion) BeforeCreate(_ *gorm.DB) error {
	if p.ID == uuid.Nil {
		p.ID = uuid.New()
	}
	if p.UpdatedAt.IsZero() {
		p.UpdatedAt = time.Now().UTC()
	}
	return nil
}

func (p *Promotion) ProductIDs() []uuid.UUID {
	if p.ProductIdsCsv == nil || strings.TrimSpace(*p.ProductIdsCsv) == "" {
		return []uuid.UUID{}
	}
	out := []uuid.UUID{}
	for _, s := range strings.Split(*p.ProductIdsCsv, ",") {
		s = strings.TrimSpace(s)
		if s == "" {
			continue
		}
		if id, err := uuid.Parse(s); err == nil {
			out = append(out, id)
		}
	}
	return out
}

func (p *Promotion) SetProductIDs(ids []uuid.UUID) {
	if len(ids) == 0 {
		p.ProductIdsCsv = nil
		return
	}
	parts := make([]string, len(ids))
	for i, id := range ids {
		parts[i] = id.String()
	}
	csv := strings.Join(parts, ",")
	p.ProductIdsCsv = &csv
}

// ---- DTOs ----

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

func (p *Promotion) ToResponse() PromotionResponse {
	return PromotionResponse{
		ID: p.ID, EventID: p.EventID, Name: p.Name, Type: p.Type,
		PercentOff: p.PercentOff, AmountOff: p.AmountOff, Scope: p.Scope,
		ProductIDs: p.ProductIDs(), MinQty: p.MinQty, MinSubtotal: p.MinSubtotal,
		StartsAt: p.StartsAt, EndsAt: p.EndsAt, Active: p.Active, Priority: p.Priority,
	}
}

func (p *Promotion) ToSnapshotDto() PromotionSnapshotDto {
	return PromotionSnapshotDto{
		ID: p.ID, Name: p.Name, Type: p.Type,
		PercentOff: p.PercentOff, AmountOff: p.AmountOff, Scope: p.Scope,
		ProductIDs: p.ProductIDs(), MinQty: p.MinQty, MinSubtotal: p.MinSubtotal,
		StartsAt: p.StartsAt, EndsAt: p.EndsAt,
	}
}
