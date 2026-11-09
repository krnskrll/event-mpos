package app_test

import (
	"net/http"
	"strings"
	"testing"

	"github.com/google/uuid"

	productiface "github.com/krnskrll/event-mpos/internal/product/iface"
	"github.com/krnskrll/event-mpos/internal/testutil"
)

func TestExtraBarcodes(t *testing.T) {
	a := testutil.NewApp(t)
	pt := newProdukToko(a)
	lipstick, serum := uuid.New(), uuid.New()
	pt.put(lipstick, "", "PRD-102", true, "active", str("8990001002"))
	pt.put(serum, "", "PRD-103", true, "active", str("8990001003"))
	pt.sync(t)
	store := pt.store
	ids := ok[productiface.AddItemsResponse](t, a.Back(http.MethodPost, "/items", J{"productIds": []uuid.UUID{lipstick, serum}}, store), http.StatusCreated).IDs
	var lip, ser uuid.UUID
	for _, id := range ids {
		item := ok[productiface.CatalogueItemDetail](t, a.Back(http.MethodGet, "/items/"+id.String(), nil, store), http.StatusOK)
		if *item.Source.Barcode == "8990001002" {
			lip = id
		} else {
			ser = id
		}
	}
	loc := ok[J](t, a.Back(http.MethodPost, "/locations", J{"type": "outlet", "name": "Outlet Kemang", "address": "Kemang Raya 8"}, store), http.StatusCreated)
	locID := uuid.MustParse(loc["id"].(string))
	save := func(id uuid.UUID, name string, extras any) *productiface.CatalogueItemDetail {
		t.Helper()
		body := J{"displayName": name, "status": "active", "basePrice": 79000, "locations": []J{{"id": locID, "sold": true}}}
		if extras != nil {
			body["extraBarcodes"] = extras
		}
		rec := a.Back(http.MethodPut, "/items/"+id.String(), body, store)
		if rec.Code != http.StatusOK {
			t.Logf("save %s: %d %s", name, rec.Code, rec.Body.String())
			return nil
		}
		return ok[*productiface.CatalogueItemDetail](t, rec, http.StatusOK)
	}
	save(ser, "Hydra Glow Serum", nil)
	ok[J](t, a.Back(http.MethodPatch, "/locations/"+locID.String()+"/status", J{"status": "ACTIVE"}, store), http.StatusOK)
	version := func() string {
		v, err := a.Events.SnapshotVersion(locID)
		if err != nil {
			t.Fatal(err)
		}
		return v
	}

	item := save(lip, "Velvet Lipstick", []J{{"barcode": " 18990001002 ", "label": "Kemasan isi 6"}, {"barcode": "DIST-77", "label": " "}})
	if item == nil || len(item.ExtraBarcodes) != 2 || item.ExtraBarcodes[0].Barcode != "18990001002" ||
		*item.ExtraBarcodes[0].Label != "Kemasan isi 6" || item.ExtraBarcodes[1].Label != nil {
		t.Fatalf("extras after save: %+v", item)
	}
	posStoreID := posStore(t, a, store)
	if p, err := a.Products.FindAvailableByBarcode(posStoreID, locID, "DIST-77"); err != nil || p.ID != lip {
		t.Errorf("cashier lookup by extra barcode: %+v %v", p, err)
	}
	before := version()
	if save(lip, "Velvet Lipstick", nil) == nil || len(get(t, a, store, lip).ExtraBarcodes) != 2 {
		t.Error("save without extraBarcodes must keep them")
	}
	save(lip, "Velvet Lipstick", []J{{"barcode": "18990001002"}, {"barcode": "DIST-78"}})
	if version() == before {
		t.Error("swapping an extra barcode must change the cashier snapshot version")
	}

	for name, extras := range map[string][]J{
		"empty":      {{"barcode": "  "}},
		"space":      {{"barcode": "899 0001"}},
		"twice":      {{"barcode": "A1"}, {"barcode": "A1"}},
		"too long":   {{"barcode": "123456789012345678901234567890123456789012345678901234567890123456"}},
		"long label": {{"barcode": "A1", "label": "Kemasan lama untuk bazaar dengan keterangan yang terlalu panjang"}},
	} {
		if rec := a.Back(http.MethodPut, "/items/"+lip.String(), J{"displayName": "Velvet Lipstick", "status": "active", "basePrice": 79000, "extraBarcodes": extras}, store); rec.Code != http.StatusBadRequest {
			t.Errorf("%s: got %d, want 400", name, rec.Code)
		}
	}
	for code, owner := range map[string]string{"8990001003": "Hydra Glow Serum", "8990001002": "Velvet Lipstick"} {
		rec := a.Back(http.MethodPut, "/items/"+lip.String(), J{"displayName": "Velvet Lipstick", "status": "active", "basePrice": 79000,
			"extraBarcodes": []J{{"barcode": code}}}, store)
		if rec.Code != http.StatusConflict || !strings.Contains(rec.Body.String(), owner) {
			t.Errorf("extra %s: %d %s, want 409 naming %s", code, rec.Code, rec.Body.String(), owner)
		}
	}
	save(ser, "Hydra Glow Serum", []J{{"barcode": "SER-1"}})

	// Produk Toko wins: the serum gets the lipstick's extra as its own barcode.
	pt.put(serum, "", "PRD-103", true, "active", str("DIST-78"))
	pt.sync(t)
	if got := get(t, a, store, lip).ExtraBarcodes; len(got) != 1 || got[0].Barcode != "18990001002" {
		t.Errorf("lipstick extras %+v, want DIST-78 removed", got)
	}
	if got := get(t, a, store, ser).ExtraBarcodes; len(got) != 1 || got[0].Barcode != "SER-1" {
		t.Errorf("serum extras %+v, want SER-1 kept by the sync", got)
	}
	if p, err := a.Products.FindAvailableByBarcode(posStoreID, locID, "DIST-78"); err != nil || p.ID != ser {
		t.Errorf("DIST-78 now scans as %+v %v, want the serum", p, err)
	}
}

func get(t *testing.T, a *testutil.TestApp, store, id uuid.UUID) productiface.CatalogueItemDetail {
	t.Helper()
	return ok[productiface.CatalogueItemDetail](t, a.Back(http.MethodGet, "/items/"+id.String(), nil, store), http.StatusOK)
}
