package app_test

import (
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/krnskrll/event-mpos/internal/integration/bhoomy"
	memberiface "github.com/krnskrll/event-mpos/internal/member/iface"
	orderiface "github.com/krnskrll/event-mpos/internal/order/iface"
	productiface "github.com/krnskrll/event-mpos/internal/product/iface"
	promotioniface "github.com/krnskrll/event-mpos/internal/promotion/iface"
	reportiface "github.com/krnskrll/event-mpos/internal/report/iface"
	"github.com/krnskrll/event-mpos/internal/testutil"
)

// inCategory files a Produk Toko product under a category (on the next pull).
func (pt *produkToko) inCategory(product, category uuid.UUID, name string) {
	mock := pt.a.Bhoomy.(*bhoomy.Mock)
	pt.at = pt.at.Add(time.Second)
	for i := range mock.Variants[pt.store] {
		if v := &mock.Variants[pt.store][i]; v.BhoomyProductID == product {
			v.CategoryID, v.CategoryName, v.ChangedAt = &category, &name, pt.at
		}
	}
}

func posStore(t *testing.T, a *testutil.TestApp, bhoomyStore uuid.UUID) uuid.UUID {
	t.Helper()
	ok[[]J](t, a.Back(http.MethodGet, "/locations", nil, bhoomyStore), http.StatusOK)
	var id string
	if err := a.DB.Raw("SELECT id FROM pos_stores WHERE bhoomy_store_id = ?", bhoomyStore).Scan(&id).Error; err != nil {
		t.Fatal(err)
	}
	return uuid.MustParse(id)
}

