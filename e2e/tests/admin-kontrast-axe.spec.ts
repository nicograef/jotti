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

test.describe('WCAG-AA-Kontrast auf Recovery-/Compliance-Screens', () => {
  for (const theme of ['light', 'dark'] as const) {
    test(`${theme}: Druckstationen, Finanzamt, Kassenabschluss erfüllen AA`, async ({
      page,
      request,
    }) => {
      const zugang = await resetAndSeed(request)
      await anmelden(page, zugang.admin)

      for (const { url, sichtbar } of screens) {
        await pruefeKontrast(page, theme, url, sichtbar)
      }
    })
  }
})
