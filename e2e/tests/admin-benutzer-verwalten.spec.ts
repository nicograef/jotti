import { expect, test } from '@playwright/test'

import { anmelden } from '../support/anmelden'
import { resetAndSeed } from '../support/seed'

// Deckt das Onboarding eines Helfers ab: der Admin legt ihn an (Anzeige des
// Einmalpasswort-Codes) und schaltet ihn frei, der Helfer setzt mit dem Code
// sein Passwort und meldet sich an. Als Servicekraft bleibt ihm der
// Admin-Bereich verschlossen. Die Seite „Helfer & Zugänge" zeigt die Helfer
// als Tabelle mit Status-Switch je Zeile.

const PETRA_PASSWORT = 'petra-passwort-42'

test.describe('Admin verwaltet Helfer', () => {
  test('neuer Helfer setzt mit dem Einmalpasswort sein Passwort und meldet sich an', async ({
    page,
    request,
  }) => {
    const zugangsdaten = await resetAndSeed(request)
    await anmelden(page, zugangsdaten.admin)

    await page.goto('/admin/benutzer')
    await expect(
      page.getByRole('heading', { name: 'Helfer & Zugänge' }),
    ).toBeVisible()

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
    await expect(
      createdDialog.getByText('petra', { exact: true }),
    ).toBeVisible()
    const code = (
      (await createdDialog.getByTestId('onetime-password').textContent()) ?? ''
    ).trim()
    expect(code, 'Einmalpasswort muss sechsstelliger Zifferncode sein').toMatch(
      /^\d{6}$/,
    )
    await createdDialog.getByRole('button', { name: 'Okay' }).click()

    // Neue Helfer starten inaktiv: der Admin schaltet Petra frei. Die
    // Tabellenzeile wird über den Namen und ihren Status-Switch aufgelöst.
    const userRow = page
      .locator('div')
      .filter({ has: page.getByRole('switch') })
      .filter({ hasText: 'Petra Neumann' })
      .last()
    const userSwitch = userRow.getByRole('switch')
    await expect(userSwitch).not.toBeChecked()
    await userSwitch.click()
    await expect(userSwitch).toBeChecked()
    // Am Desktop meldet die Sidebar ab (das Benutzermenü gibt es nur mobil).
    await page.getByRole('button', { name: 'Abmelden' }).click()
    await expect(page).toHaveURL(/\/login$/)

    // Petra setzt mit dem Einmalpasswort ihr eigenes Passwort …
    await page.getByRole('link', { name: 'Neues Passwort festlegen' }).click()
    await expect(page).toHaveURL(/\/set-password$/)
    await page.getByPlaceholder('Benutzername').fill('petra')
    await page.getByPlaceholder('Neues Passwort').fill(PETRA_PASSWORT)
    // Das OTP-Feld (input-otp) trägt keinen zugänglichen Namen.
    await page.locator('input[autocomplete="one-time-code"]').fill(code)
    await page.getByRole('button', { name: 'Passwort festlegen' }).click()
    await expect(page).toHaveURL(/\/login$/)

    // … und meldet sich damit als Servicekraft an.
    await anmelden(page, {
      username: 'petra',
      password: PETRA_PASSWORT,
      startseite: /\/service\/tische$/,
    })

    // Der Admin-Bereich bleibt ihr verschlossen: AdminGuard leitet zurück in
    // den Service.
    await page.goto('/admin/kasse')
    await expect(page).toHaveURL(/\/service\/tische$/)
    await expect(
      page.getByRole('button', { name: 'Alle Tische' }),
    ).toBeVisible()
  })
})
