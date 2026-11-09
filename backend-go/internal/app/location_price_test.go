package app_test

import (
	"bytes"
	"image"
	"image/png"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"

	productiface "github.com/krnskrll/event-mpos/internal/product/iface"
	productrepo "github.com/krnskrll/event-mpos/internal/product/repository"
	"github.com/krnskrll/event-mpos/internal/testutil"
)

// Location price of the variant → variant price → default variant price (at that location).
func TestLocationPriceRule(t *testing.T) {
	a := testutil.NewApp(t)
	bhoomyStore := uuid.New()
	store := posStore(t, a, bhoomyStore)
	p := a.ProductIn(store, "10000")
	plumPrice := decimal.RequireFromString("15000")
	rose := productrepo.ProductVariant{ItemID: p.ID, Name: "Rose", SKU: testutil.Unique("SKU-"), SellInPOS: true}
	plum := productrepo.ProductVariant{ItemID: p.ID, Name: "Plum", SKU: testutil.Unique("SKU-"), Price: &plumPrice, SellInPOS: true}
	for _, v := range []*productrepo.ProductVariant{&rose, &plum} {
		if err := a.DB.Create(v).Error; err != nil {
			t.Fatal(err)
		}
	}
	ev := a.ActiveEventIn(store)
	a.Allocate(ev.ID, p.ID, 10)
	save := func(rows ...J) int {
		return a.Back(http.MethodPut, "/locations/"+ev.ID.String()+"/items", J{"rows": rows}, bhoomyStore).Code
	}
	row := func(id uuid.UUID, sold bool, price any) J { return J{"variantId": id, "sold": sold, "price": price} }
	hash := func() string { h, _ := a.Events.SnapshotVersion(ev.ID); return h }

	if got := save(row(p.Default().ID, true, 9000), row(rose.ID, true, nil), row(plum.ID, true, nil)); got != http.StatusOK {
		t.Fatalf("save rows: %d", got)
	}
	before := hash()
	save(row(plum.ID, true, 14000))
	if hash() == before {
		t.Error("snapshot version did not change with a variant's location price")
	}

	prices := snapshotPrices(t, a, ev.ID, p.ID)
	want := map[string]string{"": "9000", "Rose": "9000", "Plum": "14000"}
	for name, price := range want {
		if prices[name] != price {
			t.Errorf("snapshot %q = %s, want %s", name, prices[name], price)
		}
	}
	if line := sellOne(t, a, ev.ID, p.ID, nil).Items[0]; line.UnitPrice.String() != "9000" {
		t.Errorf("default line: %s, want 9000", line.UnitPrice)
	}
	if line := sellOne(t, a, ev.ID, p.ID, &plum.ID).Items[0]; line.UnitPrice.String() != "14000" {
		t.Errorf("plum line: %s, want 14000", line.UnitPrice)
	}

	// Rose switched off here leaves the snapshot; an empty price clears Plum's location price.
	save(row(rose.ID, false, nil), row(plum.ID, true, nil))
	if prices := snapshotPrices(t, a, ev.ID, p.ID); prices["Rose"] != "" || prices["Plum"] != "15000" {
		t.Errorf("after changes: %v", prices)
	}
	if got := save(row(plum.ID, true, 1.5)); got != http.StatusBadRequest {
		t.Errorf("price with decimals: %d", got)
	}
}

func TestPhotoUpload(t *testing.T) {
	a := testutil.NewApp(t)
	bhoomyStore := uuid.New()
	p := a.ProductIn(posStore(t, a, bhoomyStore), "1000")
	var img bytes.Buffer
	if err := png.Encode(&img, image.NewRGBA(image.Rect(0, 0, 2, 2))); err != nil {
		t.Fatal(err)
	}

	rec := upload(a, p.ID, img.Bytes(), bhoomyStore)
	if rec.Code != http.StatusOK {
		t.Fatalf("upload: %d %s", rec.Code, rec.Body.String())
	}
	url := testutil.Decode[map[string]any](t, rec)["imageUrl"].(string)
	if !strings.HasPrefix(url, "http://pos.test/media/") || !strings.HasSuffix(url, ".png") {
		t.Fatalf("imageUrl %q", url)
	}
	got := a.Do(http.MethodGet, strings.TrimPrefix(url, "http://pos.test"), nil, "")
	if got.Code != http.StatusOK || got.Header().Get("Content-Type") != "image/png" || got.Header().Get("X-Content-Type-Options") != "nosniff" {
		t.Errorf("serve: %d %v", got.Code, got.Header())
	}

	for name, body := range map[string][]byte{"text file": []byte("<html>hi</html>"), "over 2 MB": append(img.Bytes(), make([]byte, 2<<20)...)} {
		if rec := upload(a, p.ID, body, bhoomyStore); rec.Code != http.StatusBadRequest {
			t.Errorf("%s: got %d, want 400", name, rec.Code)
		}
	}
	for _, path := range []string{"/media/..%2Fsecret", "/media/nope.png"} {
		if rec := a.Do(http.MethodGet, path, nil, ""); rec.Code != http.StatusNotFound {
			t.Errorf("%s: got %d, want 404", path, rec.Code)
		}
	}

	other := uuid.New()
	if rec := upload(a, p.ID, img.Bytes(), other); rec.Code != http.StatusNotFound {
		t.Errorf("store B uploading to store A's product: got %d, want 404", rec.Code)
	}
	evB := a.ActiveEventIn(posStore(t, a, other))
	if rec := a.Back(http.MethodPut, "/locations/"+evB.ID.String()+"/items", J{"rows": []J{{"variantId": p.Default().ID, "sold": true}}}, other); rec.Code != http.StatusNotFound {
		t.Errorf("store B pricing store A's variant: got %d, want 404", rec.Code)
	}
}

func upload(a *testutil.TestApp, productID uuid.UUID, data []byte, bhoomyStore uuid.UUID) *httptest.ResponseRecorder {
	var body bytes.Buffer
	w := multipart.NewWriter(&body)
	part, _ := w.CreateFormFile("file", "photo")
	part.Write(data)
	w.Close()
	req := httptest.NewRequest(http.MethodPost, "/api/backoffice/items/"+productID.String()+"/photo", &body)
	req.Header.Set("Content-Type", w.FormDataContentType())
	req.Header.Set("Authorization", "Bearer "+testutil.ServiceToken)
	req.Header.Set("X-Bhoomy-Store-Id", bhoomyStore.String())
	rec := httptest.NewRecorder()
	a.Router.ServeHTTP(rec, req)
	return rec
}

// snapshotPrices maps variant name ("" for the product itself) to its snapshot price.
func snapshotPrices(t *testing.T, a *testutil.TestApp, eventID, productID uuid.UUID) map[string]string {
	t.Helper()
	products, err := a.Products.AvailableSnapshotProducts(eventID)
	if err != nil {
		t.Fatal(err)
	}
	out := map[string]string{}
	for _, raw := range products {
		p := raw.(productiface.ProductSnapshotDto)
		if p.ID != productID {
			continue
		}
		out[""] = p.Price.String()
		for _, v := range p.Variants {
			out[v.Name] = v.Price.String()
		}
	}
	return out
}
