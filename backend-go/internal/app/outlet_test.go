package app_test

import (
	"net/http"
	"testing"

	"github.com/google/uuid"

	"github.com/krnskrll/event-mpos/internal/testutil"
)

func TestOutletRules(t *testing.T) {
	a := testutil.NewApp(t)
	bhoomyStore := uuid.New()
	store := posStore(t, a, bhoomyStore)
	if code := a.Back(http.MethodPost, "/locations", J{"type": "kiosk", "name": "Kios"}, bhoomyStore).Code; code != http.StatusBadRequest {
		t.Errorf("unknown type: got %d, want 400", code)
	}
	ev := ok[J](t, a.Back(http.MethodPost, "/locations", J{"type": "event", "name": "Bazaar", "startDate": "2027-01-01", "endDate": "2027-01-07"}, bhoomyStore), http.StatusCreated)
	if ev["storeId"] != store.String() {
		t.Fatalf("event: %v", ev)
	}

	// The outlet comes with the Cabang; editing it still needs an address and no end date.
	shop := outletOf(t, a, bhoomyStore)
	put := func(body J) int {
		return a.Back(http.MethodPut, "/locations/"+shop["id"].(string), body, bhoomyStore).Code
	}
	if got := put(J{"endDate": "2026-12-31", "address": "Jl. Kemang 1"}); got != http.StatusBadRequest {
		t.Errorf("outlet with end date: got %d, want 400", got)
	}
	if got := put(J{"address": " Jl. Kemang 1 "}); got != http.StatusOK {
		t.Fatalf("outlet address: got %d", got)
	}
	if shop = outletOf(t, a, bhoomyStore); shop["address"] != "Jl. Kemang 1" {
		t.Errorf("address = %v", shop["address"])
	}
}
