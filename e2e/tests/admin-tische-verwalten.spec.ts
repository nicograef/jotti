import { expect, test } from '@playwright/test'

import { anmelden } from '../support/anmelden'
import { resetAndSeed } from '../support/seed'

// Deckt den Verwaltungspfad für Tische ab: Anlegen und den Namen ändern.

test.describe('Admin verwaltet Tische', () => {
  test('Tisch anlegen und Namen ändern', async ({ page, request }) => {
    const zugangsdaten = await resetAndSeed(request)
    await anmelden(page, zugangsdaten.admin)

    await page.goto('/admin/tische')
    await expect(page.getByRole('heading', { name: 'Tische verwalten' })).toBeVisible()

    // Neuen Tisch anlegen.
    await page.getByRole('button', { name: 'Neuer Tisch' }).click()
    const newDialog = page.getByRole('dialog')
    await newDialog.getByLabel('Name').fill('Tisch 99')
    await newDialog.getByRole('button', { name: 'Tisch anlegen' }).click()
    await expect(page.getByText('Tisch "Tisch 99" wurde angelegt.')).toBeVisible()

    const tischItem = page
      .locator('[data-slot="item"]')
      .filter({ hasText: 'Tisch 99' })
    await expect(tischItem).toBeVisible()

    // Namen ändern.
    await tischItem.getByRole('button', { name: 'Tisch bearbeiten' }).click()
    const editDialog = page.getByRole('dialog')
    await editDialog.getByLabel('Name').fill('Zelt B3')
    await editDialog.getByRole('button', { name: 'Speichern' }).click()

    await expect(
      page.locator('[data-slot="item"]').filter({ hasText: 'Zelt B3' }),
    ).toBeVisible()
  })
})
