package app_test

import (
	"net/http"
	"strings"
	"testing"

	"github.com/google/uuid"

	"github.com/krnskrll/event-mpos/internal/audit"
	"github.com/krnskrll/event-mpos/internal/order"
	"github.com/krnskrll/event-mpos/internal/syncconfig"
	"github.com/krnskrll/event-mpos/internal/testutil"
	"github.com/krnskrll/event-mpos/internal/user"
)

// Ported from the Kotlin BackofficeAdminTest.

func TestAdminCanEditRoleAndDeactivateUser(t *testing.T) {
	a := testutil.NewApp(t)
	created := a.CreateUser(user.RoleBA, "password1")
	role, active := user.RoleWarehouse, false
	updated, err := a.Users.Update(created.ID, user.UpdateUserRequest{Role: &role, Active: &active})
	if err != nil {
		t.Fatal(err)
	}
	if updated.Role != user.RoleWarehouse || updated.Active {
		t.Errorf("role=%s active=%v, want WAREHOUSE/false", updated.Role, updated.Active)
	}
	if updated.FullName != created.FullName {
		t.Errorf("fullName changed to %q", updated.FullName)
	}
}

func TestSyncConfigValidatesCronAndPersists(t *testing.T) {
	a := testutil.NewApp(t)
	before, err := a.SyncConfig.Get(syncconfig.TypeERP)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _, _ = a.SyncConfig.Update(syncconfig.TypeERP, before.Enabled, before.Cron) })

	cfg, err := a.SyncConfig.Update(syncconfig.TypeERP, true, "0 0 */2 * * *")
	if err != nil {
		t.Fatal(err)
	}
	if !cfg.Enabled || cfg.Cron != "0 0 */2 * * *" {
		t.Errorf("saved %+v", cfg)
	}
	reread, _ := a.SyncConfig.Get(syncconfig.TypeERP)
	if reread.Cron != "0 0 */2 * * *" {
		t.Errorf("not persisted: %+v", reread)
	}
	if _, err := a.SyncConfig.Update(syncconfig.TypeMember, true, "not a cron"); testutil.StatusOf(err) != http.StatusBadRequest {
		t.Errorf("bad cron: got %v, want 400", err)
	}
}

func TestEndOfEventReportTotalsAndCSV(t *testing.T) {
	a := testutil.NewApp(t)
	ev, p := a.ActiveEvent(), a.Product("50.00")
	a.Allocate(ev.ID, p.ID, 10)
	if _, err := submit(a, ev.ID, uuid.NewString(), line(p.ID, 3)); err != nil {
		t.Fatal(err)
	}
	if _, err := submit(a, ev.ID, uuid.NewString(), line(p.ID, 2)); err != nil {
		t.Fatal(err)
	}
	closeEvent(t, a, ev.ID)

	r, err := a.Reports.EndOfEvent(ev.ID)
	if err != nil {
		t.Fatal(err)
	}
	if r.TotalOrders != 2 || r.TotalItems != 5 || !r.TotalRevenue.Equal(dec("250")) {
		t.Errorf("orders/items/revenue = %d/%d/%s, want 2/5/250", r.TotalOrders, r.TotalItems, r.TotalRevenue)
	}
	var found bool
	for _, row := range r.Products {
		if row.ProductID != p.ID {
			continue
		}
		found = true
		if row.UnitsSold != 5 || !row.Revenue.Equal(dec("250")) || row.Allocated != 10 || row.Sold != 5 || row.Remaining != 5 {
			t.Errorf("row = %+v", row)
		}
	}
	if !found {
		t.Fatal("product missing from report")
	}

	csv := a.Reports.ToCSV(r)
	if !strings.Contains(csv, "SKU,Product,Units Sold,Revenue,Allocated,Sold,Remaining") {
		t.Errorf("csv header missing:\n%s", csv)
	}
	if want := p.SKU + ",Widget,5,250.00,10,5,5"; !strings.Contains(csv, want) {
		t.Errorf("csv row %q missing:\n%s", want, csv)
	}

	// Same report over HTTP for an admin.
	admin := a.CreateUser(user.RoleAdmin, "Admin123!")
	tok := login(t, a, admin.Email, "Admin123!").AccessToken
	if rec := a.Do(http.MethodGet, "/api/admin/events/"+ev.ID.String()+"/report.csv", nil, tok); rec.Code != http.StatusOK {
		t.Errorf("report.csv: %d", rec.Code)
	}
}

func TestOrderActivityIsInAppendOnlyAuditLog(t *testing.T) {
	a := testutil.NewApp(t)
	ev, p := a.ActiveEvent(), a.Product("20.00")
	a.Allocate(ev.ID, p.ID, 50)
	wh := a.CreateUser(user.RoleWarehouse, "Wh123456!")

	conf, err := submit(a, ev.ID, uuid.NewString(), line(p.ID, 1))
	if err != nil {
		t.Fatal(err)
	}
	if _, err := a.Orders.AdvanceStatus(conf.ID, order.StatusPacking, &wh.ID); err != nil {
		t.Fatal(err)
	}

	trail, err := a.Audit.ForOrder(conf.ID)
	if err != nil {
		t.Fatal(err)
	}
	if len(trail) != 2 || trail[0].Action != audit.ActionOrderCreated || trail[1].Action != audit.ActionStatusChanged {
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
