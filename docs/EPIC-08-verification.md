# EPIC-08 — How to verify

EPIC-08 adds **automatic receipt printing on order confirmation**, a 58mm receipt layout, local
storage for **reprint without network**, and graceful **printer-failure handling** — all behind a
`PrinterPort` abstraction. This epic is Android-only (no backend change; it consumes EPIC-07's
confirmation payload). Builds on EPIC-01–07.

> New Room table (receipts) → DB version bump; the app uses destructive migration in dev, so it
> just works on reinstall.

---

## Design
- **`PrinterPort`** (`core/printer/PrinterPort.kt`) wraps all printer calls. Two implementations:
  - **`FakePrinterPort`** (default / emulator) — renders the receipt to Logcat (tag `RECEIPT`) and
    can simulate failures via `simulateStatus`.
  - **`SunmiPrinterPort`** — documented skeleton for the real Sunmi InnerPrinter AIDL service
    (`woyou.aidlservice.jiuiv5`); bind it instead of the fake on a Sunmi device. Swapping hardware =
    one DI line, nothing else changes.
- **`ReceiptFormatter`** (pure, unit-tested) renders 58mm (32-char) lines: brand, event, order #,
  reference, customer, item lines + prices, discount line, total, cashier footer.
- **Auto-print**: `SubmitOrderWorker` calls `ReceiptManager.storeAndPrint` the moment the server
  confirms — never on a submit attempt. Status is checked before printing.
- **Reprint**: the receipt text is stored in Room; the Receipt screen reprints from cache (no network).

## Unit test (run locally)
```bash
cd android
JAVA_HOME="/Applications/Android Studio.app/Contents/jbr/Contents/Home" ./gradlew :app:testDebugUnitTest --tests '*ReceiptFormatterTest'
```
Asserts every line ≤ 32 chars, required fields present, discount line omitted when no discount, and
the total reflects the discount.

## Android (emulator) — the flow
1. Place an order (Browse → Add → View cart → Identify customer → **Confirm**).
2. On confirm the receipt **auto-prints**. With the FakePrinterPort, view it:
   ```bash
   ~/Library/Android/sdk/platform-tools/adb logcat -d -s RECEIPT
   ```
   You'll see the formatted 58mm receipt (brand / event / `#NNN` / customer / lines / discount /
   TOTAL / cashier footer).
3. **Order history → Receipt / Reprint** opens the Receipt screen showing the stored receipt
   (monospace, exactly as printed) with a **Reprint** button → reprints with no new order, no
   network; shows **Printed ✓**.
4. **Failure path**: a non-READY printer status returns a clear failure + the Reprint button retries;
   the order is unaffected (already confirmed, receipt stored). On a real device this is paper-out /
   overheated / unbound; in dev flip `FakePrinterPort.simulateStatus`.

Verified on the emulator: order **#004** auto-printed (`ORD-20260630-EV01-004`, TOTAL RM 89.10 after
10% off) and reprinted from history.

---

## Acceptance checklist (from the epic)
- [x] Receipt prints automatically on server confirmation (no extra BA action)
- [x] Includes brand, event, order number (#NNN), customer, product lines + prices, discount line, total, cashier footer
- [x] Receipt saved to Room after printing → reprint with no network
- [x] Reprint from order history (no new order created)
- [x] Printer failure → clear error + retry; order unaffected
- [x] Printer status checked before printing
- [x] `PrinterPort` abstraction wraps all Sunmi SDK calls (swap = new impl)
- [x] 58mm layout with aligned price columns
- [x] Prints only after server confirms — never on a submission attempt
