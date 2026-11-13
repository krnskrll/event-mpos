package app_test

import (
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"

	deviceiface "github.com/krnskrll/event-mpos/internal/device/iface"
	devicerepo "github.com/krnskrll/event-mpos/internal/device/repository"
	eventiface "github.com/krnskrll/event-mpos/internal/event/iface"
	staffiface "github.com/krnskrll/event-mpos/internal/staff/iface"
	"github.com/krnskrll/event-mpos/internal/testutil"
)

func pair(t *testing.T, a *testutil.TestApp, code, ip string) deviceiface.PairResponse {
	rec := a.Send(http.MethodPost, "/api/device/pair", J{"code": code, "model": "Sunmi V2"}, map[string]string{"RemoteAddr": ip + ":4000"})
	if rec.Code != http.StatusCreated {
		return deviceiface.PairResponse{}
	}
	return testutil.Decode[deviceiface.PairResponse](t, rec)
}

func TestPairing(t *testing.T) {
	a := testutil.NewApp(t)
	bhoomyStore := uuid.New()
	store := posStore(t, a, bhoomyStore)
	ev := a.ActiveEventIn(store)
	loc := "/locations/" + ev.ID.String()

	code := ok[deviceiface.PairingCode](t, a.Back(http.MethodPost, loc+"/pairing-code", nil, bhoomyStore), http.StatusCreated)
	if len(code.Code) != 6 || time.Until(code.ExpiresAt) < 9*time.Minute || time.Until(code.ExpiresAt) > 10*time.Minute {
		t.Fatalf("code %+v, want 6 digits valid 10 minutes", code)
	}
	first, second := pair(t, a, code.Code, "10.0.0.1"), pair(t, a, code.Code, "10.0.0.2")
	if first.Device.Name != "Kasir 1" || second.Device.Name != "Kasir 2" || first.Token == "" || first.Token == second.Token {
		t.Fatalf("one code, two devices: %+v / %+v", first.Device, second.Device)
	}
	if l := first.Location; l.ID != ev.ID || l.Name != ev.Name || l.Type != eventiface.TypeEvent || l.Status != eventiface.StatusActive || first.StoreName == "" {
		t.Errorf("pair response %+v", first)
	}
	var stored devicerepo.Device
	a.DB.Take(&stored, "id = ?", first.Device.ID)
	if stored.TokenHash == first.Token || len(stored.TokenHash) != 64 || stored.Model == nil || *stored.Model != "Sunmi V2" {
		t.Errorf("stored device %+v: want the token's sha256 and the model", stored)
	}

	// Names: rename, remove, and the next device takes the lowest free "Kasir N".
	ok[deviceiface.DeviceView](t, a.Back(http.MethodPatch, "/devices/"+first.Device.ID.String(), J{"name": "Depan"}, bhoomyStore), http.StatusOK)
	if rec := a.Back(http.MethodPatch, "/devices/"+first.Device.ID.String(), J{"name": "Kasir 2"}, bhoomyStore); rec.Code != http.StatusConflict {
		t.Errorf("rename to a taken name: %d, want 409", rec.Code)
	}
	if third := pair(t, a, code.Code, "10.0.0.3"); third.Device.Name != "Kasir 1" {
		t.Errorf("after rename got %q, want Kasir 1", third.Device.Name)
	}
	if rec := a.Back(http.MethodDelete, "/devices/"+second.Device.ID.String(), nil, bhoomyStore); rec.Code != http.StatusNoContent {
		t.Fatalf("remove: %d", rec.Code)
	}
	if fourth := pair(t, a, code.Code, "10.0.0.4"); fourth.Device.Name != "Kasir 2" {
		t.Errorf("after remove got %q, want Kasir 2", fourth.Device.Name)
	}

	// The device API: heartbeat, then 401 once removed.
	dev := testutil.Device{ID: first.Device.ID, Token: first.Token}
	me := ok[deviceiface.DeviceResponse](t, a.DeviceDo(http.MethodGet, "/api/device", nil, dev, uuid.Nil), http.StatusOK)
	if me.Device.Name != "Depan" || me.Location.ID != ev.ID {
		t.Errorf("GET /api/device = %+v", me)
	}
	if rec := a.DeviceDo(http.MethodGet, "/api/device", nil, testutil.Device{Token: second.Token}, uuid.Nil); rec.Code != http.StatusUnauthorized {
		t.Errorf("removed device: %d, want 401", rec.Code)
	}
	if rec := a.DeviceDo(http.MethodGet, "/api/device", nil, testutil.Device{Token: "nope"}, uuid.Nil); rec.Code != http.StatusUnauthorized {
		t.Errorf("unknown token: %d, want 401", rec.Code)
	}
	list := ok[[]deviceiface.DeviceView](t, a.Back(http.MethodGet, loc+"/devices", nil, bhoomyStore), http.StatusOK)
	if len(list) != 3 || list[0].Name != "Depan" || list[0].LastSeenAt == nil || list[1].LastSeenAt != nil {
		t.Errorf("devices %+v, want 3 live with Depan seen", list)
	}

	// A new code replaces the old one; an expired code stops working.
	fresh := ok[deviceiface.PairingCode](t, a.Back(http.MethodPost, loc+"/pairing-code", nil, bhoomyStore), http.StatusCreated)
	if fresh.Code != code.Code && pair(t, a, code.Code, "10.0.0.5").Token != "" {
		t.Error("replaced code still pairs")
	}
	a.DB.Model(&devicerepo.PairingCode{}).Where("code = ?", fresh.Code).Update("expires_at", time.Now().Add(-time.Second))
	if pair(t, a, fresh.Code, "10.0.0.6").Token != "" {
		t.Error("expired code still pairs")
	}

	// Pairing is rate limited per client.
	got := 0
	for range 11 {
		got = a.Send(http.MethodPost, "/api/device/pair", J{"code": "000000"}, map[string]string{"RemoteAddr": "10.9.9.9:1"}).Code
	}
	if got != http.StatusTooManyRequests {
		t.Errorf("11th attempt in a minute: %d, want 429", got)
	}
	// Behind Caddy (a private address) each client counts on its own.
	if c := a.Send(http.MethodPost, "/api/device/pair", J{"code": "000000"}, map[string]string{"RemoteAddr": "10.9.9.9:1", "X-Forwarded-For": "203.0.113.7"}).Code; c == http.StatusTooManyRequests {
		t.Error("a client behind the proxy shares the proxy's limit")
	}

	// Another store sees none of it; a closed location takes no new devices.
	other := uuid.New()
	for _, r := range []struct{ method, path string }{
		{http.MethodGet, loc + "/devices"}, {http.MethodPost, loc + "/pairing-code"},
		{http.MethodPatch, "/devices/" + first.Device.ID.String()}, {http.MethodDelete, "/devices/" + first.Device.ID.String()},
	} {
		if rec := a.Back(r.method, r.path, J{"name": "x"}, other); rec.Code != http.StatusNotFound {
			t.Errorf("other store %s %s: %d, want 404", r.method, r.path, rec.Code)
		}
	}
	ok[J](t, a.Back(http.MethodPatch, loc+"/status", J{"status": "CLOSED"}, bhoomyStore), http.StatusOK)
	if rec := a.Back(http.MethodPost, loc+"/pairing-code", nil, bhoomyStore); rec.Code != http.StatusConflict {
		t.Errorf("code for a closed location: %d, want 409", rec.Code)
	}
}

