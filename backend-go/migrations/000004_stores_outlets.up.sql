-- POS Phase 3: every user, event, product and member belongs to a store. Existing rows go to the
-- default store; its id is relinked to the real Bhoomy store later (ON UPDATE CASCADE).
CREATE TABLE public.pos_stores (
    id uuid PRIMARY KEY,
    name text NOT NULL,
    bhoomy_store_id uuid UNIQUE,
    created_at timestamp with time zone NOT NULL DEFAULT now()
);
INSERT INTO public.pos_stores (id, name) VALUES ('00000000-0000-0000-0000-000000000001', 'Default store');

ALTER TABLE public.users    ADD COLUMN store_id uuid NOT NULL DEFAULT '00000000-0000-0000-0000-000000000001';
ALTER TABLE public.events   ADD COLUMN store_id uuid NOT NULL DEFAULT '00000000-0000-0000-0000-000000000001';
ALTER TABLE public.products ADD COLUMN store_id uuid NOT NULL DEFAULT '00000000-0000-0000-0000-000000000001';
ALTER TABLE public.members  ADD COLUMN store_id uuid NOT NULL DEFAULT '00000000-0000-0000-0000-000000000001';
ALTER TABLE public.users    ALTER COLUMN store_id DROP DEFAULT, ADD CONSTRAINT fk_users_store    FOREIGN KEY (store_id) REFERENCES public.pos_stores (id) ON UPDATE CASCADE;
ALTER TABLE public.events   ALTER COLUMN store_id DROP DEFAULT, ADD CONSTRAINT fk_events_store   FOREIGN KEY (store_id) REFERENCES public.pos_stores (id) ON UPDATE CASCADE;
ALTER TABLE public.products ALTER COLUMN store_id DROP DEFAULT, ADD CONSTRAINT fk_products_store FOREIGN KEY (store_id) REFERENCES public.pos_stores (id) ON UPDATE CASCADE;
ALTER TABLE public.members  ALTER COLUMN store_id DROP DEFAULT, ADD CONSTRAINT fk_members_store  FOREIGN KEY (store_id) REFERENCES public.pos_stores (id) ON UPDATE CASCADE;
CREATE INDEX idx_users_store    ON public.users (store_id);
CREATE INDEX idx_events_store   ON public.events (store_id);
CREATE INDEX idx_products_store ON public.products (store_id);
CREATE INDEX idx_members_store  ON public.members (store_id);

-- An outlet is an event (temporary) or a store (permanent shop, needs an address).
ALTER TABLE public.events
    ADD COLUMN type character varying(8) NOT NULL DEFAULT 'event' CHECK (type IN ('event', 'store')),
    ADD COLUMN address text;
