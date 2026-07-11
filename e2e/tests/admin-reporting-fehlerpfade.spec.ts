import { expect, test } from '@playwright/test'

import {
  simuliereNetzabbruch,
  simuliereServerfehler,
} from '../helpers/fehlerpfade'
import { anmelden } from '../support/anmelden'
import { resetAndSeed } from '../support/seed'

// Admin-Reporting (Live-Dashboard und Kassenberichte) — der globale
// QueryCache-Handler (frontend/src/lib/queryClient.ts) fängt Query-Fehler mit
// einem Toast ab. Diese Specs erzwingen, dass er auch hier greift, statt dass
// die Seite stillschweigend wie „keine Kassensitzung offen" wirkt.

test.describe('Admin-Dashboard bei Serverfehler und Netzabbruch', () => {
  test('Serverfehler beim Live-Reporting zeigt einen sichtbaren Fehlerhinweis', async ({
    page,
    request,
  }) => {
    const zugangsdaten = await resetAndSeed(request)
    await anmelden(page, zugangsdaten.admin)

    await simuliereServerfehler(page, ['admin/get-live-reporting'])
    await page.goto('/admin/auswertung')

    await expect(
      page.getByText('Daten konnten nicht geladen werden. Bitte Verbindung prüfen und erneut versuchen.'),
    ).toBeVisible()
  })

  test('Netzabbruch bei der Kassensitzungsliste zeigt einen sichtbaren Fehlerhinweis', async ({
    page,
    request,
  }) => {
    const zugangsdaten = await resetAndSeed(request)
    await anmelden(page, zugangsdaten.admin)

    await simuliereNetzabbruch(page, ['admin/get-abgeschlossene-kassensitzungen'])
    await page.goto('/admin/kassenberichte')

    await expect(
      page.getByText('Daten konnten nicht geladen werden. Bitte Verbindung prüfen und erneut versuchen.'),
    ).toBeVisible()
  })
})
