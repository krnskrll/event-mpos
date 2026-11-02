// Package seed loads dev-only starter data on boot when SEED_ENABLED=true. It is idempotent: it
// only seeds when the users table is empty. Expanded per phase; Phase 1 seeds the three login users.
package seed

import (
	"log"
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	"gorm.io/gorm"

	eventiface "github.com/krnskrll/event-mpos/internal/event/iface"
	eventrepo "github.com/krnskrll/event-mpos/internal/event/repository"
	"github.com/krnskrll/event-mpos/internal/integration/erp"
	membersrc "github.com/krnskrll/event-mpos/internal/integration/membersrc"
	"github.com/krnskrll/event-mpos/internal/platform/common"
	"github.com/krnskrll/event-mpos/internal/platform/security"
	productrepo "github.com/krnskrll/event-mpos/internal/product/repository"
	promotioniface "github.com/krnskrll/event-mpos/internal/promotion/iface"
	promotionsvc "github.com/krnskrll/event-mpos/internal/promotion/service"
	"github.com/krnskrll/event-mpos/internal/stock"
	useriface "github.com/krnskrll/event-mpos/internal/user/iface"
	userrepo "github.com/krnskrll/event-mpos/internal/user/repository"
)

// Seeded credentials (CHANGE before any real deployment):
//
//	admin@eventmpos.local / Admin123!  (ADMIN)
//	ba@eventmpos.local    / Ba123456!  (BA)
//	wh@eventmpos.local    / Wh123456!  (WAREHOUSE)
func Run(db *gorm.DB, erpSync *erp.SyncService, memberSync *membersrc.Service, stockSvc *stock.Service, promoSvc *promotionsvc.Service) error {
	var count int64
	if err := db.Model(&userrepo.User{}).Count(&count).Error; err != nil {
		return err
	}
	if count > 0 {
		log.Println("seed skipped — users already exist")
		return nil
	}
	log.Println("WARNING: seeding DEV data with default credentials — do NOT use in production")

	seedUsers := []struct {
		email, password, name string
		role                  useriface.Role
	}{
		{"admin@eventmpos.local", "Admin123!", "System Admin", useriface.RoleAdmin},
		{"ba@eventmpos.local", "Ba123456!", "Bella Advisor", useriface.RoleBA},
		{"wh@eventmpos.local", "Wh123456!", "Warren House", useriface.RoleWarehouse},
	}
	var baID uuid.UUID
	for _, su := range seedUsers {
		hash, err := security.HashPassword(su.password)
		if err != nil {
			return err
		}
		u := userrepo.User{
			Email:        su.email,
			PasswordHash: hash,
			FullName:     su.name,
			Role:         su.role,
			Active:       true,
		}
		if err := db.Create(&u).Error; err != nil {
			return err
		}
		if su.role == useriface.RoleBA {
			baID = u.ID
		}
	}

	// EV01 — the active demo event, with the BA assigned to it.
	start := common.NewDate(2026, time.July, 1)
	end := common.NewDate(2026, time.July, 14)
	ev := eventrepo.Event{
		Code:      "EV01",
		Name:      "Mid-Year Beauty Roadshow",
		StartDate: &start,
		EndDate:   &end,
		Status:    eventiface.StatusActive,
	}
	if err := db.Create(&ev).Error; err != nil {
		return err
	}
	if err := db.Create(&eventrepo.BaEventAssignment{UserID: baID, EventID: ev.ID, Active: true}).Error; err != nil {
		return err
	}

	// Pull the initial catalogue from the (mock) ERP and make every product available for EV01.
	erpSync.Sync()
	var products []productrepo.Product
	if err := db.Find(&products).Error; err != nil {
		return err
	}
	for i := range products {
		if err := db.Create(&productrepo.EventProduct{EventID: ev.ID, ProductID: products[i].ID, Available: true}).Error; err != nil {
			return err
		}
		// Allocate event stock; SKU-1004 gets a small amount to demo the low-stock warning.
		qty := min(products[i].WarehouseStock, 60)
		if products[i].SKU == "SKU-1004" {
			qty = 8
		}
		if err := stockSvc.ForceAllocate(ev.ID, products[i].ID, qty); err != nil {
			return err
		}
	}

	// Give the lipstick a few shade variants (one with a price override) to exercise variants.
	var lipstick productrepo.Product
	if err := db.Where("sku = ?", "SKU-1002").First(&lipstick).Error; err == nil {
		plum := decimal.RequireFromString("89000")
		variants := []productrepo.ProductVariant{
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

	// Seed a few event-scoped promotions for EV01 (so the cart/quote has something to apply).
	pct := decimal.RequireFromString("10")
	off20k := decimal.RequireFromString("20000")
	min200k := decimal.RequireFromString("200000")
	off30k := decimal.RequireFromString("30000")
	pri1, pri2, pri3 := 1, 2, 3
	bundleQty := 3
	specific := promotioniface.ScopeSpecific
	if _, err := promoSvc.Create(ev.ID, promotioniface.PromotionRequest{
		Name: "10% off everything", Type: promotioniface.TypePercentage, PercentOff: &pct, Priority: &pri1,
	}); err != nil {
		return err
	}
	if _, err := promoSvc.Create(ev.ID, promotioniface.PromotionRequest{
		Name: "Rp20.000 off when you spend Rp200.000", Type: promotioniface.TypeFixed,
		AmountOff: &off20k, MinSubtotal: &min200k, Priority: &pri2,
	}); err != nil {
		return err
	}
	if lipstick.ID != uuid.Nil {
		if _, err := promoSvc.Create(ev.ID, promotioniface.PromotionRequest{
			Name: "Lipstick bundle: buy 3, Rp30.000 off", Type: promotioniface.TypeBundle,
			AmountOff: &off30k, Scope: &specific, ProductIDs: []uuid.UUID{lipstick.ID},
			MinQty: &bundleQty, Priority: &pri3,
		}); err != nil {
			return err
		}
	}

	// Pull the member directory from the (mock) external member system.
	memberSync.Sync()
	return nil
}
