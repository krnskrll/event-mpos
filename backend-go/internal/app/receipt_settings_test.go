package app_test

import (
	"net/http"
	"strings"
	"testing"

	"github.com/google/uuid"

	deviceiface "github.com/krnskrll/event-mpos/internal/device/iface"
	receiptiface "github.com/krnskrll/event-mpos/internal/receipt/iface"
	"github.com/krnskrll/event-mpos/internal/testutil"
)

// Pengaturan Struk: one receipt setting per store, edited in Bhoomy, sent to the device with its pairing data.
func TestReceiptSettings(t *testing.T) {
	a := testutil.NewApp(t)
	bhoomyStore, other := uuid.New(), uuid.New()
	store := posStore(t, a, bhoomyStore)
	get := func(s uuid.UUID) receiptiface.Settings {
		return ok[receiptiface.Settings](t, a.Back(http.MethodGet, "/receipt", nil, s), http.StatusOK)
	}

	if got := get(bhoomyStore); got != (receiptiface.Settings{ShowAddress: true, ShowStaff: true, ShowCustomer: true}) {
		t.Fatalf("defaults %+v: want empty contact and note, everything shown", got)
	}
	saved := ok[receiptiface.Settings](t, a.Back(http.MethodPut, "/receipt", J{"phone": " 0812-3456-7890 ", "instagram": "@glowmart.bdg",
		"note": "Tukar barang maks. 3 hari", "showAddress": false, "showStaff": true, "showCustomer": false}, bhoomyStore), http.StatusOK)
	want := receiptiface.Settings{Phone: "0812-3456-7890", Instagram: "glowmart.bdg", Note: "Tukar barang maks. 3 hari", ShowStaff: true}
	if saved != want || get(bhoomyStore) != want {
		t.Fatalf("saved %+v, read back %+v: want %+v (trimmed, no @)", saved, get(bhoomyStore), want)
	}
	for _, bad := range []J{{"note": strings.Repeat("x", 65)}, {"phone": strings.Repeat("1", 21)}, {"instagram": strings.Repeat("a", 31)}} {
		if rec := a.Back(http.MethodPut, "/receipt", bad, bhoomyStore); rec.Code != http.StatusBadRequest {
			t.Errorf("%v: got %d, want 400", bad, rec.Code)
		}
	}
	if got := get(other); got.Phone != "" || !got.ShowAddress {
		t.Errorf("another store sees %+v", got)
	}

	ev := a.ActiveEventIn(store)
	code := ok[deviceiface.PairingCode](t, a.Back(http.MethodPost, "/locations/"+ev.ID.String()+"/pairing-code", nil, bhoomyStore), http.StatusCreated)
	paired := pair(t, a, code.Code, "10.0.0.9")
	if paired.Receipt != want {
		t.Errorf("pair reply receipt %+v, want %+v", paired.Receipt, want)
	}
	device := testutil.Decode[deviceiface.DeviceResponse](t, a.Send(http.MethodGet, "/api/device", nil, map[string]string{"Authorization": "Device " + paired.Token}))
	if device.Receipt != want {
		t.Errorf("GET /api/device receipt %+v, want %+v", device.Receipt, want)
	}
}
