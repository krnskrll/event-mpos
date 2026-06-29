import type { ReactNode } from "react";
import { useNavigate } from "react-router-dom";
import { useAuth } from "../auth/AuthContext";

/** Shared shell with a top bar showing the signed-in user and a logout button. */
export default function Layout({ title, children }: { title: string; children: ReactNode }) {
  const { user, logout } = useAuth();
  const navigate = useNavigate();

  async function onLogout() {
    await logout();
    navigate("/login", { replace: true });
  }

  return (
    <>
      <div className="topbar">
        <h1>{title}</h1>
        <div style={{ display: "flex", alignItems: "center", gap: 12 }}>
          <span className="muted">
            {user?.fullName} <span className="badge">{user?.role}</span>
          </span>
          <button className="secondary" onClick={onLogout}>Log out</button>
        </div>
      </div>
      <div className="container">{children}</div>
    </>
  );
}
