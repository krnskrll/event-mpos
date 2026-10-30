# EPIC-12 — How to verify

EPIC-12 closes the Phase 1 gaps: Go becomes the only backend with CI, back-office fixes, and the
selling flow for real venues — offline confirmation, queued orders after an event closes, order
cancel and cart drafts. No new product scope (no payments).

| Milestone | PR | Status |
|---|---|---|
| A1 Go is the only backend | #3 | ✅ |
| A2 CI + baseline tests | #4 | ✅ |
| D Back-office fixes | #6 | ✅ (see [EPIC-11](EPIC-11-verification.md)) |
| C Selling flow (C1–C4) | #7 | ✅ |
| B Real Sunmi printer + scanner | — | ⏸ waiting for a physical Sunmi V2 Pro |
| E Docs and release | this PR | ✅ |

> Schema: Postgres migrations 000002 (audit index) and 000003 (number blocks, order flags,
> `events.closed_at`) apply on startup. Android Room goes v10 → v11 without wiping data.
> The EPIC-01–09 docs predate the Go migration; their API calls still apply.

---

## 1. Backend + CI
```bash
docker compose down -v && docker compose up --build      # Go API on :8080
cd backend-go && TEST_DB_URL=postgres://… go test ./...   # or rely on CI
```
CI (`.github/workflows/ci.yml`) runs on every PR: Go (vet, migrations, Postgres integration tests,
build), Android (lint, unit tests, debug build), web (type check, build, vitest).
- **Promotion parity**: both engines run `testdata/promotion-cases.json`. Changing only one engine
  fails CI (checked on throwaway PR #5).
- **Concurrency**: 300 goroutines against 100 units and 40 buyers for the last 8 → no oversell.

## 2. Offline confirmation (C1, decision D1)
On the emulator, start a session on EV01 (the app fetches a block of 50 order numbers:
`POST /api/me/events/{id}/number-blocks`).
1. `adb shell cmd connectivity airplane-mode enable`.
2. Complete 3 sales (scan → cart → member → **Confirm order**). Each gets a **final number** (e.g.
   #142, #143, #144) and the receipt prints immediately (`adb logcat -s RECEIPT`). Orders show **SYNCING**.
3. `adb shell cmd connectivity airplane-mode disable` → all 3 upload and appear on the warehouse
   board with **the same numbers**.
- Server numbers and device blocks come off one locked sequence, so they never collide; a number
  outside the device's block → **400**, a reused number → **409**.
- An offline sale past the allocation is accepted and flagged **STOCK CONFLICT** on the board;
  remaining shows the shortfall (e.g. `-1 · OVERSOLD`).
- With no numbers left while offline, checkout is blocked with a clear message.

## 3. Event close with queued orders (C2, decision D2)
- Closing an event sets `closed_at`. Orders created before it are still accepted for 24h and marked
  **LATE-SYNCED** (board badge, report count, audit text). Created after the close, or uploaded more
  than 24h later → **409** with an explanation.
- On the device, a banner shows while queued orders are still syncing after the close; uploads keep
  retrying; the event's drafts are dropped.

## 4. Order cancel (C4)
- Warehouse board: **Cancel order** on a NEW/PACKING card (optional reason) → card shows CANCELLED,
  stock tab returns the units. Filter **CANCELLED**.
- Android: Orders → **Cancel** on a synced order → order and stored receipt show CANCELLED. Once the
  warehouse is packing, the BA gets a clear "can't cancel" message (server **409**).
- Check: for every product, sold = units in non-cancelled orders.

## 5. Cart drafts (C3)
Cart → **Save draft** → cart empties, **Drafts (1)**. Add another item, then
`adb shell am force-stop com.eventmpos.mpos`, relaunch, start the session: the cart item and the
draft are both back. **Resume** over a non-empty cart asks "Replace current cart?".

---

## Acceptance checklist (from the epic)
- [x] `docker compose up` runs Go; web board and Android app work end to end; docs say Go
- [x] CI green; changing only one promotion engine fails CI
- [x] Final report for CLOSED events only; paginated audit; RECEIPT_PRINTED audit; stock summary cache — each with a Go test
- [x] Airplane mode: 3 orders confirmed and 3 receipts printed with final numbers
- [x] On reconnect all 3 appear on the warehouse board with the same numbers
- [x] A cancelled order restores stock
- [x] Drafts survive the app being killed
- [ ] Real Sunmi V2 Pro: hardware scan, 58mm print, out-of-paper retry (Milestone B — needs the device)
