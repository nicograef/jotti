import tailwindcss from '@tailwindcss/vite'
import react from '@vitejs/plugin-react'
import path from 'path'
import { defineConfig } from 'vite'

// https://vite.dev/config/
export default defineConfig({
  plugins: [react(), tailwindcss()],
  resolve: {
    alias: {
      '@': path.resolve(__dirname, './src'),
    },
  },
  server: {
    // Proxy /api to Go backend (replaces nginx reverse proxy in dev)
    proxy: {
      '/api': {
        target: `http://${process.env.BACKEND_HOST ?? 'localhost'}:3000`,
        rewrite: (path) => path.replace(/^\/api/, ''),
      },
    },
  },
})
