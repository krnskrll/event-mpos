# EPIC-02 — How to verify

EPIC-02 adds the **event lifecycle** (DRAFT → ACTIVE → CLOSED), event **edit**, and the **event
snapshot sync** the Android app downloads when a BA selects their event. Builds on EPIC-01.

> Schema changed since EPIC-01 (`Event.active` → `Event.status`). If you have an old dev DB,
> start clean: `docker compose down -v && docker compose up --build`.

---

## 1. Start the backend
```bash
docker compose up --build      # from repo root; API on http://localhost:8080
```

## 2. API checks (curl)

Get tokens:
```bash
B=http://localhost:8080
ADMIN=$(curl -s -X POST $B/api/auth/login -H 'Content-Type: application/json' -d '{"email":"admin@eventmpos.local","password":"Admin123!"}' | python3 -c "import sys,json;print(json.load(sys.stdin)['accessToken'])")
BA=$(curl -s -X POST $B/api/auth/login -H 'Content-Type: application/json' -d '{"email":"ba@eventmpos.local","password":"Ba123456!"}' | python3 -c "import sys,json;print(json.load(sys.stdin)['accessToken'])")
```

### a) Lifecycle: create (DRAFT) → activate → close  ✅
```bash
# new events start DRAFT
curl -s -X POST $B/api/admin/events -H "Authorization: Bearer $ADMIN" -H 'Content-Type: application/json' \
  -d '{"code":"EV02","name":"Year-End Glow Fair"}' | python3 -m json.tool   # status: DRAFT
# grab its id, then drive the lifecycle:
NID=... # id from above
curl -s -X PATCH $B/api/admin/events/$NID/status -H "Authorization: Bearer $ADMIN" -H 'Content-Type: application/json' -d '{"status":"ACTIVE"}'
curl -s -X PATCH $B/api/admin/events/$NID/status -H "Authorization: Bearer $ADMIN" -H 'Content-Type: application/json' -d '{"status":"CLOSED"}'
# illegal transitions are rejected with 409, e.g. ACTIVE->DRAFT or reopening a CLOSED event
```

### b) Edit an event  ✅
```bash
curl -s -X PUT $B/api/admin/events/$NID -H "Authorization: Bearer $ADMIN" -H 'Content-Type: application/json' \
  -d '{"name":"Renamed","startDate":"2026-12-01","endDate":"2026-12-20"}'
```

### c) Event snapshot + version gate  ✅
```bash
EV01=$(curl -s $B/api/admin/events -H "Authorization: Bearer $ADMIN" | python3 -c "import sys,json;print([e['id'] for e in json.load(sys.stdin) if e['code']=='EV01'][0])")
# full snapshot (EV01 is seeded ACTIVE):
curl -s "$B/api/me/events/$EV01/snapshot" -H "Authorization: Bearer $BA" | python3 -m json.tool
# re-sync with the version you got back -> {"changed": false} (no re-download)
curl -s "$B/api/me/events/$EV01/snapshot?version=THEVERSION" -H "Authorization: Bearer $BA"
# edit the event, fetch the snapshot again -> the version changes (device knows to re-sync)
```

### d) ACTIVE enforcement  ✅
```bash
# a BA assigned to a DRAFT (or CLOSED) event cannot pull a snapshot -> 409
curl -s -o /dev/null -w "draft snapshot: %{http_code}\n" "$B/api/me/events/$DRAFT_EVENT_ID/snapshot" -H "Authorization: Bearer $BA"
```

### Automated
```bash
cd backend && gradle test   # AuthFlowTest (epic-01) + EventLifecycleTest (epic-02 transitions)
```

---

## 3. Web dashboard (needs Node 18+)
```bash
cd web && npm install && npm run dev   # http://localhost:5173
```
Sign in as admin → in **Events** you can now:
- See each event's **status** badge (DRAFT/ACTIVE/CLOSED)
- **Activate** a DRAFT event, **Close** an ACTIVE one (illegal transitions are blocked server-side)
- **Edit** name/dates inline (closed events are read-only)

## 4. Android app (needs Android Studio + emulator)
Sign in as a BA:
- If you're assigned to **one** event, it's **auto-selected** and the snapshot downloads automatically.
- If assigned to **several**, you get an event **picker**.
- While syncing you see "Downloading event data…"; on success "✓ Synced (vXXXX). Ready to sell."
- Select an event that isn't ACTIVE → you're **blocked** with a clear message (lifecycle rule).
- Kill the venue Wi-Fi mid-sync → you get a **failure message + Retry** (the cached snapshot, if any, stays).
- The snapshot **version** is cached in Room; re-selecting an unchanged event shows "Up to date" without re-downloading.

---

## Acceptance checklist (from the epic)
- [x] Admin can create, edit, and close events on the web dashboard
- [x] Event lifecycle DRAFT → ACTIVE → CLOSED (illegal transitions rejected)
- [x] BAs can only sell during an ACTIVE event (enforced on server + Android)
- [x] After login the Android app shows the assigned event and lets the BA select it
- [x] Selecting an event downloads a full snapshot before any selling
- [x] Single assigned event is auto-selected
- [x] Snapshot version stored in Room; re-sync only when the version changed
- [x] Sync failure shows a clear, retryable message
