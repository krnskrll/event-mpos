ALTER TABLE pos_item_barcodes ADD COLUMN source TEXT NOT NULL DEFAULT 'pos'
    CHECK (source IN ('produk_toko', 'pos'));

UPDATE pos_item_barcodes SET source = 'produk_toko'
WHERE item_id IN (SELECT id FROM pos_items WHERE bhoomy_product_id IS NOT NULL);
