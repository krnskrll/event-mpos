# EPIC-07 — How to verify

EPIC-07 is the core selling workflow: a **local cart**, **atomic order creation** (stock + promo +
customer + race-safe daily number + idempotency), **offline retry** via WorkManager, and **order
history**. Builds on EPIC-01–06.

> New tables since EPIC-06. Start clean: `docker compose down -v && docker compose up --build`.

---

## Atomic correctness (automated)
The order transaction is pinned by a test suite (run locally):
```bash
cd backend
JAVA_HOME="/Applications/Android Studio.app/Contents/jbr/Contents/Home" ./gradlew test --tests '*OrderSubmissionTest'
```
Covers: order number/reference + stock decrement, **idempotent submit** (double-tap returns the same
order, decrements once), **daily number increments**, **oversell rejected + rolled back**, and the
**best promotion applied** at submit. (Stock concurrency itself is proven by `StockConcurrencyTest`.)

Key design points:
- Submit runs in ONE `@Transactional`: idempotency check → validate → `StockService.decrement`
  (PESSIMISTIC_WRITE locks) → apply best promotion → allocate daily number → persist. Any failure
  rolls the whole thing back.
- **Daily order number** uses a per-(event, day) row locked with `SELECT … FOR UPDATE`
  (`OrderDailySequence`) → race-safe, strictly increasing #001, #002, …
- **Idempotency**: a unique `idempotency_key`; a retried/concurrent submit returns the original order.
- Three IDs per .cursorrules: internal UUID, `referenceId` `ORD-YYYYMMDD-EVxx-NNN`, daily `#042`.
- The confirmation payload carries everything EPIC-08's receipt needs (brand/event/number/customer/
  lines/discount/total/cashier note).

---

## Browser tester (no installs)
```bash
cd tools/manual-testers && python3 -m http.server 5500
```
Open **http://localhost:5500/epic7-tester.html** (BA): build a cart, submit → see the order number +
reference + total; **submit again with the same key** → same order returned (idempotent); view history.

## API checks (curl)
```bash
B=http://localhost:8080
BA=$(curl -s -X POST $B/api/auth/login -H 'Content-Type: application/json' -d '{"email":"ba@eventmpos.local","password":"Ba123456!"}' | python3 -c "import sys,json;print(json.load(sys.stdin)['accessToken'])")
ADMIN=$(curl -s -X POST $B/api/auth/login -H 'Content-Type: application/json' -d '{"email":"admin@eventmpos.local","password":"Admin123!"}' | python3 -c "import sys,json;print(json.load(sys.stdin)['accessToken'])")
EV01=$(curl -s $B/api/admin/events -H "Authorization: Bearer $ADMIN" | python3 -c "import sys,json;print([e['id'] for e in json.load(sys.stdin) if e['code']=='EV01'][0])")
SERUM=$(curl -s $B/api/admin/products -H "Authorization: Bearer $ADMIN" | python3 -c "import sys,json;print([p['id'] for p in json.load(sys.stdin) if p['sku']=='SKU-1001'][0])")

# submit
curl -s -X POST $B/api/me/events/$EV01/orders -H "Authorization: Bearer $BA" -H 'Content-Type: application/json' \
  -d "{\"idempotencyKey\":\"demo-1\",\"customer\":{\"name\":\"Aisha\",\"qrCode\":\"MBR-00001\",\"tier\":\"PLATINUM\"},\"lines\":[{\"productId\":\"$SERUM\",\"qty\":2}]}" | python3 -m json.tool
# repeat with the SAME idempotencyKey -> "duplicate": true, same id, stock not decremented again
# history
curl -s $B/api/me/events/$EV01/orders -H "Authorization: Bearer $BA" | python3 -m json.tool
```

---

## Android (emulator) — the full flow
Sign in as BA → **Browse products** → tap **Add** on items (scanning a barcode also adds) →
**View cart (N)**:
- adjust quantities (− / +); the **promotion auto-applies as a preview** (e.g. "Discount (10% off
  everything)"); the cart is **entirely local** — no server calls while editing.
- Confirm is disabled until a customer is identified → tap **Identify**, scan/type `MBR-00001`
  (resolved offline from Room) → back to cart → **Confirm order**.
- The order is queued and submitted via **WorkManager** (idempotency key). **Order history** shows it
  **Pending → Confirmed** with `#003` / `ORD-20260630-EV01-003` / total once the server confirms.
- Offline: kill Wi-Fi before Confirm → it stays **Pending** and WorkManager retries when network
  returns; a server rejection (e.g. sold out) shows **Failed**.

> Draft-to-resume and receipt printing: the cart persists in-session; the receipt print arrives in
> EPIC-08 (this epic's confirmation payload is its input).

---

## Acceptance checklist (from the epic)
- [x] BA can add / remove / adjust cart quantities (local state)
- [x] Applicable promotions appear automatically on the cart (preview from EPIC-06 engine)
- [x] Confirm submits to the backend; order created only on the server
- [x] Server validates stock, applies the final promotion, decrements allocation, creates the order **atomically**
- [x] Server generates `order_number` (daily #) + full `reference_id`
- [x] Network failure → WorkManager retry, BA sees **Pending**
- [x] Duplicate submissions handled with an **idempotency key**
- [x] BA can view order history for the event session
- [x] Daily sequence is race-condition safe (`SELECT … FOR UPDATE`)
- [x] Confirmation payload designed for the EPIC-08 receipt
