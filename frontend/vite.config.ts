import { defineConfig } from "vite";
import react from "@vitejs/plugin-react";

// Proxy /api to the Go server so the session cookie is first-party to the dev
// origin (localhost:5173) — avoids cross-site cookie issues entirely.
export default defineConfig({
  plugins: [react()],
  server: {
    port: 5173,
    proxy: {
      "/api": "http://localhost:8080",
    },
  },
});
