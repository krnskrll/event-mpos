# event-mpos

Event-based Android Mobile POS for retail beauty events.

## System Overview

| Component | Tech | Purpose |
|---|---|---|
| Android mPOS | Kotlin + Jetpack Compose | Beauty Advisor selling tool |
| Backend API | Go (Gin + GORM) + PostgreSQL | Source of truth |
| Web Dashboard | React + TypeScript | Warehouse order execution |

## Project Structure

```
event-mpos/
├── android/          # Android mPOS app (Kotlin)
├── backend-go/       # Go API (Gin, GORM, golang-migrate, native WebSocket)
├── backend/          # Legacy Kotlin/Spring Boot API — no longer run; removed once its tests are ported to Go
├── web/              # React web dashboard
└── prompts/          # Cursor development prompts (one per epic)
```

## Quick Start (local dev)

```bash
# 1. Backend + Postgres (no local Go toolchain needed — uses Docker)
docker compose up --build          # Go API on http://localhost:8080, seeded with demo users
# After pulling a schema change, reset the dev DB first: docker compose down -v

# 2. Web dashboard
cd web && npm install && npm run dev   # http://localhost:5173

# 3. Android app
# open the android/ folder in Android Studio and Run on an emulator
```

## Backend (Go)

`backend-go/` is the only backend. Stack: **Gin** (HTTP), **GORM** on PostgreSQL,
**golang-migrate** (schema in `backend-go/migrations/`, applied on startup — add new numbered
`00000N_*.up.sql` / `.down.sql` files, never edit `000001_init`), JWT auth with rotating refresh
tokens, a DB-driven **robfig/cron** sync scheduler, and a native **WebSocket** for the warehouse
board at `GET /ws/events/{id}/orders?token=<jwt>` (the token is redacted from access logs).
Configuration is environment variables — see `backend-go/.env.example`; the only value that
changes per environment is `DB_URL`. Behavioural parity with the retired Kotlin backend is recorded
in [backend-go/PARITY-NOTES.md](backend-go/PARITY-NOTES.md).

```bash
cd backend-go && go test ./...     # or: docker run --rm -v "$PWD":/app -w /app golang:1.25-alpine go test ./...
```

Seeded demo logins: `admin@eventmpos.local` / `Admin123!` (admin),
`ba@eventmpos.local` / `Ba123456!` (BA, assigned to EV01),
`wh@eventmpos.local` / `Wh123456!` (warehouse).

## Epic Status
- **EPIC-01 — Foundation & Authentication**: ✅ implemented. JWT login + refresh, role-based
  access (Admin/BA/Warehouse), admin user creation + BA→event assignment, Android secure-token
  login, web role-based routing. See [docs/EPIC-01-verification.md](docs/EPIC-01-verification.md).
- **EPIC-02 — Event Management**: ✅ implemented. Event lifecycle DRAFT→ACTIVE→CLOSED, admin
  create/edit/close UI, BA event selection with one-call snapshot download cached in Room,
  version-gated re-sync, auto-select for a single event, and ACTIVE-only selling enforced on
  server + app. See [docs/EPIC-02-verification.md](docs/EPIC-02-verification.md).
- **EPIC-03 — Product Catalogue & ERP Sync**: ✅ implemented. Pluggable ERP adapter (mock/file/REST)
  with scheduled + on-demand sync, products with multiple barcodes, admin product add/edit, per-event
  product availability, products carried in the event snapshot, barcode→product lookup, and an
  Android offline product grid (Room + Coil) with a `ScannerPort` abstraction for the Sunmi scanner.
  See [docs/EPIC-03-verification.md](docs/EPIC-03-verification.md).
- **EPIC-04 — Member / Customer Integration**: ✅ implemented. Pluggable member-system adapter
  (mock/file/REST) with scheduled + on-demand sync and region filtering, members carried in the
  event snapshot, indexed QR lookup, and an Android customer-identify flow (offline Room cache →
  online fallback → manual name). See [docs/EPIC-04-verification.md](docs/EPIC-04-verification.md).
