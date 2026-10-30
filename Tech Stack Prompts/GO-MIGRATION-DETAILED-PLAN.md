# Detailed Plan — Kotlin/Spring Boot → Go/Gin Backend Rewrite

> Source of truth: the **current Kotlin backend** (`backend/src`), NOT the June-30 epic prompts.
> The prompts are used only as a contract cross-check.
> Goal: byte-compatible API so Android needs **zero** changes and web needs **one** change (WebSocket client).

---

## 1. What actually exists today (verified inventory)

- **19 JPA entities**: User, RefreshToken, Event, BaEventAssignment, Product, ProductVariant,
  ProductBarcode, EventProduct, EventStockAllocation, Member, MemberSyncStatus, Order, OrderItem,
  OrderDailySequence, Promotion, ReceiptLog, OrderAuditLog, ErpSyncStatus, SyncConfig.
- **~60 endpoints** across 11 controllers, grouped by role prefix:
  - `/api/auth/*` — public (login, refresh, logout)
  - `/api/me/*` — BA (catalogue, snapshot, snapshot hash, orders, promo quote, members, receipts)
  - `/api/admin/*` — ADMIN (users, events, products, ERP sync, members, promotions, stock, reports, audit, sync config)
  - `/api/warehouse/*` — WAREHOUSE (order board, status transitions, stock view)
  - `/api/health` — public
- **Real-time**: STOMP over SockJS at `/ws`, broadcasts to `/topic/events/{eventId}/orders`,
  JWT-authenticated on CONNECT. Only the **web** dashboard consumes it (Android does not).
- **Scheduling**: `DynamicSyncScheduler` (SchedulingConfigurer) drives ERP + member sync from the
  DB-stored `SyncConfig` (cron editable at runtime via `/api/admin/sync`).
- **Schema**: Hibernate `ddl-auto=update` — **no Flyway, no SQL files**. Schema is derived from entities.
- **Pluggable adapters**: ERP + Member sources = mock | file(CSV) | rest, chosen by env var.
- **Post-June refinements that MUST survive** (not in the epic prompts): product variants,
  Asia/Jakarta order-number timezone, Rupiah seed data, append-only audit trigger, receipt logs,
  member `phone` field, snapshot-hash endpoint, bcrypt strength 12, WS JWT auth, rotating
  SHA-256-hashed refresh tokens.

---

## 2. Target Go project layout

```
backend/
  cmd/server/main.go            # manual wiring: config → db → migrate → seed → router → cron → hub → serve
  internal/
    config/config.go            # env → Config struct (mirrors application.yml)
    database/db.go              # GORM + pgx connection
    database/schema.go          # schema management (see §7 — decision pending)
    auth/                       # login/refresh/logout, JWT, rotating hashed refresh tokens, middleware
    user/                       # User, roles (ADMIN/BA/WAREHOUSE)
    event/                      # events, BA assignments, snapshot + snapshot hash
    product/                    # products, variants, barcodes, event-product availability
      erp/                      # ErpProductSource iface + mock/file/rest adapters + sync service
    member/                     # members + sync (mock/file/rest)
    stock/                      # allocations, FOR UPDATE decrement
    promotion/                  # promotion engine (percentage/fixed/bundle)
    order/                      # atomic order creation, daily sequence, warehouse board, status
    receipt/                    # receipt logs
    audit/                      # order audit log + append-only trigger installer
    sync/                       # SyncConfig + dynamic cron rescheduler
    report/                     # end-of-event report + CSV stream + audit list
    admin/                      # admin controller aggregation
    ws/                         # gorilla/websocket hub + JWT-authed upgrade handler
    common/                     # error types, HTTP error mapping, health
  migrations/ (if golang-migrate chosen)
  Dockerfile
  go.mod / go.sum
  .env.example
```

Wiring is explicit in `main.go` — no annotations, no component scanning.

---

## 3. Library mapping

