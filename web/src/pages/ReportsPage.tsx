import { useEffect, useState } from "react";
import { AxiosError } from "axios";
import Layout from "./Layout";
import { api } from "../api/client";
import { rupiah } from "../format";
import { AUDIT_ACTIONS } from "../types";
import type {
  AuditPage,
  EndOfEventReport,
  EventResponse,
  SyncOverview,
  SyncTypeView,
  UserResponse,
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

const PAGE_SIZES = [50, 100, 200];

interface AuditFilters {
  actor: string;
  action: string;
  page: number;
  size: number;
}

const NO_FILTERS: AuditFilters = { actor: "", action: "", page: 0, size: 50 };

function EventReport() {
  const [events, setEvents] = useState<EventResponse[]>([]);
  const [users, setUsers] = useState<UserResponse[]>([]);
  const [eventId, setEventId] = useState("");
  const [report, setReport] = useState<EndOfEventReport | null>(null);
  // Why the final report isn't available (e.g. the event is still ACTIVE) — not an error.
  const [reportNote, setReportNote] = useState<string | null>(null);
  const [error, setError] = useState<string | null>(null);
  const [busy, setBusy] = useState(false);

  useEffect(() => {
    api.get<EventResponse[]>("/api/admin/events").then((r) => setEvents(r.data)).catch(() => {});
    api.get<UserResponse[]>("/api/admin/users").then((r) => setUsers(r.data)).catch(() => {});
  }, []);

  async function load(id: string) {
    setEventId(id); setReport(null); setReportNote(null); setError(null);
    if (!id) return;
    setBusy(true);
    try {
      setReport((await api.get<EndOfEventReport>(`/api/admin/events/${id}/report`)).data);
    } catch (err) {
      if ((err as AxiosError).response?.status === 409) setReportNote(errMsg(err));
      else setError(errMsg(err));
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
      {reportNote && (
        <div className="notice" role="status">
          <strong>Final report not available yet.</strong> {reportNote} Sales can still change until the
          event is closed, so totals would not be final. The audit trail below is live.
        </div>
      )}

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
        </>
      )}

      {eventId && <AuditTrail key={eventId} eventId={eventId} users={users} />}
    </div>
  );
}

/** The event's audit log, one server-side page at a time, filterable by actor and action. */
function AuditTrail({ eventId, users }: { eventId: string; users: UserResponse[] }) {
  const [filters, setFilters] = useState<AuditFilters>(NO_FILTERS);
  const [page, setPage] = useState<AuditPage | null>(null);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    const params = new URLSearchParams({ page: String(filters.page), size: String(filters.size) });
    if (filters.actor) params.set("actor", filters.actor);
    if (filters.action) params.set("action", filters.action);
    let cancelled = false;
    setError(null);
    api.get<AuditPage>(`/api/admin/events/${eventId}/audit?${params}`)
      .then((r) => { if (!cancelled) setPage(r.data); })
      .catch((err) => { if (!cancelled) setError(errMsg(err)); });
    return () => { cancelled = true; };
  }, [eventId, filters]);

  // Changing a filter or the page size starts again from the first page.
  const refine = (patch: Partial<AuditFilters>) => setFilters((f) => ({ ...f, ...patch, page: 0 }));
  const pageCount = page ? Math.max(1, Math.ceil(page.total / page.size)) : 1;
  const from = page && page.total > 0 ? page.page * page.size + 1 : 0;
  const to = page ? page.page * page.size + page.items.length : 0;

  return (
    <>
      <h3>Order audit trail{page ? ` (${page.total})` : ""}</h3>
      <div className="pager">
        <select aria-label="Filter by user" value={filters.actor} onChange={(e) => refine({ actor: e.target.value })}>
          <option value="">All users</option>
          {users.map((u) => <option key={u.id} value={u.id}>{u.fullName} ({u.role})</option>)}
        </select>
        <select aria-label="Filter by action" value={filters.action} onChange={(e) => refine({ action: e.target.value })}>
          <option value="">All actions</option>
          {AUDIT_ACTIONS.map((a) => <option key={a} value={a}>{a}</option>)}
        </select>
        <select aria-label="Page size" value={filters.size} onChange={(e) => refine({ size: Number(e.target.value) })}>
          {PAGE_SIZES.map((n) => <option key={n} value={n}>{n} per page</option>)}
        </select>
      </div>
      {error && <div className="error">{error}</div>}

      {page && page.total === 0 ? (
        <p className="muted">No audited activity matches.</p>
      ) : page && (
        <>
          <table>
            <thead>
              <tr><th>When</th><th>Order</th><th>Action</th><th>Change</th><th>By</th></tr>
            </thead>
            <tbody>
              {page.items.map((e) => (
                <tr key={e.id}>
                  <td className="muted">{new Date(e.at).toLocaleString()}</td>
                  <td>{e.orderReference}</td>
                  <td><span className="badge">{e.action}</span></td>
                  <td>{e.fromStatus ? `${e.fromStatus} → ${e.toStatus}` : (e.detail ?? e.toStatus ?? "—")}</td>
                  <td>{e.actorName}</td>
                </tr>
              ))}
            </tbody>
          </table>
          <div className="pager">
            <button className="secondary" disabled={page.page === 0}
              onClick={() => setFilters((f) => ({ ...f, page: f.page - 1 }))}>← Newer</button>
            <span className="muted">
              {from}–{to} of {page.total} · page {page.page + 1} of {pageCount}
            </span>
            <button className="secondary" disabled={page.page + 1 >= pageCount}
              onClick={() => setFilters((f) => ({ ...f, page: f.page + 1 }))}>Older →</button>
          </div>
        </>
      )}
    </>
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
