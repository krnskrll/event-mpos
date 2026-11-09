package app_test

import (
	"net/http"
	"testing"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"

	orderiface "github.com/krnskrll/event-mpos/internal/order/iface"
	"github.com/krnskrll/event-mpos/internal/platform/tenant"
	productrepo "github.com/krnskrll/event-mpos/internal/product/repository"
	promotioniface "github.com/krnskrll/event-mpos/internal/promotion/iface"
	"github.com/krnskrll/event-mpos/internal/testutil"
	useriface "github.com/krnskrll/event-mpos/internal/user/iface"
)

// Store B may reuse store A's codes; a store may not repeat its own.
func TestCodesAreUniquePerStore(t *testing.T) {
	a := testutil.NewApp(t)
	const pw = "Passw0rd!"
	tokA := login(t, a, a.CreateUser(useriface.RoleAdmin, pw).Email, pw).AccessToken
	storeB := a.NewStore("Store B")
	tokB := login(t, a, a.CreateUserIn(storeB, useriface.RoleAdmin, pw).Email, pw).AccessToken

	barcode := testutil.Unique("899")
	product := map[string]any{"sku": testutil.Unique("SKU-"), "name": "Serum", "price": 10000, "barcodes": []map[string]string{{"barcode": barcode}}}
	event := map[string]any{"code": testutil.Unique("E"), "name": "Bazaar"}
	for name, tok := range map[string]string{"store A": tokA, "store B": tokB} {
		if rec := a.Do(http.MethodPost, "/api/admin/products", product, tok); rec.Code != http.StatusCreated {
			t.Fatalf("%s product: %d %s", name, rec.Code, rec.Body.String())
		}
		if rec := a.Do(http.MethodPost, "/api/admin/events", event, tok); rec.Code != http.StatusCreated {
			t.Fatalf("%s event: %d %s", name, rec.Code, rec.Body.String())
		}
	}

	sameBarcode := map[string]any{"sku": testutil.Unique("SKU-"), "name": "Toner", "price": 1, "barcodes": []map[string]string{{"barcode": barcode}}}
	for name, rec := range map[string]int{
		"same SKU":        a.Do(http.MethodPost, "/api/admin/products", product, tokA).Code,
		"same barcode":    a.Do(http.MethodPost, "/api/admin/products", sameBarcode, tokA).Code,
		"same event code": a.Do(http.MethodPost, "/api/admin/events", event, tokA).Code,
	} {
		if rec != http.StatusConflict {
			t.Errorf("%s in one store: got %d, want 409", name, rec)
		}
	}

	code := testutil.Unique("M-")
	addMember := func(store uuid.UUID) error {
		return a.DB.Exec(`INSERT INTO members (id, member_code, qr_code, name, tier, active, updated_at, store_id)
			VALUES (?, ?, ?, 'M', 'GOLD', true, now(), ?)`, uuid.New(), code, code, store).Error
	}
	if err := addMember(tenant.DefaultStoreID); err != nil {
		t.Fatal(err)
	}
	if err := addMember(storeB); err != nil {
		t.Errorf("member code reused by another store: %v", err)
	}
	if addMember(storeB) == nil {
		t.Error("member code repeated within one store")
	}
}

