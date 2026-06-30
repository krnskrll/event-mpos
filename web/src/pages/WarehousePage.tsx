import { useEffect, useRef, useState } from "react";
import Layout from "./Layout";
import { api } from "../api/client";
import type { EventResponse, StockLevelRow } from "../types";

/**
 * Warehouse remaining-stock view. Polls every 4s for near-real-time levels (live WebSocket push
 * arrives in EPIC-09). Low-stock rows are highlighted; sold-out rows flagged.
 */
export default function WarehousePage() {
  const [events, setEvents] = useState<EventResponse[]>([]);
  const [eventId, setEventId] = useState("");
  const [rows, setRows] = useState<StockLevelRow[]>([]);
  const [updatedAt, setUpdatedAt] = useState<string | null>(null);
  const timer = useRef<number | null>(null);

  useEffect(() => {
    api.get<EventResponse[]>("/api/warehouse/events").then((r) => setEvents(r.data));
  }, []);

  useEffect(() => {
    if (timer.current) window.clearInterval(timer.current);
    if (!eventId) {
      setRows([]);
      return;
    }
    const load = async () => {
      const r = await api.get<StockLevelRow[]>(`/api/warehouse/events/${eventId}/stock`);
      setRows(r.data);
      setUpdatedAt(new Date().toLocaleTimeString());
    };
    load();
    timer.current = window.setInterval(load, 4000);
    return () => {
      if (timer.current) window.clearInterval(timer.current);
    };
  }, [eventId]);

  return (
    <Layout title="Warehouse">
      <div className="card">
        <h2>Remaining stock</h2>
        <p className="muted" style={{ marginTop: 0 }}>
          Live remaining quantity per product for an event. Refreshes every few seconds
          {updatedAt ? ` · last updated ${updatedAt}` : ""}.
        </p>
        <label>Event</label>
        <select value={eventId} onChange={(e) => setEventId(e.target.value)}>
          <option value="">Select an event…</option>
          {events.map((ev) => (
            <option key={ev.id} value={ev.id}>{ev.code} — {ev.name} ({ev.status})</option>
          ))}
        </select>

        {eventId && (
          <table>
            <thead>
              <tr><th>SKU</th><th>Product</th><th>Allocated</th><th>Sold</th><th>Remaining</th></tr>
            </thead>
            <tbody>
              {rows.map((r) => (
                <tr key={r.productId}>
                  <td><span className="badge">{r.sku}</span></td>
                  <td>{r.name}</td>
                  <td>{r.allocatedQty}</td>
                  <td>{r.soldQty}</td>
                  <td style={{ fontWeight: 600, color: r.remaining === 0 ? "#ef4444" : r.lowStock ? "#e0b341" : undefined }}>
                    {r.remaining}
                    {r.remaining === 0 ? " · SOLD OUT" : r.lowStock ? " · LOW" : ""}
                  </td>
                </tr>
              ))}
              {rows.length === 0 && (
                <tr><td colSpan={5} className="muted">No stock allocated for this event.</td></tr>
              )}
            </tbody>
          </table>
        )}
      </div>
    </Layout>
  );
}
