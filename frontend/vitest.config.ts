import react from '@vitejs/plugin-react'
import path from 'path'
import { defineConfig } from 'vitest/config'

export default defineConfig({
  plugins: [react()],
  // Diese Config ersetzt vite.config.ts im Test; ohne die Konstante scheitert
  // jeder Test, der src/lib/version.ts importiert. Fest verdrahtet, weil kein
  // Test eine echte Release-Version braucht und ein gesetztes VERSION
  // (`make check VERSION=<tag>` reicht es durch) den Testlauf nicht ändern darf.
  define: {
    __CLIENT_VERSION__: JSON.stringify('dev'),
  },
  resolve: {
    alias: {
      '@': path.resolve(__dirname, './src'),
    },
  },
  test: {
    environment: 'jsdom',
    setupFiles: ['./src/test/setup.ts'],
  },
})
