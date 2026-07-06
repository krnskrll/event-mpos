import axios, { AxiosError, type InternalAxiosRequestConfig } from "axios";
import type { TokenResponse } from "../types";
import { tokenStore } from "./tokenStore";

/**
 * Axios instance that:
 *  - attaches the access token to every request
 *  - on a 401, transparently refreshes the token once and replays the request
 *  - de-duplicates concurrent refreshes so a burst of 401s triggers only one refresh
 */
export const api = axios.create({ baseURL: "/" });

// Bare client for the refresh call itself, so the interceptor below can't recurse.
const bare = axios.create({ baseURL: "/" });

let refreshPromise: Promise<string> | null = null;

function doRefresh(): Promise<string> {
  const refreshToken = tokenStore.refreshToken;
  if (!refreshToken) return Promise.reject(new Error("No refresh token"));
  if (!refreshPromise) {
    refreshPromise = bare
      .post<TokenResponse>("/api/auth/refresh", { refreshToken })
      .then((res) => {
        const t = res.data;
        tokenStore.save(t.accessToken, t.refreshToken, t.user);
        return t.accessToken;
      })
      .finally(() => {
        refreshPromise = null;
      });
  }
  return refreshPromise;
}

api.interceptors.request.use((config: InternalAxiosRequestConfig) => {
  const token = tokenStore.accessToken;
  if (token) config.headers.Authorization = `Bearer ${token}`;
  return config;
});

api.interceptors.response.use(
  (res) => res,
  async (error: AxiosError) => {
    const original = error.config as (InternalAxiosRequestConfig & { _retried?: boolean }) | undefined;
    const isAuthCall = original?.url?.includes("/api/auth/");
    if (error.response?.status === 401 && original && !original._retried && !isAuthCall) {
      original._retried = true;
      try {
        const newToken = await doRefresh();
        original.headers.Authorization = `Bearer ${newToken}`;
        return api(original);
      } catch {
        // Refresh failed — session is dead. Clear and bounce to login.
        tokenStore.clear();
        if (location.pathname !== "/login") location.assign("/login");
      }
    }
    return Promise.reject(error);
  },
);
