import { useCallback, useEffect, useMemo, useRef, useState } from "react";
import Layout from "./Layout";
import { api } from "../api/client";
import { tokenStore } from "../api/tokenStore";
import type { EventResponse, OrderStatus, StockLevelRow, WarehouseOrderView } from "../types";

const STATUSES: OrderStatus[] = ["NEW", "PACKING", "READY", "COLLECTED", "CANCELLED"];
const NEXT: Record<OrderStatus, OrderStatus | null> = {
  NEW: "PACKING",
  PACKING: "READY",
  READY: "COLLECTED",
  COLLECTED: null,
  CANCELLED: null,
};
const STATUS_COLOR: Record<OrderStatus, string> = {
  NEW: "#e0457b",
  PACKING: "#e0b341",
  READY: "#22c55e",
  COLLECTED: "#6b7280",
  CANCELLED: "#9ca3af",
};

export default function WarehousePage() {
  const [events, setEvents] = useState<EventResponse[]>([]);
  const [eventId, setEventId] = useState("");
  const [tab, setTab] = useState<"orders" | "stock">("orders");

  useEffect(() => {
    api.get<EventResponse[]>("/api/warehouse/events").then((r) => {
      setEvents(r.data);
      // default to the first ACTIVE event
      const active = r.data.find((e) => e.status === "ACTIVE") ?? r.data[0];
      if (active) setEventId(active.id);
    });
  }, []);

  return (
    <Layout title="Warehouse">
      <div className="card">
        <div style={{ display: "flex", gap: 16, alignItems: "center", flexWrap: "wrap" }}>
          <div style={{ flex: 1, minWidth: 240 }}>
            <label>Event</label>
            <select value={eventId} onChange={(e) => setEventId(e.target.value)}>
              <option value="">Select an event…</option>
              {events.map((ev) => (
                <option key={ev.id} value={ev.id}>{ev.code} — {ev.name} ({ev.status})</option>
              ))}
            </select>
          </div>
          <div style={{ display: "flex", gap: 8 }}>
            <button className={tab === "orders" ? "" : "secondary"} onClick={() => setTab("orders")}>Orders</button>
            <button className={tab === "stock" ? "" : "secondary"} onClick={() => setTab("stock")}>Stock</button>
          </div>
        </div>
      </div>

      {eventId && tab === "orders" && <OrderBoard eventId={eventId} />}
      {eventId && tab === "stock" && <StockPanel eventId={eventId} />}
    </Layout>
  );
}

/** Plays a short beep via WebAudio (no asset needed) — alerts staff who aren't watching the screen. */
function beep() {
  try {
    const ctx = new (window.AudioContext || (window as unknown as { webkitAudioContext: typeof AudioContext }).webkitAudioContext)();
    const osc = ctx.createOscillator();
    const gain = ctx.createGain();
    osc.connect(gain);
    gain.connect(ctx.destination);
    osc.frequency.value = 880;
    gain.gain.setValueAtTime(0.15, ctx.currentTime);
    osc.start();
    osc.stop(ctx.currentTime + 0.25);
  } catch {
    /* audio not available */
  }
}