func TestStafPOS(t *testing.T) {
	a := testutil.NewApp(t)
	bhoomyStore := uuid.New()
	store := posStore(t, a, bhoomyStore)
	ev, outlet, closed, foreign := a.ActiveEventIn(store), a.ActiveEventIn(store), a.ActiveEventIn(store), a.ActiveEvent()
	if _, err := a.Events.ChangeStatus(store, closed.ID, eventiface.StatusClosed); err != nil {
		t.Fatal(err)
	}
	body := func(name, role, pin string, locs ...uuid.UUID) J {
		b := J{"name": name, "role": role, "locationIds": locs}
		if pin != "" {
			b["pin"] = pin
		}
		return b
	}
	for name, c := range map[string]struct {
		body J
		want int
	}{
		"no pin":         {body("Sari", "KASIR", ""), http.StatusBadRequest},
		"short pin":      {body("Sari", "KASIR", "123"), http.StatusBadRequest},
		"long pin":       {body("Sari", "KASIR", "1234567"), http.StatusBadRequest},
		"letters in pin": {body("Sari", "KASIR", "12ab"), http.StatusBadRequest},
		"bad role":       {body("Sari", "BA", "1234"), http.StatusBadRequest},
		"no name":        {body(" ", "KASIR", "1234"), http.StatusBadRequest},
		"long name":      {body(strings.Repeat("a", 41), "KASIR", "1234"), http.StatusBadRequest},
		"other store":    {body("Sari", "KASIR", "1234", foreign.ID), http.StatusNotFound},
		"closed":         {body("Sari", "KASIR", "1234", closed.ID), http.StatusBadRequest},
	} {
		if rec := a.Back(http.MethodPost, "/staff", c.body, bhoomyStore); rec.Code != c.want {
			t.Errorf("%s: %d, want %d (%s)", name, rec.Code, c.want, rec.Body.String())
		}
	}

	rec := a.Back(http.MethodPost, "/staff", body("Sari", "SUPERVISOR", "246810", ev.ID, ev.ID), bhoomyStore)
	sari := ok[staffiface.StaffResponse](t, rec, http.StatusCreated)
	if sari.Role != staffiface.RoleSupervisor || !sari.Active || len(sari.LocationIDs) != 1 || strings.Contains(rec.Body.String(), "pin") {
		t.Fatalf("created %s", rec.Body.String())
	}
	budi := ok[staffiface.StaffResponse](t, a.Back(http.MethodPost, "/staff", body("Budi", "KASIR", "1111", ev.ID), bhoomyStore), http.StatusCreated)
	a.NewStaff(store, staffiface.RoleKasir, outlet.ID)

	// The device sees its location's active staff with PIN hashes, to sign in offline.
	dev := a.PairDevice(ev.ID)
	staff := ok[[]staffiface.DeviceStaff](t, a.DeviceDo(http.MethodGet, "/api/device/staff", nil, dev, uuid.Nil), http.StatusOK)
	if len(staff) != 2 || staff[0].Name != "Budi" || staff[1].ID != sari.ID || bcrypt.CompareHashAndPassword([]byte(staff[1].PinHash), []byte("246810")) != nil {
		t.Fatalf("device staff %+v", staff)
	}
	if cost, _ := bcrypt.Cost([]byte(staff[1].PinHash)); cost != 10 {
		t.Errorf("bcrypt cost %d, want 10", cost)
	}

	// Update: PIN reset, deactivate, locations replaced; a closed location stays but cannot be added.
	upd := body("Budi S", "KASIR", "9999", outlet.ID)
	upd["active"] = false
	got := ok[staffiface.StaffResponse](t, a.Back(http.MethodPut, "/staff/"+budi.ID.String(), upd, bhoomyStore), http.StatusOK)
	if got.Name != "Budi S" || got.Active || len(got.LocationIDs) != 1 || got.LocationIDs[0] != outlet.ID {
		t.Errorf("updated %+v", got)
	}
	if staff := ok[[]staffiface.DeviceStaff](t, a.DeviceDo(http.MethodGet, "/api/device/staff", nil, dev, uuid.Nil), http.StatusOK); len(staff) != 1 {
		t.Errorf("after deactivating Budi the device sees %d staff, want 1", len(staff))
	}
	if rec := a.Back(http.MethodPut, "/staff/"+budi.ID.String(), body("Budi", "KASIR", "", outlet.ID, closed.ID), bhoomyStore); rec.Code != http.StatusBadRequest {
		t.Errorf("adding a closed location: %d, want 400", rec.Code)
	}
	a.DB.Exec("INSERT INTO pos_staff_locations (staff_id, location_id) VALUES (?, ?)", budi.ID, closed.ID)
	ok[staffiface.StaffResponse](t, a.Back(http.MethodPut, "/staff/"+budi.ID.String(), body("Budi", "KASIR", "", outlet.ID, closed.ID), bhoomyStore), http.StatusOK)

	list := ok[[]staffiface.StaffResponse](t, a.Back(http.MethodGet, "/staff", nil, bhoomyStore), http.StatusOK)
	if len(list) != 3 {
		t.Errorf("store has %d staff, want 3", len(list))
	}
	if got := ok[staffiface.StaffResponse](t, a.Back(http.MethodGet, "/staff/"+sari.ID.String(), nil, bhoomyStore), http.StatusOK); got.Name != "Sari" {
		t.Errorf("get %+v", got)
	}

	// Another store: 404 and an empty list.
	other := uuid.New()
	for _, r := range []struct{ method, path string }{{http.MethodGet, "/staff/" + sari.ID.String()}, {http.MethodPut, "/staff/" + sari.ID.String()}} {
		if rec := a.Back(r.method, r.path, body("X", "KASIR", ""), other); rec.Code != http.StatusNotFound {
			t.Errorf("other store %s: %d, want 404", r.method, rec.Code)
		}
	}
	if list := ok[[]staffiface.StaffResponse](t, a.Back(http.MethodGet, "/staff", nil, other), http.StatusOK); len(list) != 0 {
		t.Errorf("other store lists %d staff", len(list))
	}
}

