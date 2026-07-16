import { expect, test } from '@playwright/test'

import { anmelden } from '../support/anmelden'
import { resetAndSeed } from '../support/seed'

// Deckt den Verwaltungspfad für Helfer ab: Anlegen (inkl. Anzeige des
// Einmalpasswort-Codes) und Deaktivieren. Die Seite „Helfer & Zugänge" zeigt
// die Helfer als Tabelle mit Status-Switch je Zeile.

test.describe('Admin verwaltet Helfer', () => {
  test('Helfer anlegen zeigt das Einmalpasswort und der Helfer lässt sich deaktivieren', async ({
    page,
    request,
  }) => {
    const zugangsdaten = await resetAndSeed(request)
    await anmelden(page, zugangsdaten.admin)

    await page.goto('/admin/benutzer')
    await expect(
      page.getByRole('heading', { name: 'Helfer & Zugänge' }),
    ).toBeVisible()

    // Neuen Helfer anlegen.
    await page.getByRole('button', { name: 'Neuer Helfer' }).click()
    const newDialog = page.getByRole('dialog')
    await newDialog.getByLabel('Name', { exact: true }).fill('Petra Neumann')
    await newDialog.getByLabel('Benutzername').fill('petra')
    await newDialog.getByRole('button', { name: 'Helfer anlegen' }).click()
    await expect(
      page.getByText('Neuer Helfer "Petra Neumann" wurde erstellt.'),
    ).toBeVisible()

    // Der Einmalpasswort-Dialog zeigt Benutzername und Code an.
    const createdDialog = page.getByRole('dialog').filter({
      hasText: 'Helfer wurde angelegt!',
    })
    await expect(createdDialog).toBeVisible()
    await expect(createdDialog.getByText('petra', { exact: true })).toBeVisible()
    const code = await createdDialog
      .getByTestId('onetime-password')
      .textContent()
    // Das Einmalpasswort ist ein sechsstelliger Zifferncode (siehe
    // ADMIN-EINMALPASSWORT-Format); die Assertion prüft genau dieses Format.
    expect(code?.trim(), 'Einmalpasswort muss sechsstelliger Zifferncode sein').toMatch(
      /^\d{6}$/,
    )
    await createdDialog.getByRole('button', { name: 'Okay' }).click()

    // Neue Helfer starten inaktiv (Passwort noch nicht gesetzt): erst
    // aktivieren, dann den Deaktivieren-Pfad prüfen. Die Tabellenzeile wird
    // über den Namen und ihren Status-Switch aufgelöst.
    const userRow = page
      .locator('div')
      .filter({ has: page.getByRole('switch') })
      .filter({ hasText: 'Petra Neumann' })
      .last()
    await expect(userRow).toBeVisible()
    const userSwitch = userRow.getByRole('switch')
    await expect(userSwitch).not.toBeChecked()

    await userSwitch.click()
    await expect(userSwitch).toBeChecked()

    await userSwitch.click()
    await expect(userSwitch).not.toBeChecked()
  })
})
