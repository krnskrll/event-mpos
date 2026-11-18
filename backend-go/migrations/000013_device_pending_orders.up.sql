-- Phase 6e: each device reports how many orders it still has to upload (X-Pending-Orders), so the
-- back office can warn before an event's leftover stock is returned.
ALTER TABLE public.pos_devices ADD COLUMN pending_orders INTEGER NOT NULL DEFAULT 0;
