import { expect, test } from '@playwright/test'

import { anmelden } from '../support/anmelden'
import { resetAndSeed } from '../support/seed'

// Deckt den Verwaltungspfad für Produkte samt Varianten ab: Anlegen,
// Bearbeiten, eine Variante aktivieren und wieder deaktivieren. Die Seite
// „Produkte & Preise" listet Produkte je Kategorie mit Varianten als Chips
// (Name, Preis, Mini-Switch); eine inaktive Variante trägt die „aus"-Markierung.

test.describe('Admin verwaltet Produkte und Varianten', () => {
  test('Produkt samt Variante anlegen, ändern und Variante deaktivieren', async ({
    page,
    request,
  }) => {
    const zugangsdaten = await resetAndSeed(request)
    await anmelden(page, zugangsdaten.admin)

    await page.goto('/admin/produkte')
    await expect(
      page.getByRole('heading', { name: 'Produkte & Preise' }),
    ).toBeVisible()

    // Neues Produkt anlegen.
    await page.getByRole('button', { name: 'Neues Produkt' }).click()
    const newProductDialog = page.getByRole('dialog')
    await newProductDialog.getByLabel('Name').fill('Eistee')
    await newProductDialog.getByRole('button', { name: 'Produkt anlegen' }).click()
    await expect(page.getByText('Produkt "Eistee" wurde angelegt.')).toBeVisible()

    // Die Produktzeile über den Namen und ihren Bearbeiten-Button auflösen und
    // eine Variante anlegen (der gestrichelte „Variante"-Button je Zeile).
    const produktItem = page
      .locator('div')
      .filter({ has: page.getByRole('button', { name: 'Produkt bearbeiten' }) })
      .filter({ hasText: 'Eistee' })
      .last()
    await produktItem.getByRole('button', { name: 'Variante', exact: true }).click()

    const newVariantDialog = page.getByRole('dialog')
    await newVariantDialog.getByLabel('Name').fill('0,5l')
    await newVariantDialog.getByLabel('Preis').fill('2,80')
    await newVariantDialog.getByRole('button', { name: 'Variante anlegen' }).click()
    await expect(page.getByText('Variante "0,5l" wurde angelegt.')).toBeVisible()

    // Die neue Variante ist zunächst deaktiviert: Preis-Chip sichtbar, Switch
    // aus und die „aus"-Markierung gesetzt.
    await expect(produktItem.getByText('2,80')).toBeVisible()
    const ausMarker = produktItem.getByText('aus', { exact: true })
    await expect(ausMarker).toBeVisible()
    const variantSwitch = produktItem.getByRole('switch').last()
    await expect(variantSwitch).not.toBeChecked()

    // Variante aktivieren: Switch an, „aus"-Markierung verschwindet.
    await variantSwitch.click()
    await expect(variantSwitch).toBeChecked()
    await expect(ausMarker).toBeHidden()

    // Variante wieder deaktivieren.
    await variantSwitch.click()
    await expect(variantSwitch).not.toBeChecked()
    await expect(ausMarker).toBeVisible()

    // Produkt bearbeiten: Namen ändern.
    await produktItem
      .getByRole('button', { name: 'Produkt bearbeiten' })
      .click()
    const editDialog = page.getByRole('dialog')
    await editDialog.getByLabel('Name').fill('Eistee Pfirsich')
    await editDialog.getByRole('button', { name: 'Speichern' }).click()

    await expect(page.getByText('Eistee Pfirsich')).toBeVisible()
  })
})
