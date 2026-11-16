ALTER TABLE public.order_items DROP COLUMN IF EXISTS bhoomy_variant_id;
DROP INDEX IF EXISTS public.idx_orders_sync_pending;
ALTER TABLE public.orders
    DROP COLUMN IF EXISTS bhoomy_place_id,
    DROP COLUMN IF EXISTS stock_status,
    DROP COLUMN IF EXISTS sync_state,
    DROP COLUMN IF EXISTS sync_attempts,
    DROP COLUMN IF EXISTS next_sync_at,
    DROP COLUMN IF EXISTS sync_error,
    DROP COLUMN IF EXISTS synced_at;
