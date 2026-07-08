package database

import (
	"gorm.io/gorm"

	"github.com/krnskrll/event-mpos/internal/auth"
	"github.com/krnskrll/event-mpos/internal/event"
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
	}
}

// AutoMigrate reconciles the schema with the registered models (dev/local behaviour, equivalent to
// the Kotlin ddl-auto=update). Production deploys run the generated golang-migrate SQL instead.
func AutoMigrate(db *gorm.DB) error {
	return db.AutoMigrate(Models()...)
}
