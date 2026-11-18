-- Phase 6f: the old dashboard is gone. Its logins (users, refresh_tokens) and the stock
-- allocations it managed (event_stock_allocations, pos_items.warehouse_stock) are dropped:
-- stock lives in Bhoomy since 6c, staff sign in with name + PIN since Phase 5.
DROP TABLE IF EXISTS public.refresh_tokens;
DROP TABLE IF EXISTS public.users;
DROP TABLE IF EXISTS public.event_stock_allocations;
ALTER TABLE public.pos_items DROP COLUMN IF EXISTS warehouse_stock;
