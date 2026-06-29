# Epic 07 — Cart & Order Creation

## Context
Read `.cursorrules` and review what was built in Epics 01–06 before starting.
This is the core selling workflow. The cart is built locally on the device; the order is only created when submitted to the server and confirmed.

## Goal
Enable BAs to build a cart, submit an order, and receive server confirmation — with proper offline handling and duplicate protection.

## What Needs to Work When This Epic Is Done
- A BA can add, remove, and adjust quantities of products in the cart
- Applicable promotions appear automatically on the cart (from Epic 06)
- A BA can save a cart as a draft to resume later — this stays local, no server order is created
- When the BA confirms, the order is submitted to the backend
- The server validates stock, applies the final promotion, decrements allocation, and creates the order atomically
- The server generates an `order_number` (daily sequence, e.g. #042) and a full `reference_id` (e.g. ORD-20260629-EV01-042)
- If submission fails due to network issues, WorkManager queues a retry and the BA sees a "Pending" status
- Duplicate submissions (e.g. double taps) are handled safely with an idempotency key
- A BA can view their order history for the current event session

## Things Worth Paying Attention To
- The cart should be entirely local state — no server calls while the BA is browsing or adjusting items
- The daily sequence order number must be race-condition safe — explore the right database approach
- Receipt printing (Epic 08) depends on this epic's order confirmation response — design the response payload with that in mind
- Think about what "Pending" means in the UI and how to recover from a permanently failed submission

## When Done

```bash
git checkout main && git pull origin main
git checkout -b feature/epic-07-cart-order
git add .
git commit -m "feat: epic-07 cart management and order creation"
git push origin feature/epic-07-cart-order
```

Then open a Pull Request to `main` on https://github.com/krnskrll/event-mpos
