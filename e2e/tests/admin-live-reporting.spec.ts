import { expect, test } from '@playwright/test'

import { anmelden } from '../support/anmelden'
import { resetAndSeed } from '../support/seed'

// Deckt das Live-Dashboard ab: die Umsätze der laufenden (Sonntags-)Sitzung
// aus den Seed-Daten sind sichtbar. Das Dashboard ist eine einzelne scrollbare
// Seite ohne Tabs — alle Blöcke sind direkt erreichbar.

test.describe('Admin sieht das Live-Dashboard', () => {
  test('Live-Dashboard zeigt Umsätze der Seed-Daten', async ({
    page,
    request,
  }) => {
    const zugangsdaten = await resetAndSeed(request)
    await anmelden(page, zugangsdaten.admin)

    await page.goto('/admin/auswertung')
    const liveSection = page.getByTestId('live-reporting-section')
    await expect(
      liveSection.getByRole('heading', { name: 'Live-Dashboard' }),
    ).toBeVisible()

    // Single-Scroll: keine Tabs mehr.
    await expect(liveSection.getByRole('tab')).toHaveCount(0)

    // Kennzahlen der laufenden Sitzung sind sichtbar (Seed hat Bestellungen,
    // Direktverkäufe und Stornierungen am Sonntag).
    await expect(liveSection.getByText('Gesamtumsatz')).toBeVisible()
    await expect(liveSection.getByText('Offene Saldi')).toBeVisible()
    await expect(liveSection.getByText('Bestellungen')).toBeVisible()

    // Aktualitäts-Anzeige und manueller Refresh sind vorhanden.
    await expect(liveSection.getByText(/^Stand \d{2}:\d{2}$/)).toBeVisible()
    await expect(
      liveSection.getByRole('button', { name: 'Aktualisieren' }),
    ).toBeVisible()

    // Offene Tische aus dem Seed-Drehbuch sind gelistet.
    await expect(
      liveSection.getByRole('heading', { name: 'Offene Tische' }),
    ).toBeVisible()

    // Servicekräfte-Block zeigt die aktiven Bediener des Sonntags.
    await expect(
      liveSection.getByRole('heading', { name: 'Servicekräfte' }),
    ).toBeVisible()
    await expect(liveSection.getByText('maria (Maria Schmidt)')).toBeVisible()

    // Stornierungen-Block zeigt die dokumentierten Stornos des Sonntags.
    await expect(
      liveSection.getByRole('heading', { name: 'Stornierungen' }),
    ).toBeVisible()
    await expect(
      liveSection.getByText('Reklamation Tagesgericht, Kulanz'),
    ).toBeVisible()

    // Stornierungen pro Servicekraft: felix hat kassiert und storniert, seine
    // Servicekraft-Zeile trägt den roten Storno-Marker.
    await expect(liveSection.getByText('1 Storno')).toBeVisible()
  })
})
