package app_test

import (
	"net/http"
	"strings"
	"testing"

	"github.com/google/uuid"

	auditiface "github.com/krnskrll/event-mpos/internal/audit/iface"
	orderiface "github.com/krnskrll/event-mpos/internal/order/iface"
	"github.com/krnskrll/event-mpos/internal/platform/tenant"
	syncconfigiface "github.com/krnskrll/event-mpos/internal/syncconfig/iface"
	"github.com/krnskrll/event-mpos/internal/testutil"
	useriface "github.com/krnskrll/event-mpos/internal/user/iface"
)

// Ported from the Kotlin BackofficeAdminTest.

func TestAdminCanEditRoleAndDeactivateUser(t *testing.T) {
	a := testutil.NewApp(t)
	created := a.CreateUser(useriface.RoleBA, "password1")
	role, active := useriface.RoleWarehouse, false
	updated, err := a.Users.Update(tenant.DefaultStoreID, created.ID, useriface.UpdateUserRequest{Role: &role, Active: &active})
	if err != nil {
		t.Fatal(err)
	}
	if updated.Role != useriface.RoleWarehouse || updated.Active {
		t.Errorf("role=%s active=%v, want WAREHOUSE/false", updated.Role, updated.Active)
	}
	if updated.FullName != created.FullName {
		t.Errorf("fullName changed to %q", updated.FullName)
	}
}

func TestSyncConfigValidatesCronAndPersists(t *testing.T) {
	a := testutil.NewApp(t)
	before, err := a.SyncConfig.Get(syncconfigiface.TypeMember)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _, _ = a.SyncConfig.Update(syncconfigiface.TypeMember, before.Enabled, before.Cron) })

	cfg, err := a.SyncConfig.Update(syncconfigiface.TypeMember, true, "0 0 */2 * * *")
	if err != nil {
		t.Fatal(err)
	}
	if !cfg.Enabled || cfg.Cron != "0 0 */2 * * *" {
		t.Errorf("saved %+v", cfg)
	}
	reread, _ := a.SyncConfig.Get(syncconfigiface.TypeMember)
	if reread.Cron != "0 0 */2 * * *" {
		t.Errorf("not persisted: %+v", reread)
	}
	if _, err := a.SyncConfig.Update(syncconfigiface.TypeMember, true, "not a cron"); testutil.StatusOf(err) != http.StatusBadRequest {
		t.Errorf("bad cron: got %v, want 400", err)
	}
}

func TestOrderActivityIsInAppendOnlyAuditLog(t *testing.T) {
	a := testutil.NewApp(t)
	ev, p := a.ActiveEvent(), a.Product("20.00")
	a.Allocate(ev.ID, p.ID, 50)
	wh := a.CreateUser(useriface.RoleWarehouse, "Wh123456!")

	conf, err := submit(a, ev.ID, uuid.NewString(), line(p.ID, 1))
	if err != nil {
		t.Fatal(err)
	}
	if _, err := a.Orders.AdvanceStatus(conf.ID, orderiface.StatusPacking, &wh.ID); err != nil {
		t.Fatal(err)
	}

	trail, err := a.Audit.ForOrder(conf.ID)
	if err != nil {
		t.Fatal(err)
	}
	if len(trail) != 2 || trail[0].Action != auditiface.ActionOrderCreated || trail[1].Action != auditiface.ActionStatusChanged {
		t.Fatalf("trail = %+v", trail)
	}
	if *trail[1].FromStatus != "NEW" || *trail[1].ToStatus != "PACKING" || trail[1].ActorName != wh.FullName {
		t.Errorf("status entry = %+v", trail[1])
	}
	if all, _ := a.Audit.ForEvent(ev.ID); len(all) != 2 {
		t.Errorf("event trail has %d entries, want 2", len(all))
	}

	// The DB trigger refuses to rewrite history.
	if err := a.DB.Exec("UPDATE order_audit_log SET action = 'X' WHERE order_id = ?", conf.ID).Error; err == nil ||
		!strings.Contains(err.Error(), "append-only") {
		t.Errorf("UPDATE on audit log: %v, want append-only error", err)
	}
	if err := a.DB.Exec("DELETE FROM order_audit_log WHERE order_id = ?", conf.ID).Error; err == nil {
		t.Error("DELETE on audit log succeeded")
	}
}
