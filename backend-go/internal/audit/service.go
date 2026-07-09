package audit

import (
	"log"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/krnskrll/event-mpos/internal/user"
)

// Service records order activity into the append-only OrderAuditLog, resolving the actor's name at
// write time. Purely additive — never updates or deletes.
type Service struct {
	db    *gorm.DB
	users *user.Service
}

func NewService(db *gorm.DB, users *user.Service) *Service {
	return &Service{db: db, users: users}
}

func (s *Service) actorName(userID *uuid.UUID) string {
	if userID == nil {
		return "system"
	}
	u, err := s.users.FindByID(*userID)
	if err != nil || u == nil {
		return "system"
	}
	return u.FullName
}

// RecordOrderCreated writes an ORDER_CREATED entry using the provided tx (same transaction as the
// order create).
func (s *Service) RecordOrderCreated(tx *gorm.DB, eventID, orderID uuid.UUID, orderReference string, actorUserID *uuid.UUID) error {
	to := "NEW"
	detail := "Order created"
	return tx.Create(&OrderAuditLog{
		EventID: eventID, OrderID: orderID, OrderReference: orderReference,
		Action: ActionOrderCreated, ActorUserID: actorUserID, ActorName: s.actorName(actorUserID),
		ToStatus: &to, Detail: &detail,
	}).Error
}

// RecordStatusChange writes a STATUS_CHANGED entry using the provided tx.
func (s *Service) RecordStatusChange(tx *gorm.DB, eventID, orderID uuid.UUID, orderReference, from, to string, actorUserID *uuid.UUID) error {
	detail := "Status " + from + " → " + to
	return tx.Create(&OrderAuditLog{
		EventID: eventID, OrderID: orderID, OrderReference: orderReference,
		Action: ActionStatusChanged, ActorUserID: actorUserID, ActorName: s.actorName(actorUserID),
		FromStatus: &from, ToStatus: &to, Detail: &detail,
	}).Error
}

func (s *Service) ForEvent(eventID uuid.UUID) ([]AuditEntryResponse, error) {
	var rows []OrderAuditLog
	if err := s.db.Where("event_id = ?", eventID).Order("created_at DESC").Find(&rows).Error; err != nil {
		return nil, err
	}
	return mapEntries(rows), nil
}

func (s *Service) ForOrder(orderID uuid.UUID) ([]AuditEntryResponse, error) {
	var rows []OrderAuditLog
	if err := s.db.Where("order_id = ?", orderID).Order("created_at ASC").Find(&rows).Error; err != nil {
		return nil, err
	}
	return mapEntries(rows), nil
}

func mapEntries(rows []OrderAuditLog) []AuditEntryResponse {
	out := make([]AuditEntryResponse, 0, len(rows))
	for i := range rows {
		out = append(out, rows[i].ToResponse())
	}
	return out
}

// InstallAppendOnlyGuard installs a Postgres BEFORE UPDATE/DELETE trigger enforcing append-only at
// the DB level (in addition to code-level immutability). Skipped with a warning on non-Postgres.
func InstallAppendOnlyGuard(db *gorm.DB) {
	stmts := []string{
		`CREATE OR REPLACE FUNCTION reject_order_audit_mutation() RETURNS trigger AS $$
		 BEGIN RAISE EXCEPTION 'order_audit_log is append-only'; END;
		 $$ LANGUAGE plpgsql;`,
		`DROP TRIGGER IF EXISTS trg_order_audit_append_only ON order_audit_log;`,
		`CREATE TRIGGER trg_order_audit_append_only
		 BEFORE UPDATE OR DELETE ON order_audit_log
		 FOR EACH ROW EXECUTE FUNCTION reject_order_audit_mutation();`,
	}
	for _, s := range stmts {
		if err := db.Exec(s).Error; err != nil {
			log.Printf("append-only DB trigger not installed (%v); code-level immutability still applies", err)
			return
		}
	}
	log.Println("Installed append-only trigger on order_audit_log")
}
