# Epic 05 — Warehouse Allocation & Stock

## Context
Read `.cursorrules` and review what was built in Epics 01–04 before starting.

Stock comes from ERP, but once allocated to an event it is owned by this platform — the platform tracks and decrements it, not ERP. This separation is intentional.

## Goal
Enable Admins to allocate warehouse stock to specific events, show real-time remaining stock on both the web dashboard and the Android app, and prevent overselling through atomic server-side validation.

## What Needs to Work When This Epic Is Done
- A backend sync job pulls current stock quantities from ERP (can share the sync infrastructure from Epic 03)
- Admins can allocate specific quantities of each product to an event on the web dashboard
- Once an event is ACTIVE, allocations are locked — no further changes
- The warehouse team can see remaining stock per product for an active event on the web dashboard, updating in near real-time
- The Android app shows BAs the remaining event stock per product, refreshing periodically
- When a BA submits an order, the server atomically validates and decrements the allocated stock — overselling must be impossible even with multiple BAs submitting simultaneously

## Things Worth Paying Attention To
- The stock decrement and order creation must happen in a single database transaction
- Explore the right PostgreSQL locking strategy for concurrent order submissions — this is a genuine concurrency problem
- The Android app should show a low-stock warning when a product is running low
- Think about what happens when stock hits zero — clear feedback to the BA

## When Done

```bash
git checkout main && git pull origin main
git checkout -b feature/epic-05-allocation-stock
git add .
git commit -m "feat: epic-05 warehouse allocation and stock management"
git push origin feature/epic-05-allocation-stock
```

Then open a Pull Request to `main` on https://github.com/krnskrll/event-mpos
