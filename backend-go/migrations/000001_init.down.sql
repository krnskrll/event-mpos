-- Rollback of the initial schema. Drops every table (CASCADE clears FKs/indexes/triggers) and the
-- append-only guard function.
DROP TABLE IF EXISTS order_audit_log CASCADE;
DROP TABLE IF EXISTS receipt_logs CASCADE;
DROP TABLE IF EXISTS order_items CASCADE;
DROP TABLE IF EXISTS orders CASCADE;
DROP TABLE IF EXISTS order_daily_sequences CASCADE;
DROP TABLE IF EXISTS promotions CASCADE;
DROP TABLE IF EXISTS event_stock_allocations CASCADE;
DROP TABLE IF EXISTS event_products CASCADE;
DROP TABLE IF EXISTS product_barcodes CASCADE;
DROP TABLE IF EXISTS product_variants CASCADE;
DROP TABLE IF EXISTS products CASCADE;
DROP TABLE IF EXISTS members CASCADE;
DROP TABLE IF EXISTS member_sync_status CASCADE;
DROP TABLE IF EXISTS erp_sync_status CASCADE;
DROP TABLE IF EXISTS ba_event_assignments CASCADE;
DROP TABLE IF EXISTS events CASCADE;
DROP TABLE IF EXISTS refresh_tokens CASCADE;
DROP TABLE IF EXISTS sync_config CASCADE;
DROP TABLE IF EXISTS users CASCADE;
DROP FUNCTION IF EXISTS reject_order_audit_mutation() CASCADE;
