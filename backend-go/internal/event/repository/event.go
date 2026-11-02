package repository

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/krnskrll/event-mpos/internal/event/iface"
	"github.com/krnskrll/event-mpos/internal/platform/common"
)

// Event is the central domain boundary of the system.
type Event struct {
	ID        uuid.UUID        `gorm:"type:uuid;primaryKey"`
	StoreID   uuid.UUID        `gorm:"column:store_id;type:uuid;not null"`
	Type      iface.OutletType `gorm:"type:varchar(8);not null"`
	Address   *string
	Code      string            `gorm:"uniqueIndex;not null"`
	Name      string            `gorm:"not null"`
	StartDate *common.Date      `gorm:"column:start_date;type:date"`
	EndDate   *common.Date      `gorm:"column:end_date;type:date"`
	Status    iface.EventStatus `gorm:"type:varchar(20);not null"`
	CreatedAt time.Time         `gorm:"column:created_at;not null"`
	// Bumped on any edit/status change; feeds the snapshot version so devices know when to re-sync.
	UpdatedAt time.Time `gorm:"column:updated_at;not null"`
	// When the event moved to CLOSED; queued orders created before it are still accepted for 24h.
	ClosedAt *time.Time `gorm:"column:closed_at"`
}

func (Event) TableName() string { return "events" }

func (e *Event) BeforeCreate(_ *gorm.DB) error {
	if e.ID == uuid.Nil {
		e.ID = uuid.New()
	}
	now := time.Now().UTC()
	if e.CreatedAt.IsZero() {
		e.CreatedAt = now
	}
	if e.UpdatedAt.IsZero() {
		e.UpdatedAt = now
	}
	return nil
}

// BaEventAssignment maps a BA (user) to the event they are working. At most one assignment is
// active at a time — the active one is what the Android app loads after login.
type BaEventAssignment struct {
	ID         uuid.UUID `gorm:"type:uuid;primaryKey"`
	UserID     uuid.UUID `gorm:"column:user_id;type:uuid;not null;index"`
	EventID    uuid.UUID `gorm:"column:event_id;type:uuid;not null;index"`
	Active     bool      `gorm:"not null;default:true"`
	AssignedAt time.Time `gorm:"column:assigned_at;not null"`
}

func (BaEventAssignment) TableName() string { return "ba_event_assignments" }

func (a *BaEventAssignment) BeforeCreate(_ *gorm.DB) error {
	if a.ID == uuid.Nil {
		a.ID = uuid.New()
	}
	if a.AssignedAt.IsZero() {
		a.AssignedAt = time.Now().UTC()
	}
	return nil
}

func (e *Event) ToResponse() iface.EventResponse {
	return iface.EventResponse{
		ID:        e.ID,
		StoreID:   e.StoreID,
		Type:      e.Type,
		Address:   e.Address,
		Code:      e.Code,
		Name:      e.Name,
		StartDate: e.StartDate,
		EndDate:   e.EndDate,
		ClosedAt:  e.ClosedAt,
		Status:    e.Status,
	}
}
