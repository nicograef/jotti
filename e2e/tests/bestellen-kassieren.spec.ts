import { expect, test } from '@playwright/test'

import { anmelden } from '../support/anmelden'
import { resetAndSeed } from '../support/seed'
import { bestellePosition, oeffneTisch, zeileMit } from '../support/servicekraft'

// Tracer-Bullet-Spec: der Kernpfad einer Servicekraft am Tisch — anmelden,
// eine Bestellung aufnehmen, kassieren und den sichtbaren Betrag prüfen. Läuft
// nur im Mobile-Service-Projekt (kein admin-*-Dateiname, siehe
// playwright.config.ts), damit der Durchstich Backend, Reverse-Proxy, Frontend
// und Seed-Reset gemeinsam abdeckt.

// „Tisch 1" ist im Demo-Drehbuch der Frühschoppen-Stammtisch: jede Runde wird
// direkt bezahlt, der Tisch startet also ausgeglichen (Saldo 0,00 €) — das hält
// die Beträge im Test deterministisch.
const TISCH = 'Tisch 1'
// Bratwurst XXL kostet 5,00 € (Seed-Stammdaten).
const PRODUKT = 'Bratwurst'
const VARIANTE = 'XXL'
const PREIS = '5,00'

test.describe('Servicekraft nimmt eine Bestellung auf und kassiert', () => {
  test('Bestellung aufnehmen und kassieren zeigt den erwarteten Betrag', async ({
    page,
    request,
  }) => {
    // Jede Spec startet vom bekannten Seed-Zustand.
    const zugangsdaten = await resetAndSeed(request)

    await anmelden(page, zugangsdaten.service)

    // Tischauswahl öffnen und Tisch 1 aufrufen (oeffneTisch wartet auf die
    // sichtbare Tab-Leiste des Tisches).
    await oeffneTisch(page, TISCH)

    // Eine Bratwurst XXL bestellen.
    await bestellePosition(page, PRODUKT, VARIANTE)

    // Auf den Kassieren-Tab wechseln und die eben bestellte Position auswählen.
    await page.getByRole('tab', { name: 'Kassieren' }).click()

    const position = zeileMit(page, `${PRODUKT} ${VARIANTE}`, 'Produkt hinzufügen')
    await expect(position).toBeVisible()
    await position.getByRole('button', { name: 'Produkt hinzufügen' }).click()

    // Kassieren-Leiste zeigt den Betrag und öffnet den Zahlungs-Drawer.
    const kassierenLeiste = page.getByRole('button', { name: /Kassieren/ })
    await expect(kassierenLeiste).toContainText(PREIS)
    await kassierenLeiste.click()

    // Der Zahlungs-Drawer bestätigt den sichtbaren Betrag …
    const drawer = page.getByRole('dialog')
    await expect(drawer.getByText(new RegExp(`${PREIS}\\s*€`)).first()).toBeVisible()

    // … und die Zahlung wird kassiert.
    await drawer.getByRole('button', { name: 'Kassieren' }).click()

    await expect(page.getByText('Zahlung erfolgreich.')).toBeVisible()

    // Nach dem Kassieren ist der Tisch wieder ausgeglichen. Auf das Saldo-Element
    // im Tisch-Header gescopt (die einzige item-description mit text-2xl, siehe
    // TablePage) statt auf ein beliebiges „0,00 €" irgendwo auf der Seite.
    const tischSaldo = page.locator('[data-slot="item-description"].text-2xl')
    await expect(tischSaldo).toHaveText('0,00 €')
  })
})
