import { useEffect, useState, type FormEvent, type ReactNode } from "react";
import { AxiosError } from "axios";
import Layout from "./Layout";
import { api } from "../api/client";
import type { EventResponse, EventStatus, Role, UserResponse } from "../types";
import { Icon } from "./icons";

const ROLES: Role[] = ["ADMIN", "BA", "WAREHOUSE"];

function errMsg(err: unknown): string {
  const ax = err as AxiosError<{ message?: string }>;
  return ax.response?.data?.message ?? "Request failed";
}

export default function AdminDashboard() {
  const [users, setUsers] = useState<UserResponse[]>([]);
  const [events, setEvents] = useState<EventResponse[]>([]);

  async function reload() {
    const [u, e] = await Promise.all([
      api.get<UserResponse[]>("/api/admin/users"),
      api.get<EventResponse[]>("/api/admin/events"),
    ]);
    setUsers(u.data);
    setEvents(e.data);
  }

  useEffect(() => {
    reload();
  }, []);

  const activeEvents = events.filter((e) => e.status === "ACTIVE").length;
  const baCount = users.filter((u) => u.role === "BA").length;

  return (
    <Layout title="Dashboard">
      <div className="page-head">
        <h2 style={{ marginBottom: 0 }}>Overview</h2>
        <p className="muted">Users, events, and BA assignments at a glance.</p>
      </div>

      <div className="metric-grid">
        <StatCard icon={Icon.users} label="Total users" value={users.length} />
        <StatCard icon={Icon.users} label="Beauty Advisors" value={baCount} />
        <StatCard icon={Icon.calendar} label="Events" value={events.length} />
        <StatCard icon={Icon.chart} label="Active events" value={activeEvents} />
      </div>

      <div className="grid-2">
        <CreateUserCard onDone={reload} />
        <CreateEventCard onDone={reload} />
      </div>

      <div style={{ height: 20 }} />
      <AssignBaCard users={users} events={events} onDone={reload} />

      <div style={{ height: 20 }} />
      <UsersTable users={users} onDone={reload} />
      <div style={{ height: 20 }} />
      <EventsTable events={events} onDone={reload} />
    </Layout>
  );
}

function StatCard({ icon, label, value }: { icon: ReactNode; label: string; value: number }) {
  return (
    <div className="pdx-metric">
      <div className="pdx-metric__icon">{icon}</div>
      <div>
        <div className="pdx-metric__label">{label}</div>
        <div className="pdx-metric__value">{value}</div>
      </div>
    </div>
  );
}

const STATUS_COLOR: Record<EventStatus, string> = {
  DRAFT: "#9aa3b2",
  ACTIVE: "#22c55e",
  CLOSED: "#ef4444",
};

function StatusBadge({ status }: { status: EventStatus }) {
  return (
    <span className="badge" style={{ color: STATUS_COLOR[status], borderColor: STATUS_COLOR[status] }}>
      {status}
    </span>
  );
}

function CreateUserCard({ onDone }: { onDone: () => void }) {
  const [email, setEmail] = useState("");
  const [password, setPassword] = useState("");
  const [fullName, setFullName] = useState("");
  const [role, setRole] = useState<Role>("BA");
  const [feedback, setFeedback] = useState<{ ok: boolean; text: string } | null>(null);
  const [busy, setBusy] = useState(false);

  async function submit(e: FormEvent) {
    e.preventDefault();
    setBusy(true);
    setFeedback(null);
    try {
      await api.post("/api/admin/users", { email, password, fullName, role });
      setFeedback({ ok: true, text: `Created ${email}` });
      setEmail("");
      setPassword("");
      setFullName("");
      onDone();
    } catch (err) {
      setFeedback({ ok: false, text: errMsg(err) });
    } finally {
      setBusy(false);
    }
  }

  return (
    <form className="card" onSubmit={submit}>
      <h2>Create user</h2>
      <label>Full name</label>
      <input value={fullName} onChange={(e) => setFullName(e.target.value)} required />
      <label>Email</label>
      <input type="email" value={email} onChange={(e) => setEmail(e.target.value)} required />
      <label>Password (min 8 chars)</label>
      <input type="text" value={password} onChange={(e) => setPassword(e.target.value)} required minLength={8} />
      <label>Role</label>
      <select value={role} onChange={(e) => setRole(e.target.value as Role)}>
        {ROLES.map((r) => (
          <option key={r} value={r}>{r}</option>
        ))}
      </select>
      {feedback && <div className={feedback.ok ? "ok" : "error"}>{feedback.text}</div>}
      <button type="submit" disabled={busy} style={{ marginTop: 8 }}>
        {busy ? "Creating…" : "Create user"}
      </button>
    </form>
  );
}

