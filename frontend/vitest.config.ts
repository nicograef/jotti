import react from '@vitejs/plugin-react'
import path from 'path'
import { defineConfig } from 'vitest/config'

// https://vitest.dev/config/
export default defineConfig({
  plugins: [react()],
  // Diese Config ersetzt vite.config.ts im Test; ohne die Konstante hier
  // scheitert jeder Test, der src/lib/version.ts importiert. Der Wert ist
  // bewusst fest verdrahtet: Kein Test braucht eine echte Release-Version
  // (VersionsHinweis.test.tsx mockt sie), und ein gesetztes VERSION in der
  // Umgebung — `make check VERSION=<tag>` reicht es durch — darf den Testlauf
  // nicht verändern.
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
