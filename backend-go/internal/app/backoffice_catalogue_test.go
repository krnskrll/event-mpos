package app_test

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/google/uuid"

	catalogiface "github.com/krnskrll/event-mpos/internal/catalog/iface"
	"github.com/krnskrll/event-mpos/internal/integration/bhoomy"
	productiface "github.com/krnskrll/event-mpos/internal/product/iface"
	"github.com/krnskrll/event-mpos/internal/testutil"
)

type J = map[string]any

// produkToko is one Bhoomy store's Produk Toko as the mock feed serves it.
type produkToko struct {
	a     *testutil.TestApp
	store uuid.UUID
	at    time.Time
}

func newProdukToko(a *testutil.TestApp) *produkToko {
	mock := a.Bhoomy.(*bhoomy.Mock)
	if mock.Variants == nil {
		mock.Variants = map[uuid.UUID][]catalogiface.Variant{}
	}
	return &produkToko{a: a, store: uuid.New(), at: time.Now().UTC().Truncate(time.Second)}
}

func (pt *produkToko) put(product uuid.UUID, name, sku string, isDefault bool, status string, barcode *string) uuid.UUID {
	mock := pt.a.Bhoomy.(*bhoomy.Mock)
	pt.at = pt.at.Add(time.Second)
	for i := range mock.Variants[pt.store] {
		if v := &mock.Variants[pt.store][i]; v.BhoomyProductID == product && v.VariantName == name {
			v.SKU, v.Status, v.Barcode, v.ChangedAt = sku, status, barcode, pt.at
			return v.BhoomyVariantID
		}
	}
	v := catalogiface.Variant{BhoomyVariantID: uuid.New(), BhoomyProductID: product, ProductName: "Velvet Matte Lipstick",
		VariantName: name, SKU: sku, Barcode: barcode, IsDefault: isDefault, Status: status, ChangedAt: pt.at}
	mock.Variants[pt.store] = append(mock.Variants[pt.store], v)
	return v.BhoomyVariantID
}

func (pt *produkToko) sync(t *testing.T) {
	t.Helper()
	if rec := pt.a.Back(http.MethodPost, "/catalog/sync", nil, pt.store); rec.Code != http.StatusOK {
		t.Fatalf("sync: %d %s", rec.Code, rec.Body.String())
	}
}

func ok[T any](t *testing.T, rec *httptest.ResponseRecorder, want int) T {
	t.Helper()
	if rec.Code != want {
		t.Fatalf("got %d, want %d: %s", rec.Code, want, rec.Body.String())
	}
	return testutil.Decode[T](t, rec)
}

func str(s string) *string { return &s }

func TestBackOfficeAccess(t *testing.T) {
	a := testutil.NewApp(t)
	call := func(token, store string) int {
		req := httptest.NewRequest(http.MethodGet, "/api/backoffice/items", nil)
		req.Header.Set("Authorization", "Bearer "+token)
		req.Header.Set("X-Bhoomy-Store-Id", store)
		rec := httptest.NewRecorder()
		a.Router.ServeHTTP(rec, req)
		return rec.Code
	}
	bhoomyStore := uuid.New()
	for name, c := range map[string]struct {
		token, store string
		want         int
	}{
		"no token":    {"", bhoomyStore.String(), http.StatusUnauthorized},
		"wrong token": {"nope", bhoomyStore.String(), http.StatusUnauthorized},
		"user token":  {adminToken(t, a), bhoomyStore.String(), http.StatusUnauthorized},
		"no store":    {testutil.ServiceToken, "", http.StatusBadRequest},
	} {
		if got := call(c.token, c.store); got != c.want {
			t.Errorf("%s: got %d, want %d", name, got, c.want)
		}
	}

	// The first visit creates the POS store; later visits reuse it.
	count := func() (n int64) {
		a.DB.Table("pos_stores").Where("bhoomy_store_id = ?", bhoomyStore).Count(&n)
		return n
	}
	for range 2 {
		if got := call(testutil.ServiceToken, bhoomyStore.String()); got != http.StatusOK {
			t.Fatalf("visit: %d", got)
		}
	}
	if count() != 1 {
		t.Errorf("POS stores for one Bhoomy store: %d, want 1", count())
	}

	// Another store cannot see or change the first store's items and locations.
	pt := newProdukToko(a)
	pt.store = bhoomyStore
	product := uuid.New()
	pt.put(product, "", "PRD-1", true, "active", nil)
	pt.sync(t)
	item := ok[productiface.AddItemsResponse](t, a.Back(http.MethodPost, "/items", J{"productIds": []uuid.UUID{product}}, bhoomyStore), http.StatusCreated).IDs[0]
	loc := outletOf(t, a, bhoomyStore)["id"].(string)
	other := uuid.New()
	for path, method := range map[string]string{
		"/items/" + item.String():      http.MethodGet,
		"/locations/" + loc:            http.MethodGet,
		"/locations/" + loc + "/items": http.MethodGet,
	} {
		if rec := a.Back(method, path, nil, other); rec.Code != http.StatusNotFound {
			t.Errorf("other store %s: got %d, want 404", path, rec.Code)
		}
	}
	if rec := a.Back(http.MethodPost, "/items", J{"productIds": []uuid.UUID{product}}, other); rec.Code != http.StatusNotFound {
		t.Errorf("other store adds a product it does not have: got %d, want 404", rec.Code)
	}
}

