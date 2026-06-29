# Manual browser testers

Zero-install HTML pages for exercising the backend by hand — no Node, no Android Studio.
Useful for a quick smoke test or a demo when only the backend is running.

## Use
1. Start the backend: `docker compose up --build` (from repo root → API on http://localhost:8080).
2. Serve this folder over HTTP (the pages call the API cross-origin; the backend allows
   `http://localhost:*`, so they must be served from `localhost`, not opened as `file://`):
   ```bash
   cd tools/manual-testers
   python3 -m http.server 5500
   ```
3. Open in a browser:
   - **Login (EPIC-01):** http://localhost:5500/login-tester.html
     Login as any role, see JWT + role, fetch the BA's assigned event.
   - **Event management + snapshot (EPIC-02):** http://localhost:5500/epic2-tester.html
     Admin: event lifecycle (create/activate/close/rename) + **map a BA to an event**.
     BA: download an event snapshot and see the version gate (`changed:false`) and ACTIVE-only blocking.
   - **Products + ERP sync (EPIC-03):** http://localhost:5500/epic3-tester.html
     Admin: ERP sync + status, **add/edit products**, toggle per-event product availability.
     BA: load the snapshot (lists products) and **scan a barcode** to resolve a product.
   - **Members + QR identification (EPIC-04):** http://localhost:5500/epic4-tester.html
     Admin: member sync + status + directory.
     BA: identify a customer by **QR** (online lookup), with a manual-name fallback when not found.

Seeded dev logins: `admin@eventmpos.local` / `Admin123!`, `ba@eventmpos.local` / `Ba123456!`,
`wh@eventmpos.local` / `Wh123456!`.

> These are dev/QA aids, not part of the shipped product. The real UIs are the React web
> dashboard (`web/`) and the Android app (`android/`).