func TestPromosi(t *testing.T) {
	a := testutil.NewApp(t)
	pt := newProdukToko(a)
	serum, lipstick, toner, serumCat := uuid.New(), uuid.New(), uuid.New(), uuid.New()
	pt.put(serum, "", "PRD-101", true, "active", nil)
	pt.put(lipstick, "", "PRD-102", true, "active", nil)
	pt.inCategory(serum, serumCat, "Serum")
	pt.sync(t)
	store := pt.store
	ok[productiface.AddItemsResponse](t, a.Back(http.MethodPost, "/items", J{"productIds": []uuid.UUID{serum, lipstick}}, store), http.StatusCreated)
	item := map[string]uuid.UUID{}
	for _, it := range ok[[]productiface.CatalogueItem](t, a.Back(http.MethodGet, "/items", nil, store), http.StatusOK) {
		item[it.Code] = it.ID
	}
	serumItem, lipstickItem := item["PRD-101"], item["PRD-102"]

	cats := ok[[]productiface.Category](t, a.Back(http.MethodGet, "/categories", nil, store), http.StatusOK)
	if len(cats) != 1 || cats[0].Name != "Serum" || len(cats[0].ItemIDs) != 1 || cats[0].ItemIDs[0] != serumItem {
		t.Fatalf("categories: %+v", cats)
	}
	newLocation := func(name string) uuid.UUID {
		loc := ok[J](t, a.Back(http.MethodPost, "/locations", J{"type": "event", "name": name, "startDate": "2027-01-01", "endDate": "2027-01-07"}, store), http.StatusCreated)
		return uuid.MustParse(loc["id"].(string))
	}
	kemang, bintaro := newLocation("Outlet Kemang"), newLocation("Outlet Bintaro")

	everywhere := ok[promotioniface.PromotionResponse](t, a.Back(http.MethodPost, "/promotions",
		J{"name": "Diskon 10%", "type": "PERCENTAGE", "percentOff": 10}, store), http.StatusCreated)
	serum15 := ok[promotioniface.PromotionResponse](t, a.Back(http.MethodPost, "/promotions",
		J{"name": "Serum 15%", "type": "PERCENTAGE", "percentOff": 15, "scope": "SPECIFIC", "categoryId": serumCat, "locationId": kemang}, store), http.StatusCreated)
	if everywhere.LocationID != nil || serum15.LocationID == nil || *serum15.CategoryID != serumCat {
		t.Fatalf("created: %+v %+v", everywhere, serum15)
	}

	for name, body := range map[string]J{
		"bundle for all products":   {"name": "x", "type": "BUNDLE", "amountOff": 1000, "minQty": 2},
		"over 100%":                 {"name": "x", "type": "PERCENTAGE", "percentOff": 120},
		"ends before it starts":     {"name": "x", "type": "FIXED", "amountOff": 1000, "startsAt": "2026-11-10T00:00:00Z", "endsAt": "2026-11-01T00:00:00Z"},
		"another store's product":   {"name": "x", "type": "FIXED", "amountOff": 1000, "scope": "SPECIFIC", "productIds": []uuid.UUID{a.Product("100").ID}},
		"category and products":     {"name": "x", "type": "FIXED", "amountOff": 1000, "scope": "SPECIFIC", "categoryId": serumCat, "productIds": []uuid.UUID{lipstickItem}},
		"unknown category":          {"name": "x", "type": "FIXED", "amountOff": 1000, "scope": "SPECIFIC", "categoryId": uuid.New()},
		"specific without products": {"name": "x", "type": "FIXED", "amountOff": 1000, "scope": "SPECIFIC"},
	} {
		if rec := a.Back(http.MethodPost, "/promotions", body, store); rec.Code != http.StatusBadRequest {
			t.Errorf("%s: got %d, want 400", name, rec.Code)
		}
	}
	other := a.ActiveEvent()
	if rec := a.Back(http.MethodPost, "/promotions", J{"name": "x", "type": "FIXED", "amountOff": 1000, "locationId": other.ID}, store); rec.Code != http.StatusNotFound {
		t.Errorf("another store's location: got %d, want 404", rec.Code)
	}

	// The cashier app gets the store-wide promotion everywhere and the category one, expanded, at Kemang.
	snapshot := func(loc uuid.UUID) map[string][]uuid.UUID {
		t.Helper()
		got, err := a.Promotions.SnapshotPromotions(loc)
		if err != nil {
			t.Fatal(err)
		}
		out := map[string][]uuid.UUID{}
		for _, p := range got {
			dto := p.(promotioniface.PromotionSnapshotDto)
			out[dto.Name] = dto.ProductIDs
		}
		return out
	}
	if got := snapshot(kemang); len(got) != 2 || len(got["Serum 15%"]) != 1 || got["Serum 15%"][0] != serumItem {
		t.Errorf("Kemang snapshot: %v", got)
	}
	if got := snapshot(bintaro); len(got) != 1 || got["Diskon 10%"] == nil {
		t.Errorf("Bintaro snapshot: %v", got)
	}
	best, err := a.Promotions.BestFor(kemang, []promotioniface.CartLine{{ProductID: serumItem, UnitPrice: dec("100000"), Qty: 1}}, time.Now())
	if err != nil || best == nil || best.Name != "Serum 15%" || !best.Discount.Equal(dec("15000")) {
		t.Errorf("best at Kemang: %+v %v", best, err)
	}

	// A product filed under Serum later joins the promotion, and the snapshot version moves.
	before, _ := a.Promotions.Fingerprint(kemang)
	pt.put(toner, "", "PRD-105", true, "active", nil)
	pt.inCategory(toner, serumCat, "Serum")
	pt.sync(t)
	tonerItem := ok[productiface.AddItemsResponse](t, a.Back(http.MethodPost, "/items", J{"productIds": []uuid.UUID{toner}}, store), http.StatusCreated).IDs[0]
	if after, _ := a.Promotions.Fingerprint(kemang); after == before || len(snapshot(kemang)["Serum 15%"]) != 2 {
		t.Errorf("toner in Serum: snapshot %v, fingerprint changed=%v", snapshot(kemang), after != before)
	}

	list := func(q string) []promotioniface.PromotionResponse {
		return ok[[]promotioniface.PromotionResponse](t, a.Back(http.MethodGet, "/promotions"+q, nil, store), http.StatusOK)
	}
	if got := list("?locationId=" + bintaro.String()); len(got) != 1 || got[0].ID != everywhere.ID {
		t.Errorf("at Bintaro: %+v", got)
	}
	upd := ok[promotioniface.PromotionResponse](t, a.Back(http.MethodPut, "/promotions/"+serum15.ID.String(),
		J{"name": "Serum & toner", "type": "FIXED", "amountOff": 5000, "scope": "SPECIFIC", "productIds": []uuid.UUID{serumItem, tonerItem, serumItem}, "active": false}, store), http.StatusOK)
	if upd.LocationID != nil || upd.CategoryID != nil || len(upd.ProductIDs) != 2 || upd.Active {
		t.Errorf("updated: %+v, want every location, 2 chosen products, off", upd)
	}
	if got := list(""); len(got) != 2 || got[0].ID != everywhere.ID {
		t.Errorf("list: %+v, want the active one first", got)
	}

	stranger := uuid.New()
	if rec := a.Back(http.MethodGet, "/promotions/"+serum15.ID.String(), nil, stranger); rec.Code != http.StatusNotFound {
		t.Errorf("another store reads it: %d", rec.Code)
	}
	if rec := a.Back(http.MethodDelete, "/promotions/"+serum15.ID.String(), nil, stranger); rec.Code != http.StatusNotFound {
		t.Errorf("another store deletes it: %d", rec.Code)
	}
	if rec := a.Back(http.MethodDelete, "/promotions/"+serum15.ID.String(), nil, store); rec.Code != http.StatusNoContent {
		t.Errorf("delete: %d", rec.Code)
	}
}

