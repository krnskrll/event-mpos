import { Navigate } from "react-router-dom";
import type { ReactNode } from "react";
import type { Role } from "../types";
import { useAuth } from "./AuthContext";

/**
 * Gate a route on authentication and (optionally) on role. Unauthenticated users go to /login;
 * authenticated users lacking the required role are sent to their own home via /.
 */
export function ProtectedRoute({ allow, children }: { allow?: Role[]; children: ReactNode }) {
  const { user } = useAuth();
  if (!user) return <Navigate to="/login" replace />;
  if (allow && !allow.includes(user.role)) return <Navigate to="/" replace />;
  return <>{children}</>;
}

/** Maps a role to its landing route — used for post-login redirect. */
export function homePathForRole(role: Role): string {
  switch (role) {
    case "ADMIN":
      return "/admin";
    case "WAREHOUSE":
      return "/warehouse";
    case "BA":
      return "/ba";
  }
}
