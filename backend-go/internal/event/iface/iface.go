package iface

import (
	"net/http"
	"time"

	"github.com/google/uuid"

	"github.com/krnskrll/event-mpos/internal/platform/common"
)

// EventStatus is the location lifecycle. Cashiers can only sell while ACTIVE.
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

// LocationType: an outlet is a permanent shop; an event is temporary.
type LocationType string

const (
	TypeEvent  LocationType = "event"
	TypeOutlet LocationType = "outlet"
)

func ParseLocationType(s string) (LocationType, bool) {
	switch LocationType(s) {
	case "", TypeEvent:
		return TypeEvent, true
	case TypeOutlet:
		return TypeOutlet, true
	}
	return "", false
}

type EventResponse struct {
	ID        uuid.UUID    `json:"id"`
	StoreID   uuid.UUID    `json:"storeId"`
	Type      LocationType `json:"type"`
	Address   *string      `json:"address,omitempty"`
	Code      string       `json:"code"`
	Name      string       `json:"name"`
	StartDate *common.Date `json:"startDate,omitempty"`
	EndDate   *common.Date `json:"endDate,omitempty"`
	Status    EventStatus  `json:"status"`
	ClosedAt  *time.Time   `json:"closedAt,omitempty"`
}

type CreateEventRequest struct {
	Code      string       `json:"code" binding:"required"`
	Name      string       `json:"name" binding:"required"`
	StartDate *common.Date `json:"startDate"`
	EndDate   *common.Date `json:"endDate"`
	Type      LocationType `json:"type"`
	Address   *string      `json:"address"`
}

type UpdateEventRequest struct {
	Name      *string      `json:"name"`
	StartDate *common.Date `json:"startDate"`
	EndDate   *common.Date `json:"endDate"`
	Address   *string      `json:"address"`
}

type ChangeStatusRequest struct {
	Status EventStatus `json:"status" binding:"required"`
}

// EventSnapshot is the one-call payload a device downloads for an event. Stock stays empty:
// remaining stock is carried on each product.
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

type ProductProvider interface {
	AvailableSnapshotProducts(eventID uuid.UUID) ([]any, error)
	AvailabilityFingerprint(eventID uuid.UUID) (string, error)
}

type MemberProvider interface {
	ActiveSnapshotMembers(storeID uuid.UUID) ([]any, error)
	Fingerprint(storeID uuid.UUID) (string, error)
}

type StockProvider interface {
	Remaining(eventID, productID uuid.UUID) (int, error)
	AllocationFingerprint(eventID uuid.UUID) (string, error)
}

type PromotionProvider interface {
	SnapshotPromotions(eventID uuid.UUID) ([]any, error)
	Fingerprint(eventID uuid.UUID) (string, error)
}

// Providers contribute to the snapshot and its version; a nil provider contributes nothing.
type Providers struct {
	Products   ProductProvider
	Members    MemberProvider
	Stock      StockProvider
	Promotions PromotionProvider
}

// Events is the event module's contract. Get is unscoped for internal use; callers acting for a
// user check RequireInStore first.
type Events interface {
	Get(id uuid.UUID) (*EventResponse, error)
	RequireInStore(storeID, eventID uuid.UUID) error
	ListEvents(storeID uuid.UUID) ([]EventResponse, error)
	SnapshotForEvent(eventID uuid.UUID) (*EventSnapshot, error)
	SnapshotVersion(eventID uuid.UUID) (string, error)
}

// CreateLocationRequest creates an outlet (address) or an event (dates); POS makes the code.
type CreateLocationRequest struct {
	Type      LocationType `json:"type" binding:"required,oneof=outlet event"`
	Name      string       `json:"name" binding:"required"`
	Address   *string      `json:"address"`
	StartDate *common.Date `json:"startDate"`
	EndDate   *common.Date `json:"endDate"`
}

// LocationResponse is a location in the back office, with how many Katalog POS items it sells.
type LocationResponse struct {
	EventResponse
	Items int `json:"items"`
}
