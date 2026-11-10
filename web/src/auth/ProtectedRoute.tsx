import { Navigate } from "react-router-dom";
import type { ReactNode } from "react";
import { useAuth } from "./AuthContext";

/** Only signed-in admins get past; everyone else goes to /login. */
export function ProtectedRoute({ children }: { children: ReactNode }) {
  const { user } = useAuth();
  if (user?.role !== "ADMIN") return <Navigate to="/login" replace />;
  return <>{children}</>;
}
