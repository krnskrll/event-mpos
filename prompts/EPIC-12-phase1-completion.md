# Epic 12 — Phase 1 Completion

## Context
Read `.cursorrules`, `README.md`, `backend-go/PARITY-NOTES.md` and the EPIC-01..11 prompts before starting.
Epics 01–11 are built, but some promised pieces were never finished, and the backend is mid-migration from
Kotlin/Spring Boot (`backend/`) to Go (`backend-go/`). This epic closes those gaps. It adds **no new product
scope**.

**Out of scope — do not build:** payments of any kind (QRIS, cash, card, EDC, refunds, split tender), loyalty
points, multi-outlet, delivery/marketplace integrations. Customers still pay at the host retailer's cashier.

## Goal
Make event-mpos usable at a real event on a real Sunmi V2 Pro, including when the venue's internet drops. Make
Go the only backend, with CI protecting it.

## Ground Rules
- **Go is the backend from Milestone A onwards.** Do all new backend work in `backend-go/`. Do not add features
  to `backend/` (Kotlin).
- Schema changes go in new numbered golang-migrate files (`backend-go/migrations/00000N_*.up.sql` + `.down.sql`).
  Never edit `000001_init`.
- The promotion engine exists twice: `backend-go/internal/promotion/engine.go` and
  `android/.../domain/promotion/PromotionEngine.kt`. They must give identical results. Change both or neither.
- Stock changes (decrement **and** restore) must use the row-locked path in the stock service. Never write a
  bare `UPDATE`.
- Every order state change writes to `order_audit_log` in the **same transaction**. The log is append-only, and
  a DB trigger enforces this.
- Hardware stays behind `PrinterPort` / `ScannerPort`. The app must still run on an emulator with no Sunmi
  device.
- Work in this order: **A1 → A2 → D → C3/C4 → C1/C2 → B → E**. B can run in parallel whenever a physical V2 Pro
  is available.
- One branch + one PR per milestone (see "Git Workflow"). Never push directly to `main`.

## Local Environment (this machine)
- No system Node, JDK, Gradle or Homebrew. Docker Desktop is installed (start it with `open -a Docker`).
- Backend: `docker compose up --build` from the repo root.
- Android build: run
  `cd android && JAVA_HOME="/Applications/Android Studio.app/Contents/jbr/Contents/Home" ANDROID_HOME=~/Library/Android/sdk ./gradlew :app:assembleDebug`.
  The emulator AVD is `Pixel7_API34`, and the app reaches the backend at `10.0.2.2:8080`.
- Web: Node is not installed locally. Use `docker run --rm -v "$PWD/web":/app -w /app node:20 npm ...` or rely
  on CI.
- `gh` is authenticated as `krnskrll`.

---

## Milestone A — Go cut-over and safety net

### A1. Make Go the only backend
1. Push `feature/go-backend-migration` (it has unpushed local commits) and open its PR to `main`. Merge it before
   starting anything else.
2. In `docker-compose.yml`, point the `backend` service at `./backend-go`. Set the env vars from
   `backend-go/.env.example`:
   - `DB_URL=postgres://eventmpos:eventmpos@db:5432/eventmpos?sslmode=disable`
   - `PORT`, `JWT_SECRET`, `SEED_ENABLED=true`, `ERP_SOURCE=mock`, `MEMBER_SOURCE=mock`,
     `CORS_ALLOWED_ORIGINS`.
3. Run `docker compose down -v && docker compose up --build`, then re-run the checks in
   `backend-go/PARITY-NOTES.md` and an Android emulator smoke test:
   login → sync → scan → cart → submit → warehouse board → receipt.
4. Remove `@stomp/stompjs`, `sockjs-client` and `@types/sockjs-client` from `web/package.json`. Nothing in
   `web/src` uses them.
5. Update `.cursorrules` and `README.md` so they describe the Go backend (Gin, GORM, golang-migrate, native
   WebSocket) instead of Spring Boot/STOMP.
6. Do **not** delete `backend/` yet. Delete it at the end of A2, once its tests are ported.

**Done when:** `docker compose up` runs Go, the web board and Android app work end to end, and the docs say Go.

### A2. CI and baseline tests
1. Add `.github/workflows/ci.yml`, triggered on PRs and on pushes to `main`, with three jobs:
   - **go:**
     - Use a `postgres:16` service.
     - Run `go vet ./...`.
     - Apply migrations.
     - Run `go test ./...`.
     - Run `go build ./cmd/server`.
   - **android:** JDK 17, then `./gradlew lint testDebugUnitTest assembleDebug`.
   - **web:** Node 20, then `npm ci`, `npx tsc --noEmit`, `npm run build` and `npx vitest run`. Add vitest if it
     is missing.
