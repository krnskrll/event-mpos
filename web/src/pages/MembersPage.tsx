import { useEffect, useState } from "react";
import { AxiosError } from "axios";
import Layout from "./Layout";
import { api } from "../api/client";
import type { MemberResponse, MemberSyncStatus } from "../types";

function errMsg(err: unknown): string {
  const ax = err as AxiosError<{ message?: string }>;
  return ax.response?.data?.message ?? "Request failed";
}

const TIER_COLOR: Record<string, string> = {
  PLATINUM: "#c0c0e0",
  GOLD: "#e0b341",
  SILVER: "#9aa3b2",
  STANDARD: "#6b7280",
};

export default function MembersPage() {
  const [members, setMembers] = useState<MemberResponse[]>([]);
  const [status, setStatus] = useState<MemberSyncStatus | null>(null);
  const [busy, setBusy] = useState(false);
  const [msg, setMsg] = useState<string | null>(null);
  const [query, setQuery] = useState("");

  async function reload() {
    const [m, s] = await Promise.all([
      api.get<MemberResponse[]>("/api/admin/members"),
      api.get<MemberSyncStatus>("/api/admin/members/sync/status"),
    ]);
    setMembers(m.data);
    setStatus(s.data);
  }

  useEffect(() => {
    reload();
  }, []);

  async function syncNow() {
    setBusy(true);
    setMsg(null);
    try {
      const res = await api.post("/api/admin/members/sync");
      const d = res.data as { created: number; updated: number; total: number; source: string };
      setMsg(`Synced from ${d.source}: ${d.created} new, ${d.updated} updated, ${d.total} total.`);
      reload();
    } catch (err) {
      setMsg(errMsg(err));
    } finally {
      setBusy(false);
    }
  }

  const filtered = members.filter(
    (m) =>
      !query ||
      m.name.toLowerCase().includes(query.toLowerCase()) ||
      m.qrCode.toLowerCase().includes(query.toLowerCase()),
  );

  return (
    <Layout title="Members">
      <div className="card">
        <h2>Member sync</h2>
        <p className="muted" style={{ marginTop: 0 }}>
          Customer data is pulled from the external member system (source:{" "}
          <strong>{status?.source ?? "—"}</strong>). The Android app caches it for instant offline QR
          lookup. Last sync:{" "}
          {status?.lastSyncAt
            ? `${new Date(status.lastSyncAt).toLocaleString()} · ${status.lastResult} · ${status.memberCount} members`
            : "never"}
        </p>
        <button onClick={syncNow} disabled={busy}>{busy ? "Syncing…" : "Sync members now"}</button>
        {msg && <div className="ok" style={{ marginTop: 8 }}>{msg}</div>}
      </div>

      <div style={{ height: 20 }} />
      <div className="card">
        <h2>Directory ({members.length})</h2>
        <input
          placeholder="Search by name or QR…"
          value={query}
          onChange={(e) => setQuery(e.target.value)}
        />
        <table>
          <thead>
            <tr><th>Member code</th><th>QR</th><th>Name</th><th>Tier</th><th>Region</th><th>Active</th></tr>
          </thead>
          <tbody>
            {filtered.map((m) => (
              <tr key={m.id}>
                <td className="muted">{m.memberCode}</td>
                <td><span className="badge">{m.qrCode}</span></td>
                <td>{m.name}</td>
                <td><span className="badge" style={{ color: TIER_COLOR[m.tier], borderColor: TIER_COLOR[m.tier] }}>{m.tier}</span></td>
                <td className="muted">{m.region ?? "—"}</td>
                <td>{m.active ? "✓" : "—"}</td>
              </tr>
            ))}
          </tbody>
        </table>
      </div>
    </Layout>
  );
}
