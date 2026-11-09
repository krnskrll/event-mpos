package app_test

import (
	"net/http"
	"testing"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"

	eventiface "github.com/krnskrll/event-mpos/internal/event/iface"
	eventrepo "github.com/krnskrll/event-mpos/internal/event/repository"
	orderiface "github.com/krnskrll/event-mpos/internal/order/iface"
	"github.com/krnskrll/event-mpos/internal/platform/tenant"
	productiface "github.com/krnskrll/event-mpos/internal/product/iface"
	productrepo "github.com/krnskrll/event-mpos/internal/product/repository"
	promotioniface "github.com/krnskrll/event-mpos/internal/promotion/iface"
	"github.com/krnskrll/event-mpos/internal/testutil"
	useriface "github.com/krnskrll/event-mpos/internal/user/iface"
)

// Store B may reuse store A's codes; a store may not repeat its own.
func TestCodesAreUniquePerStore(t *testing.T) {
	a := testutil.NewApp(t)
	storeB := a.NewStore("Store B")
	sku, barcode, code, member := testutil.Unique("SKU-"), testutil.Unique("899"), testutil.Unique("E"), testutil.Unique("M-")
	inserts := map[string]func(store uuid.UUID) error{
		"SKU": func(store uuid.UUID) error {
			return a.DB.Create(&productrepo.Product{StoreID: store, SKU: sku, DisplayName: "Serum", Source: productiface.SourceManual, Status: productrepo.StatusActive}).Error
		},
		"barcode": func(store uuid.UUID) error {
			return a.DB.Create(&productrepo.ProductBarcode{ItemID: a.ProductIn(store, "1000").ID, Barcode: barcode}).Error
		},
		"location code": func(store uuid.UUID) error {
			return a.DB.Create(&eventrepo.Event{StoreID: store, Type: eventiface.TypeEvent, Code: code, Name: "Bazaar", Status: eventiface.StatusDraft}).Error
		},
		"member code": func(store uuid.UUID) error {
			return a.DB.Exec(`INSERT INTO members (id, member_code, qr_code, name, tier, active, updated_at, store_id)
				VALUES (?, ?, ?, 'M', 'GOLD', true, now(), ?)`, uuid.New(), member, member, store).Error
		},
	}
	for name, insert := range inserts {
		if err := insert(tenant.DefaultStoreID); err != nil {
			t.Fatalf("%s: %v", name, err)
		}
		if err := insert(storeB); err != nil {
			t.Errorf("%s reused by another store: %v", name, err)
		}
		if insert(storeB) == nil {
			t.Errorf("%s repeated within one store", name)
		}
	}
}

// Two stores sell the same barcode; each cashier scans their own store's item.
func TestBarcodeLookupStaysInTheStore(t *testing.T) {
	a := testutil.NewApp(t)
	const pw = "Passw0rd!"
	barcode := testutil.Unique("899")
	scan := func(store uuid.UUID) (uuid.UUID, string) {
		p := a.ProductIn(store, "5000")
		if err := a.DB.Create(&productrepo.ProductBarcode{ItemID: p.ID, Barcode: barcode}).Error; err != nil {
			t.Fatal(err)
		}
		ev := a.ActiveEventIn(store)
		if err := a.DB.Create(&productrepo.LocationItem{LocationID: ev.ID, ItemVariantID: p.Default().ID, Active: true}).Error; err != nil {
			t.Fatal(err)
		}
		ba := a.CreateUserIn(store, useriface.RoleBA, pw)
		if _, err := a.Events.AssignBaToEvent(store, ba.ID, ev.ID); err != nil {
			t.Fatal(err)
		}
		rec := a.Do(http.MethodGet, "/api/me/events/"+ev.ID.String()+"/products/by-barcode/"+barcode, nil, login(t, a, ba.Email, pw).AccessToken)
		if rec.Code != http.StatusOK {
			t.Fatalf("scan: %d %s", rec.Code, rec.Body.String())
		}
		return p.ID, testutil.Decode[map[string]any](t, rec)["id"].(string)
	}
	for _, store := range []uuid.UUID{tenant.DefaultStoreID, a.NewStore("Store B")} {
		if want, got := scan(store); got != want.String() {
			t.Errorf("store %s scanned %s, want its own item %s", store, got, want)
		}
	}
}

