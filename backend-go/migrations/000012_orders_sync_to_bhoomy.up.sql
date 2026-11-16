-- Phase 6c: each order carries its own sending status to Bhoomy (no outbox table), the storage
-- place it was sold from and, per line, the Produk Toko variant. Orders from before 6c are
-- 'skipped': they never reach Bhoomy.
ALTER TABLE public.orders
    ADD COLUMN bhoomy_place_id UUID,
    ADD COLUMN stock_status    VARCHAR(16),
    ADD COLUMN sync_state      VARCHAR(12) NOT NULL DEFAULT 'pending',
    ADD COLUMN sync_attempts   INTEGER     NOT NULL DEFAULT 0,
    ADD COLUMN next_sync_at    TIMESTAMPTZ NOT NULL DEFAULT now(),
    ADD COLUMN sync_error      TEXT,
    ADD COLUMN synced_at       TIMESTAMPTZ;

UPDATE public.orders SET sync_state = 'skipped';

CREATE INDEX idx_orders_sync_pending ON public.orders (next_sync_at) WHERE sync_state = 'pending';

ALTER TABLE public.order_items ADD COLUMN bhoomy_variant_id UUID;
