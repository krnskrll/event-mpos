CREATE TABLE public.ba_event_assignments (
    id uuid NOT NULL,
    user_id uuid NOT NULL,
    event_id uuid NOT NULL,
    active boolean DEFAULT true NOT NULL,
    assigned_at timestamp with time zone NOT NULL,
    CONSTRAINT ba_event_assignments_pkey PRIMARY KEY (id)
);
CREATE INDEX idx_ba_event_assignments_event_id ON public.ba_event_assignments USING btree (event_id);
CREATE INDEX idx_ba_event_assignments_user_id ON public.ba_event_assignments USING btree (user_id);

UPDATE public.orders SET ba_user_id = COALESCE(staff_id, '00000000-0000-0000-0000-000000000000') WHERE ba_user_id IS NULL;
ALTER TABLE public.orders
    ALTER COLUMN ba_user_id SET NOT NULL,
    DROP COLUMN staff_id;
DROP TABLE public.pos_pairing_codes;
DROP TABLE public.pos_devices;
DROP TABLE public.pos_staff_locations;
DROP TABLE public.pos_staff;
