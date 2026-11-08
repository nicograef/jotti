import { expect, test } from '@playwright/test'

import { anmelden } from '../support/anmelden'
import { resetAndSeed } from '../support/seed'

// Deckt den Verwaltungspfad für die Bondrucker ab: die Drucker-IP einer
// Station ändern und einen Zusatzdrucker mit Produkt einrichten. Name und IP
// werden on-blur gespeichert (kein Speichern-Button).

test.describe('Admin verwaltet Druckstationen', () => {
  test('Drucker-IP einer Station ändern', async ({ page, request }) => {
    const zugangsdaten = await resetAndSeed(request)
    await anmelden(page, zugangsdaten.admin)

    await page.goto('/admin/druckstationen')
    await expect(
      page.getByRole('heading', { name: 'Bondrucker' }),
    ).toBeVisible()

    // Die Essen-Station ist per Seed-Drehbuch mit 192.168.8.51 konfiguriert.
    const essenCard = page
      .locator('div')
      .filter({ hasText: 'Bons für die Essensausgabe' })
      .filter({ has: page.getByLabel('Drucker-IP') })
      .last()
    const ipInput = essenCard.getByLabel('Drucker-IP')
    await expect(ipInput).toHaveValue('192.168.8.51')

    // Neue IP eintragen und das Feld verlassen — das speichert on-blur.
    await ipInput.fill('192.168.8.99')
    await ipInput.blur()

    await expect(
      page.getByText(/Drucker-IP für .Essen. gespeichert\./),
    ).toBeVisible()

    // Nach dem Neuladen kommt der Wert vom Backend, nicht aus dem Eingabefeld.
    await page.reload()
    await expect(
      page.getByRole('heading', { name: 'Bondrucker' }),
    ).toBeVisible()
    await expect(ipInput).toHaveValue('192.168.8.99')
  })

  test('Zusatzdrucker mit Produkt einrichten', async ({ page, request }) => {
    const zugangsdaten = await resetAndSeed(request)
    await anmelden(page, zugangsdaten.admin)

    await page.goto('/admin/druckstationen')
    await page.getByRole('button', { name: 'Drucker hinzufügen' }).click()
    const anlegenDialog = page.getByRole('dialog')
    await anlegenDialog.getByLabel('Name').fill('Kaffee & Kuchen')
    await anlegenDialog
      .getByRole('button', { name: 'Drucker hinzufügen' })
      .click()
    await expect(
      page.getByText('Drucker „Kaffee & Kuchen“ angelegt.'),
    ).toBeVisible()

    // Die Karte ist das innerste div mit Titelzusatz und IP-Feld.
    const karte = page
      .locator('div')
      .filter({ hasText: '— Zusätzlicher Drucker' })
      .filter({ has: page.getByLabel('Drucker-IP') })
      .last()
    const ipInput = karte.getByLabel('Drucker-IP')
    await ipInput.fill('192.168.8.60')
    await ipInput.blur()
    await expect(
      page.getByText('Drucker-IP für „Kaffee & Kuchen“ gespeichert.'),
    ).toBeVisible()

    await karte.getByRole('button', { name: 'Produkt hinzufügen' }).click()
    const auswahl = page.getByRole('dialog')
    await auswahl
      .getByRole('listitem')
      .filter({ has: page.getByText('Kuchen', { exact: true }) })
      .getByRole('button')
      .click()
    await expect(
      page.getByText('„Kuchen“ druckt jetzt an „Kaffee & Kuchen“.'),
    ).toBeVisible()
    await expect(auswahl).toBeHidden()
    await expect(karte.getByText('Kuchen', { exact: true })).toBeVisible()

    await karte.getByRole('button', { name: 'Testbon' }).click()
    await expect(
      page.getByText('Testbon an „Kaffee & Kuchen“ gesendet.'),
    ).toBeVisible()

    // Nach dem Neuladen kommt die Zuordnung vom Backend.
    await page.reload()
    await expect(
      page.getByRole('heading', { name: 'Bondrucker' }),
    ).toBeVisible()
    await expect(karte.getByText('Kuchen', { exact: true })).toBeVisible()
  })
})
