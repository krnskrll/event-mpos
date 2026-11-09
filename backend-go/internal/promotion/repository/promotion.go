package repository

import (
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	"gorm.io/gorm"

	"github.com/krnskrll/event-mpos/internal/promotion/iface"
)

// Promotion is evaluated by the shared PromotionEngine. EventID is the one location it runs at
// (nil = every location of the store); a SPECIFIC promotion targets CategoryID or Items.
type Promotion struct {
	ID          uuid.UUID            `gorm:"type:uuid;primaryKey"`
	StoreID     uuid.UUID            `gorm:"column:store_id;type:uuid;not null"`
	EventID     *uuid.UUID           `gorm:"column:event_id;type:uuid"`
	CategoryID  *uuid.UUID           `gorm:"column:category_id;type:uuid"`
	Name        string               `gorm:"not null"`
	Type        iface.PromotionType  `gorm:"type:varchar(20);not null"`
	PercentOff  *decimal.Decimal     `gorm:"column:percent_off;type:numeric(5,2)"`
	AmountOff   *decimal.Decimal     `gorm:"column:amount_off;type:bigint"`
	Scope       iface.PromotionScope `gorm:"type:varchar(20);not null"`
	Items       []PromotionItem      `gorm:"foreignKey:PromotionID"`
	MinQty      *int                 `gorm:"column:min_qty"`
	MinSubtotal *decimal.Decimal     `gorm:"column:min_subtotal;type:bigint"`
	StartsAt    *time.Time           `gorm:"column:starts_at"`
	EndsAt      *time.Time           `gorm:"column:ends_at"`
	Active      bool                 `gorm:"not null"`
	Priority    int                  `gorm:"not null"`
	UpdatedAt   time.Time            `gorm:"column:updated_at;not null"`
	// productIDs are the items the engine matches: Items, or the category's items.
	productIDs []uuid.UUID
}

type PromotionItem struct {
	PromotionID uuid.UUID `gorm:"column:promotion_id;type:uuid;primaryKey"`
	ItemID      uuid.UUID `gorm:"column:item_id;type:uuid;primaryKey"`
}

func (Promotion) TableName() string     { return "promotions" }
func (PromotionItem) TableName() string { return "pos_promotion_items" }

func (p *Promotion) BeforeCreate(_ *gorm.DB) error {
	if p.ID == uuid.Nil {
		p.ID = uuid.New()
	}
	if p.UpdatedAt.IsZero() {
		p.UpdatedAt = time.Now().UTC()
	}
	return nil
}

func (p *Promotion) ItemIDs() []uuid.UUID {
	out := make([]uuid.UUID, len(p.Items))
	for i, it := range p.Items {
		out[i] = it.ItemID
	}
	return out
}

func (p *Promotion) ProductIDs() []uuid.UUID {
	if p.productIDs == nil {
		return []uuid.UUID{}
	}
	return p.productIDs
}

func (p *Promotion) SetProductIDs(ids []uuid.UUID) { p.productIDs = ids }

func (p *Promotion) ToResponse() iface.PromotionResponse {
	return iface.PromotionResponse{
		ID: p.ID, LocationID: p.EventID, CategoryID: p.CategoryID, Name: p.Name, Type: p.Type,
		PercentOff: p.PercentOff, AmountOff: p.AmountOff, Scope: p.Scope,
		ProductIDs: p.ItemIDs(), MinQty: p.MinQty, MinSubtotal: p.MinSubtotal,
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