// A closed location's devices keep 24 hours to upload waiting orders, then stop working.
func TestClosingRemovesDevicesAfter24h(t *testing.T) {
	a := testutil.NewApp(t)
	bhoomyStore := uuid.New()
	store := posStore(t, a, bhoomyStore)
	// Pausing the outlet keeps its devices: only a closed event starts the 24 hours.
	outlet := outletOf(t, a, bhoomyStore)
	paired := a.PairDevice(uuid.MustParse(outlet["id"].(string)))
	ok[J](t, a.Back(http.MethodPatch, "/locations/"+outlet["id"].(string)+"/status", J{"status": "ACTIVE"}, bhoomyStore), http.StatusOK)
	ok[J](t, a.Back(http.MethodPatch, "/locations/"+outlet["id"].(string)+"/status", J{"status": "DRAFT"}, bhoomyStore), http.StatusOK)
	if me := ok[deviceiface.DeviceResponse](t, a.DeviceDo(http.MethodGet, "/api/device", nil, paired, uuid.Nil), http.StatusOK); me.Device.RemovesAt != nil {
		t.Errorf("paused outlet: removesAt %v, want none", me.Device.RemovesAt)
	}
	for _, id := range []uuid.UUID{a.ActiveEventIn(store).ID} {
		dev, still := a.PairDevice(id), a.PairDevice(id)
		ok[J](t, a.Back(http.MethodPatch, "/locations/"+id.String()+"/status", J{"status": "CLOSED"}, bhoomyStore), http.StatusOK)
		me := ok[deviceiface.DeviceResponse](t, a.DeviceDo(http.MethodGet, "/api/device", nil, dev, uuid.Nil), http.StatusOK)
		if me.Device.RemovesAt == nil || time.Until(*me.Device.RemovesAt) < 23*time.Hour {
			t.Fatalf("after close: removesAt %v, want about 24h ahead", me.Device.RemovesAt)
		}
		a.DB.Model(&devicerepo.Device{}).Where("id = ?", dev.ID).Update("removed_at", time.Now().Add(-time.Second))
		if rec := a.DeviceDo(http.MethodGet, "/api/device", nil, dev, uuid.Nil); rec.Code != http.StatusUnauthorized {
			t.Errorf("24h after close: %d, want 401", rec.Code)
		}
		if rec := a.DeviceDo(http.MethodGet, "/api/device", nil, still, uuid.Nil); rec.Code != http.StatusOK {
			t.Errorf("within the window: %d, want 200", rec.Code)
		}
	}
}
