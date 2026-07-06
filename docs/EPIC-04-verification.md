# EPIC-04 — How to verify

EPIC-04 adds **member/customer sync** from an external system, members in the **event snapshot**,
fast **QR identification** (offline cache → online fallback → manual), and the data needed to link
an order to a customer. Builds on EPIC-01–03.

> New table since EPIC-03. Start clean if needed: `docker compose down -v && docker compose up --build`.

---

## 1. Start the backend
```bash
docker compose up --build
```
The seeder syncs 6 members from the (mock) member system (5 active, 1 inactive).

## 2. Easiest check — browser tester (no installs)
```bash
cd tools/manual-testers && python3 -m http.server 5500
```
Open **http://localhost:5500/epic4-tester.html**:
- **Admin**: member **sync status** + "Sync now" + the member **directory**.
- **BA**: type a **QR** (e.g. `MBR-00001`) → name + tier; try an unknown QR or inactive `MBR-00006`
  → "not found" → **manual name** fallback.

## 3. API checks (curl)
```bash
B=http://localhost:8080
ADMIN=$(curl -s -X POST $B/api/auth/login -H 'Content-Type: application/json' -d '{"email":"admin@eventmpos.local","password":"Admin123!"}' | python3 -c "import sys,json;print(json.load(sys.stdin)['accessToken'])")
BA=$(curl -s -X POST $B/api/auth/login -H 'Content-Type: application/json' -d '{"email":"ba@eventmpos.local","password":"Ba123456!"}' | python3 -c "import sys,json;print(json.load(sys.stdin)['accessToken'])")
```

### a) Member sync — scheduled + on-demand  ✅
```bash
curl -s $B/api/admin/members/sync/status -H "Authorization: Bearer $ADMIN"
curl -s -X POST $B/api/admin/members/sync -H "Authorization: Bearer $ADMIN"   # idempotent
curl -s $B/api/admin/members -H "Authorization: Bearer $ADMIN"               # full directory (incl inactive)
```
- Scheduled: `MEMBER_SCHEDULED=true` + `MEMBER_SYNC_CRON` (default daily 02:00).
- Swap source: `MEMBER_SOURCE=file` (reads `classpath:members/members.csv`) or `rest` + `MEMBER_REST_URL`.
- Filter large datasets: `MEMBER_REGION_FILTER=Central` only syncs that region.

### b) Members in the snapshot + active-only on device  ✅
```bash
EV01=$(curl -s $B/api/admin/events -H "Authorization: Bearer $ADMIN" | python3 -c "import sys,json;print([e['id'] for e in json.load(sys.stdin) if e['code']=='EV01'][0])")
curl -s "$B/api/me/events/$EV01/snapshot" -H "Authorization: Bearer $BA" | python3 -c "import sys,json;d=json.load(sys.stdin);print('members:',len(d['members']),'memberCount:',d['memberCount'])"  # 5 (inactive excluded)
curl -s "$B/api/me/members" -H "Authorization: Bearer $BA" | python3 -c "import sys,json;print('active members:',len(json.load(sys.stdin)))"  # 5
```

### c) QR lookup (online fallback)  ✅
```bash
curl -s "$B/api/me/members/by-qr/MBR-00001" -H "Authorization: Bearer $BA"                      # Aisha Rahman / PLATINUM
curl -s -o /dev/null -w "%{http_code}\n" "$B/api/me/members/by-qr/MBR-99999" -H "Authorization: Bearer $BA"  # 404 unknown
curl -s -o /dev/null -w "%{http_code}\n" "$B/api/me/members/by-qr/MBR-00006" -H "Authorization: Bearer $BA"  # 404 inactive
```

### d) Version stability  ✅
An idempotent member re-sync does **not** change the snapshot version (content-based fingerprint),
so devices don't re-download for nothing — but editing a member's name/tier does change it.

### Automated
```bash
cd backend && gradle test
```

---

## 4. Web dashboard (needs Node 18+)
`cd web && npm install && npm run dev` → admin → **Members** tab: sync status + "Sync members now",
searchable directory with tier badges.

## 5. Android app (needs Android Studio + emulator)
Sign in as BA → event auto-selects, the snapshot downloads (**members now cached to Room**) → tap
**Identify customer**:
- Type a QR (e.g. `MBR-00001`) → resolves **instantly from the Room cache** (works offline).
- Unknown QR with network → **online fallback** lookup.
- Neither → type the **customer name manually** to proceed.
- The chosen customer is what an order links to in EPIC-07 (no order without one).

---

## Acceptance checklist (from the epic)
- [x] Backend sync job pulls member data on a configurable schedule
- [x] Member data included in the event snapshot (and available as a separate `/api/me/members` step)
- [x] BA scans a QR → instant name + tier from local cache
- [x] Online fallback lookup when not in cache and network is available
- [x] Manual name entry when neither works
- [x] Customer can be linked to an order (selection produced; enforced at submit in EPIC-07)
- [x] QR column indexed for fast lookup (PostgreSQL unique index + Room primary key)
- [x] Minimal member model — display + linking only, no financial data
- [x] Large datasets handled — active-only on device + optional region filter at sync
