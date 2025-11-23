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
  // server: {
  //   host: '0.0.0.0',
  //   port: 5173,
  //   proxy: {
  //     // Proxy API requests to backend container; allows using relative /api base
  //     '/api': {
  //       target: 'http://backend-dev:3000',
  //       changeOrigin: true,
  //       rewrite: (p) => p.replace(/^\/api/, ''),
  //     },
  //   },
  // },
})
