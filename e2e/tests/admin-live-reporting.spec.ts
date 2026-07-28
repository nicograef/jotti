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
    // Stornos nach der Storno-Zuordnung auf, „Details" zeigt die dokumentierten
    // Stornos des Sonntags. Der Sonntag hat zwei Warenrücknahmen (Tisch 7 und
    // Tisch 9); beide Zahlungen hat lisa kassiert, ausgelöst haben die Stornos
    // sophie bzw. felix. Angerechnet werden sie deshalb lisa — ihrer Kasse ging
    // das Bargeld ab —, nicht den stellvertretend Stornierenden. Der exakte Text
    // belegt zugleich, dass lisa der einzige Eintrag ist; felix erscheint am
    // Sonntag nur in der Team-Liste, weil er an Tisch 5 selbst kassiert hat. Das
    // führende „Betroffen:" gehört zur Zusage, dass die Zeile nicht als
    // Aufteilung der darüberstehenden Storno-Kopfkennzahl gelesen wird.
    await expect(
      liveSection.getByText('Betroffen: lisa (Lisa Braun) 2', { exact: true }),
    ).toBeVisible()
    await liveSection.getByRole('button', { name: 'Details' }).click()
    await expect(
      liveSection.getByText('Reklamation Tagesgericht, Kulanz'),
    ).toBeVisible()
  })
})
