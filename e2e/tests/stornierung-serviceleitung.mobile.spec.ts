import { expect, test } from '@playwright/test'

import { anmelden } from '../support/anmelden'
import { resetAndSeed } from '../support/seed'
import {
  bestellePosition,
  kassierePosition,
  oeffneHistorienDetail,
  oeffneTisch,
  zeileMit,
} from '../support/servicekraft'

// Stornierung über die Serviceleitung, beide Arten: die geldneutrale Korrektur
// (Position noch unbezahlt) und die kassenwirksame Warenrücknahme (Position
// bereits bezahlt). „Tisch 15" ist im Demo-Drehbuch des laufenden Tages
// unbenutzt. Storno ist nur Admin/Serviceleitung erlaubt (AuthSingleton.canCancel).
// Die Historien-Karte einer Bestellung zeigt nur Art und Betrag („Bestellung
// +2,50 €"), keine Produktnamen — die beiden Bestellungen bekommen deshalb
// bewusst unterschiedliche Preise, damit sie über den Betrag eindeutig
// unterscheidbar bleiben.

const TISCH = 'Tisch 15'

test.describe('Serviceleitung storniert geldneutral und mit Warenrücknahme', () => {
  test('unbezahlte Position wird geldneutral korrigiert, bezahlte per Warenrücknahme storniert', async ({
    page,
    request,
  }) => {
    const zugangsdaten = await resetAndSeed(request)
    await anmelden(page, zugangsdaten.serviceleitung)

    await oeffneTisch(page, TISCH)

    // Pommes Klein (2,50 €) bleibt unbezahlt (geldneutrale Korrektur), Brezel
    // Normal (2,00 €) wird kassiert (Warenrücknahme).
    await bestellePosition(page, 'Pommes', 'Klein')
    await bestellePosition(page, 'Brezel', 'Normal')

    await kassierePosition(page, 'Brezel Normal')

    // Saldo zeigt nur noch die unbezahlten Pommes.
    await expect(page.getByText('2,50 €').first()).toBeVisible()

    // --- Geldneutrale Korrektur: Storno der unbezahlten Pommes-Bestellung ---
    // Historien-Zeilen tragen keine Inline-Aktionen mehr: die Zeile (per Betrag
    // eindeutig) öffnet den Detail-Drawer, dort liegt „Stornieren…".
    await page.getByRole('tab', { name: 'Historie' }).click()
    let detail = await oeffneHistorienDetail(page, /Bestellung.*\+2,50/)
    await detail.getByRole('button', { name: /Stornieren…/ }).click()

    let drawer = page.getByRole('dialog')
    await expect(
      drawer.getByRole('heading', { name: /^Bestellung ·/ }),
    ).toBeVisible()
    const pommesZeile = zeileMit(drawer, 'Pommes Klein', 'hinzufügen')
    await pommesZeile.getByRole('button', { name: /hinzufügen/ }).click()
    await drawer
      .getByPlaceholder('Kommentar (erforderlich)')
      .fill('Falsch bestellt, storniert')
    await drawer.getByRole('button', { name: 'Stornierung erteilen' }).click()

    // Die geldneutrale Korrektur gleicht den Saldo aus — der Tisch ist wieder
    // bei 0,00 €.
    await expect(page.getByText('0,00 €').first()).toBeVisible()

    // --- Kassenwirksame Warenrücknahme: Storno der bezahlten Brezel-Bestellung ---
    // Die Bestellung, aus der die kassierte Position stammt, bleibt weiterhin
    // stornierbar (Warenrücknahme storniert die Bestellposition, nicht die
    // separate Zahlung).
    detail = await oeffneHistorienDetail(page, /Bestellung.*\+2,00/)
    await detail.getByRole('button', { name: /Stornieren…/ }).click()

    drawer = page.getByRole('dialog')
    const brezelZeile = zeileMit(drawer, 'Brezel Normal', 'hinzufügen')
    await brezelZeile.getByRole('button', { name: /hinzufügen/ }).click()
    await drawer
      .getByPlaceholder('Kommentar (erforderlich)')
      .fill('Brezel reklamiert, Warenrücknahme')
    await drawer.getByRole('button', { name: 'Stornierung erteilen' }).click()

    // Die Warenrücknahme lässt den Tisch-Saldo unverändert (die Position war
    // bereits bezahlt) und erscheint als eigener „Warenrücknahme"-Eintrag mit
    // dem erstatteten Betrag in der Historie.
    await expect(page.getByText('0,00 €').first()).toBeVisible()
    const stornoEintrag = page.getByRole('button', { name: /Warenrücknahme/ })
    await expect(stornoEintrag).toBeVisible()
    await expect(stornoEintrag).toContainText('2,00')
  })
})
