# EPIC-10 — How to verify

EPIC-10 hardens the **Android app** for unstable venue internet: an offline indicator, silent
background stock refresh, pull-to-refresh, and a retry queue that uploads orders when the connection
returns. Builds on EPIC-01–09. (Since EPIC-12, orders are also *confirmed* offline with a final
number — see [EPIC-12](EPIC-12-verification.md).)

> No new server tables. `docker compose up --build`, then run the app on the emulator
> (`Pixel7_API34`, backend at `10.0.2.2:8080`).

---

## Android (emulator)
Log in as `ba@eventmpos.local` / `Ba123456!` → **Start Session** on EV01.

- **Offline indicator**: `adb shell cmd connectivity airplane-mode enable` → an orange
  "Offline — keep selling…" bar slides in and the catalogue badge shows **Offline**. Disable airplane
  mode → both disappear on their own.
- **Cart works offline**: browse, scan (Catalogue → scan → *Enter manually* → `8990001005`), adjust
  quantities and identify a member (`MBR-00002` matches from the local cache).
- **Retry queue**: confirm orders while offline → they show **SYNCING** in Orders. Reconnect → the
  queue uploads them **oldest first** (the pending query is ordered by creation time) and each turns
  **CONFIRMED**; a local notification reports each one.
- **Background stock refresh**: every 60s the app pulls `/api/me/events/{id}/stock` without touching
  the current screen — sell from another client and watch the "N left" badges change.
- **Pull-to-refresh** on the catalogue re-syncs the whole event snapshot.

## Automated checks (CI)
- `OrderSubmitterTest`: network error / 5xx keep an order PENDING and ask for a retry; 4xx marks it
  FAILED; the queue drains oldest-first and one failure doesn't block the rest.
- `OrderRepositoryImplTest`: confirming queues the order and enqueues the unique upload work.

---

## Acceptance checklist (from the epic)
- [x] Failed submissions are queued and retried automatically when connectivity returns; BA sees status and is notified
- [x] Stock refreshes silently every 60 seconds
- [x] Pull-to-refresh triggers a full event re-sync
- [x] Clear offline indicator that disappears on reconnect
- [x] Cart fully usable offline
- [x] Retry queue processes in submission order
