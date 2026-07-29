import tailwindcss from '@tailwindcss/vite'
import react from '@vitejs/plugin-react'
import path from 'path'
import { defineConfig } from 'vite'

// https://vite.dev/config/
export default defineConfig({
  plugins: [react(), tailwindcss()],
  // Clientversion zur Bauzeit eingebrannt (Deklaration: src/version.d.ts).
  // Der Release-Workflow reicht den Tag als Build-Arg durch; ohne VERSION
  // steht auf beiden Seiten der Default `dev` und der Versionsvergleich
  // bleibt still. Dieselbe Zeile steht in vitest.config.ts — die beiden
  // Configs ersetzen einander, sie ergaenzen sich nicht.
  define: {
    __CLIENT_VERSION__: JSON.stringify(process.env.VERSION ?? 'dev'),
  },
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
