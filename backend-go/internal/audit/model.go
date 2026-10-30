package audit

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
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

// OrderAuditLog is an append-only audit trail of order activity — rows are only ever inserted. A
// Postgres BEFORE UPDATE/DELETE trigger (installed on startup) enforces this at the DB level too.
type OrderAuditLog struct {
	ID             uuid.UUID   `gorm:"type:uuid;primaryKey"`
	EventID        uuid.UUID   `gorm:"column:event_id;type:uuid;not null;index:idx_audit_event;index:idx_audit_event_created,priority:1"`
	OrderID        uuid.UUID   `gorm:"column:order_id;type:uuid;not null;index:idx_audit_order"`
	OrderReference string      `gorm:"column:order_reference;not null"`
	Action         AuditAction `gorm:"type:varchar(20);not null"`
	ActorUserID    *uuid.UUID  `gorm:"column:actor_user_id;type:uuid"`
	ActorName      string      `gorm:"column:actor_name;not null"`
	FromStatus     *string     `gorm:"column:from_status"`
	ToStatus       *string     `gorm:"column:to_status"`
	Detail         *string     `gorm:"column:detail"`
	CreatedAt      time.Time   `gorm:"column:created_at;not null;index:idx_audit_event_created,priority:2,sort:desc"`
}

func (OrderAuditLog) TableName() string { return "order_audit_log" }

func (a *OrderAuditLog) BeforeCreate(_ *gorm.DB) error {
	if a.ID == uuid.Nil {
		a.ID = uuid.New()
	}
	if a.CreatedAt.IsZero() {
		a.CreatedAt = time.Now().UTC()
	}
	return nil
}

// AuditEntryResponse is the read view (per-event / per-order history).
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

func (a *OrderAuditLog) ToResponse() AuditEntryResponse {
	return AuditEntryResponse{
		ID: a.ID, OrderID: a.OrderID, OrderReference: a.OrderReference, Action: a.Action,
		ActorName: a.ActorName, FromStatus: a.FromStatus, ToStatus: a.ToStatus, Detail: a.Detail, At: a.CreatedAt,
	}
}
