package event

import (
	"net/http"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/krnskrll/event-mpos/internal/common"
)

// EventStatus is the event lifecycle. BAs can only sell while ACTIVE.
type EventStatus string

const (
	StatusDraft  EventStatus = "DRAFT"
	StatusActive EventStatus = "ACTIVE"
	StatusClosed EventStatus = "CLOSED"
)

func ParseStatus(s string) (EventStatus, bool) {
	switch EventStatus(s) {
	case StatusDraft, StatusActive, StatusClosed:
		return EventStatus(s), true
	}
	return "", false
}

// AssertCanTransitionTo returns a 409 AppError for a disallowed lifecycle transition.
//
//	DRAFT  -> ACTIVE | CLOSED
//	ACTIVE -> CLOSED
//	CLOSED -> (none)
func (s EventStatus) AssertCanTransitionTo(target EventStatus) error {
	allowed := false
	switch s {
	case StatusDraft:
		allowed = target == StatusActive || target == StatusClosed
	case StatusActive:
		allowed = target == StatusClosed
	case StatusClosed:
		allowed = false
	}
	if !allowed {
		return common.NewAppError(http.StatusConflict,
			"Cannot change event status from "+string(s)+" to "+string(target))
	}
	return nil
}

// Event is the central domain boundary of the system.
type Event struct {
	ID        uuid.UUID    `gorm:"type:uuid;primaryKey"`
	Code      string       `gorm:"uniqueIndex;not null"`
	Name      string       `gorm:"not null"`
	StartDate *common.Date `gorm:"column:start_date;type:date"`
	EndDate   *common.Date `gorm:"column:end_date;type:date"`
	Status    EventStatus  `gorm:"type:varchar(20);not null"`
	CreatedAt time.Time    `gorm:"column:created_at;not null"`
	// Bumped on any edit/status change; feeds the snapshot version so devices know when to re-sync.
	UpdatedAt time.Time `gorm:"column:updated_at;not null"`
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

// ---- DTOs ----

type EventResponse struct {
	ID        uuid.UUID    `json:"id"`
	Code      string       `json:"code"`
	Name      string       `json:"name"`
	StartDate *common.Date `json:"startDate"`
	EndDate   *common.Date `json:"endDate"`
	Status    EventStatus  `json:"status"`
}

func (e *Event) ToResponse() EventResponse {
	return EventResponse{
		ID:        e.ID,
		Code:      e.Code,
		Name:      e.Name,
		StartDate: e.StartDate,
		EndDate:   e.EndDate,
		Status:    e.Status,
	}
}

type CreateEventRequest struct {
	Code      string       `json:"code" binding:"required"`
	Name      string       `json:"name" binding:"required"`
	StartDate *common.Date `json:"startDate"`
	EndDate   *common.Date `json:"endDate"`
}

type UpdateEventRequest struct {
	Name      *string      `json:"name"`
	StartDate *common.Date `json:"startDate"`
	EndDate   *common.Date `json:"endDate"`
}

type ChangeStatusRequest struct {
	Status EventStatus `json:"status" binding:"required"`
}

// AssignBaRequest — userId/eventId are looked up (a nonexistent id yields 404, matching Kotlin's
// @NotNull semantics which reject only a missing field, not a zero-valued UUID).
type AssignBaRequest struct {
	UserID  uuid.UUID `json:"userId"`
	EventID uuid.UUID `json:"eventId"`
}

// ---- Snapshot ----

// EventSnapshot is the one-call payload the Android app downloads when a BA selects an event.
// products/members/promotions are filled by the registered providers (Phases 3–6); stock stays
// empty because remaining stock is folded into each product instead.
type EventSnapshot struct {
	Event       EventResponse `json:"event"`
	Version     string        `json:"version"`
	Products    []any         `json:"products"`
	Stock       []any         `json:"stock"`
	Members     []any         `json:"members"`
	Promotions  []any         `json:"promotions"`
	MemberCount int           `json:"memberCount"`
}

// SnapshotStatus is returned when the client's stored version is already current.
type SnapshotStatus struct {
	Changed bool   `json:"changed"`
	Version string `json:"version"`
}

// ---- Provider interfaces (implemented by later phases, wired in main) ----

type ProductProvider interface {
	AvailableSnapshotProducts(eventID uuid.UUID) []any
	AvailabilityFingerprint(eventID uuid.UUID) string
}

type MemberProvider interface {
	ActiveSnapshotMembers() []any
	Fingerprint() string
}

type StockProvider interface {
	Remaining(eventID, productID uuid.UUID) int
	AllocationFingerprint(eventID uuid.UUID) string
}

type PromotionProvider interface {
	SnapshotPromotions(eventID uuid.UUID) []any
	Fingerprint(eventID uuid.UUID) string
}

// Providers bundles the optional cross-domain contributors to the snapshot + its version. Each is
// nil until its phase is wired; nil contributes an empty fingerprint and empty snapshot data.
type Providers struct {
	Products   ProductProvider
	Members    MemberProvider
	Stock      StockProvider
	Promotions PromotionProvider
}
