-- POS Phase 4c: an item added from Produk Toko starts "Tidak dijual" without a price; it needs a
-- base price before it can sell (checked by the service).
ALTER TABLE public.pos_item_variants DROP CONSTRAINT pos_item_variants_default_price;
