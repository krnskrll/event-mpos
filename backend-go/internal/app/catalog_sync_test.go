package app_test

import (
	"errors"
	"net/http"
	"testing"
	"time"

	"github.com/google/uuid"

	catalogiface "github.com/krnskrll/event-mpos/internal/catalog/iface"
	"github.com/krnskrll/event-mpos/internal/integration/bhoomy"
	"github.com/krnskrll/event-mpos/internal/testutil"
	useriface "github.com/krnskrll/event-mpos/internal/user/iface"
)

func TestCatalogSyncPullsProdukTokoChanges(t *testing.T) {
	a := testutil.NewApp(t)
	mock := a.Bhoomy.(*bhoomy.Mock)
	store, bhoomyStore := a.NewStore("Linked"), uuid.New()
	a.DB.Exec("UPDATE pos_stores SET bhoomy_store_id = ? WHERE id = ?", bhoomyStore, store)
	tok := login(t, a, a.CreateUserIn(store, useriface.RoleAdmin, "Passw0rd!").Email, "Passw0rd!").AccessToken
	t0 := time.Now().UTC().Truncate(time.Second)
	variant := func(name string, at time.Duration) catalogiface.Variant {
		return catalogiface.Variant{BhoomyVariantID: uuid.New(), BhoomyProductID: uuid.New(), ProductName: name,
			VariantName: "Default", SKU: testutil.Unique("B-"), IsDefault: true, Status: "active", ChangedAt: t0.Add(at)}
	}
	mock.Variants = map[uuid.UUID][]catalogiface.Variant{}
	for i := range 1200 {
		mock.Variants[bhoomyStore] = append(mock.Variants[bhoomyStore], variant("Item", time.Duration(i)*time.Millisecond))
	}
	status := func() map[string]any {
		t.Helper()
		rec := a.Do(http.MethodGet, "/api/admin/catalog/sync", nil, tok)
		if rec.Code != http.StatusOK {
			t.Fatalf("status: %d %s", rec.Code, rec.Body.String())
		}
		return testutil.Decode[map[string]any](t, rec)
	}

	if rec := a.Do(http.MethodPost, "/api/admin/catalog/sync", nil, tok); rec.Code != http.StatusOK {
		t.Fatalf("sync: %d %s", rec.Code, rec.Body.String())
	}
	if st := status(); st["linked"] != true || st["cached"] != float64(1200) || st["lastError"] != nil {
		t.Fatalf("after first pull (3 pages): %v", st)
	}

	// Bhoomy down: the error is kept and the cursor stays; the next pull resumes with only new changes.
	mock.Err = errors.New("connection refused")
	if rec := a.Do(http.MethodPost, "/api/admin/catalog/sync", nil, tok); rec.Code != http.StatusBadGateway {
		t.Errorf("Bhoomy down: got %d, want 502", rec.Code)
	}
	before := status()
	if before["lastError"] == nil {
		t.Errorf("error not recorded: %v", before)
	}
	mock.Err = nil
	renamed := mock.Variants[bhoomyStore][0]
	renamed.ProductName, renamed.ChangedAt = "Renamed", t0.Add(time.Hour)
	mock.Variants[bhoomyStore] = append(mock.Variants[bhoomyStore][1:], renamed, variant("New", time.Hour+time.Second))
	a.Do(http.MethodPost, "/api/admin/catalog/sync", nil, tok)
	var names []string
	a.DB.Table("pos_catalog_cache").Where("store_id = ? AND product_name IN ?", store, []string{"Renamed", "New"}).Pluck("product_name", &names)
	if st := status(); st["cached"] != float64(1201) || st["lastError"] != nil || len(names) != 2 {
		t.Errorf("after resume: %v %v", st, names)
	}

	// Another store sees only its own state; an unlinked store cannot pull.
	other := login(t, a, a.CreateUser(useriface.RoleAdmin, "Passw0rd!").Email, "Passw0rd!").AccessToken
	if st := testutil.Decode[map[string]any](t, a.Do(http.MethodGet, "/api/admin/catalog/sync", nil, other)); st["cached"] != float64(0) || st["linked"] != false {
		t.Errorf("unlinked store status: %v", st)
	}
	if rec := a.Do(http.MethodPost, "/api/admin/catalog/sync", nil, other); rec.Code != http.StatusConflict {
		t.Errorf("unlinked store sync: got %d, want 409", rec.Code)
	}
}
