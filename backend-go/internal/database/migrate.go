package database

import (
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

// AutoMigrate reconciles the schema with the registered models (dev/local behaviour, equivalent to
// the Kotlin ddl-auto=update). Production deploys run the generated golang-migrate SQL instead.
func AutoMigrate(db *gorm.DB) error {
	return db.AutoMigrate(Models()...)
}
