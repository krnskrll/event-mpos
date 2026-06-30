import { useEffect, useState, type FormEvent } from "react";
import { AxiosError } from "axios";
import Layout from "./Layout";
import { api } from "../api/client";
import type { AllocationRow, ErpSyncStatus, EventProductResponse, EventResponse, ProductResponse } from "../types";

function errMsg(err: unknown): string {
  const ax = err as AxiosError<{ message?: string }>;
  return ax.response?.data?.message ?? "Request failed";
}

export default function ProductsPage() {
  const [products, setProducts] = useState<ProductResponse[]>([]);
  const [sync, setSync] = useState<ErpSyncStatus | null>(null);

  async function reload() {
    const [p, s] = await Promise.all([
      api.get<ProductResponse[]>("/api/admin/products"),
      api.get<ErpSyncStatus>("/api/admin/erp/status"),
    ]);
    setProducts(p.data);
    setSync(s.data);
  }

  useEffect(() => {
    reload();
  }, []);

  return (
    <Layout title="Products">
      <ErpSyncCard sync={sync} onSynced={reload} />
      <div style={{ height: 20 }} />
      <AddProductCard onDone={reload} />
      <div style={{ height: 20 }} />
      <ProductsTable products={products} onDone={reload} />
      <div style={{ height: 20 }} />
      <EventAvailabilityCard />
      <div style={{ height: 20 }} />
      <StockAllocationCard />
    </Layout>
  );
}

function StockAllocationCard() {
  const [events, setEvents] = useState<EventResponse[]>([]);
  const [eventId, setEventId] = useState("");
  const [rows, setRows] = useState<AllocationRow[]>([]);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    api.get<EventResponse[]>("/api/admin/events").then((r) => setEvents(r.data));
  }, []);

  const selected = events.find((e) => e.id === eventId);
  const locked = selected ? selected.status !== "DRAFT" : true;

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
    <div className="card">
      <h2>Stock allocation per event</h2>
      <p className="muted" style={{ marginTop: 0 }}>
        Allocate warehouse stock to an event. Editable only while the event is <strong>DRAFT</strong> —
        allocations lock once it goes ACTIVE.
      </p>
      <label>Event</label>
      <select value={eventId} onChange={(e) => load(e.target.value)}>
        <option value="">Select an event…</option>
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
  );
}

function ErpSyncCard({ sync, onSynced }: { sync: ErpSyncStatus | null; onSynced: () => void }) {
  const [busy, setBusy] = useState(false);
  const [msg, setMsg] = useState<string | null>(null);

  async function syncNow() {
    setBusy(true);
    setMsg(null);
    try {
      const res = await api.post("/api/admin/erp/sync");
      const d = res.data as { created: number; updated: number; total: number; source: string };
      setMsg(`Synced from ${d.source}: ${d.created} new, ${d.updated} updated, ${d.total} total.`);
      onSynced();
    } catch (err) {
      setMsg(errMsg(err));
    } finally {
      setBusy(false);
    }
  }

  return (
    <div className="card">
      <h2>ERP sync</h2>
      <p className="muted" style={{ marginTop: 0 }}>
        Product data is pulled from the ERP (current source: <strong>{sync?.source ?? "—"}</strong>).
        Last sync:{" "}
        {sync?.lastSyncAt ? `${new Date(sync.lastSyncAt).toLocaleString()} · ${sync.lastResult} · ${sync.productCount} products` : "never"}
      </p>
      <button onClick={syncNow} disabled={busy}>{busy ? "Syncing…" : "Sync from ERP now"}</button>
      {msg && <div className="ok" style={{ marginTop: 8 }}>{msg}</div>}
    </div>
  );
}

function AddProductCard({ onDone }: { onDone: () => void }) {
  const [sku, setSku] = useState("");
  const [name, setName] = useState("");
  const [brand, setBrand] = useState("");
  const [price, setPrice] = useState("");
  const [barcodes, setBarcodes] = useState("");
  const [feedback, setFeedback] = useState<{ ok: boolean; text: string } | null>(null);
  const [busy, setBusy] = useState(false);

  async function submit(e: FormEvent) {
    e.preventDefault();
    setBusy(true);
    setFeedback(null);
    try {
      await api.post("/api/admin/products", {
        sku,
        name,
        brand: brand || null,
        price: Number(price),
        barcodes: barcodes
          .split(",")
          .map((b) => b.trim())
          .filter(Boolean)
          .map((b) => ({ barcode: b })),
      });
      setFeedback({ ok: true, text: `Created ${sku}` });
      setSku(""); setName(""); setBrand(""); setPrice(""); setBarcodes("");
      onDone();
    } catch (err) {
      setFeedback({ ok: false, text: errMsg(err) });
    } finally {
      setBusy(false);
    }
  }

  return (
    <form className="card" onSubmit={submit}>
      <h2>Add a product</h2>
      <div className="grid-2">
        <div>
          <label>SKU</label>
          <input value={sku} onChange={(e) => setSku(e.target.value)} required />
        </div>
        <div>
          <label>Name</label>
          <input value={name} onChange={(e) => setName(e.target.value)} required />
        </div>
        <div>
          <label>Brand</label>
          <input value={brand} onChange={(e) => setBrand(e.target.value)} />
        </div>
        <div>
          <label>Price</label>
          <input type="number" step="0.01" min="0" value={price} onChange={(e) => setPrice(e.target.value)} required />
        </div>
      </div>
      <label>Barcodes (comma-separated)</label>
      <input value={barcodes} onChange={(e) => setBarcodes(e.target.value)} placeholder="8990000001, 8990000002" />
      {feedback && <div className={feedback.ok ? "ok" : "error"}>{feedback.text}</div>}
      <button type="submit" disabled={busy}>{busy ? "Adding…" : "Add product"}</button>
    </form>
  );
}

