import { useEffect, useState } from "react";
import { AxiosError } from "axios";
import Layout from "./Layout";
import { api } from "../api/client";
import { canAllocate } from "../stock";
import type { AllocationRow, EventResponse } from "../types";

function errMsg(err: unknown): string {
  const ax = err as AxiosError<{ message?: string }>;
  return ax.response?.data?.message ?? "Request failed";
}

export default function StockPage() {
  const [events, setEvents] = useState<EventResponse[]>([]);
  const [eventId, setEventId] = useState("");
  const [rows, setRows] = useState<AllocationRow[]>([]);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    api.get<EventResponse[]>("/api/admin/events").then((r) => setEvents(r.data));
  }, []);

  const selected = events.find((e) => e.id === eventId);
  const locked = !selected || !canAllocate(selected);

  async function load(id: string) {
    setEventId(id);
    setError(null);
    if (!id) { setRows([]); return; }
    const r = await api.get<AllocationRow[]>(`/api/admin/events/${id}/stock`);
    setRows(r.data);
  }

  async function setQty(productId: string, allocatedQty: number) {
    setError(null);
    try {
      await api.put(`/api/admin/events/${eventId}/stock/${productId}`, { allocatedQty });
      load(eventId);
    } catch (err) {
      setError(errMsg(err));
    }
  }

  return (
    <Layout title="Stock allocation">
      <div className="card">
        <h2>Stock allocation per location</h2>
        <p className="muted" style={{ marginTop: 0 }}>
          Allocate warehouse stock to a location. An event takes allocations while <strong>DRAFT</strong>;
          an outlet keeps restocking until it is CLOSED.
        </p>
        <label>Location</label>
        <select value={eventId} onChange={(e) => load(e.target.value)}>
          <option value="">Select a location…</option>
          {events.map((ev) => (
            <option key={ev.id} value={ev.id}>{ev.code} — {ev.name} ({ev.status})</option>
          ))}
        </select>
        {selected && locked && (
          <div className="muted" style={{ marginBottom: 8 }}>🔒 {selected.status} — allocations are locked.</div>
        )}
        {error && <div className="error">{error}</div>}

        {eventId && (
          <table>
            <thead>
              <tr><th>SKU</th><th>Product</th><th>Warehouse</th><th>Allocated</th><th>Sold</th><th>Remaining</th></tr>
            </thead>
            <tbody>
              {rows.map((r) => (
                <tr key={r.productId}>
                  <td><span className="badge">{r.sku}</span></td>
                  <td>{r.name}</td>
                  <td className="muted">{r.warehouseStock}</td>
                  <td>
                    <input
                      type="number"
                      min={r.soldQty}
                      max={r.warehouseStock}
                      defaultValue={r.allocatedQty}
                      disabled={locked}
                      style={{ width: 90, margin: 0 }}
                      onBlur={(e) => {
                        const v = Number(e.target.value);
                        if (v !== r.allocatedQty) setQty(r.productId, v);
                      }}
                    />
                  </td>
                  <td>{r.soldQty}</td>
                  <td style={{ fontWeight: 600 }}>{r.remaining}</td>
                </tr>
              ))}
            </tbody>
          </table>
        )}
      </div>
    </Layout>
  );
}
