import { useEffect, useState, type FormEvent } from "react";
import { AxiosError } from "axios";
import Layout from "./Layout";
import { api } from "../api/client";
import type { EventResponse, ProductResponse, PromotionResponse, PromotionScope, PromotionType } from "../types";

function errMsg(err: unknown): string {
  const ax = err as AxiosError<{ message?: string }>;
  return ax.response?.data?.message ?? "Request failed";
}

export default function PromotionsPage() {
  const [events, setEvents] = useState<EventResponse[]>([]);
  const [products, setProducts] = useState<ProductResponse[]>([]);
  const [eventId, setEventId] = useState("");
  const [promos, setPromos] = useState<PromotionResponse[]>([]);

  useEffect(() => {
    api.get<EventResponse[]>("/api/admin/events").then((r) => setEvents(r.data));
    api.get<ProductResponse[]>("/api/admin/products").then((r) => setProducts(r.data));
  }, []);

  async function loadPromos(id: string) {
    setEventId(id);
    if (!id) { setPromos([]); return; }
    const r = await api.get<PromotionResponse[]>(`/api/admin/events/${id}/promotions`);
    setPromos(r.data);
  }

  return (
    <Layout title="Promotions">
      <div className="card">
        <h2>Event promotions</h2>
        <p className="muted" style={{ marginTop: 0 }}>
          Promotions are scoped to an event and <strong>auto-apply</strong> to the BA's cart. Phase 1:
          one promotion per order — the server applies whichever gives the biggest discount.
        </p>
        <label>Event</label>
        <select value={eventId} onChange={(e) => loadPromos(e.target.value)}>
          <option value="">Select an event…</option>
          {events.map((ev) => (
            <option key={ev.id} value={ev.id}>{ev.code} — {ev.name} ({ev.status})</option>
          ))}
        </select>
      </div>

      {eventId && (
        <>
          <div style={{ height: 20 }} />
          <CreatePromotion eventId={eventId} products={products} onDone={() => loadPromos(eventId)} />
          <div style={{ height: 20 }} />
          <PromoTable promos={promos} products={products} onDone={() => loadPromos(eventId)} />
        </>
      )}
    </Layout>
  );
}

