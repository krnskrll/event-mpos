# Android mPOS — event-mpos

Kotlin + Jetpack Compose, Clean Architecture + MVVM, Hilt, Retrofit/OkHttp. Runs on the Sunmi V2 Pro.

## What EPIC-01 delivers
- **Login screen** ([`ui/login`](app/src/main/java/com/eventmpos/mpos/ui/login)) — email + password
- **Secure token storage** via EncryptedSharedPreferences
  ([`SecureTokenStorage`](app/src/main/java/com/eventmpos/mpos/core/secure/SecureTokenStorage.kt)):
  keys in the Android Keystore, values AES-256 encrypted at rest
- **Automatic token refresh** on 401 through an OkHttp `Authenticator`
  ([`TokenAuthenticator`](app/src/main/java/com/eventmpos/mpos/data/auth/TokenAuthenticator.kt)),
  single-flight so concurrent 401s refresh once
- **Assigned event after login** — the home screen calls `GET /api/me/event` and shows only the
  BA's event ([`ui/home`](app/src/main/java/com/eventmpos/mpos/ui/home))
- Returning users skip the login screen (tokens already stored)

## Architecture
```
domain/   models + AuthRepository interface (framework-free)
data/     Retrofit AuthApi, DTOs, interceptor, authenticator, AuthRepositoryImpl
core/     SecureTokenStorage (EncryptedSharedPreferences)
di/       Hilt modules (network + bindings)
ui/       Compose screens + ViewModels (MVVM), navigation
```

## Run it
Open the `android/` folder in **Android Studio** (Hedgehog or newer) and Run on an emulator.
Android Studio generates the Gradle wrapper automatically; from the CLI you can run
`gradle wrapper` once, then `./gradlew :app:assembleDebug`.

- Emulator → backend: `API_BASE_URL` defaults to `http://10.0.2.2:8080/` (host as seen from emulator).
- Physical Sunmi device → set `API_BASE_URL` to the backend's LAN IP (e.g. `http://192.168.1.20:8080/`)
  in [`app/build.gradle.kts`](app/build.gradle.kts).

Sign in with a **BA** account (e.g. `ba@eventmpos.local` / `Ba123456!`) to see the assigned event.

## Notes
- `usesCleartextTraffic=true` is for local-dev HTTP only — use HTTPS in production.
- Printer/scanner (Sunmi SDKs) and the event snapshot download come in later epics; EPIC-01 is auth.
