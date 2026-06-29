# Epic 04 — Member / Customer Integration

## Context
Read `.cursorrules` and review what was built in Epics 01–03 before starting.

Customer data lives in an external member system. The platform syncs it on a schedule — the Android app does not call the member system directly. When a BA scans a customer's QR code, the lookup must be instant and work without network.

## Goal
Sync member/customer data from the external system into the platform, pre-load it to the device at event login, and enable fast QR-based customer identification on the mPOS.

## What Needs to Work When This Epic Is Done
- A backend sync job pulls member data from the external system on a configurable schedule
- Member data is included in the event snapshot downloaded at login (or as a separate step if the dataset is large)
- A BA can scan a customer QR code and instantly see the customer's name and member tier from local cache
- If a member is not found in the local cache and there is network, the app falls back to an online lookup
- If neither works, the BA can manually type the customer's name to proceed
- The order must be linked to a customer before it can be submitted

## Things Worth Paying Attention To
- Think about the QR code content format — the lookup needs to match exactly what is encoded in the QR
- Member datasets can be large — consider whether to sync all members or filter (active only, by region, etc.)
- The member data model should be minimal: only what is needed for display and order linking — no financial data
- The QR code column must be indexed for fast lookup in both PostgreSQL and Room

## When Done

```bash
git checkout main && git pull origin main
git checkout -b feature/epic-04-member-integration
git add .
git commit -m "feat: epic-04 member sync and customer QR identification"
git push origin feature/epic-04-member-integration
```

Then open a Pull Request to `main` on https://github.com/krnskrll/event-mpos
