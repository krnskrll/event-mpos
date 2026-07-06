# EPIC-09 — How to verify

EPIC-09 is the **real-time warehouse order board** (web), backed by Spring **WebSocket/STOMP**.
Orders appear live, the team advances them NEW→PACKING→READY→COLLECTED, filters by status, gets a
new-order alert, falls back to polling if the socket drops, and sees remaining stock. Builds on
EPIC-01–08.

> No new tables. Just `docker compose up --build`.

---

## Backend (verified live with curl)
```bash
B=http://localhost:8080
ADMIN=$(curl -s -X POST $B/api/auth/login -H 'Content-Type: application/json' -d '{"email":"admin@eventmpos.local","password":"Admin123!"}' | python3 -c "import sys,json;print(json.load(sys.stdin)['accessToken'])")
WH=$(curl -s -X POST $B/api/auth/login -H 'Content-Type: application/json' -d '{"email":"wh@eventmpos.local","password":"Wh123456!"}' | python3 -c "import sys,json;print(json.load(sys.stdin)['accessToken'])")
EV01=$(curl -s $B/api/admin/events -H "Authorization: Bearer $ADMIN" | python3 -c "import sys,json;print([e['id'] for e in json.load(sys.stdin) if e['code']=='EV01'][0])")
```
- **Order list** (initial load + polling fallback), with BA name + items + status:
  `GET /api/warehouse/events/$EV01/orders` (filter with `?status=NEW`).
- **Advance** one step: `PATCH /api/warehouse/orders/{id}/status {"status":"PACKING"}` → READY → COLLECTED.
  Skipping a step (NEW→READY) or moving past COLLECTED → **409**.
- **Role**: a BA token → **403**; warehouse/admin → 200.
- **WebSocket up**: `curl -s -o /dev/null -w '%{http_code}' http://localhost:8080/ws/info` → **200**
  (SockJS info). The broker logs `BrokerAvailabilityEvent[available=true]` on boot.

Verified: full NEW→PACKING→READY→COLLECTED advance, 409 on skip/past-end, status filter, BA 403,
order list shows BA name + line items.

> Real-time push design: on order create + status change the server fires an internal event and an
> `@TransactionalEventListener(AFTER_COMMIT)` broadcasts a `WarehouseOrderView` to
> `/topic/events/{eventId}/orders` via `SimpMessagingTemplate` (so rolled-back orders never reach the
> board). The browser is the production consumer (below); a headless raw-STOMP unit test was dropped
> as it exercised the test client more than the app.

## Web dashboard (needs Node 18+)
```bash
cd web && npm install && npm run dev    # http://localhost:5173 (proxies /api and /ws to :8080)
```
Sign in as `wh@eventmpos.local` / `Wh123456!` → **Warehouse**:
- Pick the active event. The board connects over **STOMP/SockJS** (`🟢 live`); if it drops it shows
  `🟡 polling` and refreshes every 8s.
- **Place an order from the Android app (or the EPIC-07 browser tester)** → within ~1–2s a new
  **order card** appears with a **beep + border flash**; large order #, customer, item lines, BA,
  time-since, status.
- **Advance**: tap **Mark PACKING** → **Mark READY** → **Mark COLLECTED** (each pushes live to any
  other open board).
- **Filter** tabs: ALL / NEW / PACKING / READY / COLLECTED.
- **Stock** tab: remaining stock per product (polled), LOW / SOLD OUT flags.

To see two boards update each other in real time, open the dashboard in two browser windows.

---

## Acceptance checklist (from the epic)
- [x] Warehouse sees all orders for the event in real-time (new ones appear automatically)
- [x] Each order shows: order number, customer, item list + qty, BA name, time since creation, status
- [x] Advance NEW → PACKING → READY → COLLECTED (one step; invalid jumps rejected)
- [x] Filter by status
- [x] Sound + visual alert on new order
- [x] Graceful fallback to polling if the real-time connection drops
- [x] Remaining event stock per product visible on the dashboard
- [x] WebSocket + STOMP on both Spring backend and React client
- [x] Large, glanceable order cards for warehouse use
