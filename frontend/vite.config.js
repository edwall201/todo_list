import { defineConfig } from 'vite'
import react from '@vitejs/plugin-react'

// During development the React app runs on http://localhost:5173.
// The `proxy` below forwards any request starting with /api to the Go
// service on http://localhost:8080, so the frontend can just call
// "/api/todos" and never worry about CORS or ports.
export default defineConfig({
  plugins: [react()],
  server: {
    port: 5173,
    proxy: {
      '/api': 'http://localhost:8080',
    },
  },
})
