--
-- PostgreSQL database dump
--


-- Dumped from database version 16.14
-- Dumped by pg_dump version 16.14

SET statement_timeout = 0;
SET lock_timeout = 0;
SET idle_in_transaction_session_timeout = 0;
SET client_encoding = 'UTF8';
SET standard_conforming_strings = on;
SELECT pg_catalog.set_config('search_path', '', false);
SET check_function_bodies = false;
SET xmloption = content;
SET client_min_messages = warning;
SET row_security = off;

--
-- Name: reject_order_audit_mutation(); Type: FUNCTION; Schema: public; Owner: -
--

CREATE FUNCTION public.reject_order_audit_mutation() RETURNS trigger
    LANGUAGE plpgsql
    AS $$
		 BEGIN RAISE EXCEPTION 'order_audit_log is append-only'; END;
		 $$;


SET default_tablespace = '';

SET default_table_access_method = heap;

--
-- Name: ba_event_assignments; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.ba_event_assignments (
    id uuid NOT NULL,
    user_id uuid NOT NULL,
    event_id uuid NOT NULL,
    active boolean DEFAULT true NOT NULL,
    assigned_at timestamp with time zone NOT NULL
);


--
-- Name: erp_sync_status; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.erp_sync_status (
    id bigint NOT NULL,
    last_sync_at timestamp with time zone,
    last_result text,
    product_count bigint,
    source text,
    message text
);


--
-- Name: erp_sync_status_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.erp_sync_status_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: erp_sync_status_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.erp_sync_status_id_seq OWNED BY public.erp_sync_status.id;


--
-- Name: event_products; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.event_products (
    id uuid NOT NULL,
    event_id uuid NOT NULL,
    product_id uuid NOT NULL,
    available boolean NOT NULL
);


--
-- Name: event_stock_allocations; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.event_stock_allocations (
    id uuid NOT NULL,
    event_id uuid NOT NULL,
    product_id uuid NOT NULL,
    allocated_qty bigint NOT NULL,
    sold_qty bigint NOT NULL,
    version bigint NOT NULL
);


--
-- Name: events; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.events (
    id uuid NOT NULL,
    code text NOT NULL,
    name text NOT NULL,
    start_date date,
    end_date date,
    status character varying(20) NOT NULL,
    created_at timestamp with time zone NOT NULL,
    updated_at timestamp with time zone NOT NULL
);


--
-- Name: member_sync_status; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.member_sync_status (
    id bigint NOT NULL,
    last_sync_at timestamp with time zone,
    last_result text,
    member_count bigint,
    source text,
    message text
);


--
-- Name: member_sync_status_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.member_sync_status_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: member_sync_status_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.member_sync_status_id_seq OWNED BY public.member_sync_status.id;


--
-- Name: members; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.members (
    id uuid NOT NULL,
    member_code text NOT NULL,
    qr_code text NOT NULL,
    name text NOT NULL,
    tier character varying(20) NOT NULL,
    region text,
    phone text,
    active boolean NOT NULL,
    updated_at timestamp with time zone NOT NULL
);


--
-- Name: order_audit_log; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.order_audit_log (
    id uuid NOT NULL,
    event_id uuid NOT NULL,
    order_id uuid NOT NULL,
    order_reference text NOT NULL,
    action character varying(20) NOT NULL,
    actor_user_id uuid,
    actor_name text NOT NULL,
    from_status text,
    to_status text,
    detail text,
    created_at timestamp with time zone NOT NULL
);


--
-- Name: order_daily_sequences; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.order_daily_sequences (
    id uuid NOT NULL,
    event_id uuid NOT NULL,
    date_key text NOT NULL,
    last_number bigint NOT NULL,
    version bigint NOT NULL
);


--
-- Name: order_items; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.order_items (
    id uuid NOT NULL,
    order_id uuid NOT NULL,
    product_id uuid NOT NULL,
    variant_id uuid,
    variant_name text,
    sku text NOT NULL,
    name text NOT NULL,
    unit_price numeric(12,2) NOT NULL,
    qty bigint NOT NULL,
    line_total numeric(12,2) NOT NULL
);


--
-- Name: orders; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.orders (
    id uuid NOT NULL,
    event_id uuid NOT NULL,
    ba_user_id uuid NOT NULL,
    idempotency_key text NOT NULL,
    order_number bigint NOT NULL,
    reference_id text NOT NULL,
    customer_name text NOT NULL,
    customer_qr text,
    customer_tier text,
    subtotal numeric(12,2) NOT NULL,
    discount numeric(12,2) NOT NULL,
    total numeric(12,2) NOT NULL,
    promotion_id uuid,
    promotion_name text,
    status character varying(20) NOT NULL,
    created_at timestamp with time zone NOT NULL
);


--
-- Name: product_barcodes; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.product_barcodes (
    id uuid NOT NULL,
    product_id uuid NOT NULL,
    barcode text NOT NULL,
    label text
);


--
-- Name: product_variants; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.product_variants (
    id uuid NOT NULL,
    product_id uuid NOT NULL,
    name text NOT NULL,
    sku text NOT NULL,
    price_override numeric(12,2),
    active boolean DEFAULT true NOT NULL
);


