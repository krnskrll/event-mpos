# Epic 11 — Back Office Administration

## Context
Read `.cursorrules` and review everything built in Epics 01–10 before starting.
This is the administrative layer that sits on top of the entire system. All features here are Admin-only.

## Goal
Give Admins full visibility and control over the system: user accounts, sync schedules, and post-event reporting.

## What Needs to Work When This Epic Is Done
- Admins can create, edit, deactivate, and assign roles to user accounts
- Admins can configure the ERP product/stock sync schedule and trigger a manual sync; last sync time and status are visible
- Admins can configure the member sync schedule and trigger a manual sync
- After an event closes, Admins can generate an end-of-event report showing: total orders, total revenue, items sold per product, and remaining stock vs allocated — exportable as CSV
- Admins can view an audit log of all order activity for any event: who created it, when, what status changes happened

## Things Worth Paying Attention To
- All admin routes and API endpoints must be protected — Admins only
- The end-of-event report query could be slow on large datasets — think about indexing strategy before writing the query
- The audit log should be append-only — explore the right data model for this; it should never be updated or deleted
- Sync configuration should be stored in the database, not in config files, so it can be changed at runtime without a redeploy

## When Done

```bash
git checkout main && git pull origin main
git checkout -b feature/epic-11-backoffice-admin
git add .
git commit -m "feat: epic-11 back office administration and reporting"
git push origin feature/epic-11-backoffice-admin
```

Then open a Pull Request to `main` on https://github.com/krnskrll/event-mpos

---

## After All Epics Are Complete

```bash
# Final merge and tag
git checkout main
git pull origin main
git tag -a v1.0.0 -m "Phase 1 complete"
git push origin v1.0.0
```
