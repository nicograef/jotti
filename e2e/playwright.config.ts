import { defineConfig, devices } from '@playwright/test'

// E2E_BASE_URL ist BASE_URL-agnostisch: Default ist der lokale Dev-Stack
// (http://localhost, Frontend :80). Die CI und der eigene E2E-Stack setzen die
// Variable auf die jeweils gemappte Adresse (z. B. http://localhost:8081).
const baseURL = process.env.E2E_BASE_URL ?? 'http://localhost'

// CI ist strenger: keine Retries verschleiern Flakiness, Trace/Screenshots nur
// beim ersten Fehlversuch, damit die Artefakte klein bleiben.
const isCI = !!process.env.CI

export default defineConfig({
  testDir: './tests',
  // Keine festen Wartezeiten in Specs: großzügige, aber endliche Timeouts, die
  // Playwrights Auto-Waiting den Vortritt lassen.
  timeout: 60_000,
  expect: { timeout: 15_000 },
  fullyParallel: true,
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
      // Admin arbeitet am Desktop (großer Viewport).
      name: 'desktop-admin',
      use: { ...devices['Desktop Chrome'] },
    },
    {
      // Servicekräfte arbeiten mobil (BYOD-Smartphone, Hochkant-Viewport).
      name: 'mobile-service',
      use: { ...devices['Pixel 7'] },
    },
  ],
})
