DROP INDEX public.uk_pos_item_variants_bhoomy;
ALTER TABLE public.pos_item_variants DROP COLUMN bhoomy_variant_id;
DROP INDEX public.uk_pos_items_bhoomy;
ALTER TABLE public.pos_items DROP COLUMN bhoomy_product_id;
DROP TABLE public.pos_sync_cursors;
DROP TABLE public.pos_catalog_cache;