2. Port the Kotlin backend tests to Go integration tests against the CI Postgres:
   - AuthFlow: login, refresh rotation, revoked refresh is rejected.
   - EventLifecycle: legal and illegal transitions; selling only while ACTIVE.
   - OrderSubmission: atomic submit; the same idempotency key returns the original order; daily numbering.
   - StockConcurrency: N goroutines buying the last units → no oversell, `remaining >= 0`.
   - BackofficeAdmin: user edit, sync config, report, audit.
3. **Promotion parity test:**
   - Create one shared fixture file, `testdata/promotion-cases.json`, listing carts, promotions and the expected
     chosen promotion and discount.
   - Run it from both a Go test and an Android unit test.
   - Load it from the same repo path in both; copy it in at build time if Gradle needs that.
4. Android unit tests for `CartStore`, `OrderRepositoryImpl` (queueing, PENDING → CONFIRMED/FAILED) and
   `ReceiptManager` (store + print + reprint, print failure does not affect the order).
5. Delete `backend/` and remove any remaining Kotlin-backend references.

**Done when:** CI is green, and changing only one of the two promotion engines makes CI fail.

---

## Milestone D — Back-office fixes (EPIC-11 leftovers, Go only)
1. **Final report for CLOSED events only.** `report.Service.EndOfEvent` and the CSV endpoint return 409 with a
   clear message if the event is not CLOSED. The web Reports page explains why the report is unavailable.
2. **Paginated audit log.**
   - `GET /api/admin/events/{id}/audit` accepts `page`, `size` (default 50, max 200), `actor` (user id) and
     `action`.
   - It returns `{items, page, size, total}`.
   - Add pagination and filters to `web/src/pages/ReportsPage.tsx`.
3. **RECEIPT_PRINTED audit entry.** `POST /api/me/orders/{id}/receipts` writes both `receipt_logs` and an audit
   entry in one transaction.
4. **Allocation summary cache.** Add a 10–15 s in-memory TTL cache on the allocation/stock summary endpoint.
   Clear it for that event on order submit and cancel.

**Done when:** each item has a Go test, and the Reports page pages through a large audit log.

---

## Milestone C — Selling flow

### Decisions (defaults recommended to the user)
If the user has not confirmed these, **stop and ask before starting C1/C2**.

- **D1. Offline confirmation gives a final number.**
  - On each sync, the server gives every device its own block of order numbers for the day, for example 50
    numbers.
  - The device assigns numbers from its block and prints the receipt immediately.
  - On upload, the server checks stock again. If an order would oversell, it is still accepted but flagged
    `stock_conflict=true` for the warehouse instead of being rejected.
- **D2. Orders still queued when an event closes.** The server accepts orders whose device timestamp is before
  the event's `closed_at`, for up to 24 h after closing, and marks them `late_synced=true`. Anything older or
  later is rejected with a clear message.

### C3. Cart drafts (EPIC-07 leftover)
1. Add a Room entity `cart_draft` (id, eventId, label, customerJson, linesJson, createdAtEpochMs) with a DAO and
   a Room migration.
2. Build the `CartRepository` that `CartStore.kt` already refers to: `saveDraft(label)`, `observeDrafts(eventId)`,
   `resume(id)`, `delete(id)`.
3. Cart screen changes:
   - A "Save as draft" action.
   - A drafts list.
   - Resuming a draft asks to confirm if the current cart is not empty.
   - Resuming loads the lines and the customer, then deletes the draft.
4. Delete an event's drafts when the BA switches event or the event closes.
5. Persist the live cart too, so it survives the app process being killed.

### C4. Order cancel
1. Go: add status `CANCELLED`, allowed only from `NEW` or `PACKING`. In one transaction:
   - Restore the stock through the locked stock path.
   - Write an `ORDER_CANCELLED` audit entry with the actor and an optional reason.
   - Push the change over WebSocket after commit.
2. Permissions: a BA can cancel only their own orders, and only while `NEW`. WAREHOUSE and ADMIN can cancel any
   order in an allowed state. An illegal cancel returns 409.
3. Web warehouse board: a cancel action with a confirm dialog, a "Cancelled" filter, and remaining stock
   refreshes.
4. Android: cancel from order history while the order is `NEW`. The order and its stored receipt show
   CANCELLED.
5. The report excludes cancelled orders from revenue and units and shows a cancelled count.