| Concern | Kotlin/Spring | Go |
|---|---|---|
| HTTP | Spring MVC | Gin |
| ORM | JPA/Hibernate | GORM + `gorm.io/driver/postgres` |
| JWT | jjwt | `golang-jwt/jwt/v5` |
| Password | BCrypt(12) | `x/crypto/bcrypt` cost 12 |
| **Money** | `BigDecimal(12,2)` | **`shopspring/decimal`** mapped to `numeric(12,2)` — NOT float64 |
| UUID | `java.util.UUID` | `google/uuid` |
| Cron | SchedulingConfigurer | `robfig/cron/v3` (store EntryIDs for dynamic reschedule) |
| WebSocket | STOMP/SockJS | `gorilla/websocket` hub |
| Optimistic lock | `@Version` | GORM `optimistic lock` via version column + retry |
| Pessimistic lock | `SELECT … FOR UPDATE` | `tx.Clauses(clause.Locking{Strength: "UPDATE"})` |
| Transactions | `@Transactional` | `db.Transaction(func(tx) error {...})` |
| Roles | SecurityFilterChain | Gin middleware groups |
| JSON null-omit | `non_null` inclusion | `omitempty` + pointer fields |
| Scheduled events | ApplicationReadyEvent | called explicitly after wiring in main |

---

## 4. Contract-parity rules (non-negotiable — this is what keeps clients working)

1. **JSON shape is camelCase and identical.** GORM struct tags: `json:"orderNumber"` etc.
2. **Null fields are omitted**, matching Jackson `non_null`. Optional fields = pointers + `omitempty`.
   (Android/web use `.get()` / optional decoding that depend on this.)
3. **All URL paths unchanged**, including `/api/me`, `/api/admin`, `/api/warehouse` prefixes.
4. **HTTP status codes preserved**: 401 unauth, 403 wrong role, 409 insufficient stock / conflict,
   400 validation, 404 not found, 204 for receipt log.
5. **Money serialized as JSON number** (e.g. `79000` / `79000.00`) exactly as today.
6. **Order number**: generated in `Asia/Jakarta`; daily sequence + display format preserved.
7. **Timestamps**: same ISO-8601/offset format Jackson emits.
8. **Auth tokens**: same `TokenResponse` shape; refresh tokens rotated and stored **SHA-256-hashed**;
   access-token TTL 15m, refresh 7d (env-configurable).

A parity test harness (see §8) diffs live Kotlin responses vs Go responses per endpoint.

---

## 5. Trickiest pieces (where bugs hide)

- **Atomic order creation** (`order/service.go`): one transaction — idempotency check →
  `FOR UPDATE` stock decrement per line (variants share product stock) → promotion calc →
  order-number generation → persist → fire post-commit broadcast. Insufficient stock → 409.
- **Dynamic sync scheduler**: cron entries keyed by sync type; editing config removes+re-adds the
  entry (equivalent to SchedulingConfigurer refresh).
- **Append-only audit**: install the Postgres `BEFORE UPDATE/DELETE` trigger on `order_audit_log`
  at startup (skip if not Postgres). Same guard as today.
- **WebSocket JWT**: preserve auth — validate JWT on the WS upgrade (query param or header) and
  reject if invalid. Do **not** ship the plan's `CheckOrigin: return true` without auth.
- **Promotion engine**: percentage / fixed / bundle with scope ALL|SPECIFIC, minQty/minSubtotal,
  priority ordering — port rule-for-rule and cover with unit tests.
- **Snapshot + snapshot hash**: hash must be computed over the same normalized snapshot bytes so the
  Android conditional-sync (`If hash unchanged, skip`) keeps working.

---

## 6. The one web change (WebSocket)

`web/src/pages/WarehousePage.tsx` currently uses `@stomp/stompjs` + `sockjs-client` →
subscribes to `/topic/events/{eventId}/orders`. Switch to native WS:

```ts
const ws = new WebSocket(`${wsBase}/ws/events/${eventId}/orders?token=${accessToken}`);
ws.onmessage = (e) => onOrder(JSON.parse(e.data));
ws.onclose = () => setTimeout(reconnect, 3000);
```
Remove the two STOMP/SockJS deps. Message payload (the `WarehouseOrderView` JSON) is unchanged.

