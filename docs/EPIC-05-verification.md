# EPIC-05 — How to verify

EPIC-05 adds **warehouse stock from ERP**, **per-event allocation** (locked once ACTIVE),
**remaining-stock views** (warehouse + Android), and an **atomic, oversell-proof decrement** for
order submission. Builds on EPIC-01–04.

> New tables + a new product column. Start clean: `docker compose down -v && docker compose up --build`.

---

## The headline guarantee: no overselling (automated)
The core requirement — stock can't be oversold even under simultaneous submissions — is proven by a
real concurrency test (300 threads racing on an allocation of 100):
```bash
cd backend
JAVA_HOME="/Applications/Android Studio.app/Contents/jbr/Contents/Home" ./gradlew test --tests '*StockConcurrencyTest'
```
It asserts exactly 100 decrements succeed, 200 are cleanly rejected, and `soldQty == 100` (never more).
Mechanism: `decrement()` runs in one transaction holding a **PESSIMISTIC_WRITE** row lock
(`SELECT … FOR UPDATE`) on the allocation row, plus an `@Version` optimistic guard. EPIC-07's order
submission calls this inside the same transaction that creates the order.

---

## Browser tester (no installs)
```bash
cd tools/manual-testers && python3 -m http.server 5500
```
Open **http://localhost:5500/epic5-tester.html**:
- **Admin**: pick an event → allocate stock per product (locked unless DRAFT); create a DRAFT event to edit.
- **Warehouse**: live remaining-stock table (auto-refreshes) with LOW / SOLD-OUT flags.

## API checks (curl)
```bash
B=http://localhost:8080
ADMIN=$(curl -s -X POST $B/api/auth/login -H 'Content-Type: application/json' -d '{"email":"admin@eventmpos.local","password":"Admin123!"}' | python3 -c "import sys,json;print(json.load(sys.stdin)['accessToken'])")
WH=$(curl -s -X POST $B/api/auth/login -H 'Content-Type: application/json' -d '{"email":"wh@eventmpos.local","password":"Wh123456!"}' | python3 -c "import sys,json;print(json.load(sys.stdin)['accessToken'])")
EV01=$(curl -s $B/api/admin/events -H "Authorization: Bearer $ADMIN" | python3 -c "import sys,json;print([e['id'] for e in json.load(sys.stdin) if e['code']=='EV01'][0])")
```
- **Warehouse stock from ERP**: `GET /api/admin/products` → each has `warehouseStock`.
- **Allocations**: `GET /api/admin/events/$EV01/stock` → allocated/sold/remaining per product.
- **Lock on ACTIVE**: `PUT /api/admin/events/$EV01/stock/<productId> {"allocatedQty":99}` → **409** (EV01 is ACTIVE).
- **DRAFT works**: create an event, `PUT .../stock/<pid> {"allocatedQty":30}` → 200; over-warehouse → 400; after activating → 409.
- **Warehouse view**: `GET /api/warehouse/events/$EV01/stock` (as WH) → remaining + `lowStock`; a BA token → **403**.
- **BA live levels**: `GET /api/me/events/$EV01/stock` (as BA) → `[{productId, remaining, lowStock}]`.
- **Snapshot**: each snapshot product now carries `remainingStock`.

---

## Web (needs Node 18+)
- **Warehouse** page: pick an event → remaining-stock table that **polls every 4s**, LOW (amber) / SOLD OUT (red).
- **Products** page → **Stock allocation per event**: pick an event, edit allocated quantities
  (inputs disabled + 🔒 when the event isn't DRAFT).

## Android (emulator)
BA → **Browse products**: each card shows **N in stock**, **Low stock · N left** (amber/red), or
**Sold out**. Levels **refresh every 10s** from `/api/me/events/{id}/stock` and are cached in Room
(so the last-known level shows offline). Seeded demo: Nourishing Hair Oil shows "Low stock · 8 left".

---

## Acceptance checklist (from the epic)
- [x] Backend sync pulls stock quantities from ERP (shares EPIC-03 sync infra)
- [x] Admins allocate quantities of each product to an event
- [x] Allocations locked once the event is ACTIVE
- [x] Warehouse sees remaining stock per product, near real-time (4s polling; WebSocket in EPIC-09)
- [x] Android shows BAs remaining event stock per product, refreshing periodically (10s)
- [x] Atomic server-side validate + decrement; overselling impossible under concurrency (PESSIMISTIC_WRITE + test)
- [x] Decrement + order creation will share one transaction (decrement() is transactional; wired in EPIC-07)
- [x] Low-stock warning on the device; clear sold-out feedback
