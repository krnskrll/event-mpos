-- Items still without a price get 0; they stay "Tidak dijual" (archived), so nothing sells at 0.
UPDATE public.pos_item_variants SET price = 0 WHERE is_default AND price IS NULL;
ALTER TABLE public.pos_item_variants ADD CONSTRAINT pos_item_variants_default_price CHECK (NOT is_default OR price IS NOT NULL);