function CreateEventCard({ onDone }: { onDone: () => void }) {
  const [code, setCode] = useState("");
  const [name, setName] = useState("");
  const [startDate, setStartDate] = useState("");
  const [endDate, setEndDate] = useState("");
  const [feedback, setFeedback] = useState<{ ok: boolean; text: string } | null>(null);
  const [busy, setBusy] = useState(false);

  async function submit(e: FormEvent) {
    e.preventDefault();
    setBusy(true);
    setFeedback(null);
    try {
      await api.post("/api/admin/events", {
        code,
        name,
        startDate: startDate || null,
        endDate: endDate || null,
      });
      setFeedback({ ok: true, text: `Created event ${code}` });
      setCode("");
      setName("");
      setStartDate("");
      setEndDate("");
      onDone();
    } catch (err) {
      setFeedback({ ok: false, text: errMsg(err) });
    } finally {
      setBusy(false);
    }
  }

  return (
    <form className="card" onSubmit={submit}>
      <h2>Create event</h2>
      <label>Code (e.g. EV02)</label>
      <input value={code} onChange={(e) => setCode(e.target.value)} required />
      <label>Name</label>
      <input value={name} onChange={(e) => setName(e.target.value)} required />
      <label>Start date</label>
      <input type="date" value={startDate} onChange={(e) => setStartDate(e.target.value)} />
      <label>End date</label>
      <input type="date" value={endDate} onChange={(e) => setEndDate(e.target.value)} />
      {feedback && <div className={feedback.ok ? "ok" : "error"}>{feedback.text}</div>}
      <button type="submit" disabled={busy} style={{ marginTop: 8 }}>
        {busy ? "Creating…" : "Create event"}
      </button>
    </form>
  );
}

function AssignBaCard({
  users,
  events,
  onDone,
}: {
  users: UserResponse[];
  events: EventResponse[];
  onDone: () => void;
}) {
  const bas = users.filter((u) => u.role === "BA");
  const [userId, setUserId] = useState("");
  const [eventId, setEventId] = useState("");
  const [feedback, setFeedback] = useState<{ ok: boolean; text: string } | null>(null);
  const [busy, setBusy] = useState(false);

  async function submit(e: FormEvent) {
    e.preventDefault();
    setBusy(true);
    setFeedback(null);
    try {
      await api.post("/api/admin/assignments", { userId, eventId });
      const ba = bas.find((b) => b.id === userId);
      const ev = events.find((x) => x.id === eventId);
      setFeedback({ ok: true, text: `Assigned ${ba?.fullName} → ${ev?.code}` });
      onDone();
    } catch (err) {
      setFeedback({ ok: false, text: errMsg(err) });
    } finally {
      setBusy(false);
    }
  }

  return (
    <form className="card" onSubmit={submit}>
      <h2>Assign a BA to an event</h2>
      <p className="muted" style={{ marginTop: 0 }}>
        The BA will see this event when they sign in on the Android device.
      </p>
      <div className="grid-2">
        <div>
          <label>Beauty Advisor</label>
          <select value={userId} onChange={(e) => setUserId(e.target.value)} required>
            <option value="">Select a BA…</option>
            {bas.map((b) => (
              <option key={b.id} value={b.id}>{b.fullName} ({b.email})</option>
            ))}
          </select>
        </div>
        <div>
          <label>Event</label>
          <select value={eventId} onChange={(e) => setEventId(e.target.value)} required>
            <option value="">Select an event…</option>
            {events.map((ev) => (
              <option key={ev.id} value={ev.id}>{ev.code} — {ev.name}</option>
            ))}
          </select>
        </div>
      </div>
      {feedback && <div className={feedback.ok ? "ok" : "error"}>{feedback.text}</div>}
      <button type="submit" disabled={busy || !userId || !eventId}>
        {busy ? "Assigning…" : "Assign"}
      </button>
    </form>
  );
}