--
-- Name: products; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.products (
    id uuid NOT NULL,
    sku text NOT NULL,
    name text NOT NULL,
    brand text,
    price numeric(12,2) NOT NULL,
    image_url text,
    warehouse_stock bigint NOT NULL,
    source character varying(20) NOT NULL,
    active boolean DEFAULT true NOT NULL,
    created_at timestamp with time zone NOT NULL,
    updated_at timestamp with time zone NOT NULL
);


--
-- Name: promotions; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.promotions (
    id uuid NOT NULL,
    event_id uuid NOT NULL,
    name text NOT NULL,
    type character varying(20) NOT NULL,
    percent_off numeric(5,2),
    amount_off numeric(12,2),
    scope character varying(20) NOT NULL,
    product_ids character varying(4000),
    min_qty bigint,
    min_subtotal numeric(12,2),
    starts_at timestamp with time zone,
    ends_at timestamp with time zone,
    active boolean NOT NULL,
    priority bigint NOT NULL,
    updated_at timestamp with time zone NOT NULL
);


--
-- Name: receipt_logs; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.receipt_logs (
    id uuid NOT NULL,
    order_id text NOT NULL,
    print_count bigint NOT NULL,
    first_printed_at timestamp with time zone NOT NULL,
    last_printed_at timestamp with time zone NOT NULL
);


--
-- Name: refresh_tokens; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.refresh_tokens (
    id uuid NOT NULL,
    user_id uuid NOT NULL,
    token_hash text NOT NULL,
    expires_at timestamp with time zone NOT NULL,
    revoked boolean DEFAULT false NOT NULL,
    created_at timestamp with time zone NOT NULL
);


--
-- Name: sync_config; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.sync_config (
    sync_type text NOT NULL,
    enabled boolean NOT NULL,
    cron text NOT NULL,
    updated_at timestamp with time zone NOT NULL
);


--
-- Name: users; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.users (
    id uuid NOT NULL,
    email text NOT NULL,
    password_hash text NOT NULL,
    full_name text NOT NULL,
    role character varying(20) NOT NULL,
    active boolean DEFAULT true NOT NULL,
    created_at timestamp with time zone NOT NULL
);


--
-- Name: erp_sync_status id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.erp_sync_status ALTER COLUMN id SET DEFAULT nextval('public.erp_sync_status_id_seq'::regclass);


--
-- Name: member_sync_status id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.member_sync_status ALTER COLUMN id SET DEFAULT nextval('public.member_sync_status_id_seq'::regclass);


--
-- Name: ba_event_assignments ba_event_assignments_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.ba_event_assignments
    ADD CONSTRAINT ba_event_assignments_pkey PRIMARY KEY (id);


--
-- Name: erp_sync_status erp_sync_status_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.erp_sync_status
    ADD CONSTRAINT erp_sync_status_pkey PRIMARY KEY (id);


--
-- Name: event_products event_products_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.event_products
    ADD CONSTRAINT event_products_pkey PRIMARY KEY (id);


--
-- Name: event_stock_allocations event_stock_allocations_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.event_stock_allocations
    ADD CONSTRAINT event_stock_allocations_pkey PRIMARY KEY (id);


--
-- Name: events events_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.events
    ADD CONSTRAINT events_pkey PRIMARY KEY (id);


--
-- Name: member_sync_status member_sync_status_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.member_sync_status
    ADD CONSTRAINT member_sync_status_pkey PRIMARY KEY (id);


--
-- Name: members members_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.members
    ADD CONSTRAINT members_pkey PRIMARY KEY (id);


--
-- Name: order_audit_log order_audit_log_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.order_audit_log
    ADD CONSTRAINT order_audit_log_pkey PRIMARY KEY (id);


--
-- Name: order_daily_sequences order_daily_sequences_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.order_daily_sequences
    ADD CONSTRAINT order_daily_sequences_pkey PRIMARY KEY (id);


--
-- Name: order_items order_items_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.order_items
    ADD CONSTRAINT order_items_pkey PRIMARY KEY (id);


--
-- Name: orders orders_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.orders
    ADD CONSTRAINT orders_pkey PRIMARY KEY (id);


--
-- Name: product_barcodes product_barcodes_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.product_barcodes
    ADD CONSTRAINT product_barcodes_pkey PRIMARY KEY (id);


--
-- Name: product_variants product_variants_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.product_variants
    ADD CONSTRAINT product_variants_pkey PRIMARY KEY (id);


--
-- Name: products products_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.products
    ADD CONSTRAINT products_pkey PRIMARY KEY (id);


--
-- Name: promotions promotions_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.promotions
    ADD CONSTRAINT promotions_pkey PRIMARY KEY (id);


--
-- Name: receipt_logs receipt_logs_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.receipt_logs
    ADD CONSTRAINT receipt_logs_pkey PRIMARY KEY (id);


--
-- Name: refresh_tokens refresh_tokens_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.refresh_tokens
    ADD CONSTRAINT refresh_tokens_pkey PRIMARY KEY (id);


