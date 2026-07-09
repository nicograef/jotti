import { expect, test } from '@playwright/test'
import type { Locator, Page } from '@playwright/test'

import {
  simuliereNetzabbruch,
  simuliereServerfehler,
} from '../helpers/fehlerpfade'
import { anmelden } from '../support/anmelden'
import { resetAndSeed } from '../support/seed'

// Kassieren-Drawer (ZahlungDrawer) meldet Fehler über useActionSubmit als
// Toast (siehe frontend/src/hooks/use-action-submit.ts + lib/errorMessages.ts)
// — diese Specs bestätigen, dass ein Serverfehler bzw. Netzabbruch beim
// Absenden sichtbar bleibt statt eine stille „Zahlung erfolgreich"-Meldung
// vorzutäuschen.

// „Tisch 1" ist im Demo-Drehbuch der Frühschoppen-Stammtisch: jede Runde wird
// direkt bezahlt, der Tisch ist ausgeglichen — für diese Specs wird also erst
// eine neue Bestellung aufgenommen, damit eine unbezahlte Position entsteht.
const TISCH = 'Tisch 1'
const PRODUKT = 'Bratwurst'
const VARIANTE = 'Normal'

test.describe('Kassieren-Drawer bei Serverfehler und Netzabbruch', () => {
  test('Serverfehler beim Kassieren zeigt eine sichtbare Fehlermeldung statt „Zahlung erfolgreich"', async ({
    page,
    request,
  }) => {
    const zugangsdaten = await resetAndSeed(request)
    await anmelden(page, zugangsdaten.service)
    await zumTischMitOffenerPosition(page)

    await simuliereServerfehler(page, ['service/zahlung-kassieren'])
    await kassiereEinePosition(page)

    await expect(
      page.getByText(/unerwarteter Serverfehler|Referenz:/i),
    ).toBeVisible()
    await expect(page.getByText('Zahlung erfolgreich.')).not.toBeVisible()
  })

  test('Netzabbruch beim Kassieren zeigt eine sichtbare Fehlermeldung statt „Zahlung erfolgreich"', async ({
    page,
    request,
  }) => {
    const zugangsdaten = await resetAndSeed(request)
    await anmelden(page, zugangsdaten.service)
    await zumTischMitOffenerPosition(page)

    await simuliereNetzabbruch(page, ['service/zahlung-kassieren'])
    await kassiereEinePosition(page)

    await expect(
      page.getByText(/Zahlung kassieren fehlgeschlagen/i),
    ).toBeVisible()
    await expect(page.getByText('Zahlung erfolgreich.')).not.toBeVisible()
  })
})

// zumTischMitOffenerPosition navigiert zu Tisch 1 und nimmt eine Bratwurst
// Normal auf, damit eine unbezahlte Position zum Kassieren bereitsteht.
async function zumTischMitOffenerPosition(page: Page) {
  await page.goto('/service/tische')
  await page.getByRole('button', { name: 'Alle Tische' }).click()
  await page.getByPlaceholder('Tisch suchen...').fill(TISCH)
  await page
    .getByRole('button', { name: new RegExp(`^${TISCH}\\b.*€`) })
    .click()

  await expect(page.getByRole('tab', { name: 'Bestellen' })).toBeVisible()

  await page.getByText(PRODUKT, { exact: false }).first().click()
  const variante = zeileMit(page, VARIANTE, 'Variante hinzufügen')
  await variante.getByRole('button', { name: 'Variante hinzufügen' }).click()
  await page.getByRole('button', { name: /Bestellung überprüfen/ }).click()
  const bestellDrawer = page.getByRole('dialog')
  await bestellDrawer
    .getByRole('button', { name: 'Bestellung aufnehmen' })
    .click()
  await expect(page.getByText('Bestellung wurde aufgenommen.')).toBeVisible()
}

// kassiereEinePosition wechselt auf den Kassieren-Tab, wählt die zuvor
// aufgenommene Position aus und löst den Kassieren-Vorgang im Drawer aus.
async function kassiereEinePosition(page: Page) {
  await page.getByRole('tab', { name: 'Kassieren' }).click()

  const position = zeileMit(page, `${PRODUKT} ${VARIANTE}`, 'Produkt hinzufügen')
  await expect(position).toBeVisible()
  await position.getByRole('button', { name: 'Produkt hinzufügen' }).click()

  await page.getByRole('button', { name: /Kassieren/ }).click()
  const drawer = page.getByRole('dialog')
  await drawer.getByRole('button', { name: 'Kassieren' }).click()
}

function zeileMit(page: Page, text: string, buttonName: string): Locator {
  return page
    .locator('div')
    .filter({ hasText: text })
    .filter({ has: page.getByRole('button', { name: buttonName }) })
    .last()
}
