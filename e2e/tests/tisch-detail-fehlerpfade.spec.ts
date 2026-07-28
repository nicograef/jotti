import { expect, test } from '@playwright/test'

import {
  simuliereNetzabbruch,
  simuliereServerfehler,
} from '../helpers/fehlerpfade'
import { anmelden } from '../support/anmelden'
import { resetAndSeed } from '../support/seed'

// Tisch-Detail (TablePage) hat einen expliziten Fehlerzustand für
// get-tisch-state/get-tisch-historie (siehe frontend/src/service/TablePage.tsx)
// — diese Specs bestätigen, dass er bei Serverfehler und Netzabbruch greift und
// den Tisch nicht als „Saldo 0,00 €" ausgibt. Der Zuschnitt ist verschieden:
// Der Tischzustand trägt die ganze Seite, die Historie nur ihren eigenen Tab —
// Bestellen und Kassieren bleiben bedienbar.

// „Tisch 3" hat im Demo-Drehbuch Historie, ist für diese Specs aber nur ein
// beliebiger aktiver Tisch — der Zustand selbst wird ja abgefangen.
const TISCH_ID = 3

test.describe('Tisch-Detail bei Serverfehler und Netzabbruch', () => {
  test('Serverfehler beim Laden des Tischzustands zeigt einen sichtbaren Fehlerhinweis statt Saldo 0,00 €', async ({
    page,
    request,
  }) => {
    const zugangsdaten = await resetAndSeed(request)
    await anmelden(page, zugangsdaten.service)

    await simuliereServerfehler(page, ['service/get-tisch-state'])
    await page.goto(`/service/tische/${String(TISCH_ID)}`)

    await expect(
      page.getByText('Tischdaten konnten nicht geladen werden'),
    ).toBeVisible()
    await expect(page.getByRole('button', { name: 'Erneut versuchen' })).toBeVisible()
    // Kein stiller Leer-Default: der ausgeglichene Saldo darf nicht als
    // scheinbar echtes Ergebnis erscheinen.
    await expect(page.getByText('0,00 €')).not.toBeVisible()
  })

  test('Netzabbruch beim Laden der Tisch-Historie zeigt den Fehlerhinweis nur im Historie-Tab', async ({
    page,
    request,
  }) => {
    const zugangsdaten = await resetAndSeed(request)
    await anmelden(page, zugangsdaten.service)

    await simuliereNetzabbruch(page, ['service/get-tisch-historie'])
    await page.goto(`/service/tische/${String(TISCH_ID)}`)

    // Der Tischzustand ist geladen: Die Seite steht, der Bestellen-Tab ist
    // bedienbar. Nur die Historie fehlt.
    await expect(
      page.getByText('Tischdaten konnten nicht geladen werden'),
    ).not.toBeVisible()
    await page.getByRole('tab', { name: 'Historie' }).click()

    await expect(
      page.getByText('Historie konnte nicht geladen werden'),
    ).toBeVisible()
    await expect(page.getByRole('button', { name: 'Erneut versuchen' })).toBeVisible()
  })
})
