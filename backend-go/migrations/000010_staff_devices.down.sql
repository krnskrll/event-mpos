UPDATE public.orders SET ba_user_id = COALESCE(staff_id, '00000000-0000-0000-0000-000000000000') WHERE ba_user_id IS NULL;
ALTER TABLE public.orders
    ALTER COLUMN ba_user_id SET NOT NULL,
    DROP COLUMN staff_id;
DROP TABLE public.pos_pairing_codes;
DROP TABLE public.pos_devices;
DROP TABLE public.pos_staff_locations;
DROP TABLE public.pos_staff;
