import tailwindcss from '@tailwindcss/vite'
import react from '@vitejs/plugin-react'
import path from 'path'
import { defineConfig } from 'vite'

export default defineConfig({
  plugins: [react(), tailwindcss()],
  // Der Release-Workflow reicht den Tag als Build-Arg durch; ohne VERSION steht
  // auf beiden Seiten der Default `dev` und der Vergleich bleibt still. Die
  // gleiche Zeile in vitest.config.ts ersetzt diese, sie ergaenzt sie nicht.
  define: {
    __CLIENT_VERSION__: JSON.stringify(process.env.VERSION ?? 'dev'),
  },
  resolve: {
    alias: {
      '@': path.resolve(__dirname, './src'),
    },
  },
  server: {
    proxy: {
      '/api': {
        target: `http://${process.env.BACKEND_HOST ?? 'localhost'}:3000`,
        rewrite: (path) => path.replace(/^\/api/, ''),
      },
    },
  },
})
