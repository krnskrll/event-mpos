package repository

import (
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	"gorm.io/gorm"

	"github.com/krnskrll/event-mpos/internal/promotion/iface"
)

// Promotion is an event-scoped promotion evaluated by the shared PromotionEngine.
type Promotion struct {
	ID            uuid.UUID            `gorm:"type:uuid;primaryKey"`
	EventID       uuid.UUID            `gorm:"column:event_id;type:uuid;not null;index"`
	Name          string               `gorm:"not null"`
	Type          iface.PromotionType  `gorm:"type:varchar(20);not null"`
	PercentOff    *decimal.Decimal     `gorm:"column:percent_off;type:numeric(5,2)"`
	AmountOff     *decimal.Decimal     `gorm:"column:amount_off;type:bigint"`
	Scope         iface.PromotionScope `gorm:"type:varchar(20);not null"`
	ProductIdsCsv *string              `gorm:"column:product_ids;type:varchar(4000)"`
	MinQty        *int                 `gorm:"column:min_qty"`
	MinSubtotal   *decimal.Decimal     `gorm:"column:min_subtotal;type:bigint"`
	StartsAt      *time.Time           `gorm:"column:starts_at"`
	EndsAt        *time.Time           `gorm:"column:ends_at"`
	Active        bool                 `gorm:"not null"`
	Priority      int                  `gorm:"not null"`
	UpdatedAt     time.Time            `gorm:"column:updated_at;not null"`
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

func (p *Promotion) ToResponse() iface.PromotionResponse {
	return iface.PromotionResponse{
		ID: p.ID, EventID: p.EventID, Name: p.Name, Type: p.Type,
		PercentOff: p.PercentOff, AmountOff: p.AmountOff, Scope: p.Scope,
		ProductIDs: p.ProductIDs(), MinQty: p.MinQty, MinSubtotal: p.MinSubtotal,
		StartsAt: p.StartsAt, EndsAt: p.EndsAt, Active: p.Active, Priority: p.Priority,
	}
}

func (p *Promotion) ToSnapshotDto() iface.PromotionSnapshotDto {
	return iface.PromotionSnapshotDto{
		ID: p.ID, Name: p.Name, Type: p.Type,
		PercentOff: p.PercentOff, AmountOff: p.AmountOff, Scope: p.Scope,
		ProductIDs: p.ProductIDs(), MinQty: p.MinQty, MinSubtotal: p.MinSubtotal,
		StartsAt: p.StartsAt, EndsAt: p.EndsAt, Priority: p.Priority,
	}
}