func TestMoneyIsWholeRupiah(t *testing.T) {
	a := testutil.NewApp(t)
	ev := a.ActiveEvent()
	half := decimal.RequireFromString("100.5")
	if _, err := a.Promotions.Create(ev.StoreID, promotioniface.PromotionRequest{LocationID: &ev.ID, Name: "x", Type: promotioniface.TypeFixed, AmountOff: &half}); testutil.StatusOf(err) != http.StatusBadRequest {
		t.Errorf("promotion amount with decimals: got %v, want 400", err)
	}

	// 7% of Rp15.550 is Rp1.088,5 → Rp1.089 off, total Rp14.461.
	p := a.Product("15550")
	a.Allocate(ev.ID, p.ID, 5)
	pct := decimal.RequireFromString("7")
	if _, err := a.Promotions.Create(ev.StoreID, promotioniface.PromotionRequest{LocationID: &ev.ID, Name: "7%", Type: promotioniface.TypePercentage, PercentOff: &pct}); err != nil {
		t.Fatal(err)
	}
	order := sellOne(t, a, ev.ID, p.ID, nil)
	if order.Subtotal.String() != "15550" || order.Discount.String() != "1089" || order.Total.String() != "14461" {
		t.Errorf("totals %s - %s = %s, want 15550 - 1089 = 14461", order.Subtotal, order.Discount, order.Total)
	}
}

// Every item has one default variant carrying its price; real variants without a price follow it.
func TestDefaultVariantCarriesThePrice(t *testing.T) {
	a := testutil.NewApp(t)
	id := a.Product("10000").ID
	var defaults []productrepo.ProductVariant
	a.DB.Where("item_id = ? AND is_default", id).Find(&defaults)
	if len(defaults) != 1 || defaults[0].Price.String() != "10000" || defaults[0].StoreID != tenant.DefaultStoreID {
		t.Fatalf("default variants: %+v", defaults)
	}

	shade := productrepo.ProductVariant{ItemID: id, Name: "Rose", SKU: testutil.Unique("SKU-"), SellInPOS: true}
	if err := a.DB.Create(&shade).Error; err != nil {
		t.Fatal(err)
	}
	ev := a.ActiveEvent()
	a.Allocate(ev.ID, id, 5)
	if line := sellOne(t, a, ev.ID, id, nil).Items[0]; line.UnitPrice.String() != "10000" || line.VariantName != nil {
		t.Errorf("line without a variant: %+v, want the default price and no variant", line)
	}
	if line := sellOne(t, a, ev.ID, id, &shade.ID).Items[0]; line.UnitPrice.String() != "10000" || line.VariantName == nil || *line.VariantName != "Rose" {
		t.Errorf("line with a variant: %+v", line)
	}
}

func sellOne(t *testing.T, a *testutil.TestApp, eventID, productID uuid.UUID, variantID *uuid.UUID) *orderiface.OrderConfirmation {
	t.Helper()
	ba := a.CreateUser(useriface.RoleBA, "Passw0rd!")
	o, err := a.Orders.Submit(eventID, ba.ID, sellRequest(productID, variantID))
	if err != nil {
		t.Fatal(err)
	}
	return o
}

func sellRequest(productID uuid.UUID, variantID *uuid.UUID) orderiface.SubmitOrderRequest {
	return orderiface.SubmitOrderRequest{
		IdempotencyKey: testutil.Unique("cs-"), Customer: orderiface.CustomerInfo{Name: "C"},
		Lines: []orderiface.OrderLineRequest{{ProductID: productID, VariantID: variantID, Qty: 1}},
	}
}
