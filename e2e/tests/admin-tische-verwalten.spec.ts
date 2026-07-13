import { expect, test } from '@playwright/test'

import { anmelden } from '../support/anmelden'
import { resetAndSeed } from '../support/seed'

// Deckt den Verwaltungspfad für Tische ab: Anlegen und den Namen ändern. Die
// Tische erscheinen als präfixgruppierte Kacheln; ein Klick auf die Kachel
// öffnet den Bearbeiten-Dialog (Umbenennen, Löschen).

test.describe('Admin verwaltet Tische', () => {
  test('Tisch anlegen und Namen ändern', async ({ page, request }) => {
    const zugangsdaten = await resetAndSeed(request)
    await anmelden(page, zugangsdaten.admin)

    await page.goto('/admin/tische')
    await expect(page.getByRole('heading', { name: 'Tische' })).toBeVisible()

    // Neuen Tisch anlegen.
    await page.getByRole('button', { name: 'Neuer Tisch' }).click()
    const newDialog = page.getByRole('dialog')
    await newDialog.getByLabel('Name').fill('Tisch 99')
    await newDialog.getByRole('button', { name: 'Tisch anlegen' }).click()
    await expect(page.getByText('Tisch "Tisch 99" wurde angelegt.')).toBeVisible()

    // Die Kachel öffnet per Klick den Bearbeiten-Dialog.
    const tischKachel = page.getByRole('button', { name: /Tisch 99/ })
    await expect(tischKachel).toBeVisible()
    await tischKachel.getByText('Tisch 99').click()

    // Namen ändern.
    const editDialog = page.getByRole('dialog')
    await editDialog.getByLabel('Name').fill('Zelt B3')
    await editDialog.getByRole('button', { name: 'Speichern' }).click()

    await expect(page.getByRole('button', { name: /Zelt B3/ })).toBeVisible()
  })
})
