import { expect, test } from '@playwright/test'

import { anmelden } from '../support/anmelden'
import { resetAndSeed } from '../support/seed'

// Deckt den Verwaltungspfad für Benutzer ab: Anlegen (inkl. Anzeige des
// Einmalpasswort-Codes) und Deaktivieren.

test.describe('Admin verwaltet Benutzer', () => {
  test('Benutzer anlegen zeigt das Einmalpasswort und der Benutzer lässt sich deaktivieren', async ({
    page,
    request,
  }) => {
    const zugangsdaten = await resetAndSeed(request)
    await anmelden(page, zugangsdaten.admin)

    await page.goto('/admin/benutzer')
    await expect(page.getByRole('heading', { name: 'Benutzer verwalten' })).toBeVisible()

    // Neuen Benutzer anlegen.
    await page.getByRole('button', { name: 'Neuer Benutzer' }).click()
    const newDialog = page.getByRole('dialog')
    await newDialog.getByLabel('Name', { exact: true }).fill('Petra Neumann')
    await newDialog.getByLabel('Benutzername').fill('petra')
    await newDialog.getByRole('button', { name: 'Benutzer anlegen' }).click()
    await expect(
      page.getByText('Neuer Benutzer "Petra Neumann" wurde erstellt.'),
    ).toBeVisible()

    // Der Einmalpasswort-Dialog zeigt Benutzername und Code an.
    const createdDialog = page.getByRole('dialog').filter({
      hasText: 'Benutzer wurde angelegt!',
    })
    await expect(createdDialog).toBeVisible()
    await expect(createdDialog.getByText('petra', { exact: true })).toBeVisible()
    const code = await createdDialog
      .locator('p.text-3xl.tracking-widest')
      .textContent()
    expect(code, 'Einmalpasswort-Code muss angezeigt werden').toBeTruthy()
    expect(code?.trim().length).toBeGreaterThan(0)
    await createdDialog.getByRole('button', { name: 'Okay' }).click()

    // Neue Benutzer starten inaktiv (Passwort noch nicht gesetzt): erst
    // aktivieren, dann den Deaktivieren-Pfad prüfen.
    const userItem = page
      .locator('[data-slot="item"]')
      .filter({ hasText: 'Petra Neumann' })
    await expect(userItem).toBeVisible()
    const userSwitch = userItem.getByRole('switch')
    await expect(userSwitch).not.toBeChecked()

    await userSwitch.click()
    await expect(userSwitch).toBeChecked()

    await userSwitch.click()
    await expect(userSwitch).not.toBeChecked()
  })
})
