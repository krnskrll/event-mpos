-- POS Phase 4a-1: tables follow the data model wording, every item has one default variant carrying
-- its base price, money is whole rupiah (BIGINT) and codes are unique per store.

-- Locations: an outlet (permanent shop) or an event (temporary sale).
ALTER TABLE public.events RENAME TO pos_locations;
ALTER TABLE public.pos_locations RENAME CONSTRAINT events_pkey TO pos_locations_pkey;
ALTER TABLE public.pos_locations RENAME CONSTRAINT fk_events_store TO fk_pos_locations_store;
ALTER INDEX public.idx_events_store RENAME TO idx_pos_locations_store;
ALTER TABLE public.pos_locations DROP CONSTRAINT events_type_check;
UPDATE public.pos_locations SET type = 'outlet' WHERE type = 'store';
ALTER TABLE public.pos_locations ADD CONSTRAINT pos_locations_type_check CHECK (type IN ('outlet', 'event'));

-- Items (Katalog POS entries).
ALTER TABLE public.products RENAME TO pos_items;
ALTER TABLE public.pos_items RENAME CONSTRAINT products_pkey TO pos_items_pkey;
ALTER TABLE public.pos_items RENAME CONSTRAINT fk_products_store TO fk_pos_items_store;
ALTER INDEX public.idx_products_store RENAME TO idx_pos_items_store;
ALTER TABLE public.pos_items RENAME COLUMN name TO display_name;
ALTER TABLE public.pos_items RENAME COLUMN image_url TO photo_url;
ALTER TABLE public.pos_items ADD COLUMN status character varying(16) NOT NULL DEFAULT 'active' CHECK (status IN ('active', 'archived'));
UPDATE public.pos_items SET status = 'archived' WHERE NOT active;
ALTER TABLE public.pos_items DROP COLUMN active;

-- Variants: a store_id copy (for per-store SKUs) and one default variant per item with the item's price.
ALTER TABLE public.product_variants RENAME TO pos_item_variants;
ALTER TABLE public.pos_item_variants RENAME CONSTRAINT product_variants_pkey TO pos_item_variants_pkey;
ALTER TABLE public.pos_item_variants RENAME CONSTRAINT fk_products_variants TO fk_pos_item_variants_item;
ALTER TABLE public.pos_item_variants RENAME COLUMN product_id TO item_id;
ALTER TABLE public.pos_item_variants RENAME COLUMN active TO sell_in_pos;
ALTER TABLE public.pos_item_variants RENAME COLUMN price_override TO price;
ALTER TABLE public.pos_item_variants ADD COLUMN is_default boolean NOT NULL DEFAULT false, ADD COLUMN store_id uuid;
UPDATE public.pos_item_variants v SET store_id = i.store_id FROM public.pos_items i WHERE i.id = v.item_id;
INSERT INTO public.pos_item_variants (id, item_id, name, sku, price, sell_in_pos, is_default, store_id)
    SELECT gen_random_uuid(), id, '', sku, price, true, true, store_id FROM public.pos_items;
ALTER TABLE public.pos_item_variants
    ALTER COLUMN store_id SET NOT NULL,
    ADD CONSTRAINT fk_pos_item_variants_store FOREIGN KEY (store_id) REFERENCES public.pos_stores (id) ON UPDATE CASCADE,
    ADD CONSTRAINT pos_item_variants_default_price CHECK (NOT is_default OR price IS NOT NULL);
CREATE UNIQUE INDEX uk_pos_item_variants_default ON public.pos_item_variants (item_id) WHERE is_default;
CREATE INDEX idx_pos_item_variants_item ON public.pos_item_variants (item_id);
ALTER TABLE public.pos_items DROP COLUMN price;

-- Barcodes (interim until products move into Produk Toko).
ALTER TABLE public.product_barcodes RENAME TO pos_item_barcodes;
ALTER TABLE public.pos_item_barcodes RENAME CONSTRAINT product_barcodes_pkey TO pos_item_barcodes_pkey;
ALTER TABLE public.pos_item_barcodes RENAME CONSTRAINT fk_products_barcodes TO fk_pos_item_barcodes_item;
ALTER TABLE public.pos_item_barcodes RENAME COLUMN product_id TO item_id;
ALTER TABLE public.pos_item_barcodes ADD COLUMN store_id uuid;
UPDATE public.pos_item_barcodes b SET store_id = i.store_id FROM public.pos_items i WHERE i.id = b.item_id;
ALTER TABLE public.pos_item_barcodes
    ALTER COLUMN store_id SET NOT NULL,
    ADD CONSTRAINT fk_pos_item_barcodes_store FOREIGN KEY (store_id) REFERENCES public.pos_stores (id) ON UPDATE CASCADE;

-- Money in whole rupiah; .5 rounds up. Totals are rebuilt from the rounded parts so they still add up.
ALTER TABLE public.pos_item_variants ALTER COLUMN price TYPE bigint USING round(price);
ALTER TABLE public.orders
    ALTER COLUMN subtotal TYPE bigint USING round(subtotal),
    ALTER COLUMN discount TYPE bigint USING round(discount),
    ALTER COLUMN total TYPE bigint USING greatest(round(subtotal) - round(discount), 0);
ALTER TABLE public.order_items
    ALTER COLUMN unit_price TYPE bigint USING round(unit_price),
    ALTER COLUMN line_total TYPE bigint USING round(unit_price) * qty;
ALTER TABLE public.promotions
    ALTER COLUMN amount_off TYPE bigint USING round(amount_off),
    ALTER COLUMN min_subtotal TYPE bigint USING round(min_subtotal);

-- Codes unique per store instead of globally (login email stays global).
DROP INDEX public.idx_events_code;
DROP INDEX public.idx_products_sku;
DROP INDEX public.idx_product_variants_sku;
DROP INDEX public.idx_product_barcodes_barcode;
DROP INDEX public.idx_members_member_code;
DROP INDEX public.idx_members_qr_code;
CREATE UNIQUE INDEX uk_pos_locations_code ON public.pos_locations (store_id, code);
CREATE UNIQUE INDEX uk_pos_items_sku ON public.pos_items (store_id, sku);
CREATE UNIQUE INDEX uk_pos_item_variants_sku ON public.pos_item_variants (store_id, sku) WHERE NOT is_default;
CREATE UNIQUE INDEX uk_pos_item_barcodes_barcode ON public.pos_item_barcodes (store_id, barcode);
CREATE UNIQUE INDEX uk_members_member_code ON public.members (store_id, member_code);
CREATE UNIQUE INDEX uk_members_qr_code ON public.members (store_id, qr_code);
