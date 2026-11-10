# mobile — Flutter mPOS

Flutter rewrite of `android/` (see `Tech Stack Prompts/FLUTTER-MIGRATION-DETAILED-PLAN.md`).

| | |
|---|---|
| Flutter | 3.47.9 (pinned in `.fvmrc` and CI) |
| State / DI | Riverpod |
| Navigation | go_router |
| HTTP | dio |
| Local DB | drift (`lib/core/db/`) |

## Run

```bash
docker compose up --build            # backend on :8080 (repo root)
cd mobile
flutter pub get
flutter run                          # emulator reaches the backend at 10.0.2.2:8080
flutter run --dart-define=API_BASE_URL=http://<mac-lan-ip>:8080/   # real device
```

First run: get a pairing code in Bhoomy (Outlet & Event › location › Tambah perangkat) and enter it on
the pairing screen. Afterwards staff sign in with name + PIN (checked on the device, works offline).

| Stored | Where |
|---|---|
| Device token, device, store, location | Keystore/Keychain (`DeviceStorage`) |
| Staff list with PIN hashes | drift `staff` table, refreshed at start, on resume and every minute |
| Who is signed in | memory only (relaunch or 5 idle minutes lock the device) |

## Check

```bash
flutter analyze && flutter test   # see docs/FLUTTER-verification.md
dart run build_runner build          # after changing drift tables (commit the .g.dart output)
```

## Layout

```
lib/
  app/        router, shell, theme, shared widgets
  core/       config, network, db, printers (Sunmi, Bluetooth), notifications
  domain/     pure Dart models + rules (promotion engine, receipt formatter, cart pricing)
  data/       repositories, order upload, background sync
  features/   one folder per screen
android/…/MainActivity.kt   Sunmi printer + scanner bridge
```
