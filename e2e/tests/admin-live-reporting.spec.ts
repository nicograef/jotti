import { expect, test } from '@playwright/test'

import { anmelden } from '../support/anmelden'
import { resetAndSeed } from '../support/seed'

// Deckt die Übersicht (Live-Dashboard) ab: die Umsätze der laufenden
// (Sonntags-)Sitzung aus den Seed-Daten sind sichtbar. Die Übersicht ist eine
// einzelne scrollbare Seite ohne Tabs — alle Blöcke sind direkt erreichbar.

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
      liveSection.getByRole('heading', { name: 'Übersicht' }),
    ).toBeVisible()

    // Single-Scroll: keine Tabs mehr.
    await expect(liveSection.getByRole('tab')).toHaveCount(0)

    // Kennzahlen der laufenden Sitzung sind sichtbar (Seed hat Bestellungen,
    // Direktverkäufe und Stornierungen am Sonntag): Hero-Kennzahl plus Nebenkarten.
    await expect(liveSection.getByText('Kassierter Umsatz')).toBeVisible()
    await expect(liveSection.getByText('Noch offen')).toBeVisible()
    await expect(liveSection.getByText('Bestellt gesamt')).toBeVisible()

    // Aktualitäts-Anzeige und manueller Refresh sind vorhanden.
    await expect(liveSection.getByText(/aktualisiert \d{2}:\d{2}/)).toBeVisible()
    await expect(
      liveSection.getByRole('button', { name: 'Aktualisieren' }),
    ).toBeVisible()

    // Offene Tische aus dem Seed-Drehbuch sind gelistet.
    await expect(
      liveSection.getByRole('heading', { name: 'Offene Tische' }),
    ).toBeVisible()

    // Team-Block zeigt die aktiven Servicekräfte des Sonntags.
    await expect(
      liveSection.getByRole('heading', { name: 'Team' }),
    ).toBeVisible()
    await expect(liveSection.getByText('maria (Maria Schmidt)')).toBeVisible()

    // Der Stornierungen-Block ist eingeklappt: die Zusammenfassung schlüsselt die
    // Stornos pro Servicekraft auf (felix hat kassiert und storniert), „Details"
    // zeigt die dokumentierten Stornos des Sonntags.
    await expect(
      liveSection.getByText(/felix \(Felix Weber\) 1/),
    ).toBeVisible()
    await liveSection.getByRole('button', { name: 'Details' }).click()
    await expect(
      liveSection.getByText('Reklamation Tagesgericht, Kulanz'),
    ).toBeVisible()
  })
})
