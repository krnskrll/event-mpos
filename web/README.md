# Web Dashboard — event-mpos

React + TypeScript (Vite). Used by the warehouse team and admins.

## What EPIC-01 delivers
- Login page → stores tokens, routes by role
- **Role-based routing**: ADMIN → `/admin`, WAREHOUSE → `/warehouse`, BA → `/ba`
- Automatic token refresh on 401 (single-flight) via the axios interceptor in
  [`src/api/client.ts`](src/api/client.ts)
- **Admin dashboard** ([`src/pages/AdminDashboard.tsx`](src/pages/AdminDashboard.tsx)): create users,
  create events, and assign a BA to an event

## Run it
Needs Node 18+.
```bash
cd web
npm install
npm run dev
```
Opens http://localhost:5173. API calls to `/api/*` are proxied to `http://localhost:8080`
(the backend) — so start the backend first (see `../backend/README.md`).

Sign in with the seeded accounts (e.g. `admin@eventmpos.local` / `Admin123!`).

## Build
```bash
npm run build
```

## Notes
- Tokens live in `localStorage` for this internal dashboard; the Android app uses
  EncryptedSharedPreferences instead (shared physical device).
- The warehouse and BA pages are intentionally thin in EPIC-01 — they prove auth + routing work.
  The live packing board arrives in EPIC-09.
