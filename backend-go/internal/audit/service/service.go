package service

import (
	"log"
	"strconv"
	"strings"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/krnskrll/event-mpos/internal/audit/iface"
	"github.com/krnskrll/event-mpos/internal/audit/repository"
	"github.com/krnskrll/event-mpos/internal/platform/database"
	useriface "github.com/krnskrll/event-mpos/internal/user/iface"
)

// Service records order activity into the append-only OrderAuditLog, resolving the actor's name at
// write time. Purely additive — never updates or deletes.
type Service struct {
	db    *gorm.DB
	users useriface.Users
}

var _ iface.Audit = (*Service)(nil)

func NewService(db *gorm.DB, users useriface.Users) *Service {
	return &Service{db: db, users: users}
}

// actorName resolves the actor's display name through the caller's transaction. It must not use a
// separate pool connection: the caller may hold stock row locks, and if every pooled connection is
// held by a transaction waiting on those locks, a second connection never frees up (deadlock).
func (s *Service) actorName(tx *gorm.DB, userID *uuid.UUID) string {
	if userID == nil {
		return "system"
	}
	name, err := s.users.NameOf(tx, *userID)
	if err != nil {
		return "system"
	}
	return name
}

// RecordOrderCreated writes an ORDER_CREATED entry using the provided tx (same transaction as the
// order create).
func (s *Service) RecordOrderCreated(tx database.Tx, eventID, orderID uuid.UUID, orderReference string, actorUserID *uuid.UUID, detail string) error {
	to := "NEW"
	return tx.Create(&repository.OrderAuditLog{
		EventID: eventID, OrderID: orderID, OrderReference: orderReference,
		Action: iface.ActionOrderCreated, ActorUserID: actorUserID, ActorName: s.actorName(tx, actorUserID),
		ToStatus: &to, Detail: &detail,
	}).Error
}

// RecordStatusChange writes a STATUS_CHANGED entry using the provided tx.
func (s *Service) RecordStatusChange(tx database.Tx, eventID, orderID uuid.UUID, orderReference, from, to string, actorUserID *uuid.UUID) error {
	detail := "Status " + from + " → " + to
	return tx.Create(&repository.OrderAuditLog{
		EventID: eventID, OrderID: orderID, OrderReference: orderReference,
		Action: iface.ActionStatusChanged, ActorUserID: actorUserID, ActorName: s.actorName(tx, actorUserID),
		FromStatus: &from, ToStatus: &to, Detail: &detail,
	}).Error
}

// RecordReceiptPrinted writes a RECEIPT_PRINTED entry using the provided tx. printCount is the
// order's print count after this print (1 = first print, 2+ = reprints).
func (s *Service) RecordReceiptPrinted(tx database.Tx, eventID, orderID uuid.UUID, orderReference string, actorUserID *uuid.UUID, printCount int) error {
	detail := "Receipt printed"
	if printCount > 1 {
		detail = "Receipt reprinted (print #" + strconv.Itoa(printCount) + ")"
	}
	return tx.Create(&repository.OrderAuditLog{
		EventID: eventID, OrderID: orderID, OrderReference: orderReference,
		Action: iface.ActionReceiptPrinted, ActorUserID: actorUserID, ActorName: s.actorName(tx, actorUserID),
		Detail: &detail,
	}).Error
}

// RecordCancelled writes an ORDER_CANCELLED entry (with the optional reason) using the provided tx.
func (s *Service) RecordCancelled(tx database.Tx, eventID, orderID uuid.UUID, orderReference, from string, actorUserID *uuid.UUID, reason *string) error {
	to := "CANCELLED"
	detail := "Order cancelled"
	if reason != nil && strings.TrimSpace(*reason) != "" {
		detail += ": " + strings.TrimSpace(*reason)
	}
	return tx.Create(&repository.OrderAuditLog{
		EventID: eventID, OrderID: orderID, OrderReference: orderReference,
		Action: iface.ActionOrderCancelled, ActorUserID: actorUserID, ActorName: s.actorName(tx, actorUserID),
		FromStatus: &from, ToStatus: &to, Detail: &detail,
	}).Error
}

// Search returns one page of an event's audit log, newest first.
func (s *Service) Search(q iface.Query) (*iface.Page, error) {
	tx := s.db.Model(&repository.OrderAuditLog{}).Where("event_id = ?", q.EventID)
	if q.OrderID != nil {
		tx = tx.Where("order_id = ?", *q.OrderID)
	}
	if q.ActorID != nil {
		tx = tx.Where("actor_user_id = ?", *q.ActorID)
	}
	if q.Action != nil {
		tx = tx.Where("action = ?", *q.Action)
	}
	var total int64
	if err := tx.Count(&total).Error; err != nil {
		return nil, err
	}
	var rows []repository.OrderAuditLog
	if err := tx.Order("created_at DESC").Order("id DESC").
		Offset(q.Page * q.Size).Limit(q.Size).Find(&rows).Error; err != nil {
		return nil, err
	}
	return &iface.Page{Items: mapEntries(rows), Page: q.Page, Size: q.Size, Total: total}, nil
}

func (s *Service) ForEvent(eventID uuid.UUID) ([]iface.AuditEntryResponse, error) {
	var rows []repository.OrderAuditLog
	if err := s.db.Where("event_id = ?", eventID).Order("created_at DESC").Find(&rows).Error; err != nil {
		return nil, err
	}
	return mapEntries(rows), nil
}

func (s *Service) ForOrder(orderID uuid.UUID) ([]iface.AuditEntryResponse, error) {
	var rows []repository.OrderAuditLog
	if err := s.db.Where("order_id = ?", orderID).Order("created_at ASC").Find(&rows).Error; err != nil {
		return nil, err
	}
	return mapEntries(rows), nil
}

func mapEntries(rows []repository.OrderAuditLog) []iface.AuditEntryResponse {
	out := make([]iface.AuditEntryResponse, 0, len(rows))
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
