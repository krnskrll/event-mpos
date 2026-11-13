-- Phase 6a: the outlet is the Cabang. One open outlet per POS store; extra open outlets
-- (staging only, no real data) are closed, keeping the oldest.
UPDATE public.pos_locations p
   SET status = 'CLOSED', closed_at = now(), updated_at = now()
 WHERE p.type = 'outlet' AND p.status <> 'CLOSED'
   AND EXISTS (SELECT 1 FROM public.pos_locations q
                WHERE q.store_id = p.store_id AND q.type = 'outlet' AND q.status <> 'CLOSED'
                  AND (q.created_at, q.id) < (p.created_at, p.id));

CREATE UNIQUE INDEX uq_pos_locations_open_outlet
    ON public.pos_locations (store_id) WHERE type = 'outlet' AND status <> 'CLOSED';
