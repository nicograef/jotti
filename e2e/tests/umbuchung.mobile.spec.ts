import { expect, test } from '@playwright/test'

import { anmelden } from '../support/anmelden'
import { resetAndSeed } from '../support/seed'
import {
  bestellePosition,
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

    await page.getByRole('tab', { name: 'Historie' }).click()
    const bestellungsEintrag = zeileMit(page, 'Bestellung', 'Umbuchen')
    await expect(bestellungsEintrag).toBeVisible()
    await bestellungsEintrag.getByRole('button', { name: 'Umbuchen' }).click()

    const drawer = page.getByRole('dialog')
    await expect(drawer.getByText(/Umbuchung aus Vorgang/)).toBeVisible()

    const wasserZeile = zeileMit(drawer, 'Wasser Still 0,5l', 'hinzufügen')
    // Die Position ist standardmäßig bereits mit der vollen Menge vorbelegt
    // (createDefaultMengen), daher wird hier nichts weiter angeklickt.
    await expect(wasserZeile).toBeVisible()

    await drawer.getByRole('combobox').selectOption({ label: ZIEL_TISCH })
    await drawer
      .getByRole('button', { name: 'Umbuchung ausführen' })
      .click()
    await expect(page.getByText('Bestellung umgebucht.')).toBeVisible()

    // Der Quelltisch ist danach ausgeglichen …
    await expect(page.getByText('0,00 €').first()).toBeVisible()
    await expect(page.getByText('Alles ausgegeben!')).toBeVisible()

    // … und der Zieltisch zeigt den übernommenen Saldo.
    await oeffneTisch(page, ZIEL_TISCH)
    await expect(page.getByText('2,00 €').first()).toBeVisible()
  })
})