// Two stores sell the same barcode; each cashier scans their own store's item.
func TestBarcodeLookupStaysInTheStore(t *testing.T) {
	a := testutil.NewApp(t)
	const pw = "Passw0rd!"
	barcode := testutil.Unique("899")
	scan := func(store uuid.UUID) (string, string) {
		admin := login(t, a, a.CreateUserIn(store, useriface.RoleAdmin, pw).Email, pw).AccessToken
		rec := a.Do(http.MethodPost, "/api/admin/products",
			map[string]any{"sku": testutil.Unique("SKU-"), "name": "Mask", "price": 5000, "barcodes": []map[string]string{{"barcode": barcode}}}, admin)
		if rec.Code != http.StatusCreated {
			t.Fatalf("product: %d %s", rec.Code, rec.Body.String())
		}
		pid := testutil.Decode[map[string]any](t, rec)["id"].(string)
		ev := a.ActiveEventIn(store)
		ba := a.CreateUserIn(store, useriface.RoleBA, pw)
		if _, err := a.Events.AssignBaToEvent(store, ba.ID, ev.ID); err != nil {
			t.Fatal(err)
		}
		if _, err := a.Products.SetAvailability(store, ev.ID, uuid.MustParse(pid), true); err != nil {
			t.Fatal(err)
		}
		rec = a.Do(http.MethodGet, "/api/me/events/"+ev.ID.String()+"/products/by-barcode/"+barcode, nil, login(t, a, ba.Email, pw).AccessToken)
		if rec.Code != http.StatusOK {
			t.Fatalf("scan: %d %s", rec.Code, rec.Body.String())
		}
		return pid, testutil.Decode[map[string]any](t, rec)["id"].(string)
	}
	for _, store := range []uuid.UUID{tenant.DefaultStoreID, a.NewStore("Store B")} {
		if want, got := scan(store); got != want {
			t.Errorf("store %s scanned %s, want its own item %s", store, got, want)
		}
	}
}

func TestMoneyIsWholeRupiah(t *testing.T) {
	a := testutil.NewApp(t)
	tok := adminToken(t, a)
	for name, rec := range map[string]int{
		"create with decimals": a.Do(http.MethodPost, "/api/admin/products", map[string]any{"sku": testutil.Unique("SKU-"), "name": "x", "price": 10000.5}, tok).Code,
		"update with decimals": a.Do(http.MethodPut, "/api/admin/products/"+a.Product("1000").ID.String(), map[string]any{"price": 1.5}, tok).Code,
	} {
		if rec != http.StatusBadRequest {
			t.Errorf("%s: got %d, want 400", name, rec)
		}
	}
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

// Every item has one default variant carrying its price; the API shows only the real variants.
func TestDefaultVariantCarriesThePrice(t *testing.T) {
	a := testutil.NewApp(t)
	tok := adminToken(t, a)
	rec := a.Do(http.MethodPost, "/api/admin/products", map[string]any{"sku": testutil.Unique("SKU-"), "name": "Lipstick", "price": 10000}, tok)
	created := testutil.Decode[map[string]any](t, rec)
	if rec.Code != http.StatusCreated || created["price"] != float64(10000) || len(created["variants"].([]any)) != 0 {
		t.Fatalf("create: %d %v", rec.Code, created)
	}
	id := uuid.MustParse(created["id"].(string))
	var defaults []productrepo.ProductVariant
	a.DB.Where("item_id = ? AND is_default", id).Find(&defaults)
	if len(defaults) != 1 || defaults[0].Price.String() != "10000" || defaults[0].StoreID != tenant.DefaultStoreID {
		t.Fatalf("default variants: %+v", defaults)
	}

	// A real variant without its own price follows the default one.
	shade := productrepo.ProductVariant{ItemID: id, Name: "Rose", SKU: testutil.Unique("SKU-"), SellInPOS: true}
	if err := a.DB.Create(&shade).Error; err != nil {
		t.Fatal(err)
	}
	updated := testutil.Decode[map[string]any](t, a.Do(http.MethodPut, "/api/admin/products/"+id.String(), map[string]any{"price": 12000}, tok))
	variants := updated["variants"].([]any)
	if updated["price"] != float64(12000) || len(variants) != 1 || variants[0].(map[string]any)["price"] != float64(12000) {
		t.Fatalf("update: %v", updated)
	}

	ev := a.ActiveEvent()
	a.Allocate(ev.ID, id, 5)
	if line := sellOne(t, a, ev.ID, id, nil).Items[0]; line.UnitPrice.String() != "12000" || line.VariantName != nil {
		t.Errorf("line without a variant: %+v, want the default price and no variant", line)
	}
	if line := sellOne(t, a, ev.ID, id, &shade.ID).Items[0]; line.UnitPrice.String() != "12000" || line.VariantName == nil || *line.VariantName != "Rose" {
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