--
-- Name: sync_config sync_config_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.sync_config
    ADD CONSTRAINT sync_config_pkey PRIMARY KEY (sync_type);


--
-- Name: users users_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.users
    ADD CONSTRAINT users_pkey PRIMARY KEY (id);


--
-- Name: idx_audit_event; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_audit_event ON public.order_audit_log USING btree (event_id);


--
-- Name: idx_audit_order; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_audit_order ON public.order_audit_log USING btree (order_id);


--
-- Name: idx_ba_event_assignments_event_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_ba_event_assignments_event_id ON public.ba_event_assignments USING btree (event_id);


--
-- Name: idx_ba_event_assignments_user_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_ba_event_assignments_user_id ON public.ba_event_assignments USING btree (user_id);


--
-- Name: idx_event_product_event; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_event_product_event ON public.event_products USING btree (event_id);


--
-- Name: idx_event_stock_event; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_event_stock_event ON public.event_stock_allocations USING btree (event_id);


--
-- Name: idx_events_code; Type: INDEX; Schema: public; Owner: -
--

CREATE UNIQUE INDEX idx_events_code ON public.events USING btree (code);


--
-- Name: idx_members_member_code; Type: INDEX; Schema: public; Owner: -
--

CREATE UNIQUE INDEX idx_members_member_code ON public.members USING btree (member_code);


--
-- Name: idx_members_qr_code; Type: INDEX; Schema: public; Owner: -
--

CREATE UNIQUE INDEX idx_members_qr_code ON public.members USING btree (qr_code);


--
-- Name: idx_order_event; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_order_event ON public.orders USING btree (event_id);


--
-- Name: idx_order_idem; Type: INDEX; Schema: public; Owner: -
--

CREATE UNIQUE INDEX idx_order_idem ON public.orders USING btree (idempotency_key);


--
-- Name: idx_order_item_order; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_order_item_order ON public.order_items USING btree (order_id);


--
-- Name: idx_product_barcodes_barcode; Type: INDEX; Schema: public; Owner: -
--

CREATE UNIQUE INDEX idx_product_barcodes_barcode ON public.product_barcodes USING btree (barcode);


--
-- Name: idx_product_variants_sku; Type: INDEX; Schema: public; Owner: -
--

CREATE UNIQUE INDEX idx_product_variants_sku ON public.product_variants USING btree (sku);


--
-- Name: idx_products_sku; Type: INDEX; Schema: public; Owner: -
--

CREATE UNIQUE INDEX idx_products_sku ON public.products USING btree (sku);


--
-- Name: idx_promotions_event_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_promotions_event_id ON public.promotions USING btree (event_id);


--
-- Name: idx_receipt_logs_order_id; Type: INDEX; Schema: public; Owner: -
--

CREATE UNIQUE INDEX idx_receipt_logs_order_id ON public.receipt_logs USING btree (order_id);


--
-- Name: idx_refresh_tokens_token_hash; Type: INDEX; Schema: public; Owner: -
--

CREATE UNIQUE INDEX idx_refresh_tokens_token_hash ON public.refresh_tokens USING btree (token_hash);


--
-- Name: idx_refresh_tokens_user_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_refresh_tokens_user_id ON public.refresh_tokens USING btree (user_id);


--
-- Name: idx_users_email; Type: INDEX; Schema: public; Owner: -
--

CREATE UNIQUE INDEX idx_users_email ON public.users USING btree (email);


--
-- Name: uk_event_product; Type: INDEX; Schema: public; Owner: -
--

CREATE UNIQUE INDEX uk_event_product ON public.event_products USING btree (event_id, product_id);


--
-- Name: uk_event_stock; Type: INDEX; Schema: public; Owner: -
--

CREATE UNIQUE INDEX uk_event_stock ON public.event_stock_allocations USING btree (event_id, product_id);


--
-- Name: uk_seq_event_date; Type: INDEX; Schema: public; Owner: -
--

CREATE UNIQUE INDEX uk_seq_event_date ON public.order_daily_sequences USING btree (event_id, date_key);


--
-- Name: order_audit_log trg_order_audit_append_only; Type: TRIGGER; Schema: public; Owner: -
--

CREATE TRIGGER trg_order_audit_append_only BEFORE DELETE OR UPDATE ON public.order_audit_log FOR EACH ROW EXECUTE FUNCTION public.reject_order_audit_mutation();


--
-- Name: order_items fk_orders_items; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.order_items
    ADD CONSTRAINT fk_orders_items FOREIGN KEY (order_id) REFERENCES public.orders(id) ON DELETE CASCADE;


--
-- Name: product_barcodes fk_products_barcodes; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.product_barcodes
    ADD CONSTRAINT fk_products_barcodes FOREIGN KEY (product_id) REFERENCES public.products(id) ON DELETE CASCADE;


--
-- Name: product_variants fk_products_variants; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.product_variants
    ADD CONSTRAINT fk_products_variants FOREIGN KEY (product_id) REFERENCES public.products(id) ON DELETE CASCADE;


--
-- PostgreSQL database dump complete
--


