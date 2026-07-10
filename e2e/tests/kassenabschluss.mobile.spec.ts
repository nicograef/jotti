import { expect, test } from '@playwright/test'

import { anmelden } from '../support/anmelden'
import { resetAndSeed } from '../support/seed'
import {
  abmelden,
  bestellePosition,
  kassierePosition,
  oeffneTisch,
  settleAlleOffenenTische,
} from '../support/servicekraft'

// Kassenabschluss: die Spec erzeugt selbst einen frischen Umsatz (Bestellung,
// Kassieren) auf einem im Demo-Drehbuch unbenutzten Tisch, meldet sich
// dann als Admin an und schließt die laufende Kassensitzung ab. Assertiert wird
// ausschließlich die sichtbare Abschlussmeldung. Läuft im Handy-Viewport wie
// die übrigen Service-Flows, obwohl „Kasse abschließen" eine Admin-Route ist
// (/admin/kasse) — Admins bedienen die Kasse ebenfalls am Smartphone.
//
// Der Kassenabschluss verlangt, dass jeder Tisch ausgeglichen ist
// (tische_saldo_offen); das Demo-Drehbuch des laufenden Tages hinterlässt
// mehrere Tische bewusst in offenem Zustand (Vielzustand-Abdeckung) — die Spec
// gleicht sie deshalb vor dem Abschluss selbst aus.

const TISCH = 'Tisch 12'

test.describe('Kassenabschluss beendet die laufende Kassensitzung', () => {
  test('erzeugt einen Umsatz und zeigt nach dem Abschluss die Erfolgsmeldung', async ({
    page,
    request,
  }) => {
    // Höheres Timeout als der Standard: das Ausgleichen aller vom Drehbuch
    // offen gelassenen Tische (settleAlleOffenenTische) durchläuft rund 20
    // Tische mit vielen Positionen und braucht dadurch spürbar länger als ein
    // einzelner Kernflow.
    test.setTimeout(120_000)

    const zugangsdaten = await resetAndSeed(request)

    // Frischer Umsatz durch eine vollständige Runde am Tisch.
    await anmelden(page, zugangsdaten.service)
    await oeffneTisch(page, TISCH)
    await bestellePosition(page, 'Kaffee', 'Tasse') // 2,00 €
    await kassierePosition(page, 'Kaffee Tasse')
    // Saldo-Element im Tisch-Header (die einzige item-description mit text-2xl,
    // siehe TablePage): der Tisch ist nach dem Kassieren ausgeglichen.
    await expect(
      page.locator('[data-slot="item-description"].text-2xl'),
    ).toHaveText('0,00 €')

    // Alle übrigen, vom Drehbuch offen gelassenen Tische ausgleichen, damit
    // der Kassenabschluss nicht am „tische_saldo_offen"-Gate scheitert.
    await settleAlleOffenenTische(page)

    await abmelden(page)

    // Als Admin die Kasse abschließen.
    await anmelden(page, zugangsdaten.admin)
    await page.goto('/admin/kasse')
    await expect(page.getByRole('heading', { name: 'Kassensitzung' })).toBeVisible()
    await expect(page.getByText('offen')).toBeVisible()

    await page.getByLabel('Gezählter Ist-Bestand').fill('342,00')
    await page.getByRole('button', { name: 'Kasse abschließen' }).click()

    await expect(
      page.getByRole('alertdialog', { name: 'Kasse abschließen?' }),
    ).toBeVisible()
    await page
      .getByRole('alertdialog')
      .getByRole('button', { name: 'Kasse abschließen' })
      .click()

    // Sichtbare Abschlussmeldung nach erfolgreichem Kassenabschluss.
    await expect(page.getByText(/Kasse abgeschlossen\./)).toBeVisible()
  })
})
