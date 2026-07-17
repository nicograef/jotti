import { expect, test } from '@playwright/test'

import { anmelden } from '../support/anmelden'
import { resetAndSeed } from '../support/seed'
import {
  LANGE_BESTELLUNG_POSITIONEN,
  nimmLangeBestellungAuf,
  oeffneHistorienDetail,
  oeffneTisch,
  waehleAlleVollAus,
  waehleVariante,
} from '../support/servicekraft'

// Viewport-Regression für das Drawer-Sticky-Footer-Layout. Bei einer langen
// Positionsliste (der DrawerBody scrollt) müssen Gesamtsumme, das jeweilige
// Pflichtfeld und die Primäraktion ohne Scrollen gleichzeitig sichtbar bleiben —
// sie liegen im nicht-scrollenden DrawerFooter, nur die Positionsliste (Body)
// scrollt. Ursprung: Bug 2 aus dem Praxistest 2026-07-09 (Bestellen/Kassieren),
// mit der UI-Politur (Phase 1) auf Stornierung und Umbuchung ausgeweitet.

test.describe('Drawer-Sticky-Footer bei langer Positionsliste', () => {
  // „Tisch 1" startet im Demo-Drehbuch ausgeglichen (Saldo 0,00 €) und ohne
  // unbezahlte Positionen — die Auswahl auf dem Kassieren-Tab enthält damit
  // genau die hier bestellten Positionen.
  test('Bestellung aufnehmen und mit Trinkgeld kassieren — Summe und Aktion sichtbar', async ({
    page,
    request,
  }) => {
    const zugangsdaten = await resetAndSeed(request)
    await anmelden(page, zugangsdaten.service)
    await oeffneTisch(page, 'Tisch 1')

    // Bestell-Drawer: Beleg zeigt alle Positionen, Gesamtsumme und Buttons
    // bleiben trotz langer Liste im Viewport.
    await page.getByRole('tab', { name: 'Bestellen' }).click()
    for (const [produkt, variante] of LANGE_BESTELLUNG_POSITIONEN) {
      await waehleVariante(page, produkt, variante)
    }
    await page.getByRole('button', { name: /Bestellung überprüfen/ }).click()
    const bestellDrawer = page.getByRole('dialog')
    await expect(bestellDrawer.getByText('Flammkuchen Mediterran')).toBeVisible()
    await expect(bestellDrawer.getByText('Gesamt')).toBeInViewport()
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
    // Der freie Zielbetrag steckt hinter dem „Anderer …"-Aufrunden-Chip.
    const zahlungsDrawer = page.getByRole('dialog')
    await zahlungsDrawer.getByRole('button', { name: 'Anderer …' }).click()
    await zahlungsDrawer.getByLabel('Zahlbetrag inkl. Trinkgeld').fill('55,00')
    await zahlungsDrawer.getByLabel('Erhalten').fill('60,00')
    // exact: der Trinkgeld-Buchungshinweis enthält das Wort ebenfalls.
    await expect(
      zahlungsDrawer.getByText('Trinkgeld', { exact: true }),
    ).toBeVisible()
    await expect(
      zahlungsDrawer.getByText('Rückgeld', { exact: true }),
    ).toBeVisible()

    // Der Kassieren-Button bleibt im Viewport und die Zahlung gelingt.
    const kassieren = zahlungsDrawer.getByRole('button', { name: 'Kassieren' })
    await expect(kassieren).toBeInViewport()
    await kassieren.click()
    await expect(page.getByText('Zahlung erfolgreich.').first()).toBeVisible()

    // Tisch ist danach wieder ausgeglichen (Saldo-Element im Tisch-Header).
    const tischSaldo = page.locator('[data-slot="tisch-saldo"]')
    await expect(tischSaldo).toHaveText('0,00 €')
  })

  // „Tisch 15" ist im Demo-Drehbuch unbenutzt; Storno ist nur der Serviceleitung
  // (bzw. Admin) erlaubt (AuthSingleton.canCancel).
  test('Stornierung: Summe, Pflicht-Kommentar und Aktion bleiben im Viewport', async ({
    page,
    request,
  }) => {
    const zugangsdaten = await resetAndSeed(request)
    await anmelden(page, zugangsdaten.serviceleitung)
    await oeffneTisch(page, 'Tisch 15')
    await nimmLangeBestellungAuf(page)

    // Über die Historie die Bestellung öffnen und den Storno-Drawer starten.
    await page.getByRole('tab', { name: 'Historie' }).click()
    const detail = await oeffneHistorienDetail(page, /Bestellung/)
    await detail.getByRole('button', { name: /Stornieren…/ }).click()

    const drawer = page.getByRole('dialog')
    await expect(drawer.getByText('Bratwurst Normal')).toBeVisible()

    // Alle Positionen auswählen (jede Zeile einmal „+"): erst dann steht die
    // Gesamtsumme im Footer; jede Menge ist 1, der Add-Button deaktiviert danach.
    const hinzufuegen = drawer.getByRole('button', { name: /hinzufügen$/ })
    const anzahl = await hinzufuegen.count()
    for (let i = 0; i < anzahl; i++) {
      await hinzufuegen.nth(i).click()
    }

    await drawer
      .getByPlaceholder('Kommentar (erforderlich)')
      .fill('Falsch bestellt, storniert')

    // Gesamtsumme, Pflicht-Kommentar und Primäraktion liegen zusammen im
    // sichtbaren Footer — auch bei langer, scrollender Positionsliste.
    await expect(drawer.getByText('Stornierung gesamt')).toBeInViewport()
    await expect(
      drawer.getByPlaceholder('Kommentar (erforderlich)'),
    ).toBeInViewport()
    await expect(
      drawer.getByRole('button', { name: 'Stornierung erteilen' }),
    ).toBeInViewport()
  })

  // „Tisch 3" (Quelle) und „Tisch 10" (Ziel) sind im Demo-Drehbuch unbenutzt;
  // Umbuchen ist für jede Servicekraft erlaubt.
  test('Umbuchung: Summe, Ziel-Tisch und Aktion bleiben im Viewport', async ({
    page,
    request,
  }) => {
    const zugangsdaten = await resetAndSeed(request)
    await anmelden(page, zugangsdaten.service)
    await oeffneTisch(page, 'Tisch 3')
    await nimmLangeBestellungAuf(page)

    // Über die Historie die Bestellung öffnen und den Umbuchungs-Drawer starten.
    await page.getByRole('tab', { name: 'Historie' }).click()
    const detail = await oeffneHistorienDetail(page, /Bestellung/)
    await detail.getByRole('button', { name: /Umbuchen/ }).click()

    const drawer = page.getByRole('dialog')
    await expect(drawer.getByText('Bratwurst Normal')).toBeVisible()

    // „Alle auswählen" übernimmt die volle umbuchbare Menge — erst dann steht die
    // Gesamtsumme im Footer.
    await drawer
      .getByRole('button', { name: /Alle \d+ Positionen auswählen/ })
      .click()
    await drawer.getByRole('combobox').selectOption({ label: 'Tisch 10' })

    // Gesamtsumme, Ziel-Tisch-Auswahl (Pflichtfeld) und Primäraktion liegen
    // zusammen im sichtbaren Footer — auch bei langer, scrollender Liste.
    await expect(drawer.getByText('Umbuchung gesamt')).toBeInViewport()
    await expect(drawer.getByRole('combobox')).toBeInViewport()
    await expect(
      drawer.getByRole('button', { name: 'Umbuchung ausführen' }),
    ).toBeInViewport()
  })
})
