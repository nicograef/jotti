import { AxeBuilder } from '@axe-core/playwright'
import type { Page } from '@playwright/test'
import { expect, test } from '@playwright/test'

import { anmelden } from '../support/anmelden'
import { resetAndSeed } from '../support/seed'

// WCAG-AA-Kontrast-Gate (Phase 8) für die Recovery- und Compliance-Screens.
// axe-core prüft das Kriterium 1.4.3 (Kontrast, AA) über den gerenderten DOM;
// wir fahren gezielt nur die Regel `color-contrast`, weil dies der Kontrast-
// Check ist, den diese Phase besitzt (übrige a11y-Regeln gehören nicht hierher).
// Beide Themes werden geprüft: die Recovery-Screens werden bei Außeneinsatz
// (BYOD) auch im Dark Mode bedient.
//
// axe' `color-contrast` deckt WCAG 1.4.3 (Text 4,5:1) ab, nicht 1.4.11 (UI-Ränder
// 3:1). Die 3:1-Ränder der Outline-Buttons und Drucker-IP-Felder sind daher über
// die Token-Werte (`--input`/`--border`) abgesichert, nicht über dieses Gate.

// Die drei Screens tragen die Grün-Aktionen, Outline-Buttons, Drucker-IP-Felder
// und WarnKarten, um die es in Muster 05 geht.
const screens = [
  { url: '/admin/druckstationen', sichtbar: /Bondrucker/ },
  { url: '/admin/finanzamt', sichtbar: /Finanzamt/ },
  { url: '/admin/kasse', sichtbar: /Kassentag/ },
]

