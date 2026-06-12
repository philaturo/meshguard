// File: apps/dashboard/vite.config.ts
// Purpose: Vite build configuration for React dashboard
// Connects to: package.json (dev/build scripts), index.html (entry point)
// Proxy: /api and /ws forwarded to Go backend on localhost:8080

import { defineConfig } from 'vite'
import react from '@vitejs/plugin-react'

export default defineConfig({
  plugins: [react()],
  server: {
    port: 5173,
    host: true,
    proxy: {
      '/api': {
        target: 'http://localhost:8080',
        changeOrigin: true,
      },
      '/ws': {
        target: 'ws://localhost:8080',
        ws: true,
      },
    },
  },
  build: {
    outDir: 'dist',
    emptyOutDir: true,
  },
})
