-- Pengaturan Struk: one receipt setting per store, printed by the cashier app.
ALTER TABLE public.pos_stores
    ADD COLUMN receipt_phone text NOT NULL DEFAULT '',
    ADD COLUMN receipt_instagram text NOT NULL DEFAULT '',
    ADD COLUMN receipt_note text NOT NULL DEFAULT '',
    ADD COLUMN receipt_show_address boolean NOT NULL DEFAULT true,
    ADD COLUMN receipt_show_staff boolean NOT NULL DEFAULT true,
    ADD COLUMN receipt_show_customer boolean NOT NULL DEFAULT true;
