import { expect, test } from '@playwright/test'

import { anmelden } from '../support/anmelden'
import { resetAndSeed } from '../support/seed'

// Deckt den Verwaltungspfad für die Bondrucker-Stationen ab: die Drucker-IP
// einer Station ändern. Die IP wird on-blur gespeichert (kein Speichern-Button).

test.describe('Admin verwaltet Druckstationen', () => {
  test('Drucker-IP einer Station ändern', async ({ page, request }) => {
    const zugangsdaten = await resetAndSeed(request)
    await anmelden(page, zugangsdaten.admin)

    await page.goto('/admin/druckstationen')
    await expect(page.getByRole('heading', { name: 'Bondrucker' })).toBeVisible()

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
    await expect(ipInput).toHaveValue('192.168.8.99')
  })
})
