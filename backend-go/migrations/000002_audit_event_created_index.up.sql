-- Paginated audit log: GET /api/admin/events/{id}/audit lists an event's entries newest first.
CREATE INDEX IF NOT EXISTS idx_audit_event_created ON public.order_audit_log USING btree (event_id, created_at DESC);
