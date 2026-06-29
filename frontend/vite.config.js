import { defineConfig } from 'vite'
import react from '@vitejs/plugin-react'

// The frontend now talks to the backend over NATS (WebSocket), not HTTP,
// so there is no longer an /api proxy. The NATS server URL lives in
// src/api.js (default ws://localhost:8080, see nats.conf).
export default defineConfig({
  plugins: [react()],
  server: {
    port: 5173,
  },
})
