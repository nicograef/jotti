import { expect, test } from '@playwright/test'

import { anmelden } from '../support/anmelden'
import { resetAndSeed } from '../support/seed'

// Deckt die Kassenführung aus Admin-Sicht ab: die laufende Kassensitzung wird
// angezeigt, ein Geldtransit lässt sich buchen, und der Kassenabschluss prüft
// offene Tischsalden, bevor er den Tagesabschluss zulässt (Seed-Sonntag hat
// mehrere offene Tische).

test.describe('Admin führt die Kasse', () => {
  test('Geldtransit buchen und Kassenabschluss meldet offene Tischsalden', async ({
    page,
    request,
  }) => {
    const zugangsdaten = await resetAndSeed(request)
    await anmelden(page, zugangsdaten.admin)

    await page.goto('/admin/kasse')
    await expect(page.getByRole('heading', { name: 'Kassensitzung' })).toBeVisible()

    // Die offene Sonntags-Sitzung (ZNr 3) ist sichtbar.
    await expect(page.getByText(/Kassensitzung #3/)).toBeVisible()
    await expect(page.getByText(/offen/)).toBeVisible()

    // Geldtransit-Einlage buchen.
    await page.getByRole('radio', { name: /Einlage/ }).check()
    await page.getByLabel('Betrag').fill('25,00')
    await page.getByLabel('Kommentar').fill('Zusätzliches Wechselgeld')
    await page.getByRole('button', { name: 'Geldtransit buchen' }).click()
    await expect(page.getByText('Kassenbewegung gebucht.')).toBeVisible()

    // Kasse abschließen anstoßen: Soll-Bestand zählen und Dialog öffnen.
    await page.getByLabel('Gezählter Ist-Bestand').fill('500,00')
    await page.getByRole('button', { name: 'Kasse abschließen' }).first().click()

    const confirmDialog = page.getByRole('alertdialog')
    await expect(confirmDialog.getByText('Soll-Bestand')).toBeVisible()
    await expect(confirmDialog.getByText('Z-Bon-Vorschau')).toBeVisible()
    await confirmDialog
      .getByRole('button', { name: 'Kasse abschließen' })
      .click()

    // Seed-Sonntag hat mehrere offene Tische — der Abschluss wird blockiert.
    await expect(
      page.getByText(
        'Es gibt noch offene Tische mit ausstehenden Beträgen. Bitte alle Tische abrechnen.',
      ),
    ).toBeVisible()
  })
})
