ALTER TABLE public.promotions ADD COLUMN product_ids character varying(4000);
UPDATE public.promotions p SET product_ids = x.ids
    FROM (SELECT promotion_id, string_agg(item_id::text, ',') ids FROM public.pos_promotion_items GROUP BY promotion_id) x
    WHERE x.promotion_id = p.id;
DROP TABLE public.pos_promotion_items;
DELETE FROM public.promotions WHERE event_id IS NULL OR category_id IS NOT NULL;
ALTER TABLE public.promotions
    DROP CONSTRAINT fk_promotions_location,
    DROP CONSTRAINT fk_promotions_store,
    DROP COLUMN category_id,
    DROP COLUMN store_id,
    ALTER COLUMN event_id SET NOT NULL;