func TestMemberPOS(t *testing.T) {
	a := testutil.NewApp(t)
	bhoomyStore := uuid.New()
	store := posStore(t, a, bhoomyStore)
	if _, err := a.Members.Upsert(a.DB, store, memberiface.MemberRecord{MemberCode: "MBR-0042", QRCode: "QR-AISHA-7821",
		Name: "Aisha Rahman", Tier: memberiface.TierGold, Phone: str("0812"), Active: true}); err != nil {
		t.Fatal(err)
	}
	rec := a.Back(http.MethodGet, "/members", nil, bhoomyStore)
	got := ok[[]memberiface.MemberResponse](t, rec, http.StatusOK)
	if len(got) != 1 || got[0].QRCode != "•••• 7821" || got[0].Tier != memberiface.TierGold {
		t.Fatalf("members: %+v", got)
	}
	if body := rec.Body.String(); strings.Contains(body, "0812") || strings.Contains(body, "QR-AISHA") {
		t.Errorf("phone or full QR leaked: %s", body)
	}
	if got := ok[[]memberiface.MemberResponse](t, a.Back(http.MethodGet, "/members", nil, uuid.New()), http.StatusOK); len(got) != 0 {
		t.Errorf("another store sees %d members", len(got))
	}
}

func TestLaporanPOS(t *testing.T) {
	a := testutil.NewApp(t)
	bhoomyStore := uuid.New()
	store := posStore(t, a, bhoomyStore)
	ev, p := a.ActiveEventIn(store), a.ProductIn(store, "50000")
	a.Allocate(ev.ID, p.ID, 100)
	ok[promotioniface.PromotionResponse](t, a.Back(http.MethodPost, "/promotions",
		J{"name": "Hemat", "type": "FIXED", "amountOff": 10000, "minSubtotal": 100000}, bhoomyStore), http.StatusCreated)

	for _, qty := range []int{1, 2} {
		if _, err := submit(a, ev.ID, uuid.NewString(), line(p.ID, qty)); err != nil {
			t.Fatal(err)
		}
	}
	device := "device-" + uuid.NewString()
	block, err := a.Orders.IssueNumberBlock(ev.ID, device)
	if err != nil {
		t.Fatal(err)
	}
	yesterday := time.Now().Add(-24 * time.Hour)
	if _, err := a.Orders.Submit(ev.ID, uuid.New(), orderiface.SubmitOrderRequest{IdempotencyKey: uuid.NewString(),
		Customer: orderiface.CustomerInfo{Name: "Offline"}, Lines: []orderiface.OrderLineRequest{line(p.ID, 4)},
		DeviceID: &device, OrderNumber: &block.RangeStart, ClientCreatedAt: &yesterday}); err != nil {
		t.Fatal(err)
	}

	wib := time.FixedZone("WIB", 7*3600)
	today, yday := time.Now().In(wib).Format(time.DateOnly), yesterday.In(wib).Format(time.DateOnly)
	report := func(q string) reportiface.SalesReport {
		return ok[reportiface.SalesReport](t, a.Back(http.MethodGet, "/locations/"+ev.ID.String()+"/report"+q, nil, bhoomyStore), http.StatusOK)
	}
	r := report("")
	if r.From != today || r.TotalOrders != 2 || r.TotalItems != 3 || !r.TotalRevenue.Equal(dec("140000")) ||
		!r.TotalDiscount.Equal(dec("10000")) || r.PromotionsUsed != 1 || r.ProductCount != 1 || r.Final {
		t.Errorf("today: %+v", r)
	}
	if r = report("?from=" + yday + "&to=" + today); r.TotalOrders != 3 || r.Products[0].UnitsSold != 7 {
		t.Errorf("two days, offline order counted on its sale day: %+v", r)
	}
	if r = report("?from=" + yday + "&to=" + yday); r.TotalOrders != 1 || r.Final {
		t.Errorf("yesterday: %+v", r)
	}
	for _, q := range []string{"?from=" + today + "&to=" + yday, "?from=2026-13-01", "?from=2024-01-01&to=2026-01-01"} {
		if rec := a.Back(http.MethodGet, "/locations/"+ev.ID.String()+"/report"+q, nil, bhoomyStore); rec.Code != http.StatusBadRequest {
			t.Errorf("%s: got %d, want 400", q, rec.Code)
		}
	}
	if rec := a.Back(http.MethodGet, "/locations/"+ev.ID.String()+"/report", nil, uuid.New()); rec.Code != http.StatusNotFound {
		t.Errorf("another store's report: %d", rec.Code)
	}

	page := ok[struct{ Total int64 }](t, a.Back(http.MethodGet, "/locations/"+ev.ID.String()+"/audit?from="+today+"&to="+today+"&action=ORDER_CREATED", nil, bhoomyStore), http.StatusOK)
	tomorrow := time.Now().In(wib).AddDate(0, 0, 1).Format(time.DateOnly)
	none := ok[struct{ Total int64 }](t, a.Back(http.MethodGet, "/locations/"+ev.ID.String()+"/audit?from="+tomorrow+"&to="+tomorrow, nil, bhoomyStore), http.StatusOK)
	if page.Total != 3 || none.Total != 0 {
		t.Errorf("history: today %d (want 3, recorded when uploaded), tomorrow %d", page.Total, none.Total)
	}
}
