-- POS Phase 4d: promotions belong to a store and apply to one location (event_id) or all of them
-- (NULL), to all items, a Bhoomy category, or chosen items (pos_promotion_items replaces the CSV).
ALTER TABLE public.promotions
    ADD COLUMN store_id uuid,
    ADD COLUMN category_id uuid,
    ALTER COLUMN event_id DROP NOT NULL;
UPDATE public.promotions p SET store_id = l.store_id FROM public.pos_locations l WHERE l.id = p.event_id;
DELETE FROM public.promotions WHERE store_id IS NULL;
ALTER TABLE public.promotions
    ALTER COLUMN store_id SET NOT NULL,
    ADD CONSTRAINT fk_promotions_store FOREIGN KEY (store_id) REFERENCES public.pos_stores (id) ON UPDATE CASCADE,
    ADD CONSTRAINT fk_promotions_location FOREIGN KEY (event_id) REFERENCES public.pos_locations (id) ON DELETE CASCADE;
CREATE INDEX idx_promotions_store ON public.promotions (store_id);

CREATE TABLE public.pos_promotion_items (
    promotion_id uuid NOT NULL REFERENCES public.promotions (id) ON DELETE CASCADE,
    item_id uuid NOT NULL REFERENCES public.pos_items (id) ON DELETE CASCADE,
    PRIMARY KEY (promotion_id, item_id)
);
INSERT INTO public.pos_promotion_items (promotion_id, item_id)
    SELECT DISTINCT p.id, i.id
    FROM public.promotions p, unnest(string_to_array(p.product_ids, ',')) s
    JOIN public.pos_items i ON i.id::text = trim(s)
    WHERE i.store_id = p.store_id;
ALTER TABLE public.promotions DROP COLUMN product_ids;
