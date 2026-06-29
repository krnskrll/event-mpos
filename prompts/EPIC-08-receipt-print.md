# Epic 08 — Receipt Generation & Thermal Print

## Context
Read `.cursorrules` and review what was built in Epics 01–07 before starting.
The Sunmi V2 Pro has a built-in 58mm thermal printer. Printing uses Sunmi's InnerPrinter SDK via AIDL — not Bluetooth, not a network printer.

## Goal
Automatically print a receipt the moment an order is confirmed, store it locally for reprint, and handle printer failures gracefully.

## What Needs to Work When This Epic Is Done
- When an order is confirmed by the server, the receipt prints automatically without the BA doing anything extra
- The receipt includes: brand name, event name, order number (#042), customer name, product lines with prices, discount line if applicable, total, and a footer directing the customer to the nearest cashier
- The receipt data is saved to Room after printing so it can be reprinted later without any network call
- A BA can reprint any receipt from their order history by tapping a Reprint button — this does not create a new order
- If the printer fails (paper jam, out of paper, etc.), the BA sees a clear error with a retry option — the order itself is not affected
- The printer status is checked before attempting to print

## Things Worth Paying Attention To
- Explore the Sunmi InnerPrinter SDK and how to bind to the print service via AIDL before writing any print logic
- Build a `PrinterPort` abstraction — wrap all Sunmi SDK calls behind an interface so swapping hardware later only requires a new implementation
- A receipt is 58mm wide — think about text alignment and column spacing for a clean layout
- The receipt must only print after the server confirms the order — never on a submission attempt

## When Done

```bash
git checkout main && git pull origin main
git checkout -b feature/epic-08-receipt-print
git add .
git commit -m "feat: epic-08 receipt generation and thermal printing"
git push origin feature/epic-08-receipt-print
```

Then open a Pull Request to `main` on https://github.com/krnskrll/event-mpos
