import type { Page } from '@playwright/test'
import { expect } from '@playwright/test'

import type { Zugangsdaten } from './seed'

// anmelden meldet einen Benutzer über das Anmeldeformular an. Es nutzt
// zugängliche Selektoren (Platzhalter, Button-Beschriftung) statt Test-IDs und
// wartet auf die rollenabhängige Weiterleitung nach erfolgreicher Anmeldung:
// Admins landen unter /admin, Servicekräfte unter /service.
export async function anmelden(
  page: Page,
  zugangsdaten: Zugangsdaten,
): Promise<void> {
  await page.goto('/login')

  await page.getByPlaceholder('Benutzername').fill(zugangsdaten.username)
  await page.getByPlaceholder('Passwort').fill(zugangsdaten.password)

  await page.getByRole('button', { name: 'Anmelden' }).click()

  // Die Anmeldung leitet weg von /login (auf /admin bzw. /service) — darauf
  // wartet Playwright ohne feste Verzögerung.
  await expect(page).not.toHaveURL(/\/login$/)
}
