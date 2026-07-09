# Parity Notes — Go backend vs Kotlin backend

The Go backend is validated against the live Kotlin backend with a side-by-side response diff
(same request → both servers → compare status + normalized JSON). Volatile fields (tokens, ids,
timestamps) are normalized out. This file records the few **intentional** divergences — everything
else is byte-identical.

## Intentional divergences

### 1. Malformed login/refresh body with a *missing* required field → 400 (Go) vs 401 (Kotlin)
- **When**: a JSON body omits a required key entirely, e.g. `POST /api/auth/login {"email":"x"}`
  (no `password`).
- **Kotlin**: returns `401 {"status":401,"error":"Unauthorized","message":"Authentication required"}`.
  This is incidental, not intentional: Jackson-Kotlin cannot bind the missing non-null field, Spring
  forwards the request to `/error`, and `/error` is not in the security permit-list, so the request
  (unauthenticated) is rejected with 401.
- **Go**: returns a correct `400 Bad Request` with an `ApiError` body naming the missing field.
- **Why we diverge**: 401 for a malformed body is a bug, not a contract. Replicating it would be
  wrong, and no client is affected — the Android app and web dashboard always serialize both fields,
  so the missing-key path is never exercised in practice. When a field is *present but empty/invalid*,
  both backends return an identical `400`.
- **Same family**: an invalid enum *path variable* (e.g. `PUT /api/admin/sync/FOO`) also hits
  Kotlin's `/error`-forward → 401, where Go returns a clean 400 naming the allowed values. Same
  reasoning; no client sends an invalid sync type.

### 1b. JSON money fields: numerically identical, trailing zeros may differ
- Money read back from a `numeric(12,2)` column serializes as a JSON number that is **numerically
  equal** across backends but may differ textually: Go emits e.g. `142200` where Kotlin's BigDecimal
  emits `142200.00` (Go's decimal drops trailing zeros; freshly-computed values that were rounded
  still carry 2 places). This is invisible to clients — `JSON.parse` / kotlinx both yield the same
  number, and the web formats via `rupiah()`. Human-facing CSV export **is** rendered fixed-2
  (`StringFixed(2)`) to match Kotlin exactly. (A dedicated Money type would make the JSON textually
  identical too, if ever required.)

## Deferred (not a divergence — completes in a later phase)

### Event snapshot collections + version hash — completes in Phase 6
`GET /api/me/events/{id}/snapshot` aggregates products, members, and promotions, and its `version`
hash folds in each of those domains' fingerprints. Those domains land in Phases 3–6, so until then
the Go snapshot returns empty `products`/`members`/`promotions` (memberCount 0) and a version hash
computed from event fields only. The endpoint, auth/assignment guards, version-gating
(`?version=` → `{changed:false}`), and the `event` payload are already parity-verified. Full
snapshot content + version parity is re-checked once Phase 6 completes.

### 2. Event snapshot `version` string is backend-relative (opaque cache token)
- **When**: the `version` field of `GET /api/me/events/{id}/snapshot` and the `snapshotHash` of
  `.../snapshot/hash`.
- **Behaviour**: Go computes a deterministic version (stable; changes iff the event's data changes),
  but its exact string differs from Kotlin's.
- **Why we diverge**: the version is an opaque cache token a client only ever compares against the
  *same* backend (the device stores the version it received and sends it back via `?version=`). No
  client compares a Go version to a Kotlin version. Reproducing Kotlin's exact hash would require
  mirroring Java's `UUID.compareTo` (signed-long halves) and Hibernate's unspecified `findAllById`
  ordering — fragile and of zero client value. The version-gating behaviour (`?version=current` →
  `{changed:false}`) and all snapshot *content* (products/members/stock/promotions) are verified
  identical. The parity harness normalizes `version`/`snapshotHash` out.

### 3. Real-time transport: native WebSocket instead of STOMP/SockJS (planned change)
- **Kotlin**: STOMP over SockJS at `/ws`, subscribe topic `/topic/events/{id}/orders`, JWT on the
  STOMP CONNECT frame.
- **Go**: native `gorilla/websocket` at `GET /ws/events/{id}/orders?token=<jwt>`, one message per
  order change (the same `WarehouseOrderView` JSON body). JWT validated on the HTTP upgrade.
- **Why**: gorilla/websocket doesn't speak STOMP, and the only client is the web dashboard (Android
  uses no WebSocket). This is the one deliberate contract change in the migration; the web client was
  updated in lockstep (`web/src/pages/WarehousePage.tsx`: STOMP/SockJS → native WebSocket). The
  message payload is unchanged, so `OrderBoard` rendering is identical.
- **Verified**: connect as warehouse → submit an order as BA → the order view is pushed live;
  a bad/missing token is rejected with HTTP 401 on the upgrade.

## Internal (non-contractual) difference

- **Access-token JWT `alg`**: Kotlin signs with HS384, Go with HS256. The token is opaque to clients
  (they only echo it back in `Authorization`), each backend validates its own, and no response body
  is affected — so this is invisible to the API contract.

## Verified identical (Phase 9)
- `GET /api/admin/sync` overview (schedule + last-run status for ERP + member)
- `PUT /api/admin/sync/{type}` (enabled + cron; runtime-editable); invalid cron → 400;
  dynamic robfig/cron scheduler reschedules on update (disabled by default)
- `GET /api/admin/events/{id}/report` (end-of-event JSON: totals + per-product units/revenue/stock)
- `GET /api/admin/events/{id}/report.csv` (CSV export — byte-identical incl. fixed-2 money)
- `GET /api/admin/events/{id}/audit` (+ `?orderId=`) — append-only order trail
- Divergence: invalid sync-type path → Kotlin 401 quirk vs Go 400 (see §1).