### C1. Offline order confirmation (needs D1 confirmed)
1. **Go:**
   - Migration: add a `device_number_block` table (device_id, event_id, business_date, range_start, range_end,
     next_unused).
   - Add `orders.device_id`, `orders.client_created_at`, `orders.stock_conflict` and `orders.late_synced`.
   - Issue or extend a block for the device in the snapshot response, or via
     `POST /api/me/events/{id}/number-blocks`.
   - Keep the daily sequence so server-assigned and block numbers never collide: allocate blocks from the same
     `OrderDailySequence` row under its lock.
   - `Submit` accepts an optional device-assigned number, which must fall inside that device's block, plus
     `clientCreatedAt`. Idempotency stays the deduplication guard.
2. **Android:**
   - Generate a `deviceId` UUID once and store it in DataStore.
   - Store the block in Room. `confirmCurrentCart` assigns the next number, decrements stock in the local
     snapshot, saves the pending order with its number, and prints the receipt right away.
   - Move printing out of `SubmitOrderWorker`, which now only uploads.
   - When the block is running low (fewer than 10 left) and the device is online, fetch a new block. When the
     block is empty and the device is offline, block checkout with a clear message.
3. The receipt shows the final order number whether the order was confirmed online or offline.

### C2. Event close with queued orders (needs D2 confirmed)
1. Go: add `events.closed_at`, set when the event moves to CLOSED. `Submit` applies the grace window from D2.
2. Android: when the event is CLOSED but the device still has PENDING orders, show a banner and keep retrying
   uploads.
3. The web report and warehouse board show a "late-synced" marker on those orders.

**Done when:**
- With airplane mode on, a BA completes 3 orders and prints 3 receipts with final numbers.
- When the device reconnects, all 3 appear on the warehouse board with the same numbers.
- A cancelled order restores stock.
- Drafts survive the app being killed.

---

## Milestone B — Real Sunmi hardware (needs a physical Sunmi V2 Pro)

### B1. Printer
1. Add the Sunmi InnerPrinter SDK. Use Sunmi's Maven printer library, or the vendor AIDL file under
   `app/src/main/aidl/woyou/aidlservice/jiuiv5/` as described in `SunmiPrinterPort.kt`.
2. Implement `core/printer/SunmiPrinterPort.kt`:
   - Bind the service once and handle reconnects.
   - `status()` translates the printer status codes into `PrinterStatus`.
   - `print()` prints the lines, then feeds 3 lines (the V2 Pro has no cutter), and reports failure through the
     callback.
3. In `di/AppModule.kt`, replace `@Binds FakePrinterPort` with a `@Provides` that picks the port at runtime. Use
   Sunmi when `Build.MANUFACTURER` is SUNMI and the service binds, and `FakePrinterPort` otherwise.

### B2. Hardware scanner
1. Add `core/scanner/SunmiScannerPort.kt`. It registers a BroadcastReceiver for the Sunmi scanner's
   decoded-data broadcast and emits each value on `scans`.
2. Bind one combined `ScannerPort` that merges the hardware scanner, the camera and manual entry. Scan screens
   must not change.

**Done when:** on a real V2 Pro, a hardware scan adds to the cart, a confirmed order prints a 58 mm receipt, and
running out of paper shows a retry message without affecting the order. The emulator still works with the Fake
printer and the camera scanner.

---

## Milestone E — Docs and release
1. Update the README's Epic Status to cover EPIC-01..12 and the Go architecture.
2. Write `docs/EPIC-10-verification.md`, `docs/EPIC-11-verification.md` and `docs/EPIC-12-verification.md` in
   the same format as `docs/EPIC-01..09-verification.md`.
3. **Ask the user before pushing any tag.** Then tag `v1.0.0` on the commit where EPIC-11 was merged (it was
   never tagged), and `v1.1.0` on `main` after this epic.

---

## Things Worth Paying Attention To
- Device-assigned order numbers must never collide, even across reinstalls. Treat a reinstall as a new
  `deviceId`, and never reuse a block.
- Offline local stock is a best guess. The server is still the authority, so flag conflicts instead of losing
  orders.
- WebSocket auth on Go uses `?token=` in the URL. Make sure that token never appears in server access logs.
- Keep `PromotionEngine` parity whenever promotion code is touched.
- A schema change needs `docker compose down -v` on old dev databases. Say so in the PR description.

## Git Workflow
One branch and one PR per milestone. Branch from an up-to-date `main` after the previous milestone's PR merges:

```bash
git checkout main && git pull origin main
git checkout -b feature/epic-12-<milestone>   # e.g. epic-12-a2-ci, epic-12-d-backoffice, epic-12-c-selling, epic-12-b-sunmi, epic-12-e-docs
# ... work, commit in small steps ...
git push -u origin feature/epic-12-<milestone>
gh pr create --base main --fill
```

A1 is the exception: its PR is the existing `feature/go-backend-migration` branch, plus the compose, docs and
web dependency changes from A1.
