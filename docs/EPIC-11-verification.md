# EPIC-11 — How to verify

EPIC-11 is the **Admin-only back office**: user management, runtime-editable sync schedules, the
end-of-event report (+ CSV) and an append-only audit log of order activity. Builds on EPIC-01–10.
Includes the EPIC-12 Milestone D fixes (CLOSED-only report, paginated audit, receipt-print audit).

> Schema is managed by golang-migrate (`backend-go/migrations/`). After pulling, `docker compose up
> --build` applies new migrations automatically.

---

## Backend (curl)
```bash
B=http://localhost:8080
ADMIN=$(curl -s -X POST $B/api/auth/login -H 'Content-Type: application/json' -d '{"email":"admin@eventmpos.local","password":"Admin123!"}' | python3 -c "import sys,json;print(json.load(sys.stdin)['accessToken'])")
EV01=$(curl -s $B/api/admin/events -H "Authorization: Bearer $ADMIN" | python3 -c "import sys,json;print([e['id'] for e in json.load(sys.stdin) if e['code']=='EV01'][0])")
```
- **Users**: `PATCH /api/admin/users/{id} {"role":"WAREHOUSE","active":false}` → role changed, user
  deactivated, other fields kept.
- **Sync schedules** (stored in the DB, picked up without a redeploy): `GET /api/admin/sync`;
  `PUT /api/admin/sync/ERP {"enabled":true,"cron":"0 0 */2 * * *"}`. A bad cron → **400**.
- **End-of-event report**: `GET /api/admin/events/$EV01/report` and `/report.csv`.
  - While the event is ACTIVE → **409** "…available once EV01 is CLOSED".
  - Once CLOSED → totals (orders, items, revenue, discount, cancelled / late-synced / stock-conflict
    counts) and per-product units, revenue, allocated, sold, remaining. Cancelled orders are excluded.
- **Audit log** (newest first): `GET /api/admin/events/$EV01/audit?page=0&size=50` →
  `{items, page, size, total}`. Filters: `actor=<userId>`, `action=ORDER_CREATED|STATUS_CHANGED|RECEIPT_PRINTED|ORDER_CANCELLED`,
  `orderId=`. `size` is capped at 200; a bad filter → **400**.
- **Append-only**: `docker compose exec db psql -U eventmpos -c "DELETE FROM order_audit_log"` →
  `ERROR: order_audit_log is append-only`.
- **Admin only**: a BA or warehouse token on `/api/admin/*` → **403**.

## Web dashboard
Sign in as admin → **Reports & Sync**:
- Edit a sync schedule and **Save**; **Sync now** runs it immediately and shows last run + status.
- Pick an ACTIVE event → a notice explains the final report isn't available yet; the audit trail
  below is live. Pick a CLOSED event → metrics, product table and **Download CSV**.
- Audit trail: filter by user and action, choose 50/100/200 per page, page with **← Newer / Older →**
  ("51–100 of 268 · page 2 of 6").

## Automated checks (CI)
`backoffice_admin_test.go` and `backoffice_fixes_test.go` (Postgres): user edit, sync config, report
totals + CSV, report gated on CLOSED, audit paging/filters on 125 entries, receipt prints audited in
the same transaction (10 concurrent prints → count 12), append-only trigger.

---

## Acceptance checklist (from the epic)
- [x] Admins can create, edit, deactivate and assign roles to users
- [x] ERP sync schedule configurable + manual trigger; last sync time and status visible
- [x] Member sync schedule configurable + manual trigger
- [x] End-of-event report (orders, revenue, items per product, remaining vs allocated) exportable as CSV — for CLOSED events
- [x] Audit log of all order activity per event (who, when, status changes), paginated and filterable
- [x] Admin-only routes; sync config stored in the DB; audit log append-only