func TestKatalogPOS(t *testing.T) {
	a := testutil.NewApp(t)
	pt := newProdukToko(a)
	lipstick, toner, old := uuid.New(), uuid.New(), uuid.New()
	pt.put(lipstick, "", "PRD-102", true, "active", str("8990001002"))
	pt.put(lipstick, "Rose", "PRD-102-RO", false, "active", str("8990001003"))
	pt.put(lipstick, "Coral", "PRD-102-CO", false, "active", nil)
	pt.put(toner, "", "PRD-105", true, "active", nil)
	pt.put(old, "", "PRD-900", true, "archived", nil)
	pt.sync(t)
	store := pt.store

	source := func(q string) []productiface.SourceOption {
		return ok[[]productiface.SourceOption](t, a.Back(http.MethodGet, "/catalog/source?q="+q, nil, store), http.StatusOK)
	}
	if got := source(""); len(got) != 2 || got[0].VariantCount+got[1].VariantCount != 3 {
		t.Fatalf("source: %+v, want the lipstick (2 variants) and the toner (1), not the archived product", got)
	}
	if got := source("102-ro"); len(got) != 1 || got[0].ProductID != lipstick {
		t.Errorf("search by variant SKU: %+v", got)
	}

	id := ok[productiface.AddItemsResponse](t, a.Back(http.MethodPost, "/items", J{"productIds": []uuid.UUID{lipstick}}, store), http.StatusCreated).IDs[0]
	if got := source(""); len(got) != 1 || got[0].ProductID != toner {
		t.Errorf("source after adding: %+v, want only the toner", got)
	}
	if rec := a.Back(http.MethodPost, "/items", J{"productIds": []uuid.UUID{lipstick}}, store); rec.Code != http.StatusConflict {
		t.Errorf("adding twice: got %d, want 409", rec.Code)
	}
	list := ok[[]productiface.CatalogueItem](t, a.Back(http.MethodGet, "/items", nil, store), http.StatusOK)
	if len(list) != 1 || list[0].Status != "archived" || list[0].PriceMin != nil || list[0].Variants != 2 || list[0].Code != "PRD-102" {
		t.Fatalf("new item: %+v, want Tidak dijual, no price, 2 variants", list)
	}

	item := ok[productiface.CatalogueItemDetail](t, a.Back(http.MethodGet, "/items/"+id.String(), nil, store), http.StatusOK)
	if item.Source.Barcode == nil || *item.Source.Barcode != "8990001002" || len(item.Variants) != 2 {
		t.Fatalf("detail: %+v", item)
	}
	rose, coral := item.Variants[0].ID, item.Variants[1].ID
	if item.Variants[0].Name == "Coral" {
		rose, coral = coral, rose
	}
	loc := outletOf(t, a, store)
	locID := loc["id"].(string)
	update := J{"displayName": "Velvet Lipstick", "status": "active", "basePrice": nil,
		"variants":  []J{{"id": rose, "sellInPos": false}, {"id": coral, "sellInPos": true, "price": 89000}},
		"locations": []J{{"id": locID, "sold": true}}}
	if rec := a.Back(http.MethodPut, "/items/"+id.String(), update, store); rec.Code != http.StatusBadRequest {
		t.Errorf("selling without a price: got %d, want 400", rec.Code)
	}
	update["basePrice"] = 79000
	item = ok[productiface.CatalogueItemDetail](t, a.Back(http.MethodPut, "/items/"+id.String(), update, store), http.StatusOK)
	if item.Status != "active" || item.BasePrice.String() != "79000" || len(item.Locations) != 1 || !item.Locations[0].Sold {
		t.Fatalf("after save: %+v", item)
	}

	ok[J](t, a.Back(http.MethodPatch, "/locations/"+locID+"/status", J{"status": "ACTIVE"}, store), http.StatusOK)
	sells := func() map[string]string {
		t.Helper()
		snap, err := a.Events.SnapshotForEvent(uuid.MustParse(locID))
		if err != nil || len(snap.Products) != 1 {
			t.Fatalf("snapshot: %v %+v", err, snap)
		}
		p := snap.Products[0].(productiface.ProductSnapshotDto)
		out := map[string]string{"": p.Price.String()}
		for _, v := range p.Variants {
			out[v.Name] = v.Price.String()
		}
		return out
	}
	if got := sells(); len(got) != 2 || got[""] != "79000" || got["Coral"] != "89000" {
		t.Errorf("cashier app gets %v, want base 79000 and Coral 89000 (Rose off)", got)
	}

	rows := ok[[]productiface.LocationRow](t, a.Back(http.MethodGet, "/locations/"+locID+"/items", nil, store), http.StatusOK)
	if len(rows) != 2 || rows[1].VariantName != "Coral" || !rows[1].Sold || rows[1].CatalogPrice.String() != "89000" {
		t.Fatalf("location rows: %+v, want the base row and Coral", rows)
	}
	ok[[]productiface.LocationRow](t, a.Back(http.MethodPut, "/locations/"+locID+"/items",
		J{"rows": []J{{"variantId": coral, "sold": true, "price": 85000}}}, store), http.StatusOK)
	if got := sells(); got["Coral"] != "85000" {
		t.Errorf("location price: Coral sells at %s, want 85000", got["Coral"])
	}
	before, _ := a.Events.SnapshotVersion(uuid.MustParse(locID))
	pt.put(lipstick, "Rose", "PRD-102-RO", false, "active", str("8990001009"))
	pt.sync(t)
	if after, _ := a.Events.SnapshotVersion(uuid.MustParse(locID)); after == before {
		t.Error("a new Produk Toko barcode must change the cashier snapshot version")
	}
	locs := ok[[]J](t, a.Back(http.MethodGet, "/locations", nil, store), http.StatusOK)
	if len(locs) != 1 || locs[0]["items"] != float64(1) || locs[0]["id"] != locID {
		t.Errorf("locations: %v", locs)
	}
}