function ProductsTable({ products, onDone }: { products: ProductResponse[]; onDone: () => void }) {
  const [editing, setEditing] = useState<string | null>(null);
  return (
    <div className="card">
      <h2>Catalogue ({products.length})</h2>
      <table>
        <thead>
          <tr><th>SKU</th><th>Name</th><th>Brand</th><th>Price</th><th>Source</th><th>Barcodes</th><th></th></tr>
        </thead>
        <tbody>
          {products.map((p) =>
            editing === p.id ? (
              <EditProductRow key={p.id} product={p} onClose={() => setEditing(null)} onDone={onDone} />
            ) : (
              <tr key={p.id}>
                <td><span className="badge">{p.sku}</span></td>
                <td>{p.name}</td>
                <td className="muted">{p.brand ?? "—"}</td>
                <td>{p.price.toFixed(2)}</td>
                <td><span className="badge">{p.source}</span></td>
                <td className="muted">{p.barcodes.map((b) => b.barcode).join(", ") || "—"}</td>
                <td><button className="secondary" onClick={() => setEditing(p.id)}>Edit</button></td>
              </tr>
            ),
          )}
        </tbody>
      </table>
    </div>
  );
}

function EditProductRow({ product, onClose, onDone }: { product: ProductResponse; onClose: () => void; onDone: () => void }) {
  const [name, setName] = useState(product.name);
  const [brand, setBrand] = useState(product.brand ?? "");
  const [price, setPrice] = useState(String(product.price));
  const [error, setError] = useState<string | null>(null);

  async function save() {
    setError(null);
    try {
      await api.put(`/api/admin/products/${product.id}`, { name, brand: brand || null, price: Number(price) });
      onClose();
      onDone();
    } catch (err) {
      setError(errMsg(err));
    }
  }

  return (
    <tr>
      <td><span className="badge">{product.sku}</span></td>
      <td><input value={name} onChange={(e) => setName(e.target.value)} style={{ margin: 0 }} /></td>
      <td><input value={brand} onChange={(e) => setBrand(e.target.value)} style={{ margin: 0 }} /></td>
      <td><input type="number" step="0.01" value={price} onChange={(e) => setPrice(e.target.value)} style={{ margin: 0, width: 90 }} /></td>
      <td><span className="badge">{product.source}</span></td>
      <td className="muted">{product.barcodes.map((b) => b.barcode).join(", ") || "—"}</td>
      <td style={{ display: "flex", gap: 6 }}>
        <button onClick={save}>Save</button>
        <button className="secondary" onClick={onClose}>Cancel</button>
        {error && <span className="error">{error}</span>}
      </td>
    </tr>
  );
}

function EventAvailabilityCard() {
  const [events, setEvents] = useState<EventResponse[]>([]);
  const [eventId, setEventId] = useState("");
  const [rows, setRows] = useState<EventProductResponse[]>([]);

  useEffect(() => {
    api.get<EventResponse[]>("/api/admin/events").then((r) => setEvents(r.data));
  }, []);

  async function loadRows(id: string) {
    setEventId(id);
    if (!id) { setRows([]); return; }
    const r = await api.get<EventProductResponse[]>(`/api/admin/events/${id}/products`);
    setRows(r.data);
  }

  async function toggle(productId: string, available: boolean) {
    await api.put(`/api/admin/events/${eventId}/products/${productId}`, { available });
    loadRows(eventId);
  }

  return (
    <div className="card">
      <h2>Products available per event</h2>
      <p className="muted" style={{ marginTop: 0 }}>
        Choose which products sell at an event. Only available products appear in the BA's event snapshot.
      </p>
      <label>Event</label>
      <select value={eventId} onChange={(e) => loadRows(e.target.value)}>
        <option value="">Select an event…</option>
        {events.map((ev) => (
          <option key={ev.id} value={ev.id}>{ev.code} — {ev.name} ({ev.status})</option>
        ))}
      </select>

      {eventId && (
        <table>
          <thead>
            <tr><th>SKU</th><th>Name</th><th>Price</th><th>Available</th></tr>
          </thead>
          <tbody>
            {rows.map((row) => (
              <tr key={row.product.id}>
                <td><span className="badge">{row.product.sku}</span></td>
                <td>{row.product.name}</td>
                <td>{row.product.price.toFixed(2)}</td>
                <td>
                  <input
                    type="checkbox"
                    checked={row.available}
                    style={{ width: "auto", margin: 0 }}
                    onChange={(e) => toggle(row.product.id, e.target.checked)}
                  />
                </td>
              </tr>
            ))}
          </tbody>
        </table>
      )}
    </div>
  );
}