function UsersTable({ users, onDone }: { users: UserResponse[]; onDone: () => void }) {
  const [error, setError] = useState<string | null>(null);
  const [busyId, setBusyId] = useState<string | null>(null);

  async function patch(id: string, body: Record<string, unknown>) {
    setError(null);
    setBusyId(id);
    try {
      await api.patch(`/api/admin/users/${id}`, body);
      onDone();
    } catch (err) {
      setError(errMsg(err));
    } finally {
      setBusyId(null);
    }
  }

  return (
    <div className="card">
      <h2>Users ({users.length})</h2>
      {error && <div className="error">{error}</div>}
      <table>
        <thead>
          <tr><th>Name</th><th>Email</th><th>Role</th><th>Status</th><th>Actions</th></tr>
        </thead>
        <tbody>
          {users.map((u) => (
            <tr key={u.id} style={{ opacity: u.active ? 1 : 0.55 }}>
              <td>{u.fullName}</td>
              <td className="muted">{u.email}</td>
              <td>
                <select
                  value={u.role}
                  disabled={busyId === u.id}
                  onChange={(e) => patch(u.id, { role: e.target.value })}
                  style={{ margin: 0 }}
                >
                  {ROLES.map((r) => <option key={r} value={r}>{r}</option>)}
                </select>
              </td>
              <td>
                <span className="badge" style={{ color: u.active ? "#22c55e" : "#ef4444", borderColor: u.active ? "#22c55e" : "#ef4444" }}>
                  {u.active ? "ACTIVE" : "INACTIVE"}
                </span>
              </td>
              <td>
                <button
                  className="secondary"
                  disabled={busyId === u.id}
                  onClick={() => patch(u.id, { active: !u.active })}
                >
                  {u.active ? "Deactivate" : "Reactivate"}
                </button>
              </td>
            </tr>
          ))}
        </tbody>
      </table>
    </div>
  );
}

function EventsTable({ events, onDone }: { events: EventResponse[]; onDone: () => void }) {
  const [editing, setEditing] = useState<string | null>(null);
  const [error, setError] = useState<string | null>(null);

  async function changeStatus(ev: EventResponse, status: EventStatus) {
    setError(null);
    try {
      await api.patch(`/api/admin/events/${ev.id}/status`, { status });
      onDone();
    } catch (err) {
      setError(errMsg(err));
    }
  }

  return (
    <div className="card">
      <h2>Events ({events.length})</h2>
      <p className="muted" style={{ marginTop: 0 }}>
        Lifecycle: DRAFT → ACTIVE → CLOSED. BAs can only sell while an event is ACTIVE.
      </p>
      {error && <div className="error">{error}</div>}
      <table>
        <thead>
          <tr><th>Code</th><th>Name</th><th>Dates</th><th>Status</th><th>Actions</th></tr>
        </thead>
        <tbody>
          {events.map((ev) =>
            editing === ev.id ? (
              <EditEventRow key={ev.id} event={ev} onClose={() => setEditing(null)} onDone={onDone} />
            ) : (
              <tr key={ev.id}>
                <td><span className="badge">{ev.code}</span></td>
                <td>{ev.name}</td>
                <td className="muted">{ev.startDate ?? "—"} → {ev.endDate ?? "—"}</td>
                <td><StatusBadge status={ev.status} /></td>
                <td style={{ display: "flex", gap: 6, flexWrap: "wrap" }}>
                  {ev.status === "DRAFT" && (
                    <button onClick={() => changeStatus(ev, "ACTIVE")}>Activate</button>
                  )}
                  {ev.status === "ACTIVE" && (
                    <button onClick={() => changeStatus(ev, "CLOSED")}>Close</button>
                  )}
                  {ev.status !== "CLOSED" && (
                    <button className="secondary" onClick={() => setEditing(ev.id)}>Edit</button>
                  )}
                </td>
              </tr>
            ),
          )}
        </tbody>
      </table>
    </div>
  );
}

function EditEventRow({
  event,
  onClose,
  onDone,
}: {
  event: EventResponse;
  onClose: () => void;
  onDone: () => void;
}) {
  const [name, setName] = useState(event.name);
  const [startDate, setStartDate] = useState(event.startDate ?? "");
  const [endDate, setEndDate] = useState(event.endDate ?? "");
  const [error, setError] = useState<string | null>(null);

  async function save() {
    setError(null);
    try {
      await api.put(`/api/admin/events/${event.id}`, {
        name,
        startDate: startDate || null,
        endDate: endDate || null,
      });
      onClose();
      onDone();
    } catch (err) {
      setError(errMsg(err));
    }
  }

  return (
    <tr>
      <td><span className="badge">{event.code}</span></td>
      <td><input value={name} onChange={(e) => setName(e.target.value)} style={{ margin: 0 }} /></td>
      <td>
        <input type="date" value={startDate} onChange={(e) => setStartDate(e.target.value)} style={{ margin: 0 }} />
        <input type="date" value={endDate} onChange={(e) => setEndDate(e.target.value)} style={{ margin: "4px 0 0" }} />
      </td>
      <td><StatusBadge status={event.status} /></td>
      <td style={{ display: "flex", gap: 6 }}>
        <button onClick={save}>Save</button>
        <button className="secondary" onClick={onClose}>Cancel</button>
        {error && <div className="error">{error}</div>}
      </td>
    </tr>
  );
}
