# EPIC-03 — How to verify

EPIC-03 adds the **product catalogue + ERP sync**, **per-event product availability**, **admin
product add/edit**, products inside the event **snapshot**, and **barcode scan → product** lookup.
Builds on EPIC-01/02.

> New tables since EPIC-02. Start clean if needed: `docker compose down -v && docker compose up --build`.

---

## 1. Start the backend
```bash
docker compose up --build      # API on http://localhost:8080
```
On first boot the seeder pulls the catalogue from the (mock) ERP and makes all 5 products available
for EV01.

## 2. Easiest check — the browser tester (no installs)
```bash
cd tools/manual-testers && python3 -m http.server 5500
```
Open **http://localhost:5500/epic3-tester.html**:
- **Admin**: "Sync from ERP" + last-sync status, **add a product**, **edit** a product, and toggle
  which products are **available per event**.
- **BA**: download the event snapshot (now lists products) and **scan a barcode** → resolves to a product.

## 3. API checks (curl)
```bash
B=http://localhost:8080
ADMIN=$(curl -s -X POST $B/api/auth/login -H 'Content-Type: application/json' -d '{"email":"admin@eventmpos.local","password":"Admin123!"}' | python3 -c "import sys,json;print(json.load(sys.stdin)['accessToken'])")
BA=$(curl -s -X POST $B/api/auth/login -H 'Content-Type: application/json' -d '{"email":"ba@eventmpos.local","password":"Ba123456!"}' | python3 -c "import sys,json;print(json.load(sys.stdin)['accessToken'])")
```

### a) ERP sync — scheduled + on-demand  ✅
```bash
curl -s $B/api/admin/erp/status -H "Authorization: Bearer $ADMIN"   # last sync time/result/count/source
curl -s -X POST $B/api/admin/erp/sync -H "Authorization: Bearer $ADMIN"   # trigger now (idempotent)
curl -s $B/api/admin/products -H "Authorization: Bearer $ADMIN"     # synced catalogue (note multiple barcodes)
```
- Scheduled pulls: set `ERP_SCHEDULED=true` and `ERP_SYNC_CRON` (default hourly).
- Swap the source without code changes: `ERP_SOURCE=file` (reads `classpath:erp/products.csv`) or
  `ERP_SOURCE=rest` + `ERP_REST_URL=...`.

### b) Admin add + edit product  ✅
```bash
# add (manual product)
curl -s -X POST $B/api/admin/products -H "Authorization: Bearer $ADMIN" -H 'Content-Type: application/json' \
  -d '{"sku":"SKU-2001","name":"Gift Set","price":199.00,"brand":"LumièreCo","barcodes":[{"barcode":"7770002001","label":"set"}]}'
# edit
curl -s -X PUT $B/api/admin/products/<ID> -H "Authorization: Bearer $ADMIN" -H 'Content-Type: application/json' \
  -d '{"name":"Gift Set (Holiday)","price":179.00}'
# duplicate SKU -> 409
```

### c) Per-event availability  ✅
```bash
EV01=$(curl -s $B/api/admin/events -H "Authorization: Bearer $ADMIN" | python3 -c "import sys,json;print([e['id'] for e in json.load(sys.stdin) if e['code']=='EV01'][0])")
curl -s $B/api/admin/events/$EV01/products -H "Authorization: Bearer $ADMIN"          # all products + available flag
curl -s -X PUT $B/api/admin/events/$EV01/products/<PRODUCT_ID> -H "Authorization: Bearer $ADMIN" \
  -H 'Content-Type: application/json' -d '{"available":true}'
```

### d) Products in the snapshot + version  ✅
```bash
curl -s "$B/api/me/events/$EV01/snapshot" -H "Authorization: Bearer $BA" | python3 -m json.tool
# -> products[] now populated; changing availability/price changes the snapshot version
```

### e) Barcode scan lookup (BA)  ✅
```bash
curl -s "$B/api/me/events/$EV01/products/by-barcode/8990001001" -H "Authorization: Bearer $BA"   # -> product
curl -s -o /dev/null -w "%{http_code}\n" "$B/api/me/events/$EV01/products/by-barcode/0000" -H "Authorization: Bearer $BA"  # 404
```

### Automated
```bash
cd backend && gradle test
```

---

## 4. Web dashboard (needs Node 18+)
`cd web && npm install && npm run dev` → sign in as admin → **Products** tab:
- ERP sync status + "Sync from ERP now"
- Add a product, edit a product inline
- "Products available per event": pick an event, tick products on/off

## 5. Android app (needs Android Studio + emulator)
Sign in as BA → event auto-selects and the snapshot downloads (**products now included**):
- Tap **Browse products** → a grid of products with images (Coil, cached for offline).
- Search by name/SKU/barcode.
- Type a barcode + **Scan typed barcode** → resolves to a product (a Sunmi hardware scan uses the
  same `ScannerPort`; manual entry is the emulator fallback).
- Browsing reads from **Room**, so it works with the network off.

---

## Acceptance checklist (from the epic)
- [x] Backend sync job pulls product data from ERP on a configurable schedule
- [x] Admins view synced products + last sync time
- [x] Admins select which products are available for a specific event
- [x] Admins can add & edit products (manual catalogue management)
- [x] Android shows the event product catalogue in a browsable grid
- [x] BA can scan a barcode to resolve a product (ScannerPort abstraction; Sunmi via AIDL behind it)
- [x] Product images load from cache, fetched during event sync (Coil disk cache)
- [x] ERP layer swappable between API-pull and file-import without rewriting (adapter interface)
- [x] A product can have multiple barcodes (data model + scan)
- [x] Browsing is offline-first (Room)
