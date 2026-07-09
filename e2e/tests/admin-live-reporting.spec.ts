import { expect, test } from '@playwright/test'

import { anmelden } from '../support/anmelden'
import { resetAndSeed } from '../support/seed'

// Deckt das Live-Dashboard ab: die Umsätze der laufenden (Sonntags-)Sitzung
// aus den Seed-Daten sind sichtbar.

test.describe('Admin sieht das Live-Dashboard', () => {
  test('Live-Dashboard zeigt Umsätze der Seed-Daten', async ({
    page,
    request,
  }) => {
    const zugangsdaten = await resetAndSeed(request)
    await anmelden(page, zugangsdaten.admin)

    await page.goto('/admin/auswertung')
    // Die Seite zeigt sowohl das Live-Dashboard als auch (weiter unten) eine
    // historische Auswertung mit denselben Tab-Beschriftungen — deshalb bleibt
    // jede Prüfung auf den Live-Dashboard-Bereich beschränkt.
    const liveSection = page.getByTestId('live-reporting-section')
    await expect(liveSection.getByRole('heading', { name: 'Live-Dashboard' })).toBeVisible()

    // Übersicht-Kennzahlen der laufenden Sitzung sind sichtbar (Seed hat
    // Bestellungen, Direktverkäufe und Stornierungen am Sonntag).
    await expect(liveSection.getByText('Bestellungen')).toBeVisible()
    await expect(liveSection.getByText('Gesamtumsatz')).toBeVisible()
    await expect(liveSection.getByText('Offene Saldi')).toBeVisible()

    // Offene Tische aus dem Seed-Drehbuch sind gelistet.
    await expect(
      liveSection.getByRole('heading', { name: 'Offene Tische' }),
    ).toBeVisible()

    // Servicekräfte-Tab zeigt die aktiven Bediener des Sonntags.
    await liveSection.getByRole('tab', { name: 'Servicekräfte' }).click()
    await expect(liveSection.getByText('maria (Maria Schmidt)')).toBeVisible()

    // Stornierungen-Tab zeigt die dokumentierten Stornos des Sonntags.
    await liveSection.getByRole('tab', { name: 'Stornierungen' }).click()
    await expect(
      liveSection.getByText('Reklamation Tagesgericht, Kulanz'),
    ).toBeVisible()
  })
})
