-- POS Phase 5: cashiers sign in with name + PIN on paired devices; BA/WAREHOUSE logins retire.
CREATE TABLE public.pos_staff (
    id uuid PRIMARY KEY,
    store_id uuid NOT NULL REFERENCES public.pos_stores (id) ON UPDATE CASCADE,
    name character varying(40) NOT NULL,
    role character varying(10) NOT NULL CHECK (role IN ('KASIR', 'SUPERVISOR')),
    pin_hash text NOT NULL,
    active boolean NOT NULL DEFAULT true,
    created_at timestamp with time zone NOT NULL DEFAULT now(),
    updated_at timestamp with time zone NOT NULL DEFAULT now()
);
CREATE INDEX idx_pos_staff_store ON public.pos_staff (store_id);

CREATE TABLE public.pos_staff_locations (
    staff_id uuid NOT NULL REFERENCES public.pos_staff (id) ON DELETE CASCADE,
    location_id uuid NOT NULL REFERENCES public.pos_locations (id) ON DELETE CASCADE,
    PRIMARY KEY (staff_id, location_id)
);
CREATE INDEX idx_pos_staff_locations_location ON public.pos_staff_locations (location_id);

-- A device belongs to one location; removed_at in the future is a closed event's late-sync window.
CREATE TABLE public.pos_devices (
    id uuid PRIMARY KEY,
    store_id uuid NOT NULL REFERENCES public.pos_stores (id) ON UPDATE CASCADE,
    location_id uuid NOT NULL REFERENCES public.pos_locations (id) ON DELETE CASCADE,
    name character varying(40) NOT NULL,
    model text,
    token_hash character(64) NOT NULL UNIQUE,
    paired_at timestamp with time zone NOT NULL DEFAULT now(),
    last_seen_at timestamp with time zone,
    removed_at timestamp with time zone
);
CREATE INDEX idx_pos_devices_location ON public.pos_devices (location_id);
CREATE UNIQUE INDEX uk_pos_devices_name ON public.pos_devices (location_id, name) WHERE removed_at IS NULL;

CREATE TABLE public.pos_pairing_codes (
    code character(6) PRIMARY KEY,
    store_id uuid NOT NULL REFERENCES public.pos_stores (id) ON UPDATE CASCADE,
    location_id uuid NOT NULL REFERENCES public.pos_locations (id) ON DELETE CASCADE,
    expires_at timestamp with time zone NOT NULL,
    created_at timestamp with time zone NOT NULL DEFAULT now()
);

ALTER TABLE public.orders
    ADD COLUMN staff_id uuid,
    ALTER COLUMN ba_user_id DROP NOT NULL;
