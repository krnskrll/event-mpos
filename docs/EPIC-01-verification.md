# EPIC-01 — How to verify

This walks through proving every acceptance item from the epic actually works. Three layers:
backend (API), web dashboard, Android app.

---

## 0. Prerequisites
- **Docker** (for the backend + Postgres) — no local JDK needed.
- **Node 18+** (for the web dashboard).
- **Android Studio** (for the Android app).

---

## 1. Start the backend

```bash
# from the repo root
docker compose up --build
```
Wait for `Started EventMposApplicationKt`. The DB is seeded with 3 users + event `EV01`
(see backend/README.md for credentials).

Quick liveness check:
```bash
curl -s http://localhost:8080/api/health
# {"status":"UP"}
```

---

## 2. Verify the API directly (curl)

### a) Login returns a JWT + refresh token  ✅ "log in and receive a JWT"
```bash
curl -s -X POST http://localhost:8080/api/auth/login \
  -H 'Content-Type: application/json' \
  -d '{"email":"admin@eventmpos.local","password":"Admin123!"}'
```
You get `accessToken`, `refreshToken`, `expiresInSeconds`, and `user`.

Save the admin token for the next calls:
```bash
ADMIN=$(curl -s -X POST http://localhost:8080/api/auth/login \
  -H 'Content-Type: application/json' \
  -d '{"email":"admin@eventmpos.local","password":"Admin123!"}' | jq -r .accessToken)
```

### b) Role boundaries are enforced  ✅ "appropriate access boundaries"
```bash
# Admin CAN list users:
curl -s -o /dev/null -w "admin->users: %{http_code}\n" \
  http://localhost:8080/api/admin/users -H "Authorization: Bearer $ADMIN"   # 200

# A BA CANNOT reach admin routes:
BA=$(curl -s -X POST http://localhost:8080/api/auth/login -H 'Content-Type: application/json' \
  -d '{"email":"ba@eventmpos.local","password":"Ba123456!"}' | jq -r .accessToken)
curl -s -o /dev/null -w "ba->admin: %{http_code}\n" \
  http://localhost:8080/api/admin/users -H "Authorization: Bearer $BA"      # 403

# No token at all:
curl -s -o /dev/null -w "anon->admin: %{http_code}\n" \
  http://localhost:8080/api/admin/users                                     # 401
```

### c) Admin creates a user + assigns a BA to an event  ✅ "Admin can create accounts / assign BAs"
```bash
# create a new BA
curl -s -X POST http://localhost:8080/api/admin/users -H "Authorization: Bearer $ADMIN" \
  -H 'Content-Type: application/json' \
  -d '{"email":"ba2@eventmpos.local","password":"Ba234567!","fullName":"Cara Counter","role":"BA"}'

# list events to grab EV01's id, list users to grab the new BA's id, then:
EVENT_ID=$(curl -s http://localhost:8080/api/admin/events -H "Authorization: Bearer $ADMIN" | jq -r '.[0].id')
BA2_ID=$(curl -s http://localhost:8080/api/admin/users -H "Authorization: Bearer $ADMIN" | jq -r '.[] | select(.email=="ba2@eventmpos.local") | .id')
curl -s -X POST http://localhost:8080/api/admin/assignments -H "Authorization: Bearer $ADMIN" \
  -H 'Content-Type: application/json' -d "{\"userId\":\"$BA2_ID\",\"eventId\":\"$EVENT_ID\"}"
```

### d) A BA sees only their assigned event  ✅
```bash
curl -s http://localhost:8080/api/me/event -H "Authorization: Bearer $BA" | jq
# {"code":"EV01","name":"Mid-Year Beauty Roadshow", ...}
```

### e) Token refresh works (with rotation)  ✅ "token refreshed automatically when it expires"
```bash
REFRESH=$(curl -s -X POST http://localhost:8080/api/auth/login -H 'Content-Type: application/json' \
  -d '{"email":"admin@eventmpos.local","password":"Admin123!"}' | jq -r .refreshToken)

# exchange it for a new pair
curl -s -X POST http://localhost:8080/api/auth/refresh -H 'Content-Type: application/json' \
  -d "{\"refreshToken\":\"$REFRESH\"}" | jq '{accessToken, refreshToken}'

# the OLD refresh token is now revoked (rotation) -> 401
curl -s -o /dev/null -w "reused old refresh: %{http_code}\n" \
  -X POST http://localhost:8080/api/auth/refresh -H 'Content-Type: application/json' \
  -d "{\"refreshToken\":\"$REFRESH\"}"   # 401
```

### Automated version of all the above
```bash
cd backend && gradle test    # AuthFlowTest covers login, protection, rotation, role boundaries
```

---

## 3. Verify the web dashboard

```bash
cd web && npm install && npm run dev   # http://localhost:5173
```
1. Sign in as `admin@eventmpos.local` / `Admin123!` → you land on **/admin**.
2. Create a user, create an event, assign a BA — tables update.  ✅ admin flows
3. Log out, sign in as `wh@eventmpos.local` / `Wh123456!` → you land on **/warehouse**
   (and can't reach /admin).  ✅ role-based routing
4. Sign in as a BA → **/ba** shows the assigned event.  ✅

Token refresh: the axios interceptor refreshes automatically on 401; you can watch it in the
browser Network tab (a `401` followed by a `refresh` call and a replay of the original request).

---

## 4. Verify the Android app

Open `android/` in Android Studio and Run on an emulator.
1. Sign in as `ba@eventmpos.local` / `Ba123456!`.  ✅ login screen
2. Home screen shows **EV01 — Mid-Year Beauty Roadshow** (only the assigned event).  ✅
3. Kill and relaunch the app → it skips login (tokens persisted securely).  ✅ secure storage
4. Log out → tokens cleared, back to login.

**Confirm tokens are encrypted at rest** (secure storage requirement):
```bash
adb shell run-as com.eventmpos.mpos cat /data/data/com.eventmpos.mpos/shared_prefs/empos_secure_tokens.xml
```
The keys and values are ciphertext, not the raw JWT — proving EncryptedSharedPreferences is in effect.

---

## Acceptance checklist (from the epic)
- [x] Log in with email/password → JWT
- [x] Token auto-refreshes when it expires (web interceptor + Android `Authenticator`)
- [x] Role access boundaries enforced on the API
- [x] Android login screen + secure on-device token storage
- [x] Web login + role-based routing
- [x] Admin can create users & assign BAs to events (web)
- [x] BA on Android sees only their assigned event
