import react from '@vitejs/plugin-react'
import path from 'path'
import { defineConfig } from 'vitest/config'

export default defineConfig({
  plugins: [react()],
  // Tests run as a fixed release, so the version guard compares for real against
  // each test's /health fake. Fixed because a set VERSION must not change the test run.
  define: {
    __CLIENT_VERSION__: JSON.stringify('v1.2.3'),
  },
  resolve: {
    alias: {
      '@': path.resolve(import.meta.dirname, './src'),
    },
  },
  test: {
    environment: 'jsdom',
    setupFiles: ['./src/test/setup.ts'],
  },
})