## Verified identical (Phase 7)
- `POST /api/me/events/{id}/orders` — atomic create (variant pricing, server-decided promotion,
  Asia/Jakarta daily order number, ORD-YYYYMMDD-EVxx-NNN reference); idempotency-key returns the
  original order with `duplicate:true`; insufficient stock → 409; empty cart / unknown product /
  unknown variant → 400
- `GET /api/me/events/{id}/orders` (history) + `GET /api/me/orders/{id}` (detail/reprint)
- `POST /api/me/orders/{id}/receipts` → 204 (receipt-log counter)
- `GET /api/warehouse/events/{id}/orders` (board, optional ?status) + `PATCH /api/warehouse/orders/{id}/status`
  (one-step-forward guard; illegal transition → 409)
- Append-only `order_audit_log` written in-transaction for create + status change; Postgres
  BEFORE UPDATE/DELETE trigger installed on startup
- **Concurrency (Go): 20 simultaneous 1-unit orders against an 8-unit allocation → exactly 8 sold,
  0 remaining, 12 rejected — no oversell** (SELECT … FOR UPDATE row lock verified)
- Warehouse-board broadcast (OrderChangedEvent) is deferred to Phase 8 (WebSocket); the order
  service holds a nil broadcaster hook until then.

## Verified identical (Phase 6)
- `GET/POST /api/admin/events/{id}/promotions`; `PUT/DELETE /api/admin/promotions/{id}` (204)
- Validation: percentage needs percentOff; fixed/bundle need amountOff; bundle needs minQty;
  SPECIFIC needs productIds — all → 400 with matching messages
- `POST /api/me/events/{id}/promotions/quote` — subtotal/discount/total + applied promotion; the
  pure engine (percentage/fixed/bundle, scope, minQty/minSubtotal, cap-at-subtotal, best-by
  discount→priority→name) verified against Kotlin and via Go unit tests
- Event snapshot `promotions` array — matches. **Snapshot content is now fully identical to Kotlin
  across products + members + stock + promotions**; only the opaque `version` token differs (above).

## Verified identical (Phase 5)
- `GET /api/admin/events/{id}/stock` (allocation rows); `PUT .../stock/{productId}` — DRAFT-locked
  (ACTIVE → 409), over-warehouse → 400, negative → 400, below-sold → 400
- `GET /api/warehouse/events` + `GET /api/warehouse/events/{id}/stock` (WH+ADMIN; BA → 403);
  lowStock flag (remaining 1..10)
- `GET /api/me/events/{id}/stock` (BA compact levels)
- Event snapshot `products.remainingStock` now populated (matches Kotlin); StockProvider also feeds
  the allocation fingerprint into the snapshot version.
- Atomic oversell-proof `Decrement` (SELECT … FOR UPDATE row lock) implemented; exercised under
  concurrency in Phase 7's order flow.

## Verified identical (Phase 4)
- `GET /api/admin/members` (full directory incl. inactive; no `phone` field exposed)
- `POST /api/admin/members/sync` (Result) + `GET /api/admin/members/sync/status`
- `GET /api/me/members` (active-only compact snapshot DTOs)
- `GET /api/me/members/by-qr/{qr}`; inactive → 404; unknown → 404
- Event snapshot `members` array + `memberCount` (active-only) — now matches Kotlin.
  Member fingerprint uses a Java-`String.hashCode()` port for snapshot-version parity (Phase 6).

## Verified identical (Phase 3)
- `GET /api/admin/products` + `GET /api/admin/products/{id}` — full shape incl. `price` (JSON number),
  `variants` (effective price), `barcodes`, `warehouseStock`, `source`, `active`
- `POST /api/admin/products` → 201; duplicate SKU → 409; barcode-in-use → 409; negative price → 400
- `PUT /api/admin/products/{id}` (partial update; price serialized as supplied)
- `POST /api/admin/erp/sync` (SyncResult) + `GET /api/admin/erp/status`
- `GET /api/admin/events/{id}/products` + `PUT .../products/{productId}` (set availability)
- `GET /api/me/events/{id}/products/by-barcode/{barcode}`; unknown → 404
- Event snapshot `products` array (compact snapshot DTOs incl. active variants) — matches; only
  `remainingStock` stays 0 in Go until Phase 5 wires stock.

## Verified identical (Phase 2)
- `GET /api/me` (all roles); `GET /api/me/event` + `/api/me/events` (BA); non-BA → 400
- `GET /api/me/events/{id}/snapshot` structure + `event` payload + gating; unassigned event → 403
- `GET /api/me/events/{id}/snapshot/hash` shape
- `GET/POST /api/admin/events`; duplicate code → 409
- `PUT /api/admin/events/{id}`; edit CLOSED → 409
- `PATCH /api/admin/events/{id}/status` transitions (DRAFT→ACTIVE→CLOSED); illegal transitions → 409
- `POST /api/admin/assignments`; non-BA → 400; unknown user → 404

## Verified identical (Phase 1)
- `POST /api/auth/login` success shape (accessToken, refreshToken, expiresInSeconds=900, user{})
- `POST /api/auth/refresh` rotation success + `401` on reused/expired/unknown token
- `POST /api/auth/logout` → `200` empty body
- `GET /api/admin/users` (+ `?role=` filter)
- `POST /api/admin/users` → `201`; duplicate email → `409`
- Bad credentials → `401` (full ApiError shape); no token → `401` (terse shape);
  wrong role → `403` (terse shape)
- Present-but-empty / invalid-format field validation → `400`
