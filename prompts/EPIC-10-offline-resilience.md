# Epic 10 — Background Sync & Offline Resilience

## Context
Read `.cursorrules` and review what was built in Epics 01–09 before starting.
Event venues have unstable internet. The app must handle this gracefully — BAs should not be blocked from working just because the connection is slow or drops momentarily.

## Goal
Harden the Android app so it continues to function during network instability and recovers cleanly when connectivity returns.

## What Needs to Work When This Epic Is Done
- If an order submission fails due to network issues, it is queued and retried automatically in the background when connectivity returns — the BA sees a "Pending" status and is notified when it resolves
- Stock levels refresh silently in the background every 60 seconds without disrupting the BA's current screen
- The BA can manually trigger a full event data refresh with a pull-to-refresh gesture
- The app shows a clear offline indicator when there is no network, and it disappears automatically when reconnected
- The cart remains fully usable offline — BAs can browse, scan, and build carts; only submission requires network
- The retry queue processes in the correct submission order when connectivity returns

## Things Worth Paying Attention To
- Explore WorkManager's retry and backoff strategies — what makes sense for order submission retries in an event context?
- Room should be the single source of truth for all UI — network syncs write to Room, screens observe Room; this pattern should already be in place from earlier epics but audit it here
- Think about what "Pending" orders mean if an event closes before the retry succeeds
- The connectivity state change is the trigger for multiple things at once — stock refresh, retry queue flush — coordinate these carefully

## When Done

```bash
git checkout main && git pull origin main
git checkout -b feature/epic-10-offline-resilience
git add .
git commit -m "feat: epic-10 offline resilience and background sync"
git push origin feature/epic-10-offline-resilience
```

Then open a Pull Request to `main` on https://github.com/krnskrll/event-mpos
