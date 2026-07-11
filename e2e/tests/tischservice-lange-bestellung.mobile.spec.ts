import { expect, test } from '@playwright/test'

import { anmelden } from '../support/anmelden'
import { resetAndSeed } from '../support/seed'
import {
  oeffneTisch,
  waehleAlleVollAus,
  waehleVariante,
} from '../support/servicekraft'

// Viewport-Regression für das Drawer-Scroll-Layout (Bug 2 aus dem Praxistest
// 2026-07-09): Bei einer Bestellung mit vielen Positionen und ausgefüllten
// Trinkgeld-/Erhalten-Feldern müssen die Primär-Buttons („Bestellung
// aufnehmen", „Kassieren") im mobilen Viewport sichtbar und bedienbar bleiben
// — der Mittelteil des Drawers (DrawerBody) scrollt, Header und Footer stehen
// fest.

// „Tisch 1" startet im Demo-Drehbuch ausgeglichen (Saldo 0,00 €) und ohne
// unbezahlte Positionen — die Auswahl auf dem Kassieren-Tab enthält damit
// genau die hier bestellten Positionen.
const TISCH = 'Tisch 1'

// 9 unterschiedliche Varianten → 9 Positionen im Beleg (Summe 52,00 €). Die
// Variantennamen sind über alle gewählten Produkte hinweg eindeutig, damit
// waehleVariante jede Zeile ohne Mehrdeutigkeit trifft.
const POSITIONEN: [produkt: string, variante: string][] = [
  ['Bratwurst', 'Normal'],
  ['Bratwurst', 'XXL'],
  ['Bratwurst', 'Currywurst'],
  ['Pommes', 'Klein'],
  ['Pommes', 'Groß'],
  ['Flammkuchen', 'Classic'],
  ['Flammkuchen', 'Speck & Zwiebel'],
  ['Flammkuchen', 'Mediterran'],
  ['Tagesgericht', 'Fr: Schnitzel mit Pommes'],
]

test.describe('Drawer-Scroll-Layout bei langer Positionsliste', () => {
  test('Bestellung mit 9 Positionen aufnehmen und mit Trinkgeld kassieren', async ({
    page,
    request,
  }) => {
    const zugangsdaten = await resetAndSeed(request)
    await anmelden(page, zugangsdaten.service)
    await oeffneTisch(page, TISCH)

    // Alle 9 Positionen in einer Bestellung sammeln.
    await page.getByRole('tab', { name: 'Bestellen' }).click()
    for (const [produkt, variante] of POSITIONEN) {
      await waehleVariante(page, produkt, variante)
    }

    // Bestell-Drawer: Beleg zeigt alle Positionen, die Buttons bleiben trotz
    // langer Liste im Viewport.
    await page.getByRole('button', { name: /Bestellung überprüfen/ }).click()
    const bestellDrawer = page.getByRole('dialog')
    await expect(bestellDrawer.getByText('Flammkuchen Mediterran')).toBeVisible()
    const aufnehmen = bestellDrawer.getByRole('button', {
      name: 'Bestellung aufnehmen',
    })
    await expect(aufnehmen).toBeInViewport()
    await expect(
      bestellDrawer.getByRole('button', { name: 'Abbrechen' }),
    ).toBeInViewport()
    await aufnehmen.click()
    await expect(
      page.getByText('Bestellung wurde aufgenommen.').first(),
    ).toBeVisible()

    // Alle Positionen zum Kassieren auswählen und den Zahlungs-Drawer öffnen.
    await page.getByRole('tab', { name: 'Kassieren' }).click()
    await waehleAlleVollAus(page)
    await page.getByRole('button', { name: /Kassieren/ }).click()

    // Trinkgeld und Erhalten eintragen: Die zusätzlichen Rückgeld-/
    // Trinkgeld-Zeilen samt Hinweistext verlängern den Drawer-Inhalt weiter.
    const zahlungsDrawer = page.getByRole('dialog')
    await zahlungsDrawer.getByLabel('inklusive Trinkgeld').fill('55,00')
    await zahlungsDrawer.getByLabel('Erhalten').fill('60,00')
    await expect(
      zahlungsDrawer.getByText('Trinkgeld', { exact: true }),
    ).toBeVisible()
    await expect(zahlungsDrawer.getByText('Rückgeld')).toBeVisible()

    // Der Kassieren-Button bleibt im Viewport und die Zahlung gelingt.
    const kassieren = zahlungsDrawer.getByRole('button', { name: 'Kassieren' })
    await expect(kassieren).toBeInViewport()
    await kassieren.click()
    await expect(page.getByText('Zahlung erfolgreich.').first()).toBeVisible()

    // Tisch ist danach wieder ausgeglichen (Saldo-Element im Tisch-Header).
    const tischSaldo = page.locator('[data-slot="item-description"].text-2xl')
    await expect(tischSaldo).toHaveText('0,00 €')
  })
})
