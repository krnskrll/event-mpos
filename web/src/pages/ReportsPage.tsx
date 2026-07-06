import { useEffect, useState } from "react";
import { AxiosError } from "axios";
import Layout from "./Layout";
import { api } from "../api/client";
import { rupiah } from "../format";
import type {
  AuditEntry,
  EndOfEventReport,
  EventResponse,
  SyncOverview,
  SyncTypeView,
} from "../types";

function errMsg(err: unknown): string {
  const ax = err as AxiosError<{ message?: string }>;
  return ax.response?.data?.message ?? "Request failed";
}

export default function ReportsPage() {
  return (
    <Layout title="Reports & Sync">
      <SyncSchedules />
      <div style={{ height: 20 }} />
      <EventReport />
    </Layout>
  );
}

// ---- Sync schedules (DB-stored, editable at runtime) ----

function SyncSchedules() {
  const [overview, setOverview] = useState<SyncOverview | null>(null);
  const [error, setError] = useState<string | null>(null);

  async function reload() {
    try {
      setOverview((await api.get<SyncOverview>("/api/admin/sync")).data);
    } catch (err) {
      setError(errMsg(err));
    }
  }
  useEffect(() => { reload(); }, []);

  return (
    <div className="card">
      <h2>Sync schedules</h2>
      <p className="muted" style={{ marginTop: 0 }}>
        Stored in the database — changes take effect on the next tick, no redeploy needed.
      </p>
      {error && <div className="error">{error}</div>}
      <div className="grid-2">
        {overview && <SyncCard cfg={overview.erp} onChange={reload} />}
        {overview && <SyncCard cfg={overview.member} onChange={reload} />}
      </div>
    </div>
  );
}

function SyncCard({ cfg, onChange }: { cfg: SyncTypeView; onChange: () => void }) {
  const [enabled, setEnabled] = useState(cfg.enabled);
  const [cron, setCron] = useState(cfg.cron);
  const [busy, setBusy] = useState(false);
  const [feedback, setFeedback] = useState<{ ok: boolean; text: string } | null>(null);

  const manualPath = cfg.type === "ERP" ? "/api/admin/erp/sync" : "/api/admin/members/sync";

  async function save() {
    setBusy(true); setFeedback(null);
    try {
      await api.put(`/api/admin/sync/${cfg.type}`, { enabled, cron });
      setFeedback({ ok: true, text: "Schedule saved" });
      onChange();
    } catch (err) {
      setFeedback({ ok: false, text: errMsg(err) });
    } finally { setBusy(false); }
  }

  async function syncNow() {
    setBusy(true); setFeedback(null);
    try {
      await api.post(manualPath);
      setFeedback({ ok: true, text: "Manual sync triggered" });
      onChange();
    } catch (err) {
      setFeedback({ ok: false, text: errMsg(err) });
    } finally { setBusy(false); }
  }

  return (
    <div className="card" style={{ background: "transparent" }}>
      <h3 style={{ marginTop: 0 }}>{cfg.type === "ERP" ? "Product / ERP sync" : "Member sync"}</h3>
      <label style={{ display: "flex", alignItems: "center", gap: 8 }}>
        <input type="checkbox" checked={enabled} onChange={(e) => setEnabled(e.target.checked)} style={{ width: "auto", margin: 0 }} />
        Scheduled sync enabled
      </label>
      <label>Cron (sec min hour day month weekday)</label>
      <input value={cron} onChange={(e) => setCron(e.target.value)} placeholder="0 0 * * * *" />
      <div style={{ display: "flex", gap: 8, marginTop: 8 }}>
        <button onClick={save} disabled={busy}>Save schedule</button>
        <button className="secondary" onClick={syncNow} disabled={busy}>Sync now</button>
      </div>
      {feedback && <div className={feedback.ok ? "ok" : "error"}>{feedback.text}</div>}
      <p className="muted" style={{ fontSize: "0.8rem", marginBottom: 0 }}>
        Last run: {cfg.lastSyncAt ? new Date(cfg.lastSyncAt).toLocaleString() : "never"}
        {cfg.lastResult ? ` · ${cfg.lastResult}` : ""} · {cfg.count} records
        {cfg.source ? ` · ${cfg.source}` : ""}
      </p>
    </div>
  );
}

