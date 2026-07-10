import { expect, test } from '@playwright/test'

import { anmelden } from '../support/anmelden'
import { resetAndSeed } from '../support/seed'
import {
  bestellePosition,
  kassierePosition,
  oeffneTisch,
} from '../support/servicekraft'

// Servicekraft-Kernflow am Handy: Bestellen, Kassieren — einmal als
// Teilzahlung (der Tisch bleibt danach mit Restsaldo offen) und einmal als
// Vollzahlung (Tisch danach ausgeglichen). „Tisch 13" ist im Demo-Drehbuch des
// laufenden Tages unbenutzt, damit die Beträge deterministisch bleiben.

const TISCH = 'Tisch 13'
// Zwei Positionen desselben Produkts mit unterschiedlichem Preis, damit sich
// Teil- und Vollzahlung eindeutig zuordnen lassen.
const NORMAL = 'Bratwurst Normal' // 3,50 €
const XXL = 'Bratwurst XXL' // 5,00 €

test.describe('Servicekraft bestellt und kassiert', () => {
  test('Teilzahlung lässt Restsaldo offen, Vollzahlung gleicht den Tisch aus', async ({
    page,
    request,
  }) => {
    const zugangsdaten = await resetAndSeed(request)
    await anmelden(page, zugangsdaten.service)

    await oeffneTisch(page, TISCH)

    await bestellePosition(page, 'Bratwurst', 'Normal')
    await bestellePosition(page, 'Bratwurst', 'XXL')

    // Saldo-Element im Tisch-Header (die einzige item-description mit text-2xl,
    // siehe TablePage): so prüft die Assertion den tatsächlichen Restsaldo statt
    // eines beliebigen gleichlautenden Betrags irgendwo auf der Seite.
    const tischSaldo = page.locator('[data-slot="item-description"].text-2xl')

    // Nur die Bratwurst Normal (3,50 €) kassieren — die Bratwurst XXL bleibt
    // unbezahlt, der Tisch zeigt den Restsaldo.
    await kassierePosition(page, NORMAL)

    await expect(tischSaldo).toHaveText('5,00 €')

    // Jetzt auch die Bratwurst XXL kassieren — der Tisch gleicht sich aus.
    await kassierePosition(page, XXL)

    await expect(tischSaldo).toHaveText('0,00 €')
  })
})
