// Package seed loads dev-only starter data on boot when SEED_ENABLED=true. It is idempotent: it
// only seeds when the users table is empty. Expanded per phase; Phase 1 seeds the three login users.
package seed

import (
	"log"

	"gorm.io/gorm"

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
	for _, u := range seedUsers {
		hash, err := security.HashPassword(u.password)
		if err != nil {
			return err
		}
		if err := db.Create(&user.User{
			Email:        u.email,
			PasswordHash: hash,
			FullName:     u.name,
			Role:         u.role,
			Active:       true,
		}).Error; err != nil {
			return err
		}
	}
	return nil
}
