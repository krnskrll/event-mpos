-- Back to per-product availability; location prices are dropped.
CREATE TABLE public.event_products (
    id uuid NOT NULL,
    event_id uuid NOT NULL,
    product_id uuid NOT NULL,
    available boolean NOT NULL,
    CONSTRAINT event_products_pkey PRIMARY KEY (id)
);
CREATE INDEX idx_event_product_event ON public.event_products USING btree (event_id);
CREATE UNIQUE INDEX uk_event_product ON public.event_products USING btree (event_id, product_id);

INSERT INTO public.event_products (id, event_id, product_id, available)
    SELECT gen_random_uuid(), li.location_id, v.item_id, bool_or(li.active)
    FROM public.pos_location_items li JOIN public.pos_item_variants v ON v.id = li.item_variant_id
    GROUP BY li.location_id, v.item_id;

DROP TABLE public.pos_location_items;
