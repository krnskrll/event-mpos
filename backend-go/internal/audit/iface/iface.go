package iface

import (
	"time"

	"github.com/google/uuid"

	"github.com/krnskrll/event-mpos/internal/platform/database"
)

type AuditAction string

const (
	ActionOrderCreated   AuditAction = "ORDER_CREATED"
	ActionStatusChanged  AuditAction = "STATUS_CHANGED"
	ActionReceiptPrinted AuditAction = "RECEIPT_PRINTED"
	ActionOrderCancelled AuditAction = "ORDER_CANCELLED"
)

// ParseAction validates an action name from a query string.
func ParseAction(s string) (AuditAction, bool) {
	switch a := AuditAction(s); a {
	case ActionOrderCreated, ActionStatusChanged, ActionReceiptPrinted, ActionOrderCancelled:
		return a, true
	}
	return "", false
}

type AuditEntryResponse struct {
	ID             uuid.UUID   `json:"id"`
	OrderID        uuid.UUID   `json:"orderId"`
	OrderReference string      `json:"orderReference"`
	Action         AuditAction `json:"action"`
	ActorName      string      `json:"actorName"`
	FromStatus     *string     `json:"fromStatus,omitempty"`
	ToStatus       *string     `json:"toStatus,omitempty"`
	Detail         *string     `json:"detail,omitempty"`
	At             time.Time   `json:"at"`
}

// Query filters for an event's audit log. Nil filters match everything.
type Query struct {
	EventID uuid.UUID
	OrderID *uuid.UUID
	ActorID *uuid.UUID
	Action  *AuditAction
	// From and To bound the entry time, [From, To); zero is open.
	From, To time.Time
	Page     int // 0-based
	Size     int
}

// Page is one page of audit entries, newest first, plus the total matching count.
type Page struct {
	Items []AuditEntryResponse `json:"items"`
	Page  int                  `json:"page"`
	Size  int                  `json:"size"`
	Total int64                `json:"total"`
}

// Names resolves an actor id (an admin user or a staff member) to a display name.
type Names interface {
	NameOf(tx database.Tx, id uuid.UUID) (string, error)
}

// Audit is the audit module's contract. Record* write inside the caller's tx.
type Audit interface {
	RecordOrderCreated(tx database.Tx, eventID, orderID uuid.UUID, orderReference string, actorUserID *uuid.UUID, detail string) error
	RecordStatusChange(tx database.Tx, eventID, orderID uuid.UUID, orderReference, from, to string, actorUserID *uuid.UUID) error
	RecordReceiptPrinted(tx database.Tx, eventID, orderID uuid.UUID, orderReference string, actorUserID *uuid.UUID, printCount int) error
	RecordCancelled(tx database.Tx, eventID, orderID uuid.UUID, orderReference, from string, actorUserID *uuid.UUID, reason *string, approvedBy string) error
	Search(q Query) (*Page, error)
	ForEvent(eventID uuid.UUID) ([]AuditEntryResponse, error)
	ForOrder(orderID uuid.UUID) ([]AuditEntryResponse, error)
}
