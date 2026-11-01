# Flutter mPOS — How to verify

Flutter app in `mobile/` (plan: `Tech Stack Prompts/FLUTTER-MIGRATION-DETAILED-PLAN.md`).

## Run

```bash
docker compose up -d                       # Go API on :8080 (repo root)
cd mobile && flutter test                  # 40 tests
flutter run                                # emulator → 10.0.2.2:8080
```

Login: `ba@eventmpos.local` / `Ba123456!`

## Checks (emulator, Pixel 7 API 34, against the Go backend — 2026-11-01)

| # | Check | Result |
|---|---|---|
| 1 | Wrong password → "Invalid email or password"; correct → events | ✅ |
| 2 | Expired token → one refresh, request retried | ✅ |
| 3 | Start session → snapshot (`?version=`), stock, number block, event status | ✅ |
| 4 | Catalogue: search, brand chips, stock badges, out of stock not tappable | ✅ |
| 5 | Product detail: shade changes price, quantity, Add to Cart | ✅ |
| 6 | Scan screen: typed barcode → product sheet → Add to Cart | ✅ |
| 7 | Member ID → "matched offline" from local cache → linked to cart | ✅ |
| 8 | Cart: 10% promotion preview, confirm disabled without customer | ✅ |
| 9 | **Airplane mode** → confirm → final number #001 + receipt printed + order SYNCING | ✅ |
| 10 | Reconnect → upload, receipt logged, CONFIRMED, notification shown | ✅ |
| 11 | Cancel → server cancel, CANCELLED, stock returned, receipt stamped | ✅ |
| 12 | Save draft → cart empty → resume draft | ✅ |
| 13 | Live cart survives app restart and reinstall | ✅ |
| 14 | Reprint receipt (offline-capable) | ✅ |
| 15 | Printer screen: test page, Bluetooth scan permission | ✅ |
| 16 | Tablet width → side rail + 5-column grid | ✅ |
| 17 | End session → logout → login | ✅ |
| 18 | DB upgrade v1 → v2 on an existing install | ✅ |
| 19 | iOS simulator build (`flutter build ios --simulator`) | ✅ builds |

## Not verified yet (needs hardware or accounts)

| Item | Needs |
|---|---|
| Sunmi built-in printer + 2D scanner | Sunmi V2 Pro |
| Real Bluetooth printing | BLE 58 mm ESC/POS printer |
| Upload by WorkManager while the app is closed | real device test |
| iOS on device, TestFlight | iPhone/iPad + Apple Developer account |
| Event-closed banner, out-of-numbers message | close an event / exhaust a block on a test backend |
