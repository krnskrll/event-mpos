# Epic 03 — Product Catalogue & ERP Sync

## Context
Read `.cursorrules` and review what was built in Epics 01–02 before starting.

Product data comes from an external ERP system. The platform pulls (not receives) this data on a schedule. The ERP export format may be a REST API or a file (CSV/Excel) — build an adapter that can support both.

## Goal
Sync product data from ERP into the platform, allow Admins to select which products are active for each event, and make the product catalogue browsable and barcode-scannable on the Android mPOS.

## What Needs to Work When This Epic Is Done
- A backend sync job pulls product data from ERP on a configurable schedule
- Admins can view synced products on the web dashboard and see the last sync time
- Admins can select which products are available for a specific event
- The Android app shows the event product catalogue in a browsable grid/list view
- A BA can scan a product barcode using the Sunmi built-in scanner to instantly add it to the cart
- Product images load from cache — they should be pre-fetched during event sync so they work offline

## Things Worth Paying Attention To
- Design the ERP sync layer so it is easy to swap between API-pull and file-import without rewriting everything
- The Sunmi device has a built-in barcode scanner with its own SDK — explore how to integrate it properly via AIDL
- A product may have multiple barcodes (variants, different packaging) — the data model should handle this
- Browsing should be fast and work offline — data comes from Room, not the network

## When Done

```bash
git checkout main && git pull origin main
git checkout -b feature/epic-03-product-erp-sync
git add .
git commit -m "feat: epic-03 product catalogue and ERP sync"
git push origin feature/epic-03-product-erp-sync
```

Then open a Pull Request to `main` on https://github.com/krnskrll/event-mpos
