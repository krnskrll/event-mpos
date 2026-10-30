-- EPIC-12 C1/C2: offline order confirmation (per-device number blocks) and late sync after close.

-- A range of daily order numbers reserved for one device; carved from order_daily_sequences under
-- its row lock, so device-assigned and server-assigned numbers never collide. Never reused.
CREATE TABLE public.device_number_block (
    id uuid PRIMARY KEY,
    device_id text NOT NULL,
    event_id uuid NOT NULL,
    business_date text NOT NULL,
    range_start bigint NOT NULL,
    range_end bigint NOT NULL,
    next_unused bigint NOT NULL,
    created_at timestamp with time zone NOT NULL
);
CREATE INDEX idx_number_block_device ON public.device_number_block USING btree (device_id, event_id);

ALTER TABLE public.orders
    ADD COLUMN device_id text,
    ADD COLUMN client_created_at timestamp with time zone,
    ADD COLUMN stock_conflict boolean NOT NULL DEFAULT false,
    ADD COLUMN late_synced boolean NOT NULL DEFAULT false;
-- The reference encodes event + date + number: a collision would surface here, never silently.
CREATE UNIQUE INDEX uk_order_reference ON public.orders USING btree (reference_id);

ALTER TABLE public.events ADD COLUMN closed_at timestamp with time zone;
