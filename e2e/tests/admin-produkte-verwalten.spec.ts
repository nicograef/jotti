import { expect, test } from '@playwright/test'

import { anmelden } from '../support/anmelden'
import { resetAndSeed } from '../support/seed'

// Deckt den Verwaltungspfad für Produkte samt Varianten ab: Anlegen,
// Bearbeiten, eine Variante aktivieren und wieder deaktivieren.

test.describe('Admin verwaltet Produkte und Varianten', () => {
  test('Produkt samt Variante anlegen, ändern und Variante deaktivieren', async ({
    page,
    request,
  }) => {
    const zugangsdaten = await resetAndSeed(request)
    await anmelden(page, zugangsdaten.admin)

    await page.goto('/admin/produkte')
    await expect(page.getByRole('heading', { name: 'Produkte verwalten' })).toBeVisible()

    // Neues Produkt anlegen.
    await page.getByRole('button', { name: 'Neues Produkt' }).click()
    const newProductDialog = page.getByRole('dialog')
    await newProductDialog.getByLabel('Name').fill('Eistee')
    await newProductDialog.getByRole('button', { name: 'Produkt anlegen' }).click()
    await expect(page.getByText('Produkt "Eistee" wurde angelegt.')).toBeVisible()

    // Das neue Produkt in der Liste öffnen und eine Variante anlegen.
    const produktItem = page
      .locator('[data-slot="item"]')
      .filter({ hasText: 'Eistee' })
    await produktItem.getByRole('button', { name: 'Varianten' }).click()
    await produktItem.getByRole('button', { name: 'Variante' }).click()

    const newVariantDialog = page.getByRole('dialog')
    await newVariantDialog.getByLabel('Name').fill('0,5l')
    await newVariantDialog.getByLabel('Preis').fill('2,80')
    await newVariantDialog.getByRole('button', { name: 'Variante anlegen' }).click()
    await expect(page.getByText('Variante "0,5l" wurde angelegt.')).toBeVisible()

    // Die neue Variante ist zunächst deaktiviert; die Zeile mit Preis prüfen.
    await expect(produktItem.getByText('2,80')).toBeVisible()
    const variantSwitch = produktItem.getByRole('switch').last()
    await expect(variantSwitch).not.toBeChecked()

    // Variante aktivieren.
    await variantSwitch.click()
    await expect(variantSwitch).toBeChecked()
    await expect(produktItem.getByText('1 aktiv')).toBeVisible()

    // Variante wieder deaktivieren.
    await variantSwitch.click()
    await expect(variantSwitch).not.toBeChecked()
    await expect(produktItem.getByText('0 aktiv')).toBeVisible()

    // Produkt bearbeiten: Namen ändern.
    await produktItem
      .getByRole('button', { name: 'Produkt bearbeiten' })
      .click()
    const editDialog = page.getByRole('dialog')
    await editDialog.getByLabel('Name').fill('Eistee Pfirsich')
    await editDialog.getByRole('button', { name: 'Speichern' }).click()

    await expect(
      page.locator('[data-slot="item"]').filter({ hasText: 'Eistee Pfirsich' }),
    ).toBeVisible()
  })
})
