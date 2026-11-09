import type { Page } from '@playwright/test'
import { expect } from '@playwright/test'

import type { Zugangsdaten } from './seed'

// Prüft die rollenabhängige Weiterleitung (AuthRedirect): Admin → /admin,
// Servicekraft und Serviceleitung → /service.
export async function anmelden(
  page: Page,
  zugangsdaten: Zugangsdaten,
): Promise<void> {
  await page.goto('/login')

  await page.getByPlaceholder('Benutzername').fill(zugangsdaten.username)
  await page.getByPlaceholder('Passwort').fill(zugangsdaten.password)

  await page.getByRole('button', { name: 'Anmelden' }).click()

  await expect(page).toHaveURL(zugangsdaten.startseite)
}