---

## 7. Schema management — **DECISION PENDING** (pick one)

Because there are no Flyway files, the plan's "reuse SQL as-is" is impossible; the schema must be
established fresh. Three options:

### Option A — Hybrid: golang-migrate, auto-generated init  ★ recommended
- I stand up a throwaway Postgres in Docker, let GORM AutoMigrate build the full schema from the
  ported structs, then `pg_dump --schema-only` into `000001_init.up.sql` (+ a matching down).
- The audit trigger becomes `000002_audit_trigger.up.sql`.
- App runs golang-migrate on startup; `schema_migrations` records history.
- **Pros**: production-safe history + rollback from day one; no hand-typing; no struct/SQL drift at
  init; clean path to Supabase/Railway. **Cons**: a little more setup now; future changes need a
  migration file each (small, reviewed).

### Option B — GORM AutoMigrate only
- Structs are the schema; GORM reconciles at startup (mirrors today's `ddl-auto=update`).
- **Pros**: fastest, least code, identical to current behavior. **Cons**: only ever *adds* — cannot
  rename/drop/retype columns or backfill data; no rollback; schema drift is invisible; startup DDL
  can race across instances. Risk shows up once real sales data exists.

### Option C — Pure hand-written golang-migrate
- Author every up/down SQL file by hand (what the original plan asked).
- **Pros**: most explicit/auditable. **Cons**: slowest to build; must be manually kept in sync with
  structs; higher chance of init mismatch.

---

## 8. Verification strategy (no local Go toolchain → Docker)

- Build/test in `golang:1.22` container; Postgres 16 in Docker for integration tests.
- **Unit tests**: promotion engine, order-number/Jakarta TZ, JWT, bcrypt, stock decrement.
- **Parity harness**: script hits each endpoint on the old Kotlin backend and the new Go backend
  with identical inputs and diffs JSON + status codes. Green diff = contract preserved.
- **Live smoke**: `docker compose up`, seed, run the existing browser/no-install testers and the
  Android emulator against the Go backend (points at `10.0.2.2:8080`, unchanged).

---

## 9. Execution phases (incremental, each independently verifiable)

0. Scaffold Go module, config, DB connect, health endpoint, Docker build green.
1. Auth + users (login/refresh/logout, JWT, rotating hashed refresh, middleware, roles). Parity-test.
2. Events + BA assignments + `/api/me` basics + snapshot + snapshot hash.
3. Products + variants + barcodes + ERP adapters + sync + event-product availability.
4. Members + member sync.
5. Stock allocations + FOR UPDATE decrement.
6. Promotions + engine + quote.
7. Orders (atomic create, daily sequence, warehouse board, status) + receipts + audit + trigger.
8. WebSocket hub + JWT upgrade + broadcast; update web client (§6).
9. Sync config + dynamic scheduler; reports (JSON + CSV) + audit list.
10. DataSeeder parity (admin + sample data, Rupiah, variants); full smoke + Android + web run.
11. Dockerfile swap (single binary), docker-compose update, README/DEPLOY notes.

Old Kotlin backend stays in place until Go passes parity, so we can diff side-by-side. Work on
branch `feature/go-backend-migration`; delete Kotlin only at the end.

---

## 10. Corrections to the original plan (for the record)

- No Flyway SQL exists → schema built fresh (§7), not "renamed from V1__…".
- Referenced `Deployment Prompts/` and `DEPLOY-*.md` do not exist; nothing is deployed yet
  (no Supabase/Railway/Vercel) → this migration is local-only for now.
- Epic prompts are stale (June 30) → we port from live Kotlin, using prompts only as a checklist.
- Plan's WS `CheckOrigin: true` drops JWT auth → we keep auth on the WS upgrade.
- `shopspring/decimal` (not float64) for money to avoid rounding drift.
- `omitempty`/pointers required to match Jackson's null-omission (client contract).
