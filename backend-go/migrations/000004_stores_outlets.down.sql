ALTER TABLE public.events DROP COLUMN IF EXISTS address, DROP COLUMN IF EXISTS type;
ALTER TABLE public.members  DROP COLUMN IF EXISTS store_id;
ALTER TABLE public.products DROP COLUMN IF EXISTS store_id;
ALTER TABLE public.events   DROP COLUMN IF EXISTS store_id;
ALTER TABLE public.users    DROP COLUMN IF EXISTS store_id;
DROP TABLE IF EXISTS public.pos_stores;
