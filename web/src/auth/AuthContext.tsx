import { createContext, useContext, useState, type ReactNode } from "react";
import type { TokenResponse, UserSummary } from "../types";
import { api } from "../api/client";
import { tokenStore } from "../api/tokenStore";

interface AuthState {
  user: UserSummary | null;
  login: (email: string, password: string) => Promise<UserSummary>;
  logout: () => Promise<void>;
}

const AuthContext = createContext<AuthState | undefined>(undefined);

export function AuthProvider({ children }: { children: ReactNode }) {
  const [user, setUser] = useState<UserSummary | null>(tokenStore.user);

  async function login(email: string, password: string): Promise<UserSummary> {
    const res = await api.post<TokenResponse>("/api/auth/login", { email, password });
    const t = res.data;
    tokenStore.save(t.accessToken, t.refreshToken, t.user);
    setUser(t.user);
    return t.user;
  }

  async function logout(): Promise<void> {
    try {
      await api.post("/api/auth/logout");
    } catch {
      // best-effort; clear locally regardless
    }
    tokenStore.clear();
    setUser(null);
  }

  return <AuthContext.Provider value={{ user, login, logout }}>{children}</AuthContext.Provider>;
}

export function useAuth(): AuthState {
  const ctx = useContext(AuthContext);
  if (!ctx) throw new Error("useAuth must be used within AuthProvider");
  return ctx;
}
