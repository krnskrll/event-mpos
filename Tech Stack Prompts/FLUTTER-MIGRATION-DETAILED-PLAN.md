# Flutter Migration Plan — Android Kotlin → Flutter mPOS

- **Goal:** one app for any device (Sunmi, Android phone/tablet, iPhone/iPad).
- **Unchanged:** `backend-go/`, `web/`, DB schema, `testdata/promotion-cases.json`, Android `applicationId`.
- **Repo:** stays a monorepo; Flutter app goes in `mobile/`.
- **Source of truth:** current `android/` on `main` + `docs/EPIC-*-verification.md`.
- **Status:** Phases 1–3 done (PR #9, #10, #11); Phase 0 spike waits for devices.

---

## 1. Current app (what we port)

| Item | Today |
|---|---|
| Code | ~6,850 lines Kotlin, 11 screens (Compose) |
| Local DB | Room v11, 8 tables: event_snapshot, product, member, promotion, pending_order, receipt, cart_draft, number_block |
| API | 14 calls under `/api/auth/*`, `/api/me/*` (no WebSocket) |
| Background | 60 s stock refresh + number-block top-up + event-status check; WorkManager order upload |
| Offline | Final order numbers from per-device blocks; receipt printed at confirm; upload later |
| Hardware | `PrinterPort` (Sunmi = stub), `ScannerPort` (camera + manual) |
| Tests | PromotionEngine (shared fixture), ReceiptFormatter, OrderNumbers, OrderSubmitter, Cart |

## 2. Tech mapping

| Concern | Kotlin | Flutter |
|---|---|---|
| UI / state / DI | Compose, ViewModel, Hilt | Widgets, **Riverpod** |
| Navigation | navigation-compose | `go_router` |
| HTTP / JSON | Retrofit, kotlinx.serialization | `dio`, `json_serializable` / `freezed` |
| Local DB | Room | **`drift`** (new schema v1) |
| Storage | DataStore, EncryptedPrefs | `shared_preferences`, `flutter_secure_storage` |
| Background | WorkManager | `workmanager` (Android) + foreground drain (iOS) |
| Scan / images / notif | CameraX + ML Kit, Coil, NotificationManager | `mobile_scanner`, `cached_network_image`, `flutter_local_notifications` |
| Money | BigDecimal | **`decimal`** (never `double`) |
| Sunmi printer/scanner | AIDL (stub) | Kotlin MethodChannel bridge |
| Bluetooth printer | — | ESC/POS package (picked in spike) |

## 3. Must-keep rules

1. Same API contract; optional fields nullable.
2. Idempotency key created once at confirm, reused on retry.
3. Offline confirm: number from block (DB transaction) → store receipt → print. Print failure never changes the order.
4. Upload: network/5xx → retry · other 4xx → FAILED + notify · success → CONFIRMED.
5. One upload at a time across isolates → claim rows in DB (not in-memory lock).
6. Promotion engine passes `testdata/promotion-cases.json`.
7. Receipt output identical to Kotlin tests (58 mm).
8. Rupiah format `Rp 1.032`.
9. Snapshot hash gate · 60 s refresh · top-up below 10 · CLOSED → drop carts, keep uploading.
10. Device id per install in `shared_preferences` (not iOS Keychain).
11. Drafts per event + autosaved live cart.
12. One token refresh at a time; refresh fail → logout.

## 4. Device matrix

| Device | Printer | Scanner |
|---|---|---|
| Sunmi V2 Pro | Built-in (AIDL bridge) | Built-in + camera |
| Android phone/tablet | Bluetooth ESC/POS | Camera |
| iPhone/iPad | Bluetooth **BLE** ESC/POS | Camera |
| Emulator | Fake | Manual + camera |

---

## 5. Before development (~2–3 weeks)

**P1 — Decide** (product owner + tech lead)
- [ ] Decide F1–F5 (§9)
- [ ] Definition of done = all mobile checks in `docs/EPIC-01…12-verification.md` pass on target devices
- [ ] Approve budget/timeline
- [ ] Kotlin freeze after go decision; Sunmi Milestone B built in Flutter only

**P2 — Procure** (start day 1)
- [ ] Apple Developer org account (USD 99/yr, needs D-U-N-S, 1–2+ weeks) — needed by Phase 10 (TestFlight); until then a free Apple ID runs the app on simulator + dev iPhone
- [ ] Mac + Xcode
- [ ] Devices: Sunmi V2 Pro Label, Android phone, Android tablet, iPhone, iPad
- [ ] 1–2 BLE 58 mm Bluetooth printers + paper
- [ ] Play Console only if publishing on Play (else APK/MDM)

**P3 — Team**
- [ ] Assign 1–2 devs + reviewer
- [ ] Dart/Flutter ramp-up (+2–3 weeks if new)
- [ ] Agree lint rules + folder layout

**P4 — Environment**
- [ ] Flutter pinned via fvm; Android Studio; Xcode
- [ ] Backend via `docker compose up` + seeded active event
- [ ] Devices in developer mode; iOS devices registered

**P5 — Repo & process**
- [ ] Branch `feature/flutter-migration`, empty `mobile/` app
- [ ] Flutter CI job added (keep Android job until cut-over)
- [ ] One ticket + one PR per phase

**P6 — Kotlin baseline**
- [ ] Run all verification docs on Kotlin app; record results
- [ ] Screenshots of all screens
- [ ] Sample receipts (golden files) + sample API calls

**P7 — Spike / Phase 0 (1–2 weeks)**
- [ ] Prove: Sunmi print + scanner (broadcast or keyboard mode), BT print from iPhone, QR scan, offline queue survives app kill
- [ ] Performance on Sunmi V2 Pro (Android 7.1, 2 GB RAM): cold start, catalogue scroll, scan → cart
- [ ] No Sunmi yet → do the rest first; go decision waits for the Sunmi test
- [ ] 1-page spike report → **go / no-go**

**Ready to start:** all P1–P7 boxes checked.

---

## 6. Development phases

One PR per phase. Days = one experienced Flutter dev.

| # | Phase | Key steps | Done when | Days |
|---|---|---|---|---|
| 0 | Spike | See P7 | Go decision | 5–8 |
| 1 | Foundation | Create app (minSdk 24), packages, env config, router, theme, dio, drift setup, release keystore (store safely), CI | CI green | 4–5 |
| 2 | Auth | Login, token storage, refresh interceptor, logout | EPIC-01 checks | 2–3 |
| 3 | Event + sync | Event select, snapshot → drift, hash gate, sync screen | EPIC-02 checks | 5–6 |
| 4 | Catalogue | Grid, search, detail/variants, stock flags, offline banner | EPIC-03/05 checks | 3–4 |
| 5 | Scan + customer | ScannerPort (camera/manual/Sunmi), barcode lookup, member QR | EPIC-04 checks | 4–5 |
| 6 | Cart + promo | PromotionEngine (`decimal`) + fixture, cart, drafts, autosave | Fixture 100% green | 5–6 |
| 7 | Orders + offline | Number blocks, confirm, queue, upload, background, sync loop, cancel, history, notifications | EPIC-07/10/12-C checks (airplane mode) | 7–9 |
| 8 | Receipt + print | Formatter + golden tests, store/print/reprint, Sunmi + BT adapters, printer settings | EPIC-08 + 12-B on Sunmi + BT | 6–8 |
| 9 | Shell + layouts | Home, bottom bar, tablet/iPad layouts, empty/error states | Works on phone + tablet | 3–4 |
| 10 | iOS | Signing, permissions, background fetch, TestFlight | Full sell flow on iOS | 3–5 |
| 11 | Hardening | Integration test, all verification docs on 5 devices, fixes | Device sign-off | 5–7 |
| 12 | Cut-over | See §7 | Kotlin removed, CI green | 2 |
| | **Total** | | | **54–72** |

**Timeline:** 11–15 weeks (1 dev) or 7–9 weeks (2 devs: A = data/sync/orders, B = UI/hardware), incl. Phase 0. Pre-development P1–P6 runs ~1–2 weeks before/alongside Phase 0. Estimates are re-checked after the spike.

## 7. Cut-over

1. All checks signed off on all devices.
2. No on-device data migration (no production installs). If Kotlin goes live first: 0 pending orders before upgrading.
3. Tag last Kotlin commit `android-kotlin-final`, delete `android/`.
4. CI: drop Android job, keep Flutter job.
5. Update README + add `docs/FLUTTER-verification.md`.

## 8. Risks

| Risk | Mitigation |
|---|---|
| Apple enrollment delay | Only blocks Phase 10+ (TestFlight/distribution); free Apple ID covers dev testing |
| BT printer differences | Support 1–2 tested BLE models only |
| iOS background limits | Upload on open/resume/reconnect; server accepts late sync 24 h |
| No Sunmi device | Order now; use Fake adapter meanwhile |
| Slow on Sunmi (Android 7.1, 2 GB RAM) | Measure in spike; keep images small, lazy lists; fallback = no-go |
| Lost release keystore | Can't update installed apps; back it up in a password manager |
| Rounding drift | `decimal` + shared fixture in CI |
| Duplicate uploads | DB row claim + idempotency key |
| Team new to Dart | +2–3 weeks; reviewed AI-assisted porting |

**Out of scope:** payments, Flutter web/desktop, backend/web changes, repo split.

## 9. Decisions

| # | Decision | Recommendation | Decided |
|---|---|---|---|
| F1 | Platforms | Android + iOS | ☐ |
| F2 | Printers | Sunmi + 1–2 BLE 58 mm models | ☐ |
| F3 | State management | Riverpod | ☐ |
| F4 | Distribution | APK/MDM (Android), TestFlight → Apple Business Manager (iOS) | ☐ |
| F5 | Who builds | In-house + reviewed AI-assisted porting | ☐ |
| — | Repo split | Deferred, monorepo `mobile/` | ✅ 2026-10-30 |