function CreatePromotion({ eventId, products, onDone }: { eventId: string; products: ProductResponse[]; onDone: () => void }) {
  const [name, setName] = useState("");
  const [type, setType] = useState<PromotionType>("PERCENTAGE");
  const [value, setValue] = useState("");
  const [scope, setScope] = useState<PromotionScope>("ALL");
  const [productIds, setProductIds] = useState<string[]>([]);
  const [minQty, setMinQty] = useState("");
  const [minSubtotal, setMinSubtotal] = useState("");
  const [feedback, setFeedback] = useState<{ ok: boolean; text: string } | null>(null);

  async function submit(e: FormEvent) {
    e.preventDefault();
    setFeedback(null);
    const body: Record<string, unknown> = {
      name,
      type,
      scope,
      productIds: scope === "SPECIFIC" || type === "BUNDLE" ? productIds : [],
      minQty: minQty ? Number(minQty) : null,
      minSubtotal: minSubtotal ? Number(minSubtotal) : null,
    };
    if (type === "PERCENTAGE") body.percentOff = Number(value);
    else body.amountOff = Number(value);
    try {
      await api.post(`/api/admin/events/${eventId}/promotions`, body);
      setFeedback({ ok: true, text: `Created "${name}"` });
      setName(""); setValue(""); setMinQty(""); setMinSubtotal(""); setProductIds([]);
      onDone();
    } catch (err) {
      setFeedback({ ok: false, text: errMsg(err) });
    }
  }

  const needsProducts = scope === "SPECIFIC" || type === "BUNDLE";

  return (
    <form className="card" onSubmit={submit}>
      <h2>Add a promotion</h2>
      <div className="grid-2">
        <div>
          <label>Name</label>
          <input value={name} onChange={(e) => setName(e.target.value)} required />
        </div>
        <div>
          <label>Type</label>
          <select value={type} onChange={(e) => setType(e.target.value as PromotionType)}>
            <option value="PERCENTAGE">Percentage off</option>
            <option value="FIXED">Fixed amount off</option>
            <option value="BUNDLE">Bundle (buy N, amount off)</option>
          </select>
        </div>
        <div>
          <label>{type === "PERCENTAGE" ? "Percent (e.g. 10)" : "Amount off (RM)"}</label>
          <input type="number" step="0.01" min="0" value={value} onChange={(e) => setValue(e.target.value)} required />
        </div>
        <div>
          <label>Scope</label>
          <select value={scope} onChange={(e) => setScope(e.target.value as PromotionScope)} disabled={type === "BUNDLE"}>
            <option value="ALL">All products</option>
            <option value="SPECIFIC">Specific products</option>
          </select>
        </div>
        <div>
          <label>{type === "BUNDLE" ? "Bundle size (min qty)" : "Min quantity (optional)"}</label>
          <input type="number" min="0" value={minQty} onChange={(e) => setMinQty(e.target.value)} required={type === "BUNDLE"} />
        </div>
        <div>
          <label>Min subtotal Rp (optional)</label>
          <input type="number" step="0.01" min="0" value={minSubtotal} onChange={(e) => setMinSubtotal(e.target.value)} />
        </div>
      </div>
      {needsProducts && (
        <>
          <label>Products ({type === "BUNDLE" ? "the bundle items" : "applies to these"})</label>
          <select
            multiple
            value={productIds}
            onChange={(e) => setProductIds(Array.from(e.target.selectedOptions, (o) => o.value))}
            style={{ height: 120 }}
          >
            {products.map((p) => (
              <option key={p.id} value={p.id}>{p.sku} — {p.name}</option>
            ))}
          </select>
        </>
      )}
      {feedback && <div className={feedback.ok ? "ok" : "error"}>{feedback.text}</div>}
      <button type="submit">Add promotion</button>
    </form>
  );
}

function PromoTable({ promos, products, onDone }: { promos: PromotionResponse[]; products: ProductResponse[]; onDone: () => void }) {
  const skuOf = (id: string) => products.find((p) => p.id === id)?.sku ?? id.slice(0, 6);

  function describe(p: PromotionResponse): string {
    const base = p.type === "PERCENTAGE" ? `${p.percentOff}% off` : `RM${p.amountOff} off`;
    const where = p.scope === "SPECIFIC" || p.type === "BUNDLE" ? ` [${p.productIds.map(skuOf).join(", ")}]` : " (all)";
    const cond = [p.minQty ? `min qty ${p.minQty}` : "", p.minSubtotal ? `min RM${p.minSubtotal}` : ""].filter(Boolean).join(", ");
    return `${base}${where}${cond ? ` · ${cond}` : ""}`;
  }

  async function toggle(p: PromotionResponse) {
    await api.put(`/api/admin/promotions/${p.id}`, { ...p, active: !p.active });
    onDone();
  }
  async function remove(id: string) {
    await api.delete(`/api/admin/promotions/${id}`);
    onDone();
  }

  return (
    <div className="card">
      <h2>Promotions ({promos.length})</h2>
      <table>
        <thead>
          <tr><th>Name</th><th>Type</th><th>Rule</th><th>Active</th><th></th></tr>
        </thead>
        <tbody>
          {promos.map((p) => (
            <tr key={p.id}>
              <td>{p.name}</td>
              <td><span className="badge">{p.type}</span></td>
              <td className="muted">{describe(p)}</td>
              <td>{p.active ? "✓" : "—"}</td>
              <td style={{ display: "flex", gap: 6 }}>
                <button className="secondary" onClick={() => toggle(p)}>{p.active ? "Disable" : "Enable"}</button>
                <button className="secondary" onClick={() => remove(p.id)}>Delete</button>
              </td>
            </tr>
          ))}
        </tbody>
      </table>
    </div>
  );
}
