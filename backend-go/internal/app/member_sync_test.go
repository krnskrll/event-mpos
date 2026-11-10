package app_test

import (
	"net/http"
	"testing"

	"github.com/google/uuid"

	syncconfigiface "github.com/krnskrll/event-mpos/internal/syncconfig/iface"
	"github.com/krnskrll/event-mpos/internal/testutil"
)

// Member POS sync bar: status, sync now, and a daily WIB schedule on the member sync cron.
func TestMemberSyncBar(t *testing.T) {
	a := testutil.NewApp(t)
	before, err := a.SyncConfig.Get(syncconfigiface.TypeMember)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _, _ = a.SyncConfig.Update(syncconfigiface.TypeMember, before.Enabled, before.Cron) })
	store := uuid.New()

	ok[syncconfigiface.MemberSync](t, a.Back(http.MethodGet, "/members/sync", nil, store), http.StatusOK)
	ran := ok[syncconfigiface.MemberSync](t, a.Back(http.MethodPost, "/members/sync", nil, store), http.StatusOK)
	if ran.LastSyncAt == nil || ran.LastResult == nil || *ran.LastResult != "SUCCESS" || ran.Count == 0 {
		t.Errorf("after sync now: %+v", ran)
	}

	got := ok[syncconfigiface.MemberSync](t, a.Back(http.MethodPut, "/members/sync/schedule", J{"enabled": true, "time": "02:30"}, store), http.StatusOK)
	if !got.Enabled || got.Time == nil || *got.Time != "02:30" {
		t.Errorf("schedule %+v, want enabled at 02:30", got)
	}
	if cfg, _ := a.SyncConfig.Get(syncconfigiface.TypeMember); cfg.Cron != "CRON_TZ=Asia/Jakarta 0 30 2 * * *" {
		t.Errorf("stored cron %q", cfg.Cron)
	}
	if got := ok[syncconfigiface.MemberSync](t, a.Back(http.MethodGet, "/members/sync", nil, store), http.StatusOK); got.Time == nil || *got.Time != "02:30" {
		t.Errorf("status after schedule %+v", got)
	}
	for _, bad := range []J{{"enabled": true, "time": "25:00"}, {"enabled": true, "time": "2.30"}, {"enabled": true}} {
		if rec := a.Back(http.MethodPut, "/members/sync/schedule", bad, store); rec.Code != http.StatusBadRequest {
			t.Errorf("schedule %v: %d, want 400", bad, rec.Code)
		}
	}
	if rec := a.Do(http.MethodGet, "/api/backoffice/members/sync", nil, ""); rec.Code != http.StatusUnauthorized {
		t.Errorf("without the service token: %d, want 401", rec.Code)
	}
}
