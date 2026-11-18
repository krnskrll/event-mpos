CREATE TABLE public.erp_sync_status (
    id bigserial PRIMARY KEY,
    last_sync_at timestamp with time zone,
    last_result text,
    product_count bigint,
    source text,
    message text
);
