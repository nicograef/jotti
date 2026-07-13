import type { Page } from '@playwright/test'
import { expect, test } from '@playwright/test'

import { anmelden } from '../support/anmelden'
import { resetAndSeed } from '../support/seed'
import {
  nimmLangeBestellungAuf,
  oeffneTisch,
  waehleVariante,
} from '../support/servicekraft'

// Viewport-Regression für den Raster-Basisspalten-Sweep (Phase 2). Die Listen
// der Servicekraft-Screens nutzen Grids, deren Spalten erst am Breakpoint (lg/
// 2xl) greifen; ohne Basis-Track (`grid-cols-1`) sizen die impliziten Grid-Tracks
// am Handy auf max-content und lange, nicht umbrechende Inhalte (Titel mit
// `truncate`, lange Positions-Zusammenfassungen) sprengen die Seitenbreite. Mit
// der Basis `grid-cols-1` (= `minmax(0,1fr)`) bleibt der Track auf die
// Containerbreite gedeckelt und die Inhalte kürzen sich statt überzulaufen. Der
// Test misst pro Screen den horizontalen Seiten-Überlauf am 390px-Viewport.

// Schmaler Hochkant-Viewport (390px Breite) — enger als der Pixel-7-Default und
// die im Audit dokumentierte Referenzbreite für den Überlauf-Check.
test.use({ viewport: { width: 390, height: 844 } })

// erwarteKeinenHorizontalenUeberlauf misst am gerenderten DOM, ob die Seite
// horizontal überläuft: scrollWidth des Scroll-Wurzelelements gegen die
// Viewport-Breite (innerWidth). Verhaltensbasiert statt Klassennamen-Prüfung.
async function erwarteKeinenHorizontalenUeberlauf(
  page: Page,
  screen: string,
): Promise<void> {
  const { scrollWidth, innerWidth } = await page.evaluate(() => ({
    scrollWidth: document.scrollingElement?.scrollWidth ?? 0,
    innerWidth: window.innerWidth,
  }))
  expect(
    scrollWidth,
    `${screen}: scrollWidth ${scrollWidth.toString()} darf innerWidth ${innerWidth.toString()} nicht überschreiten`,
  ).toBeLessThanOrEqual(innerWidth)
}

test.describe('Kein horizontaler Überlauf der Servicekraft-Screens bei 390px', () => {
  test('Tischauswahl, Kassieren und Historie bleiben in der Viewport-Breite', async ({
    page,
    request,
  }) => {
    const zugangsdaten = await resetAndSeed(request)
    await anmelden(page, zugangsdaten.serviceleitung)

    // Auf einem Tisch eine lange, gemischte Bestellung aufnehmen — füllt die
    // Kassieren- und Historien-Listen dieses Tisches mit genug nicht-umbrechendem
    // Text (teils lange Varianten-Namen), dass ein fehlender Basis-Track überliefe.
    await oeffneTisch(page, 'Tisch 3')

    // Bestellen-Tab: die Varianten-Liste (VariantRow) mit teils langen Namen.
    // Dieser Check ist die 390px-Überlauf-Regression für den Bestellen-Screen
    // (Akzeptanzkriterium: kein horizontaler Überlauf, keine abgeschnittenen
    // Preise). Die feste Preis-Spalte selbst (Name flex-1 truncate, Preis
    // shrink-0) deckt der Unit-Test von VariantNamePreis ab; hier zählt allein
    // die Seitenbreite.
    await expect(page.getByText('Fr: Schnitzel mit Pommes')).toBeVisible()
    await erwarteKeinenHorizontalenUeberlauf(page, 'Bestellen')

    await nimmLangeBestellungAuf(page)

    // Kassieren-Tab: Positionsliste (Zahlung-Grid) gefüllt.
    await page.getByRole('tab', { name: 'Kassieren' }).click()
    await expect(page.getByText('Bratwurst Normal').first()).toBeVisible()
    await erwarteKeinenHorizontalenUeberlauf(page, 'Kassieren')

    // Historie-Tab: Historien-Liste (TischHistorie-Grid) gefüllt.
    await page.getByRole('tab', { name: 'Historie' }).click()
    await expect(page.getByRole('button', { name: /Bestellung/ })).toBeVisible()
    await erwarteKeinenHorizontalenUeberlauf(page, 'Historie')

    // Tischauswahl: die Karten-Grids (Noch offen / Erledigt) am schmalen
    // Viewport. Der zuvor bestellte Tisch erscheint hier als eigene Karte.
    await page.goto('/service/tische')
    await expect(page.getByRole('button', { name: 'Alle Tische' })).toBeVisible()
    await erwarteKeinenHorizontalenUeberlauf(page, 'Tischauswahl')
  })

  test('Direktverkauf-Historie bleibt in der Viewport-Breite', async ({
    page,
    request,
  }) => {
    const zugangsdaten = await resetAndSeed(request)
    await anmelden(page, zugangsdaten.serviceleitung)

    // Einen Direktverkauf mit mehreren Positionen tätigen — die
    // Positions-Zusammenfassung der Historien-Zeile wird dadurch lang.
    await page.goto('/service/direktverkauf')
    await expect(page.getByRole('tab', { name: 'Verkaufen' })).toBeVisible()
    await waehleVariante(page, 'Festbändchen', 'Erwachsene')
    await waehleVariante(page, 'Festbändchen', 'Kinder')
    await page.getByRole('button', { name: /Kassieren.*8,00/ }).click()
    const kassierenDrawer = page.getByRole('dialog')
    await expect(kassierenDrawer).toBeVisible()
    await kassierenDrawer
      .getByRole('button', { name: 'Verkauf abschließen' })
      .click()
    await expect(page.getByText('Verkauf abgeschlossen.')).toBeVisible()

    // Direktverkauf-Historie: die Verkaufs-Liste (DirektverkaufHistorie-Grid).
    await page.getByRole('tab', { name: 'Historie' }).click()
    await expect(page.getByRole('button', { name: /Verkauf.*8,00/ })).toBeVisible()
    await erwarteKeinenHorizontalenUeberlauf(page, 'Direktverkauf-Historie')
  })
})
