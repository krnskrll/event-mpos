// Package seed loads dev-only starter data on boot when SEED_ENABLED=true. It is idempotent: it
// only seeds when the demo event EV01 does not exist yet.
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
	membersrc "github.com/krnskrll/event-mpos/internal/integration/membersrc"
	"github.com/krnskrll/event-mpos/internal/platform/common"
	"github.com/krnskrll/event-mpos/internal/platform/tenant"
	productiface "github.com/krnskrll/event-mpos/internal/product/iface"
	productrepo "github.com/krnskrll/event-mpos/internal/product/repository"
	promotioniface "github.com/krnskrll/event-mpos/internal/promotion/iface"
	promotionsvc "github.com/krnskrll/event-mpos/internal/promotion/service"
	staffiface "github.com/krnskrll/event-mpos/internal/staff/iface"
	staffsvc "github.com/krnskrll/event-mpos/internal/staff/service"
)

// Seeded credentials (CHANGE before any real deployment):
//
//	EV01 staff: Bella (KASIR, PIN 1234), Sari (SUPERVISOR, PIN 9999)
//	EV01 device "Kasir 1": its token and a fresh pairing code are printed in the log.
func Run(db *gorm.DB, memberSync *membersrc.Service, promoSvc *promotionsvc.Service, staff *staffsvc.Service, devices *devicesvc.Service) error {
	var count int64
	if err := db.Model(&eventrepo.Event{}).Where("store_id = ? AND code = ?", tenant.DefaultStoreID, "EV01").Count(&count).Error; err != nil {
		return err
	}
	if count > 0 {
		log.Println("seed skipped — EV01 already exists")
		return nil
	}
	log.Println("WARNING: seeding DEV data with default PINs — do NOT use in production")

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

	// Demo catalogue; outside dev, items come from Bhoomy's Katalog POS. The lipstick has shade variants.
	str := func(v string) *string { return &v }
	bc := func(codes ...string) []productrepo.ProductBarcode {
		out := make([]productrepo.ProductBarcode, 0, len(codes))
		for _, c := range codes {
			out = append(out, productrepo.ProductBarcode{Barcode: c})
		}
		return out
	}
	var lipstick productrepo.Product
	for _, d := range []struct {
		sku, name, brand, price string
		barcodes                []productrepo.ProductBarcode
	}{
		{"SKU-1001", "Hydra Glow Serum 30ml", "LumièreCo", "129000", []productrepo.ProductBarcode{{Barcode: "8990001001", Label: str("single")}, {Barcode: "18990001001", Label: str("box of 6")}}},
		{"SKU-1002", "Velvet Matte Lipstick — Rose", "LumièreCo", "79000", bc("8990001002")},
		{"SKU-1003", "Daily Mineral Sunscreen SPF50", "DermaPure", "99000", []productrepo.ProductBarcode{{Barcode: "8990001003", Label: str("single")}, {Barcode: "28990001003", Label: str("twin pack")}}},
		{"SKU-1004", "Nourishing Hair Oil 100ml", "Botané", "65000", bc("8990001004")},
		{"SKU-1005", "Gentle Foaming Cleanser 150ml", "DermaPure", "55000", bc("8990001005")},
	} {
		item := productrepo.Product{StoreID: tenant.DefaultStoreID, SKU: d.sku, DisplayName: d.name, Brand: str(d.brand),
			Source: productiface.SourceManual, Status: productrepo.StatusActive, Barcodes: d.barcodes}
		item.SetPrice(decimal.RequireFromString(d.price))
		if d.sku == "SKU-1002" {
			plum := decimal.RequireFromString("89000")
			item.Variants = append(item.Variants,
				productrepo.ProductVariant{Name: "Rose", SKU: "SKU-1002-ROSE", SellInPOS: true},
				productrepo.ProductVariant{Name: "Coral", SKU: "SKU-1002-CORAL", SellInPOS: true},
				productrepo.ProductVariant{Name: "Plum (limited)", SKU: "SKU-1002-PLUM", Price: &plum, SellInPOS: true})
		}
		if err := db.Create(&item).Error; err != nil {
			return err
		}
		if d.sku == "SKU-1002" {
			lipstick = item
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
