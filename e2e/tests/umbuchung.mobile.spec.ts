import { expect, test } from '@playwright/test'

import { anmelden } from '../support/anmelden'
import { resetAndSeed } from '../support/seed'
import {
  bestellePosition,
  oeffneHistorienDetail,
  oeffneTisch,
  zeileMit,
} from '../support/servicekraft'

// Umbuchung einer offenen Bestellung von einem Tisch auf einen anderen. Beide
// Tische ("Tisch 3" und „Tisch 10") sind im Demo-Drehbuch des laufenden Tages
// unbenutzt. Umbuchen ist geldneutral und für jede Servicekraft erlaubt.
// *.mobile.spec.ts läuft nur im mobile-service-Projekt (Handy-Viewport).

const QUELL_TISCH = 'Tisch 3'
const ZIEL_TISCH = 'Tisch 10'

test.describe('Servicekraft bucht eine Bestellung auf einen anderen Tisch um', () => {
  test('Quelltisch wird ausgeglichen, Zieltisch übernimmt den Saldo', async ({
    page,
    request,
  }) => {
    const zugangsdaten = await resetAndSeed(request)
    await anmelden(page, zugangsdaten.service)

    await oeffneTisch(page, QUELL_TISCH)
    await bestellePosition(page, 'Wasser', 'Still 0,5l') // 2,00 €

    await expect(page.getByText('2,00 €').first()).toBeVisible()

    // Historien-Zeilen tragen keine Inline-Aktionen mehr: die Bestellzeile öffnet
    // den Detail-Drawer, dort liegt „Umbuchen".
    await page.getByRole('tab', { name: 'Historie' }).click()
    const detail = await oeffneHistorienDetail(page, /Bestellung/)
    await detail.getByRole('button', { name: /Umbuchen/ }).click()

    const drawer = page.getByRole('dialog')
    await expect(
      drawer.getByRole('heading', { name: /^Bestellung ·/ }),
    ).toBeVisible()

    const wasserZeile = zeileMit(drawer, 'Wasser Still 0,5l', 'hinzufügen')
    await expect(wasserZeile).toBeVisible()

    const ausfuehren = drawer.getByRole('button', {
      name: 'Umbuchung ausführen',
    })
    // Der Drawer startet mit leerer Auswahl (wie Kassieren/Storno): ohne
    // ausgewählte Position und ohne Ziel-Tisch bleibt die Umbuchung gesperrt.
    await expect(ausfuehren).toBeDisabled()

    // „Alle auswählen" übernimmt die volle umbuchbare Menge aller Positionen.
    await drawer
      .getByRole('button', { name: /Alle \d+ Positionen auswählen/ })
      .click()

    // Optionales Benutzerkommentar erfassen (ergänzt den Richtungs-Autotext).
    await drawer.getByPlaceholder('Kommentar (optional)').fill('Gast gewechselt')

    // Auch mit gewählten Positionen bleibt gesperrt, bis der Ziel-Tisch steht.
    await expect(ausfuehren).toBeDisabled()
    await drawer.getByRole('combobox').selectOption({ label: ZIEL_TISCH })
    await expect(ausfuehren).toBeEnabled()
    await ausfuehren.click()
    await expect(page.getByText('Bestellung umgebucht.')).toBeVisible()

    // Der Quelltisch ist danach ausgeglichen …
    await expect(page.getByText('0,00 €').first()).toBeVisible()

    // … und der Zieltisch zeigt den übernommenen Saldo.
    await oeffneTisch(page, ZIEL_TISCH)
    await expect(page.getByText('2,00 €').first()).toBeVisible()

    // Der Umbuchungs-Zugang trägt das Benutzerkommentar (Autotext bleibt Titel).
    // Der Autotext lautet "Umbuchung von <Tischname>" — mit dem Seed-Namen
    // "Tisch 3" ergibt das "Umbuchung von Tisch 3", ohne Doppelung des Worts "Tisch".
    await page.getByRole('tab', { name: 'Historie' }).click()
    const zugangDetail = await oeffneHistorienDetail(
      page,
      new RegExp(`Umbuchung von ${QUELL_TISCH}`),
    )
    await expect(zugangDetail.getByRole('textbox')).toHaveValue(
      'Gast gewechselt',
    )
  })
})