function OrderBoard({ eventId }: { eventId: string }) {
  const [orders, setOrders] = useState<Record<string, WarehouseOrderView>>({});
  const [connected, setConnected] = useState(false);
  const [filter, setFilter] = useState<OrderStatus | "ALL">("ALL");
  const [flash, setFlash] = useState(false);
  const seen = useRef<Set<string>>(new Set());

  const upsert = useCallback((o: WarehouseOrderView, alertOnNew: boolean) => {
    setOrders((prev) => ({ ...prev, [o.id]: o }));
    if (alertOnNew && !seen.current.has(o.id) && o.status === "NEW") {
      beep();
      setFlash(true);
      setTimeout(() => setFlash(false), 1200);
    }
    seen.current.add(o.id);
  }, []);

  // Initial load + polling fallback (every 8s; harmless even when the socket is live).
  useEffect(() => {
    let active = true;
    const load = async (alert: boolean) => {
      const r = await api.get<WarehouseOrderView[]>(`/api/warehouse/events/${eventId}/orders`);
      if (!active) return;
      r.data.forEach((o) => upsert(o, alert));
    };
    setOrders({});
    seen.current = new Set();
    load(false);
    const poll = window.setInterval(() => {
      if (!connected) load(true);
    }, 8000);
    return () => {
      active = false;
      window.clearInterval(poll);
    };
  }, [eventId, connected, upsert]);

  // Real-time order push over a native WebSocket. The JWT goes in the ?token= query param since a
  // browser can't set an Authorization header on a WebSocket handshake. Auto-reconnects on drop.
  useEffect(() => {
    let ws: WebSocket | null = null;
    let retry: number | undefined;
    let closed = false;

    const connect = () => {
      if (closed) return;
      const proto = location.protocol === "https:" ? "wss" : "ws";
      const token = encodeURIComponent(tokenStore.accessToken ?? "");
      ws = new WebSocket(`${proto}://${location.host}/ws/events/${eventId}/orders?token=${token}`);
      ws.onopen = () => setConnected(true);
      ws.onmessage = (e) => upsert(JSON.parse(e.data) as WarehouseOrderView, true);
      ws.onerror = () => ws?.close();
      ws.onclose = () => {
        setConnected(false);
        if (!closed) retry = window.setTimeout(connect, 4000);
      };
    };
    connect();

    return () => {
      closed = true;
      if (retry) window.clearTimeout(retry);
      ws?.close();
      setConnected(false);
    };
  }, [eventId, upsert]);

  // Cancelling returns the stock (the Stock tab picks it up on its next poll). The prompt doubles as
  // the confirm step: Cancel aborts, OK proceeds with an optional reason.
  async function cancel(o: WarehouseOrderView) {
    const reason = window.prompt(`Cancel order ${o.orderNumberDisplay} for ${o.customerName}? Its stock goes back on sale.\n\nReason (optional):`, "");
    if (reason === null) return;
    try {
      upsert((await api.post<WarehouseOrderView>(`/api/warehouse/orders/${o.id}/cancel`, { reason: reason || null })).data, false);
    } catch (err) {
      window.alert((err as { response?: { data?: { message?: string } } }).response?.data?.message ?? "Cancel failed");
    }
  }

  async function advance(o: WarehouseOrderView) {
    const next = NEXT[o.status];
    if (!next) return;
    const r = await api.patch<WarehouseOrderView>(`/api/warehouse/orders/${o.id}/status`, { status: next });
    upsert(r.data, false);
  }

  const list = useMemo(
    () =>
      Object.values(orders)
        .filter((o) => filter === "ALL" || o.status === filter)
        .sort((a, b) => b.orderNumber - a.orderNumber),
    [orders, filter],
  );

  return (
    <>
      <div className="card" style={{ borderColor: flash ? "#e0457b" : undefined }}>
        <div style={{ display: "flex", justifyContent: "space-between", alignItems: "center", flexWrap: "wrap", gap: 8 }}>
          <div style={{ display: "flex", gap: 6, flexWrap: "wrap" }}>
            {(["ALL", ...STATUSES] as const).map((s) => (
              <button key={s} className={filter === s ? "" : "secondary"} onClick={() => setFilter(s)}>
                {s}
              </button>
            ))}
          </div>
          <span className="muted" style={{ fontSize: "0.85rem" }}>
            {connected ? "🟢 live" : "🟡 polling (reconnecting…)"} · {list.length} orders
          </span>
        </div>
      </div>

      <div className="grid-2">
        {list.map((o) => (
          <OrderCard key={o.id} order={o} onAdvance={() => advance(o)} onCancel={() => cancel(o)} />
        ))}
      </div>
      {list.length === 0 && <div className="card muted">No orders{filter !== "ALL" ? ` in ${filter}` : ""} yet.</div>}
    </>
  );
}

function OrderCard({ order, onAdvance, onCancel }: { order: WarehouseOrderView; onAdvance: () => void; onCancel: () => void }) {
  const next = NEXT[order.status];
  return (
    <div className="card" style={{ borderLeft: `6px solid ${STATUS_COLOR[order.status]}` }}>
      <div style={{ display: "flex", justifyContent: "space-between", alignItems: "baseline" }}>
        <h2 style={{ fontSize: "1.6rem", margin: 0 }}>{order.orderNumberDisplay}</h2>
        <span className="badge" style={{ color: STATUS_COLOR[order.status], borderColor: STATUS_COLOR[order.status], fontSize: "0.9rem" }}>
          {order.status}
        </span>
      </div>
      <div style={{ fontSize: "1.1rem", fontWeight: 600, marginTop: 4 }}>{order.customerName}</div>
      <div className="muted" style={{ fontSize: "0.8rem" }}>
        BA: {order.baName} · {timeSince(order.createdAt)}
      </div>
      <ul style={{ margin: "10px 0", paddingLeft: 18 }}>
        {order.items.map((it, i) => (
          <li key={i} style={{ fontSize: "1.05rem" }}>
            <strong>{it.qty}×</strong> {it.name}
          </li>
        ))}
      </ul>
      {next ? (
        <button onClick={onAdvance} style={{ width: "100%", fontSize: "1rem" }}>
          Mark {next}
        </button>
      ) : (
        <div className="muted" style={{ textAlign: "center" }}>{order.status === "CANCELLED" ? "✕ Cancelled" : "✓ Collected"}</div>
      )}
      {(order.status === "NEW" || order.status === "PACKING") && (
        <button className="secondary" onClick={onCancel} style={{ width: "100%", marginTop: 8 }}>Cancel order</button>
      )}
    </div>
  );
}

function timeSince(iso: string): string {
  const secs = Math.max(0, Math.floor((Date.now() - new Date(iso).getTime()) / 1000));
  if (secs < 60) return `${secs}s ago`;
  const mins = Math.floor(secs / 60);
  if (mins < 60) return `${mins}m ago`;
  return `${Math.floor(mins / 60)}h ${mins % 60}m ago`;
}

/** Remaining stock per product for the event (polled) — visible on the warehouse dashboard. */
function StockPanel({ eventId }: { eventId: string }) {
  const [rows, setRows] = useState<StockLevelRow[]>([]);
  useEffect(() => {
    const load = () => api.get<StockLevelRow[]>(`/api/warehouse/events/${eventId}/stock`).then((r) => setRows(r.data));
    load();
    const t = window.setInterval(load, 5000);
    return () => window.clearInterval(t);
  }, [eventId]);
  return (
    <div className="card">
      <h2>Remaining stock</h2>
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
                {r.remaining}{r.remaining === 0 ? " · SOLD OUT" : r.lowStock ? " · LOW" : ""}
              </td>
            </tr>
          ))}
          {rows.length === 0 && <tr><td colSpan={5} className="muted">No stock allocated.</td></tr>}
        </tbody>
      </table>
    </div>
  );
}
