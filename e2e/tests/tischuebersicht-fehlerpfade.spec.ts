import { expect, test } from '@playwright/test'

import {
  simuliereNetzabbruch,
  simuliereServerfehler,
} from '../helpers/fehlerpfade'
import { anmelden } from '../support/anmelden'
import { resetAndSeed } from '../support/seed'

// Praxistest-Befund: Serverfehler beim Laden der Tischübersicht wirkten wie
// eine leere/erledigte Kasse (Saldo 0,00 €). Der globale QueryCache-Handler
// (frontend/src/lib/queryClient.ts) fängt das inzwischen mit einem Toast ab —
// diese Specs erzwingen, dass er bei jedem betroffenen Endpunkt greift.

test.describe('Tischübersicht bei Serverfehler und Netzabbruch', () => {
  test('Serverfehler bei den Tischdaten zeigt einen sichtbaren Fehlerhinweis', async ({
    page,
    request,
  }) => {
    const zugangsdaten = await resetAndSeed(request)
    await anmelden(page, zugangsdaten.service)

    await simuliereServerfehler(page, [
      'service/get-meine-tische-state',
      'service/get-eigene-uebersicht',
    ])
    await page.goto('/service/tische')

    // Ein stiller Leer-Default ohne jeden Hinweis gälte als Fehlschlag —
    // erwartet wird ein sichtbarer Fehlerhinweis (Toast).
    await expect(
      page.getByText('Daten konnten nicht geladen werden. Bitte Verbindung prüfen und erneut versuchen.'),
    ).toBeVisible()
  })

  test('Netzabbruch bei den Tischdaten zeigt einen sichtbaren Fehlerhinweis', async ({
    page,
    request,
  }) => {
    const zugangsdaten = await resetAndSeed(request)
    await anmelden(page, zugangsdaten.service)

    await simuliereNetzabbruch(page, [
      'service/get-meine-tische-state',
      'service/get-eigene-uebersicht',
    ])
    await page.goto('/service/tische')

    await expect(
      page.getByText('Daten konnten nicht geladen werden. Bitte Verbindung prüfen und erneut versuchen.'),
    ).toBeVisible()
  })
})

test.describe('Tischübersicht (Alle-Tische-Drawer) bei Serverfehler und Netzabbruch', () => {
  test('Serverfehler beim Laden aller Tische zeigt einen sichtbaren Fehlerhinweis', async ({
    page,
    request,
  }) => {
    const zugangsdaten = await resetAndSeed(request)
    await anmelden(page, zugangsdaten.service)

    await simuliereServerfehler(page, [
      'service/get-aktive-tische-mit-favoriten',
    ])

    await page.goto('/service/tische')
    await page.getByRole('button', { name: 'Alle Tische' }).click()

    await expect(
      page.getByText('Daten konnten nicht geladen werden. Bitte Verbindung prüfen und erneut versuchen.'),
    ).toBeVisible()
  })

  test('Netzabbruch beim Laden aller Tische zeigt einen sichtbaren Fehlerhinweis', async ({
    page,
    request,
  }) => {
    const zugangsdaten = await resetAndSeed(request)
    await anmelden(page, zugangsdaten.service)

    await simuliereNetzabbruch(page, [
      'service/get-aktive-tische-mit-favoriten',
    ])

    await page.goto('/service/tische')
    await page.getByRole('button', { name: 'Alle Tische' }).click()

    await expect(
      page.getByText('Daten konnten nicht geladen werden. Bitte Verbindung prüfen und erneut versuchen.'),
    ).toBeVisible()
  })
})
