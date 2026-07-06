import { useState, type FormEvent } from "react";
import { useNavigate } from "react-router-dom";
import { AxiosError } from "axios";
import { useAuth } from "../auth/AuthContext";
import { homePathForRole } from "../auth/ProtectedRoute";

export default function LoginPage() {
  const { login } = useAuth();
  const navigate = useNavigate();
  const [email, setEmail] = useState("");
  const [password, setPassword] = useState("");
  const [error, setError] = useState<string | null>(null);
  const [busy, setBusy] = useState(false);

  async function onSubmit(e: FormEvent) {
    e.preventDefault();
    setError(null);
    setBusy(true);
    try {
      const user = await login(email, password);
      navigate(homePathForRole(user.role), { replace: true });
    } catch (err) {
      const ax = err as AxiosError<{ message?: string }>;
      setError(ax.response?.data?.message ?? "Login failed. Please try again.");
    } finally {
      setBusy(false);
    }
  }

  return (
    <div className="center">
      <form className="card login-card" onSubmit={onSubmit}>
        <div style={{ display: "flex", alignItems: "center", gap: 12, marginBottom: 20 }}>
          <span className="app-sidebar__logo">e</span>
          <div>
            <h1>event-mpos</h1>
            <p className="muted" style={{ margin: 0, fontSize: "var(--pdx-font-sm)" }}>Warehouse &amp; Admin dashboard</p>
          </div>
        </div>

        <label htmlFor="email">Email</label>
        <input
          id="email"
          type="email"
          autoComplete="username"
          value={email}
          onChange={(e) => setEmail(e.target.value)}
          required
        />

        <label htmlFor="password">Password</label>
        <input
          id="password"
          type="password"
          autoComplete="current-password"
          value={password}
          onChange={(e) => setPassword(e.target.value)}
          required
        />

        {error && <div className="error">{error}</div>}

        <button type="submit" disabled={busy} style={{ width: "100%", marginTop: 8 }}>
          {busy ? "Signing in…" : "Sign in"}
        </button>
      </form>
    </div>
  );
}
