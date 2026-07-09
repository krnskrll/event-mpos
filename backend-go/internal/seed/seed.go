// Package seed loads dev-only starter data on boot when SEED_ENABLED=true. It is idempotent: it
// only seeds when the users table is empty. Expanded per phase; Phase 1 seeds the three login users.
package seed

import (
	"log"
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	"gorm.io/gorm"

	"github.com/krnskrll/event-mpos/internal/common"
	"github.com/krnskrll/event-mpos/internal/event"
	membersync "github.com/krnskrll/event-mpos/internal/member/sync"
	"github.com/krnskrll/event-mpos/internal/product"
	"github.com/krnskrll/event-mpos/internal/product/erp"
	"github.com/krnskrll/event-mpos/internal/security"
	"github.com/krnskrll/event-mpos/internal/user"
)

// Seeded credentials (CHANGE before any real deployment):
//   admin@eventmpos.local / Admin123!  (ADMIN)
//   ba@eventmpos.local    / Ba123456!  (BA)
//   wh@eventmpos.local    / Wh123456!  (WAREHOUSE)
func Run(db *gorm.DB, erpSync *erp.SyncService, memberSync *membersync.Service) error {
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

	// Pull the initial catalogue from the (mock) ERP and make every product available for EV01.
	erpSync.Sync()
	var products []product.Product
	if err := db.Find(&products).Error; err != nil {
		return err
	}
	for i := range products {
		if err := db.Create(&product.EventProduct{EventID: ev.ID, ProductID: products[i].ID, Available: true}).Error; err != nil {
			return err
		}
	}

	// Give the lipstick a few shade variants (one with a price override) to exercise variants.
	var lipstick product.Product
	if err := db.Where("sku = ?", "SKU-1002").First(&lipstick).Error; err == nil {
		plum := decimal.RequireFromString("89000")
		variants := []product.ProductVariant{
			{ProductID: lipstick.ID, Name: "Rose", SKU: "SKU-1002-ROSE", Active: true},
			{ProductID: lipstick.ID, Name: "Coral", SKU: "SKU-1002-CORAL", Active: true},
			{ProductID: lipstick.ID, Name: "Plum (limited)", SKU: "SKU-1002-PLUM", PriceOverride: &plum, Active: true},
		}
		for i := range variants {
			if err := db.Create(&variants[i]).Error; err != nil {
				return err
			}
		}
	}

	// Pull the member directory from the (mock) external member system.
	memberSync.Sync()
	return nil
}
