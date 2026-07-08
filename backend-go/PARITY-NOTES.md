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

## Deferred (not a divergence — completes in a later phase)

### Event snapshot collections + version hash — completes in Phase 6
`GET /api/me/events/{id}/snapshot` aggregates products, members, and promotions, and its `version`
hash folds in each of those domains' fingerprints. Those domains land in Phases 3–6, so until then
the Go snapshot returns empty `products`/`members`/`promotions` (memberCount 0) and a version hash
computed from event fields only. The endpoint, auth/assignment guards, version-gating
(`?version=` → `{changed:false}`), and the `event` payload are already parity-verified. Full
snapshot content + version parity is re-checked once Phase 6 completes.

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
