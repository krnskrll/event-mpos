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
	owner, cashier := uuid.NewString()+"@toko.id", uuid.NewString()+"@toko.id"
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
	ok[J](t, a.Do(http.MethodPost, "/api/admin/users", J{"email": cashier, "password": "rahasia123", "fullName": "Budi", "role": "BA"}, token), http.StatusCreated)
	users := ok[[]J](t, a.Do(http.MethodGet, "/api/admin/users", nil, token), http.StatusOK)
	if len(users) != 2 {
		t.Errorf("store users %v, want the admin and the new cashier only", users)
	}
	if _, err := a.CreateStoreAdmin(uuid.New(), owner, "Rina", "rahasia123"); err == nil {
		t.Error("same email twice accepted")
	}
}
