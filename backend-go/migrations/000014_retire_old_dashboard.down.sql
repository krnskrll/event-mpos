-- Restores the empty tables (data comes back only from the pre-6f backup).
ALTER TABLE public.pos_items ADD COLUMN IF NOT EXISTS warehouse_stock BIGINT NOT NULL DEFAULT 0;

CREATE TABLE public.users (
    id            UUID PRIMARY KEY,
    email         TEXT NOT NULL,
    password_hash TEXT NOT NULL,
    full_name     TEXT NOT NULL,
    role          VARCHAR(20) NOT NULL,
    active        BOOLEAN NOT NULL DEFAULT true,
    created_at    TIMESTAMPTZ NOT NULL,
    store_id      UUID NOT NULL REFERENCES public.pos_stores(id) ON UPDATE CASCADE
);
CREATE UNIQUE INDEX idx_users_email ON public.users (email);
CREATE INDEX idx_users_store ON public.users (store_id);

CREATE TABLE public.refresh_tokens (
    id         UUID PRIMARY KEY,
    user_id    UUID NOT NULL,
    token_hash TEXT NOT NULL,
    expires_at TIMESTAMPTZ NOT NULL,
    revoked    BOOLEAN NOT NULL DEFAULT false,
    created_at TIMESTAMPTZ NOT NULL
);
CREATE UNIQUE INDEX idx_refresh_tokens_token_hash ON public.refresh_tokens (token_hash);
CREATE INDEX idx_refresh_tokens_user_id ON public.refresh_tokens (user_id);

CREATE TABLE public.event_stock_allocations (
    id            UUID PRIMARY KEY,
    event_id      UUID NOT NULL,
    product_id    UUID NOT NULL,
    allocated_qty BIGINT NOT NULL,
    sold_qty      BIGINT NOT NULL,
    version       BIGINT NOT NULL
);
CREATE INDEX idx_event_stock_event ON public.event_stock_allocations (event_id);
CREATE UNIQUE INDEX uk_event_stock ON public.event_stock_allocations (event_id, product_id);
