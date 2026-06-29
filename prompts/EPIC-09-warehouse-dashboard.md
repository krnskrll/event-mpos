# Epic 09 — Warehouse Order Execution Dashboard

## Context
Read `.cursorrules` and review what was built in Epics 01–08 before starting.
This epic is web-only. The warehouse team uses this dashboard to see incoming orders and pack them. Orders arrive in real-time — the team should not need to refresh the page.

## Goal
Build a real-time web dashboard where the warehouse team can see, manage, and execute orders as they come in from BAs on the floor.

## What Needs to Work When This Epic Is Done
- The warehouse team sees all orders for the active event in real-time — new orders appear automatically within a couple of seconds
- Each order shows: order number, customer name, item list with quantities, BA name, time since creation, and current status
- The warehouse team can advance an order through its statuses: NEW → PACKING → READY → COLLECTED
- The dashboard can be filtered by status so the team can focus on what needs attention
- A sound or visual alert triggers when a new order arrives — the team may not be watching the screen
- If the real-time connection drops, the dashboard falls back gracefully and still shows orders (consider a polling fallback)
- Remaining event stock per product is visible somewhere on the dashboard

## Things Worth Paying Attention To
- Explore WebSocket with STOMP for the real-time connection — both the Spring backend and the React client side
- Think about what the order card should look like for efficient warehouse use — large text, clear status, easy to act on
- The fallback polling strategy matters — what is a reasonable interval that keeps the team informed without hammering the server?
- Consider the warehouse team's environment: they may be moving around, the screen may be at a distance

## When Done

```bash
git checkout main && git pull origin main
git checkout -b feature/epic-09-warehouse-dashboard
git add .
git commit -m "feat: epic-09 real-time warehouse order execution dashboard"
git push origin feature/epic-09-warehouse-dashboard
```

Then open a Pull Request to `main` on https://github.com/krnskrll/event-mpos
