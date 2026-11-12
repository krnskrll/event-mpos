package app_test

import (
	"net/http"
	"testing"

	"github.com/google/uuid"

	"github.com/krnskrll/event-mpos/internal/testutil"
)

func TestCreateStoreAdmin(t *testing.T) {
	a := testutil.NewApp(t)
	bhoomyStore := uuid.New()
	owner := uuid.NewString() + "@toko.id"
	if _, err := a.CreateStoreAdmin(bhoomyStore, owner, "Rina", "short"); err == nil {
		t.Error("short password accepted")
	}
	u, err := a.CreateStoreAdmin(bhoomyStore, owner, "Rina", "rahasia123")
	if err != nil {
		t.Fatal(err)
	}
	if u.StoreID != posStore(t, a, bhoomyStore) {
		t.Errorf("admin in store %s, want the Bhoomy store's POS store", u.StoreID)
	}
	login := ok[J](t, a.Do(http.MethodPost, "/api/auth/login", J{"email": owner, "password": "rahasia123"}, ""), http.StatusOK)
	token, _ := login["accessToken"].(string)
	ev := a.ActiveEventIn(u.StoreID)
	var own []J
	for _, e := range ok[[]J](t, a.Do(http.MethodGet, "/api/admin/events", nil, token), http.StatusOK) {
		if e["type"] != "outlet" { // the Cabang's own outlet comes with the store
			own = append(own, e)
		}
	}
	if events := own; len(events) != 1 || events[0]["id"] != ev.ID.String() {
		t.Errorf("store events %v, want only its own location", events)
	}
	if _, err := a.CreateStoreAdmin(uuid.New(), owner, "Rina", "rahasia123"); err == nil {
		t.Error("same email twice accepted")
	}
}