// Die vier per Seed erreichbaren Lösch-Bestätigungen tragen den soliden
// destructive-solid-Button (weiße bzw. dunkelrote Schrift auf der destruktiven
// Fläche). Jede Funktion navigiert zur passenden Admin-Seite und öffnet den
// AlertDialog; die Kontrastmessung läuft anschließend auf den Dialog-Inhalt.
const loeschDialoge: {
  name: string
  oeffne: (page: Page) => Promise<void>
}[] = [
  {
    name: 'Produkt löschen',
    oeffne: async (page) => {
      await page.goto('/admin/produkte')
      await page.getByRole('heading', { name: 'Produkte & Preise' }).waitFor()
      await page.waitForLoadState('networkidle')
      await page
        .getByRole('button', { name: 'Weitere Aktionen' })
        .first()
        .click()
      await page.getByRole('menuitem', { name: /Löschen/ }).click()
    },
  },
  {
    name: 'Variante löschen',
    oeffne: async (page) => {
      await page.goto('/admin/produkte')
      await page.getByRole('heading', { name: 'Produkte & Preise' }).waitFor()
      await page.waitForLoadState('networkidle')
      // Ersten Varianten-Chip öffnen (Bearbeiten-Dialog), dann darin den
      // Lösch-Einstieg tippen, der die Bestätigung einblendet.
      await page
        .getByRole('button', { name: /^Variante „.+" bearbeiten$/ })
        .first()
        .click()
      await page.getByRole('button', { name: 'Variante löschen' }).click()
    },
  },
  {
    name: 'Tisch löschen',
    oeffne: async (page) => {
      await page.goto('/admin/tische')
      await page.getByRole('heading', { name: 'Tische' }).waitFor()
      await page.waitForLoadState('networkidle')
      // Ein frisch angelegter Tisch trägt keinen Saldo — nur dann bietet der
      // Bearbeiten-Dialog den aktiven Lösch-Einstieg (statt der gesperrten
      // Variante bei offenem Saldo).
      await page.getByRole('button', { name: 'Neuer Tisch' }).click()
      const neuerDialog = page.getByRole('dialog')
      await neuerDialog.getByLabel('Name').fill('Tisch 99')
      await neuerDialog.getByRole('button', { name: 'Tisch anlegen' }).click()
      await page.getByText('Tisch "Tisch 99" wurde angelegt.').waitFor()
      await page.getByRole('button', { name: /Tisch 99/ }).getByText('Tisch 99').click()
      await page
        .getByRole('dialog')
        .getByRole('button', { name: 'Tisch löschen' })
        .click()
    },
  },
  {
    name: 'Helfer löschen',
    oeffne: async (page) => {
      await page.goto('/admin/benutzer')
      await page.getByRole('heading', { name: 'Helfer & Zugänge' }).waitFor()
      await page.waitForLoadState('networkidle')
      // Das eigene Konto bietet kein Löschen — die „···"-Menüs durchgehen, bis
      // eines den Lösch-Eintrag zeigt (ein fremder Helfer).
      const menues = page.getByRole('button', { name: 'Weitere Aktionen' })
      const anzahl = await menues.count()
      for (let i = 0; i < anzahl; i++) {
        await menues.nth(i).click()
        const loeschen = page.getByRole('menuitem', { name: /Löschen/ })
        if (await loeschen.isVisible().catch(() => false)) {
          await loeschen.click()
          return
        }
        await page.keyboard.press('Escape')
      }
      throw new Error('Kein löschbarer Helfer gefunden')
    },
  },
]

// pruefeKontrast lädt einen Screen im gewünschten Theme und lässt axe nur die
// color-contrast-Regel laufen. Das Theme wird über den localStorage-Schlüssel
// des ThemeProviders gesetzt (siehe frontend/src/components/theme-provider.tsx);
// die anschließende Navigation mountet die App neu und liest den Wert.
async function pruefeKontrast(
  page: Page,
  theme: 'light' | 'dark',
  url: string,
  sichtbar: RegExp,
): Promise<void> {
  await page.evaluate((t) => {
    localStorage.setItem('vite-ui-theme', t)
  }, theme)
  await page.goto(url)
  await page.getByText(sichtbar).first().waitFor()
  // Die geprüften destructive-Inhalte (WarnKarten, Fehlertexte) hängen an
  // asynchronen Queries. networkidle stellt sicher, dass sie gemountet sind,
  // bevor axe misst — der synchron gerenderte Header allein garantiert das nicht.
  await page.waitForLoadState('networkidle')

  const ergebnis = await new AxeBuilder({ page })
    .withRules(['color-contrast'])
    .analyze()

  const befunde = ergebnis.violations.flatMap((v) =>
    v.nodes.map((n) => `${n.target.join(' ')} :: ${n.failureSummary ?? ''}`),
  )
  expect(
    befunde,
    `${theme} ${url}: Kontrast-Verstöße\n${befunde.join('\n')}`,
  ).toEqual([])
}

// pruefeDialogKontrast setzt das Theme, öffnet über oeffne den Lösch-Dialog und
// misst den Kontrast ausschließlich auf dem AlertDialog-Inhalt (Titel,
// Beschreibung, Abbrechen und der solide destructive-solid-Button). Die
// Einschränkung auf den Dialog isoliert die geprüfte Bestätigung vom übrigen
// Seiten- bzw. Bearbeiten-Dialog-Inhalt.
async function pruefeDialogKontrast(
  page: Page,
  theme: 'light' | 'dark',
  name: string,
  oeffne: (page: Page) => Promise<void>,
): Promise<void> {
  await page.evaluate((t) => {
    localStorage.setItem('vite-ui-theme', t)
  }, theme)
  await oeffne(page)
  const dialog = page.locator('[data-slot="alert-dialog-content"]')
  await expect(dialog, `${name}: Lösch-Dialog sichtbar`).toBeVisible()
  // Radix blendet den Dialog mit Fade-/Zoom-Animation ein; axe rechnet die
  // momentane Opacity in die Farben ein. Erst bei voller Deckkraft messen,
  // sonst verfälscht der halbtransparente Zwischenzustand den Kontrast (der
  // solide destructive-solid-Button mischt sich dann mit dem dunklen Backdrop).
  await expect(dialog, `${name}: Einblend-Animation beendet`).toHaveCSS(
    'opacity',
    '1',
  )

  const ergebnis = await new AxeBuilder({ page })
    .include('[data-slot="alert-dialog-content"]')
    .withRules(['color-contrast'])
    .analyze()

  const befunde = ergebnis.violations.flatMap((v) =>
    v.nodes.map((n) => `${n.target.join(' ')} :: ${n.failureSummary ?? ''}`),
  )
  expect(
    befunde,
    `${theme} ${name}: Kontrast-Verstöße\n${befunde.join('\n')}`,
  ).toEqual([])
}

test.describe('WCAG-AA-Kontrast auf Recovery-/Compliance-Screens', () => {
  for (const theme of ['light', 'dark'] as const) {
    test(`${theme}: Screens und Lösch-Dialoge erfüllen AA`, async ({
      page,
      request,
    }) => {
      const zugangsdaten = await resetAndSeed(request)
      await anmelden(page, zugangsdaten.admin)

      for (const { url, sichtbar } of screens) {
        await pruefeKontrast(page, theme, url, sichtbar)
      }

      for (const { name, oeffne } of loeschDialoge) {
        await pruefeDialogKontrast(page, theme, name, oeffne)
      }
    })
  }
})