- **EPIC-05 — Warehouse Allocation & Stock**: ✅ implemented. Warehouse stock synced from ERP,
  admin per-event allocation (locked once ACTIVE), warehouse + Android remaining-stock views with
  low-stock/sold-out flags (polling), and an **atomic, oversell-proof decrement** (PostgreSQL
  `SELECT … FOR UPDATE` + `@Version`) proven by a 300-thread concurrency test. See
  [docs/EPIC-05-verification.md](docs/EPIC-05-verification.md).
- **EPIC-06 — Promotions**: ✅ implemented. Event-scoped promotions (percentage / fixed / bundle)
  that auto-apply, a **shared rule engine** (server-authoritative + mirrored client preview, pinned
  by a unit test), promotions in the snapshot, admin CRUD, and an authoritative cart **quote**
  endpoint that picks the single best promotion. See [docs/EPIC-06-verification.md](docs/EPIC-06-verification.md).
- **EPIC-07 — Cart & Order Creation**: ✅ implemented. Local cart with live promo preview + customer
  requirement; **atomic order submit** (stock decrement + promotion + race-safe daily order number
  `SELECT … FOR UPDATE` + idempotency key) in one transaction; **WorkManager** offline retry with
  Pending/Confirmed/Failed status; order history. Receipt-ready confirmation payload for EPIC-08.
  See [docs/EPIC-07-verification.md](docs/EPIC-07-verification.md).
- **EPIC-08 — Receipt Generation & Thermal Print**: ✅ implemented. `PrinterPort` abstraction
  (FakePrinterPort for emulator/logcat; documented Sunmi InnerPrinter AIDL impl), a unit-tested 58mm
  `ReceiptFormatter`, **auto-print on confirm** (status checked first), receipts stored in Room for
  **reprint without network**, and a receipt screen + retry on printer failure (order unaffected).
  See [docs/EPIC-08-verification.md](docs/EPIC-08-verification.md).
- **EPIC-09 — Warehouse Order Dashboard**: ✅ implemented. Real-time order board over a **native
  WebSocket** (broadcast after commit on `/ws/events/{id}/orders`; originally STOMP on Kotlin), with status advance
  NEW→PACKING→READY→COLLECTED (validated), status filter, **new-order beep + flash**, **polling
  fallback** when the socket drops, and a remaining-stock panel. See
  [docs/EPIC-09-verification.md](docs/EPIC-09-verification.md).

## How to Use the Prompts

Each file in `prompts/` is a Cursor prompt for one epic.
Open the project in Cursor, open the relevant prompt file, and paste it into the Cursor chat.
Work through epics in order — each one builds on the previous.

```
prompts/
├── EPIC-01-foundation-auth.md
├── EPIC-02-event-management.md
├── EPIC-03-product-erp-sync.md
├── EPIC-04-member-integration.md
├── EPIC-05-allocation-stock.md
├── EPIC-06-promotions.md
├── EPIC-07-cart-order.md
├── EPIC-08-receipt-print.md
├── EPIC-09-warehouse-dashboard.md
├── EPIC-10-offline-resilience.md
└── EPIC-11-backoffice-admin.md
```

## Git Setup (First Time)

```bash
git init
git remote add origin https://github.com/krnskrll/event-mpos.git
git checkout -b main
git add .
git commit -m "chore: initial project setup"
git push -u origin main
```

## Epic Branch Workflow

Each epic runs on its own branch:

```bash
# Start a new epic
git checkout main
git pull origin main
git checkout -b feature/epic-01-auth

# During development — commit often
git add .
git commit -m "feat: epic-01 add JWT auth and login screen"

# When epic is done
git push origin feature/epic-01-auth
# Then open a Pull Request to main on GitHub
```

## Target Hardware

**Sunmi V2 Pro — Label Version** (2D barcode scanner + 58mm thermal printer)

> ⚠️ Must purchase the **Label Version** to get the 2D scanner required for QR code scanning.
