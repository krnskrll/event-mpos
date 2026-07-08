// Package seed loads dev-only starter data on boot when SEED_ENABLED=true. It is idempotent: it
// only seeds when the users table is empty. Expanded per phase; Phase 1 seeds the three login users.
package seed

import (
	"log"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/krnskrll/event-mpos/internal/common"
	"github.com/krnskrll/event-mpos/internal/event"
	"github.com/krnskrll/event-mpos/internal/security"
	"github.com/krnskrll/event-mpos/internal/user"
)

// Seeded credentials (CHANGE before any real deployment):
//   admin@eventmpos.local / Admin123!  (ADMIN)
//   ba@eventmpos.local    / Ba123456!  (BA)
//   wh@eventmpos.local    / Wh123456!  (WAREHOUSE)
func Run(db *gorm.DB) error {
	var count int64
	if err := db.Model(&user.User{}).Count(&count).Error; err != nil {
		return err
	}
	if count > 0 {
		log.Println("seed skipped — users already exist")
		return nil
	}
	log.Println("WARNING: seeding DEV data with default credentials — do NOT use in production")

	seedUsers := []struct {
		email, password, name string
		role                  user.Role
	}{
		{"admin@eventmpos.local", "Admin123!", "System Admin", user.RoleAdmin},
		{"ba@eventmpos.local", "Ba123456!", "Bella Advisor", user.RoleBA},
		{"wh@eventmpos.local", "Wh123456!", "Warren House", user.RoleWarehouse},
	}
	var baID uuid.UUID
	for _, su := range seedUsers {
		hash, err := security.HashPassword(su.password)
		if err != nil {
			return err
		}
		u := user.User{
			Email:        su.email,
			PasswordHash: hash,
			FullName:     su.name,
			Role:         su.role,
			Active:       true,
		}
		if err := db.Create(&u).Error; err != nil {
			return err
		}
		if su.role == user.RoleBA {
			baID = u.ID
		}
	}

	// EV01 — the active demo event, with the BA assigned to it.
	start := common.NewDate(2026, time.July, 1)
	end := common.NewDate(2026, time.July, 14)
	ev := event.Event{
		Code:      "EV01",
		Name:      "Mid-Year Beauty Roadshow",
		StartDate: &start,
		EndDate:   &end,
		Status:    event.StatusActive,
	}
	if err := db.Create(&ev).Error; err != nil {
		return err
	}
	if err := db.Create(&event.BaEventAssignment{UserID: baID, EventID: ev.ID, Active: true}).Error; err != nil {
		return err
	}
	return nil
}
