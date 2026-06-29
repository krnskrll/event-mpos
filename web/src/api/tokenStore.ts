import type { UserSummary } from "../types";

/**
 * Persists tokens + the current user across reloads. localStorage is acceptable for the
 * internal warehouse/admin dashboard; the short access-token TTL plus refresh rotation on the
 * server limits exposure. (The Android app uses EncryptedSharedPreferences — see android/.)
 */
const ACCESS = "empos.accessToken";
const REFRESH = "empos.refreshToken";
const USER = "empos.user";

export const tokenStore = {
  get accessToken(): string | null {
    return localStorage.getItem(ACCESS);
  },
  get refreshToken(): string | null {
    return localStorage.getItem(REFRESH);
  },
  get user(): UserSummary | null {
    const raw = localStorage.getItem(USER);
    return raw ? (JSON.parse(raw) as UserSummary) : null;
  },
  save(accessToken: string, refreshToken: string, user: UserSummary) {
    localStorage.setItem(ACCESS, accessToken);
    localStorage.setItem(REFRESH, refreshToken);
    localStorage.setItem(USER, JSON.stringify(user));
  },
  clear() {
    localStorage.removeItem(ACCESS);
    localStorage.removeItem(REFRESH);
    localStorage.removeItem(USER);
  },
};
