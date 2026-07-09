package database

import (
	"errors"
	"os"
	"path/filepath"
	"strings"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"gorm.io/gorm"

	"github.com/krnskrll/event-mpos/internal/audit"
	"github.com/krnskrll/event-mpos/internal/auth"
	"github.com/krnskrll/event-mpos/internal/event"
	"github.com/krnskrll/event-mpos/internal/member"
	membersync "github.com/krnskrll/event-mpos/internal/member/sync"
	"github.com/krnskrll/event-mpos/internal/order"
	"github.com/krnskrll/event-mpos/internal/product"
	"github.com/krnskrll/event-mpos/internal/product/erp"
	"github.com/krnskrll/event-mpos/internal/promotion"
	"github.com/krnskrll/event-mpos/internal/receipt"
	"github.com/krnskrll/event-mpos/internal/stock"
	"github.com/krnskrll/event-mpos/internal/syncconfig"
	"github.com/krnskrll/event-mpos/internal/user"
)

// Models is the full set of GORM entities. Each phase appends its models here. This single registry
// drives AutoMigrate during local development and, later, the golang-migrate init file that is
// pg_dump'd from a throwaway DB (Option A) — so the two can never drift.
func Models() []interface{} {
	return []interface{}{
		&user.User{},
		&auth.RefreshToken{},
		&event.Event{},
		&event.BaEventAssignment{},
		&product.Product{},
		&product.ProductBarcode{},
		&product.ProductVariant{},
		&product.EventProduct{},
		&erp.SyncStatus{},
		&member.Member{},
		&membersync.Status{},
		&stock.EventStockAllocation{},
		&promotion.Promotion{},
		&order.Order{},
		&order.OrderItem{},
		&order.OrderDailySequence{},
		&receipt.ReceiptLog{},
		&audit.OrderAuditLog{},
		&syncconfig.SyncConfig{},
	}
}

// AutoMigrate reconciles the schema with the registered models (dev fallback, equivalent to the
// Kotlin ddl-auto=update). Used only when no golang-migrate files are present.
func AutoMigrate(db *gorm.DB) error {
	return db.AutoMigrate(Models()...)
}

// HasMigrations reports whether the given dir contains golang-migrate files.
func HasMigrations(dir string) bool {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return false
	}
	for _, e := range entries {
		if strings.HasSuffix(e.Name(), ".up.sql") {
			return true
		}
	}
	return false
}

// RunMigrations applies the golang-migrate files in dir against the DB. This is the production
// source of truth (Option A); the init file was generated from the GORM structs so the two can't
// drift. Idempotent — a no-op when the DB is already up to date.
func RunMigrations(dbURL, dir string) error {
	abs, err := filepath.Abs(dir)
	if err != nil {
		return err
	}
	m, err := migrate.New("file://"+abs, dbURL)
	if err != nil {
		return err
	}
	defer m.Close()
	if err := m.Up(); err != nil && !errors.Is(err, migrate.ErrNoChange) {
		return err
	}
	return nil
}
