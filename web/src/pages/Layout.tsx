import type { ReactNode } from "react";
import { Link, useLocation, useNavigate } from "react-router-dom";
import { useAuth } from "../auth/AuthContext";
import { Icon } from "./icons";

// Staff, devices, members and everything else live in Bhoomy; stock stays here until Phase 6.
const NAV = [{ to: "/stock", label: "Stock", icon: Icon.box }];

function initials(name: string): string {
  return name.split(" ").map((n) => n[0]).filter(Boolean).join("").toUpperCase().slice(0, 2) || "?";
}

/** Sidebar dashboard shell: brand + nav on the left, top bar + content on the right. */
export default function Layout({ title, children }: { title: string; children: ReactNode }) {
  const { user, logout } = useAuth();
  const navigate = useNavigate();
  const location = useLocation();

  async function onLogout() {
    await logout();
    navigate("/login", { replace: true });
  }


  return (
    <div className="app-shell">
      <aside className="app-sidebar">
        <div className="app-sidebar__brand">
          <span className="app-sidebar__logo">e</span>
          <span>event-mpos</span>
        </div>
        <div className="app-sidebar__section-label">Administration</div>
        <nav>
          {NAV.map((l) => (
            <Link key={l.to} to={l.to} className={`pdx-nav-item${location.pathname.startsWith(l.to) ? " pdx-nav-item--active" : ""}`}>
              <span className="pdx-nav-item__icon">{l.icon}</span>
              <span className="pdx-nav-item__label">{l.label}</span>
            </Link>
          ))}
        </nav>
        <div className="app-sidebar__footer">
          <button className="pdx-nav-item" style={{ width: "100%", background: "transparent" }} onClick={onLogout}>
            <span className="pdx-nav-item__icon">{Icon.logout}</span>
            <span className="pdx-nav-item__label">Log out</span>
          </button>
        </div>
      </aside>

      <div className="app-main">
        <header className="app-topbar">
          <h1>{title}</h1>
          <div className="app-topbar__user">
            <div style={{ textAlign: "right", lineHeight: 1.2 }}>
              <div style={{ fontWeight: 700, fontSize: "var(--pdx-font-sm)" }}>{user?.fullName}</div>
              <div className="muted" style={{ fontSize: "var(--pdx-font-xs)" }}>{user?.role}</div>
            </div>
            <span className="pdx-avatar">{initials(user?.fullName ?? "")}</span>
          </div>
        </header>
        <main className="app-content">{children}</main>
      </div>
    </div>
  );
}
