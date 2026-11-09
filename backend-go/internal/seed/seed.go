// Package seed loads dev-only starter data on boot when SEED_ENABLED=true. It is idempotent: it
// only seeds when the users table is empty.
package seed

import (
	"log"
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	"gorm.io/gorm"

	devicesvc "github.com/krnskrll/event-mpos/internal/device/service"
	eventiface "github.com/krnskrll/event-mpos/internal/event/iface"
	eventrepo "github.com/krnskrll/event-mpos/internal/event/repository"
	"github.com/krnskrll/event-mpos/internal/integration/erp"
	membersrc "github.com/krnskrll/event-mpos/internal/integration/membersrc"
	"github.com/krnskrll/event-mpos/internal/platform/common"
	"github.com/krnskrll/event-mpos/internal/platform/security"
	"github.com/krnskrll/event-mpos/internal/platform/tenant"
	productrepo "github.com/krnskrll/event-mpos/internal/product/repository"
	promotioniface "github.com/krnskrll/event-mpos/internal/promotion/iface"
	promotionsvc "github.com/krnskrll/event-mpos/internal/promotion/service"
	staffiface "github.com/krnskrll/event-mpos/internal/staff/iface"
	staffsvc "github.com/krnskrll/event-mpos/internal/staff/service"
	stocksvc "github.com/krnskrll/event-mpos/internal/stock/service"
	useriface "github.com/krnskrll/event-mpos/internal/user/iface"
	userrepo "github.com/krnskrll/event-mpos/internal/user/repository"
)

// Seeded credentials (CHANGE before any real deployment):
//
//	admin@eventmpos.local / Admin123!  (ADMIN, old dashboard)
//	EV01 staff: Bella (KASIR, PIN 1234), Sari (SUPERVISOR, PIN 9999)
//	EV01 device "Kasir 1": its token and a fresh pairing code are printed in the log.
func Run(db *gorm.DB, erpSync *erp.SyncService, memberSync *membersrc.Service, stockSvc *stocksvc.Service, promoSvc *promotionsvc.Service, staff *staffsvc.Service, devices *devicesvc.Service) error {
	var count int64
	if err := db.Model(&userrepo.User{}).Count(&count).Error; err != nil {
		return err
	}
	if count > 0 {
		log.Println("seed skipped — users already exist")
		return nil
	}
	log.Println("WARNING: seeding DEV data with default credentials — do NOT use in production")

	hash, err := security.HashPassword("Admin123!")
	if err != nil {
		return err
	}
	if err := db.Create(&userrepo.User{StoreID: tenant.DefaultStoreID, Email: "admin@eventmpos.local", PasswordHash: hash,
		FullName: "System Admin", Role: useriface.RoleAdmin, Active: true}).Error; err != nil {
		return err
	}

	// EV01 — the active demo event, with a cashier, a supervisor and one paired device.
	start := common.NewDate(2026, time.July, 1)
	end := common.NewDate(2026, time.July, 14)
	ev := eventrepo.Event{
		StoreID:   tenant.DefaultStoreID,
		Type:      eventiface.TypeEvent,
		Code:      "EV01",
		Name:      "Mid-Year Beauty Roadshow",
		StartDate: &start,
		EndDate:   &end,
		Status:    eventiface.StatusActive,
	}
	if err := db.Create(&ev).Error; err != nil {
		return err
	}
	for _, s := range []struct {
		name, pin string
		role      staffiface.Role
	}{{"Bella", "1234", staffiface.RoleKasir}, {"Sari", "9999", staffiface.RoleSupervisor}} {
		if _, err := staff.Create(tenant.DefaultStoreID, staffiface.StaffRequest{Name: s.name, Role: s.role, Pin: &s.pin, LocationIDs: []uuid.UUID{ev.ID}}); err != nil {
			return err
		}
	}
	code, err := devices.NewCode(tenant.DefaultStoreID, ev.ID)
	if err != nil {
		return err
	}
	paired, err := devices.Pair(code.Code, "Dev seed")
	if err != nil {
		return err
	}
	if code, err = devices.NewCode(tenant.DefaultStoreID, ev.ID); err != nil {
		return err
	}
	log.Printf("DEV ONLY: EV01 device %q token: %s", paired.Device.Name, paired.Token)
	log.Printf("DEV ONLY: EV01 pairing code %s (valid until %s)", code.Code, code.ExpiresAt.Format(time.RFC3339))

	// Pull the initial catalogue from the (mock) ERP and make every product available for EV01.
	erpSync.Sync()

	// Give the lipstick a few shade variants (one with a price override) to exercise variants.
	var lipstick productrepo.Product
	if err := db.Where("sku = ?", "SKU-1002").First(&lipstick).Error; err == nil {
		plum := decimal.RequireFromString("89000")
		variants := []productrepo.ProductVariant{
			{ItemID: lipstick.ID, Name: "Rose", SKU: "SKU-1002-ROSE", SellInPOS: true},
			{ItemID: lipstick.ID, Name: "Coral", SKU: "SKU-1002-CORAL", SellInPOS: true},
			{ItemID: lipstick.ID, Name: "Plum (limited)", SKU: "SKU-1002-PLUM", Price: &plum, SellInPOS: true},
		}
		for i := range variants {
			if err := db.Create(&variants[i]).Error; err != nil {
				return err
			}
		}
	}
	var products []productrepo.Product
	if err := db.Preload("Variants").Find(&products).Error; err != nil {
		return err
	}
	for i := range products {
		for _, v := range products[i].Variants {
			if err := db.Create(&productrepo.LocationItem{LocationID: ev.ID, ItemVariantID: v.ID, Active: true}).Error; err != nil {
				return err
			}
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

	// Seed a few event-scoped promotions for EV01 (so the cart/quote has something to apply).
	pct := decimal.RequireFromString("10")
	off20k := decimal.RequireFromString("20000")
	min200k := decimal.RequireFromString("200000")
	off30k := decimal.RequireFromString("30000")
	pri1, pri2, pri3 := 1, 2, 3
	bundleQty := 3
	specific := promotioniface.ScopeSpecific
	if _, err := promoSvc.Create(tenant.DefaultStoreID, promotioniface.PromotionRequest{LocationID: &ev.ID,
		Name: "10% off everything", Type: promotioniface.TypePercentage, PercentOff: &pct, Priority: &pri1,
	}); err != nil {
		return err
	}
	if _, err := promoSvc.Create(tenant.DefaultStoreID, promotioniface.PromotionRequest{LocationID: &ev.ID,
		Name: "Rp20.000 off when you spend Rp200.000", Type: promotioniface.TypeFixed,
		AmountOff: &off20k, MinSubtotal: &min200k, Priority: &pri2,
	}); err != nil {
		return err
	}
	if lipstick.ID != uuid.Nil {
		if _, err := promoSvc.Create(tenant.DefaultStoreID, promotioniface.PromotionRequest{LocationID: &ev.ID,
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
