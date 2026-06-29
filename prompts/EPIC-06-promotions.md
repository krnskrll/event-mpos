# Epic 06 — Promotions

## Context
Read `.cursorrules` and review what was built in Epics 01–05 before starting.

Promotions are always scoped to a specific event. They auto-apply to the cart — BAs do not manually search for or apply them. The client shows a preview using cached rules; the server applies the authoritative discount at order submission.

## Goal
Allow Admins to define event-scoped promotions and have them automatically apply to the BA's cart on the mPOS, with server-side validation on order submit.

## What Needs to Work When This Epic Is Done
- Admins can create and manage promotions for a specific event on the web dashboard
- Supported promotion types for Phase 1: percentage discount, fixed amount discount, product bundle
- Promotions are included in the event snapshot downloaded to the Android app at login
- Applicable promotions automatically appear on the cart — the BA does not need to do anything
- The server re-validates and applies the authoritative discount when the order is submitted
- Phase 1 rule: apply only one promotion per order (keep it simple)

## Things Worth Paying Attention To
- Client-side promotion preview and server-side promotion application must produce the same result — explore how to share the rule logic or keep them consistent
- Think about what "applicable" means — does a promotion apply to any product, specific products, or a minimum quantity?
- Promotions have optional start and end times within the event — handle expiry

## When Done

```bash
git checkout main && git pull origin main
git checkout -b feature/epic-06-promotions
git add .
git commit -m "feat: epic-06 event-scoped promotions"
git push origin feature/epic-06-promotions
```

Then open a Pull Request to `main` on https://github.com/krnskrll/event-mpos
