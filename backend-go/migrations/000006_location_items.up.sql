-- POS Phase 4a-2: what a location sells, per variant, with an optional location price (whole rupiah).
CREATE TABLE public.pos_location_items (
    id uuid PRIMARY KEY,
    location_id uuid NOT NULL REFERENCES public.pos_locations (id),
    item_variant_id uuid NOT NULL REFERENCES public.pos_item_variants (id) ON DELETE CASCADE,
    price_override bigint CHECK (price_override >= 0),
    active boolean NOT NULL,
    CONSTRAINT uk_pos_location_items UNIQUE (location_id, item_variant_id)
);
CREATE INDEX idx_pos_location_items_variant ON public.pos_location_items (item_variant_id);

-- Every variant of a product linked to an event inherits the product's availability.
INSERT INTO public.pos_location_items (id, location_id, item_variant_id, active)
    SELECT gen_random_uuid(), ep.event_id, v.id, ep.available
    FROM public.event_products ep JOIN public.pos_item_variants v ON v.item_id = ep.product_id;

DROP TABLE public.event_products;
