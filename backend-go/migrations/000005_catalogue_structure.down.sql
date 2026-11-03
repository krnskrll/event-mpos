-- Back to Phase 3. Fails (and changes nothing) if two stores already share a code.
DROP INDEX public.uk_pos_locations_code;
DROP INDEX public.uk_pos_items_sku;
DROP INDEX public.uk_pos_item_variants_sku;
DROP INDEX public.uk_pos_item_barcodes_barcode;
DROP INDEX public.uk_members_member_code;
DROP INDEX public.uk_members_qr_code;
CREATE UNIQUE INDEX idx_events_code ON public.pos_locations (code);
CREATE UNIQUE INDEX idx_products_sku ON public.pos_items (sku);
CREATE UNIQUE INDEX idx_product_barcodes_barcode ON public.pos_item_barcodes (barcode);
CREATE UNIQUE INDEX idx_members_member_code ON public.members (member_code);
CREATE UNIQUE INDEX idx_members_qr_code ON public.members (qr_code);

ALTER TABLE public.promotions
    ALTER COLUMN amount_off TYPE numeric(12,2),
    ALTER COLUMN min_subtotal TYPE numeric(12,2);
ALTER TABLE public.order_items
    ALTER COLUMN unit_price TYPE numeric(12,2),
    ALTER COLUMN line_total TYPE numeric(12,2);
ALTER TABLE public.orders
    ALTER COLUMN subtotal TYPE numeric(12,2),
    ALTER COLUMN discount TYPE numeric(12,2),
    ALTER COLUMN total TYPE numeric(12,2);
ALTER TABLE public.pos_item_variants ALTER COLUMN price TYPE numeric(12,2);

ALTER TABLE public.pos_item_barcodes DROP COLUMN store_id;
ALTER TABLE public.pos_item_barcodes RENAME COLUMN item_id TO product_id;
ALTER TABLE public.pos_item_barcodes RENAME CONSTRAINT fk_pos_item_barcodes_item TO fk_products_barcodes;
ALTER TABLE public.pos_item_barcodes RENAME CONSTRAINT pos_item_barcodes_pkey TO product_barcodes_pkey;
ALTER TABLE public.pos_item_barcodes RENAME TO product_barcodes;

ALTER TABLE public.pos_items ADD COLUMN price numeric(12,2);
UPDATE public.pos_items i SET price = v.price FROM public.pos_item_variants v WHERE v.item_id = i.id AND v.is_default;
ALTER TABLE public.pos_items ALTER COLUMN price SET NOT NULL;
DELETE FROM public.pos_item_variants WHERE is_default;
CREATE UNIQUE INDEX idx_product_variants_sku ON public.pos_item_variants (sku);
DROP INDEX public.idx_pos_item_variants_item;
DROP INDEX public.uk_pos_item_variants_default;
ALTER TABLE public.pos_item_variants DROP CONSTRAINT pos_item_variants_default_price, DROP COLUMN store_id, DROP COLUMN is_default;
ALTER TABLE public.pos_item_variants RENAME COLUMN price TO price_override;
ALTER TABLE public.pos_item_variants RENAME COLUMN sell_in_pos TO active;
ALTER TABLE public.pos_item_variants RENAME COLUMN item_id TO product_id;
ALTER TABLE public.pos_item_variants RENAME CONSTRAINT fk_pos_item_variants_item TO fk_products_variants;
ALTER TABLE public.pos_item_variants RENAME CONSTRAINT pos_item_variants_pkey TO product_variants_pkey;
ALTER TABLE public.pos_item_variants RENAME TO product_variants;

ALTER TABLE public.pos_items ADD COLUMN active boolean NOT NULL DEFAULT true;
UPDATE public.pos_items SET active = (status = 'active');
ALTER TABLE public.pos_items DROP COLUMN status;
ALTER TABLE public.pos_items RENAME COLUMN photo_url TO image_url;
ALTER TABLE public.pos_items RENAME COLUMN display_name TO name;
ALTER INDEX public.idx_pos_items_store RENAME TO idx_products_store;
ALTER TABLE public.pos_items RENAME CONSTRAINT fk_pos_items_store TO fk_products_store;
ALTER TABLE public.pos_items RENAME CONSTRAINT pos_items_pkey TO products_pkey;
ALTER TABLE public.pos_items RENAME TO products;

ALTER TABLE public.pos_locations DROP CONSTRAINT pos_locations_type_check;
UPDATE public.pos_locations SET type = 'store' WHERE type = 'outlet';
ALTER TABLE public.pos_locations ADD CONSTRAINT events_type_check CHECK (type IN ('event', 'store'));
ALTER INDEX public.idx_pos_locations_store RENAME TO idx_events_store;
ALTER TABLE public.pos_locations RENAME CONSTRAINT fk_pos_locations_store TO fk_events_store;
ALTER TABLE public.pos_locations RENAME CONSTRAINT pos_locations_pkey TO events_pkey;
ALTER TABLE public.pos_locations RENAME TO events;
