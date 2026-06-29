# Epic 02 — Event Management

## Context
Read `.cursorrules` and review what was built in Epic 01 before starting.
The **Event** is the central domain boundary of this entire system — everything else (products, stock, orders, promotions) belongs to an event.

## Goal
Enable Admins to create and manage events through the web dashboard, and enable the Android app to let a BA select and sync their assigned event before doing anything else.

## What Needs to Work When This Epic Is Done
- Admin can create, edit, and close events on the web dashboard
- An event has a lifecycle: DRAFT → ACTIVE → CLOSED
- BAs can only sell during an ACTIVE event — the app should enforce this
- After login, the Android app shows the BA their assigned event and lets them select it
- When the BA selects an event, the app downloads a full event snapshot to the device (products, stock, members, promotions) before allowing any selling activity
- If the BA has only one assigned event, it should be selected automatically

## Things Worth Paying Attention To
- The event snapshot endpoint is important — think about what data shape makes the most sense to download efficiently in one call
- Consider storing a version/hash of the snapshot in Room so future refreshes only re-sync when something has changed
- Think about what happens if the sync fails — what should the BA see?

## When Done

```bash
git checkout main && git pull origin main
git checkout -b feature/epic-02-event-management
git add .
git commit -m "feat: epic-02 event management and BA event sync"
git push origin feature/epic-02-event-management
```

Then open a Pull Request to `main` on https://github.com/krnskrll/event-mpos
