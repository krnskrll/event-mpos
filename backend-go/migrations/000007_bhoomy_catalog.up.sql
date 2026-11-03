-- POS Phase 4b: POS's read-only copy of Produk Toko, the bookmark of the last pull, and the links
-- from POS items and variants to Bhoomy.
CREATE TABLE public.pos_catalog_cache (
    bhoomy_variant_id uuid PRIMARY KEY,
    store_id uuid NOT NULL REFERENCES public.pos_stores (id) ON UPDATE CASCADE,
    bhoomy_product_id uuid NOT NULL,
    product_name text NOT NULL,
    variant_name text NOT NULL,
    sku text NOT NULL,
    barcode text,
    is_default boolean NOT NULL,
    brand_name text,
    category_id uuid,
    category_name text,
    image_url text,
    status character varying(16) NOT NULL CHECK (status IN ('active', 'archived')),
    changed_at timestamp with time zone NOT NULL,
    synced_at timestamp with time zone NOT NULL DEFAULT now()
);
CREATE INDEX idx_pos_catalog_cache_product ON public.pos_catalog_cache (store_id, bhoomy_product_id);

CREATE TABLE public.pos_sync_cursors (
    store_id uuid NOT NULL REFERENCES public.pos_stores (id) ON UPDATE CASCADE,
    source character varying(32) NOT NULL,
    cursor_at timestamp with time zone,
    cursor_id uuid,
    last_error text,
    updated_at timestamp with time zone NOT NULL DEFAULT now(),
    PRIMARY KEY (store_id, source)
);

ALTER TABLE public.pos_items ADD COLUMN bhoomy_product_id uuid;
CREATE UNIQUE INDEX uk_pos_items_bhoomy ON public.pos_items (store_id, bhoomy_product_id) WHERE bhoomy_product_id IS NOT NULL;
ALTER TABLE public.pos_item_variants ADD COLUMN bhoomy_variant_id uuid;
CREATE UNIQUE INDEX uk_pos_item_variants_bhoomy ON public.pos_item_variants (store_id, bhoomy_variant_id) WHERE bhoomy_variant_id IS NOT NULL;
