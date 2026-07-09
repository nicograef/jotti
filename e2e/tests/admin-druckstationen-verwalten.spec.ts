import { expect, test } from '@playwright/test'

import { anmelden } from '../support/anmelden'
import { resetAndSeed } from '../support/seed'

// Deckt den Verwaltungspfad für Druckstationen ab: die Drucker-IP einer
// Station ändern und speichern.

test.describe('Admin verwaltet Druckstationen', () => {
  test('Drucker-IP einer Station ändern', async ({ page, request }) => {
    const zugangsdaten = await resetAndSeed(request)
    await anmelden(page, zugangsdaten.admin)

    await page.goto('/admin/druckstationen')
    await expect(
      page.getByRole('heading', { name: 'Druckstationen' }),
    ).toBeVisible()

    // Die Essen-Station ist per Seed-Drehbuch mit 192.168.8.51 konfiguriert.
    const essenRow = page
      .locator('div')
      .filter({ hasText: 'Arbeitsbons für bestellte Essens-Positionen.' })
      .filter({ has: page.getByLabel('Drucker-IP') })
      .last()
    const ipInput = essenRow.getByLabel('Drucker-IP')
    await expect(ipInput).toHaveValue('192.168.8.51')

    await ipInput.fill('192.168.8.99')
    await essenRow.getByRole('button', { name: 'Speichern' }).click()

    await expect(page.getByText('Druckstation „Essen“ gespeichert.')).toBeVisible()
    await expect(ipInput).toHaveValue('192.168.8.99')
  })
})