// ---- End-of-event report + audit ----

function EventReport() {
  const [events, setEvents] = useState<EventResponse[]>([]);
  const [eventId, setEventId] = useState("");
  const [report, setReport] = useState<EndOfEventReport | null>(null);
  const [audit, setAudit] = useState<AuditEntry[]>([]);
  const [error, setError] = useState<string | null>(null);
  const [busy, setBusy] = useState(false);

  useEffect(() => {
    api.get<EventResponse[]>("/api/admin/events").then((r) => setEvents(r.data)).catch(() => {});
  }, []);

  async function load(id: string) {
    setEventId(id); setReport(null); setAudit([]); setError(null);
    if (!id) return;
    setBusy(true);
    try {
      const [rep, aud] = await Promise.all([
        api.get<EndOfEventReport>(`/api/admin/events/${id}/report`),
        api.get<AuditEntry[]>(`/api/admin/events/${id}/audit`),
      ]);
      setReport(rep.data);
      setAudit(aud.data);
    } catch (err) {
      setError(errMsg(err));
    } finally { setBusy(false); }
  }

  async function downloadCsv() {
    if (!report) return;
    const res = await api.get(`/api/admin/events/${eventId}/report.csv`, { responseType: "blob" });
    const url = URL.createObjectURL(res.data as Blob);
    const a = document.createElement("a");
    a.href = url;
    a.download = `report-${report.eventCode}.csv`;
    a.click();
    URL.revokeObjectURL(url);
  }

  return (
    <div className="card">
      <h2>End-of-event report</h2>
      <label>Event</label>
      <select value={eventId} onChange={(e) => load(e.target.value)}>
        <option value="">Select an event…</option>
        {events.map((ev) => (
          <option key={ev.id} value={ev.id}>{ev.code} — {ev.name} ({ev.status})</option>
        ))}
      </select>
      {error && <div className="error">{error}</div>}
      {busy && <p className="muted">Loading…</p>}

      {report && (
        <>
          <div style={{ display: "flex", gap: 24, flexWrap: "wrap", margin: "16px 0" }}>
            <Metric label="Orders" value={String(report.totalOrders)} />
            <Metric label="Items sold" value={String(report.totalItems)} />
            <Metric label="Revenue" value={rupiah(report.totalRevenue)} />
            <Metric label="Discounts" value={rupiah(report.totalDiscount)} />
            <button onClick={downloadCsv} style={{ alignSelf: "center" }}>Download CSV</button>
          </div>
          <table>
            <thead>
              <tr><th>SKU</th><th>Product</th><th>Sold</th><th>Revenue</th><th>Allocated</th><th>Remaining</th></tr>
            </thead>
            <tbody>
              {report.products.map((p) => (
                <tr key={p.productId}>
                  <td className="muted">{p.sku}</td>
                  <td>{p.name}</td>
                  <td>{p.unitsSold}</td>
                  <td>{rupiah(p.revenue)}</td>
                  <td>{p.allocated}</td>
                  <td>{p.remaining}</td>
                </tr>
              ))}
            </tbody>
          </table>

          <h3>Order audit trail ({audit.length})</h3>
          {audit.length === 0 ? (
            <p className="muted">No audited activity yet (orders created before auditing was enabled aren't shown).</p>
          ) : (
            <table>
              <thead>
                <tr><th>When</th><th>Order</th><th>Action</th><th>Change</th><th>By</th></tr>
              </thead>
              <tbody>
                {audit.map((e) => (
                  <tr key={e.id}>
                    <td className="muted">{new Date(e.at).toLocaleString()}</td>
                    <td>{e.orderReference}</td>
                    <td><span className="badge">{e.action}</span></td>
                    <td>{e.fromStatus ? `${e.fromStatus} → ${e.toStatus}` : (e.toStatus ?? "—")}</td>
                    <td>{e.actorName}</td>
                  </tr>
                ))}
              </tbody>
            </table>
          )}
        </>
      )}
    </div>
  );
}

function Metric({ label, value }: { label: string; value: string }) {
  return (
    <div>
      <div className="muted" style={{ fontSize: "0.75rem", textTransform: "uppercase" }}>{label}</div>
      <div style={{ fontSize: "1.4rem", fontWeight: 700 }}>{value}</div>
    </div>
  );
}
