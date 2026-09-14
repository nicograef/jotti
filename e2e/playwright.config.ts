import { defineConfig, devices } from '@playwright/test'

// Default ist der lokale Dev-Stack; CI und E2E-Stack setzen E2E_BASE_URL auf
// ihre gemappte Adresse (z. B. http://localhost:8081).
const baseURL = process.env.E2E_BASE_URL ?? 'http://localhost'

// isCI schaltet nur Reporter und forbidOnly um; retries: 0 gilt auch lokal,
// damit Flakiness sichtbar bleibt.
const isCI = !!process.env.CI

export default defineConfig({
  testDir: './tests',
  // Keine festen Wartezeiten in Specs — Auto-Waiting hat Vorrang.
  timeout: 60_000,
  expect: { timeout: 15_000 },
  // POST /api/test/reset-and-seed setzt globalen DB-Zustand zurück: Spec-Dateien
  // dürfen sich nicht überlappen, daher seriell mit einem Worker.
  fullyParallel: false,
  workers: 1,
  forbidOnly: isCI,
  retries: 0,
  reporter: isCI
    ? [['github'], ['html', { open: 'never' }]]
    : [['list'], ['html', { open: 'never' }]],
  use: {
    baseURL,
    trace: 'retain-on-failure',
    screenshot: 'only-on-failure',
    video: 'off',
  },
  projects: [
    {
      // Nur Admin-Specs: sonst liefen die Service-Specs zusätzlich im
      // Desktop-Viewport.
      name: 'desktop-admin',
      use: { ...devices['Desktop Chrome'] },
      testMatch: /admin-.*\.spec\.ts$/,
    },
    {
      // Servicekräfte arbeiten mobil (BYOD); Admin-Seiten sind Desktop-only.
      name: 'mobile-service',
      use: { ...devices['Pixel 7'] },
      testIgnore: /admin-.*\.spec\.ts$/,
    },
  ],
})
