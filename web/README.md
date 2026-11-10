# Web Dashboard — event-mpos

React + TypeScript (Vite). The old admin dashboard, kept only for **stock allocation** until
Phase 6 moves stock to Bhoomy. Staff, devices, locations, products, promotions, members and reports
are managed in Bhoomy; cashiers work in the Flutter app on paired devices.

| Page | What |
|---|---|
| Login | ADMIN accounts only (`create-admin`, or the dev seed) |
| Stock ([`src/pages/StockPage.tsx`](src/pages/StockPage.tsx)) | Allocate stock per location |

Tokens refresh automatically on 401 (single-flight) via [`src/api/client.ts`](src/api/client.ts).

## Run it
Needs Node 18+.
```bash
cd web
npm install
npm run dev
```
Opens http://localhost:5173. API calls to `/api/*` are proxied to `http://localhost:8080`
(the Go backend) — start it first with `docker compose up --build` from the repo root.
Sign in with `admin@eventmpos.local` / `Admin123!`.

## Build and test
```bash
npx tsc --noEmit && npm run build && npm test
```
