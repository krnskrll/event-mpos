import { defineConfig } from "vite";
import react from "@vitejs/plugin-react";

// Backend origin for the dev proxy. Defaults to localhost:8080; override with VITE_PROXY_TARGET
// (e.g. http://host.docker.internal:8080 when the dev server runs in a container).
const proxyTarget = process.env.VITE_PROXY_TARGET ?? "http://localhost:8080";

export default defineConfig({
  plugins: [react()],
  // sockjs-client (STOMP/WebSocket for the warehouse board) references Node's `global`, which
  // doesn't exist in the browser. Alias it to globalThis so the bundle runs client-side.
  define: {
    global: "globalThis",
  },
  server: {
    port: 5173,
    // Dev only: accept any Host header (lets the app be reached via host.docker.internal etc.).
    allowedHosts: true,
    // Proxy API calls to the backend so the browser stays same-origin in dev.
    proxy: {
      "/api": {
        target: proxyTarget,
        changeOrigin: true,
      },
      // SockJS/STOMP endpoint for the real-time warehouse board (needs websocket upgrade)
      "/ws": {
        target: proxyTarget,
        changeOrigin: true,
        ws: true,
      },
    },
  },
});
