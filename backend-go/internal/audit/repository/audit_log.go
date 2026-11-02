package repository

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/krnskrll/event-mpos/internal/audit/iface"
)

// OrderAuditLog is an append-only audit trail of order activity — rows are only ever inserted. A
// Postgres BEFORE UPDATE/DELETE trigger (installed on startup) enforces this at the DB level too.
type OrderAuditLog struct {
	ID             uuid.UUID         `gorm:"type:uuid;primaryKey"`
	EventID        uuid.UUID         `gorm:"column:event_id;type:uuid;not null;index:idx_audit_event;index:idx_audit_event_created,priority:1"`
	OrderID        uuid.UUID         `gorm:"column:order_id;type:uuid;not null;index:idx_audit_order"`
	OrderReference string            `gorm:"column:order_reference;not null"`
	Action         iface.AuditAction `gorm:"type:varchar(20);not null"`
	ActorUserID    *uuid.UUID        `gorm:"column:actor_user_id;type:uuid"`
	ActorName      string            `gorm:"column:actor_name;not null"`
	FromStatus     *string           `gorm:"column:from_status"`
	ToStatus       *string           `gorm:"column:to_status"`
	Detail         *string           `gorm:"column:detail"`
	CreatedAt      time.Time         `gorm:"column:created_at;not null;index:idx_audit_event_created,priority:2,sort:desc"`
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
func (a *OrderAuditLog) ToResponse() iface.AuditEntryResponse {
	return iface.AuditEntryResponse{
		ID: a.ID, OrderID: a.OrderID, OrderReference: a.OrderReference, Action: a.Action,
		ActorName: a.ActorName, FromStatus: a.FromStatus, ToStatus: a.ToStatus, Detail: a.Detail, At: a.CreatedAt,
	}
}
