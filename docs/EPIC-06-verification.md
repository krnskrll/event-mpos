# EPIC-06 — How to verify

EPIC-06 adds **event-scoped promotions** that auto-apply (percentage / fixed / bundle), a **shared
rule engine** (client preview = server-authoritative), promotions in the **snapshot**, and an
authoritative **quote** endpoint. Builds on EPIC-01–05.

> New table since EPIC-05. Start clean: `docker compose down -v && docker compose up --build`.

---

## Shared-logic guarantee
The discount math lives in one deterministic place per side:
- Server: `backend/.../promotion/PromotionEngine.kt` (authoritative)
- Client: `android/.../domain/promotion/PromotionEngine.kt` (mirror — kept byte-for-byte equivalent)

A unit test pins the rules (run locally):
```bash
cd backend
JAVA_HOME="/Applications/Android Studio.app/Contents/jbr/Contents/Home" ./gradlew test --tests '*PromotionEngineTest'
```
Covers percentage, fixed + min-subtotal, specific scope, bundle (buy-N), expiry/not-yet-started,
best-promotion-wins (one per order), and the discount-never-exceeds-subtotal cap.

---

## Browser tester (no installs)
```bash
cd tools/manual-testers && python3 -m http.server 5500
```
Open **http://localhost:5500/epic6-tester.html**:
- **Admin**: pick an event → create/disable/delete promotions (percentage / fixed / bundle).
- **BA**: build a quick cart → **Quote** shows the server-applied best promotion + discount + total.

## API checks (curl)
```bash
B=http://localhost:8080
ADMIN=$(curl -s -X POST $B/api/auth/login -H 'Content-Type: application/json' -d '{"email":"admin@eventmpos.local","password":"Admin123!"}' | python3 -c "import sys,json;print(json.load(sys.stdin)['accessToken'])")
BA=$(curl -s -X POST $B/api/auth/login -H 'Content-Type: application/json' -d '{"email":"ba@eventmpos.local","password":"Ba123456!"}' | python3 -c "import sys,json;print(json.load(sys.stdin)['accessToken'])")
EV01=$(curl -s $B/api/admin/events -H "Authorization: Bearer $ADMIN" | python3 -c "import sys,json;print([e['id'] for e in json.load(sys.stdin) if e['code']=='EV01'][0])")
```
- **List / create / delete**: `GET|POST /api/admin/events/$EV01/promotions`, `PUT|DELETE /api/admin/promotions/{id}`.
  Validation: a PERCENTAGE without `percentOff`, or BUNDLE without `minQty`, → **400**.
- **Snapshot**: `GET /api/me/events/$EV01/snapshot` → `promotions[]` populated.
- **Authoritative quote** (auto-applies the single best promotion):
  ```bash
  SERUM=$(curl -s $B/api/admin/products -H "Authorization: Bearer $ADMIN" | python3 -c "import sys,json;print([p['id'] for p in json.load(sys.stdin) if p['sku']=='SKU-1001'][0])")
  curl -s -X POST $B/api/me/events/$EV01/promotions/quote -H "Authorization: Bearer $BA" -H 'Content-Type: application/json' \
    -d "{\"lines\":[{\"productId\":\"$SERUM\",\"qty\":2}]}"
  # subtotal 258.00, discount 25.80 (10% beats RM20), total 232.20
  ```
  3× lipstick (SKU-1002) → the **RM30 bundle** wins over 10% (23.70). One promotion per order.

---

## Web (needs Node 18+)
Admin → **Promotions** tab: pick an event, add a promotion (type-aware form: percentage / fixed /
bundle, scope ALL or specific products, min qty / min subtotal), enable-disable, delete.

## Android (emulator)
BA → **View promotions**: the event's promotions (downloaded in the snapshot, cached in Room) are
listed with human descriptions. The mirrored `PromotionEngine` powers the offline cart preview that
EPIC-07 will show; the server `quote` stays authoritative at submit.

---

## Acceptance checklist (from the epic)
- [x] Admins create & manage promotions for a specific event (web)
- [x] Types: percentage, fixed amount, product bundle
- [x] Promotions included in the event snapshot
- [x] Auto-apply (no BA action) — engine selects automatically; client preview + server quote
- [x] Server re-validates & applies the authoritative discount (quote endpoint; reused at submit in EPIC-07)
- [x] One promotion per order (engine picks the best)
- [x] Client preview == server result (shared/mirrored engine, pinned by a unit test)
- [x] "Applicable" = scope (all / specific products), min quantity, min subtotal
- [x] Optional start/end times with expiry handling