func TestLocationCodes(t *testing.T) {
	a := testutil.NewApp(t)
	store := uuid.New()
	create := func(body J, want int) string {
		t.Helper()
		got := ok[J](t, a.Back(http.MethodPost, "/locations", body, store), want)
		code, _ := got["code"].(string)
		return code
	}
	event := J{"type": "event", "name": "Pop-up Senayan City", "startDate": "2027-04-05", "endDate": "2027-04-12"}
	for i, want := range []string{"EV-SC27", "EV-SC27-2", "EV-SC27-3"} {
		if got := create(event, http.StatusCreated); got != want {
			t.Errorf("event %d: code %s, want %s", i+1, got, want)
		}
	}
	create(J{"type": "event", "name": "No dates"}, http.StatusBadRequest)
	create(J{"type": "event", "name": "Backwards", "startDate": "2027-04-12", "endDate": "2027-04-05"}, http.StatusBadRequest)
	if rec := a.Back(http.MethodPost, "/locations", event, uuid.New()); ok[J](t, rec, http.StatusCreated)["code"] != "EV-SC27" {
		t.Error("codes are per store: another store starts again at EV-SC27")
	}
}

// Produk Toko owns variants, SKUs and barcodes: linked items follow its changes on the next pull.
func TestKatalogPOSFollowsProdukToko(t *testing.T) {
	a := testutil.NewApp(t)
	pt := newProdukToko(a)
	lipstick := uuid.New()
	pt.put(lipstick, "", "PRD-102", true, "active", nil)
	pt.put(lipstick, "Rose", "PRD-102-RO", false, "active", str("8990001003"))
	pt.put(lipstick, "Coral", "PRD-102-CO", false, "active", nil)
	pt.sync(t)
	id := ok[productiface.AddItemsResponse](t, a.Back(http.MethodPost, "/items", J{"productIds": []uuid.UUID{lipstick}}, pt.store), http.StatusCreated).IDs[0]

	pt.put(lipstick, "Rose", "PRD-102-ROSE", false, "active", str("8990001009"))
	pt.put(lipstick, "Coral", "PRD-102-CO", false, "archived", nil)
	pt.put(lipstick, "Plum", "PRD-102-PL", false, "active", nil)
	pt.sync(t)

	item := ok[productiface.CatalogueItemDetail](t, a.Back(http.MethodGet, "/items/"+id.String(), nil, pt.store), http.StatusOK)
	got := map[string]productiface.CatalogueVariant{}
	for _, v := range item.Variants {
		got[v.Name] = v
	}
	if got["Rose"].SKU != "PRD-102-ROSE" || !got["Rose"].SellInPOS {
		t.Errorf("Rose: %+v, want the new SKU and still sold", got["Rose"])
	}
	if got["Coral"].SellInPOS {
		t.Error("Coral archived in Produk Toko: still sold in POS")
	}
	if v, found := got["Plum"]; !found || v.SellInPOS {
		t.Errorf("new Plum: %+v (found %v), want added and switched off", v, found)
	}
	var barcodes []string
	a.DB.Table("pos_item_barcodes").Where("item_id = ?", id).Pluck("barcode", &barcodes)
	if len(barcodes) != 1 || barcodes[0] != "8990001009" {
		t.Errorf("barcodes %v, want the new Rose barcode only", barcodes)
	}
}
