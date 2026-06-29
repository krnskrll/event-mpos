# Backend — event-mpos API

Spring Boot (Kotlin) + PostgreSQL + Spring Security + JWT. Source of truth for the whole system.

## What EPIC-01 delivers
- `POST /api/auth/login` — email + password → access JWT + opaque refresh token
- `POST /api/auth/refresh` — rotate refresh token → new access + refresh
- `POST /api/auth/logout` — revoke all refresh tokens for the caller
- `GET  /api/me` — current user
- `GET  /api/me/event` — the active event assigned to the calling BA
- `GET/POST /api/admin/users` — list / create users (ADMIN only)
- `GET/POST /api/admin/events` — list / create events (ADMIN only)
- `POST /api/admin/assignments` — assign a BA to an event (ADMIN only)
- `GET  /api/health` — liveness

Role boundaries live in [`SecurityConfig`](src/main/kotlin/com/eventmpos/config/SecurityConfig.kt).
Adding a protected endpoint later = one matcher there (or `@PreAuthorize`); the JWT filter is untouched.

## Auth design
- **Access token**: stateless JWT, 15 min TTL, carries `sub`, `email`, `role`, `name`.
- **Refresh token**: opaque random string. Only its SHA-256 hash is stored; rotated on every use;
  revocable. See [`RefreshTokenService`](src/main/kotlin/com/eventmpos/auth/RefreshTokenService.kt).
- Passwords hashed with BCrypt.

## Run it

### With Docker (no local JDK needed) — recommended
From the repo root:
```bash
docker compose up --build
```
Postgres + API come up together. API on http://localhost:8080.

### Locally (needs JDK 21 + a running Postgres)
```bash
# Point at your DB via env vars (defaults shown)
export DB_URL=jdbc:postgresql://localhost:5432/eventmpos
export DB_USERNAME=eventmpos DB_PASSWORD=eventmpos
gradle bootRun
```

## Seed data (dev)
On first boot (empty DB, `SEED_ENABLED=true`) these accounts are created:

| Role      | Email                  | Password   | Notes                    |
|-----------|------------------------|------------|--------------------------|
| ADMIN     | admin@eventmpos.local  | Admin123!  |                          |
| BA        | ba@eventmpos.local     | Ba123456!  | assigned to event `EV01` |
| WAREHOUSE | wh@eventmpos.local     | Wh123456!  |                          |

> Change/disable for any real deployment (`SEED_ENABLED=false`, set a real `JWT_SECRET`).

## Tests
```bash
gradle test
```
[`AuthFlowTest`](src/test/kotlin/com/eventmpos/auth/AuthFlowTest.kt) covers login, route protection,
refresh-token rotation, and role boundaries against an in-memory H2 DB.

## Config (env vars)
| Var | Default | Meaning |
|---|---|---|
| `DB_URL` / `DB_USERNAME` / `DB_PASSWORD` | local Postgres | datasource |
| `JWT_SECRET` | dev value | base64 256-bit signing key — **override in prod** |
| `JWT_ACCESS_TTL` | 15 | access token minutes |
| `JWT_REFRESH_TTL` | 7 | refresh token days |
| `SEED_ENABLED` | true | seed dev data on empty DB |
