ALTER TABLE public.events DROP COLUMN IF EXISTS closed_at;
DROP INDEX IF EXISTS public.uk_order_reference;
ALTER TABLE public.orders
    DROP COLUMN IF EXISTS late_synced,
    DROP COLUMN IF EXISTS stock_conflict,
    DROP COLUMN IF EXISTS client_created_at,
    DROP COLUMN IF EXISTS device_id;
DROP TABLE IF EXISTS public.device_number_block;
